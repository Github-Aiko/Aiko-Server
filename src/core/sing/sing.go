package sing

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/goccy/go-json"
	"github.com/inazumav/sing-box/constant"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
	vCore "github.com/Github-Aiko/Aiko-Server/src/core"
	"github.com/inazumav/sing-box/adapter"
	"github.com/inazumav/sing-box/inbound"
	"github.com/inazumav/sing-box/log"
	"github.com/inazumav/sing-box/option"
	"github.com/inazumav/sing-box/outbound"
	"github.com/inazumav/sing-box/route"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/service/pause"
)

var _ adapter.Service = (*Box)(nil)

type Box struct {
	createdAt  time.Time
	router     adapter.Router
	inbounds   map[string]adapter.Inbound
	outbounds  []adapter.Outbound
	logFactory log.Factory
	logger     log.ContextLogger
	hookServer *HookServer
	done       chan struct{}
}

func init() {
	vCore.RegisterCore("sing", New)
}

func New(c *conf.CoreConfig) (vCore.Core, error) {
	sc := c.SingConfig
	options := option.Options{}
	options.Log = &option.LogOptions{
		Disabled:  sc.LogConfig.Disabled,
		Level:     sc.LogConfig.Level,
		Timestamp: sc.LogConfig.Timestamp,
		Output:    sc.LogConfig.Output,
	}
	options.DNS = &option.DNSOptions{}
	os.Setenv("CORE_RUNNING", "")
	os.Setenv("SING_DNS_PATH", "")

	if sc.DnsConfigPath != "" {
		if f, err := os.Open(sc.DnsConfigPath); err != nil {
			panic("Failed to read DNS config file")
		} else {
			if err = json.NewDecoder(f).Decode(options.DNS); err != nil {
				panic("Failed to unmarshal DNS config")
			}
		}
		os.Setenv("SING_DNS_PATH", sc.DnsConfigPath)
	}
	ctx := context.Background()
	ctx = pause.ContextWithDefaultManager(ctx)
	createdAt := time.Now()
	experimentalOptions := common.PtrValueOrDefault(options.Experimental)
	applyDebugOptions(common.PtrValueOrDefault(experimentalOptions.Debug))
	var defaultLogWriter io.Writer
	logFactory, err := log.New(log.Options{
		Context:       ctx,
		Options:       common.PtrValueOrDefault(options.Log),
		DefaultWriter: defaultLogWriter,
		BaseTime:      createdAt,
	})
	if err != nil {
		return nil, E.Cause(err, "create log factory")
	}
	router, err := route.NewRouter(
		ctx,
		logFactory,
		common.PtrValueOrDefault(options.Route),
		common.PtrValueOrDefault(options.DNS),
		common.PtrValueOrDefault(options.NTP),
		options.Inbounds,
		nil,
	)
	if err != nil {
		return nil, E.Cause(err, "parse route options")
	}
	inbounds := make([]adapter.Inbound, len(options.Inbounds))
	inMap := make(map[string]adapter.Inbound, len(inbounds))
	outbounds := make([]adapter.Outbound, 0, len(options.Outbounds))
	for i, inboundOptions := range options.Inbounds {
		var in adapter.Inbound
		var tag string
		if inboundOptions.Tag != "" {
			tag = inboundOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		in, err = inbound.New(
			ctx,
			router,
			logFactory.NewLogger(F.ToString("inbound/", inboundOptions.Type, "[", tag, "]")),
			inboundOptions,
			nil,
		)
		if err != nil {
			return nil, E.Cause(err, "parse inbound[", i, "]")
		}
		inbounds[i] = in
		inMap[inboundOptions.Tag] = in
	}
	for i, outboundOptions := range options.Outbounds {
		var out adapter.Outbound
		var tag string
		if outboundOptions.Tag != "" {
			tag = outboundOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		out, err = outbound.New(
			ctx,
			router,
			logFactory.NewLogger(F.ToString("outbound/", outboundOptions.Type, "[", tag, "]")),
			tag,
			outboundOptions)
		if err != nil {
			return nil, E.Cause(err, "parse outbound[", i, "]")
		}
		outbounds = append(outbounds, out)
	}
	err = router.Initialize(inbounds, outbounds, func() adapter.Outbound {
		out, oErr := outbound.New(ctx, router, logFactory.NewLogger("outbound/direct"), "direct", option.Outbound{Type: "direct", Tag: "default"})
		common.Must(oErr)
		outbounds = append(outbounds, out)
		return out
	})
	if err != nil {
		return nil, err
	}
	server := NewHookServer(logFactory.NewLogger("Hook-Server"))
	if err != nil {
		return nil, E.Cause(err, "create v2ray api server")
	}
	router.SetClashServer(server)
	return &Box{
		router:     router,
		inbounds:   inMap,
		outbounds:  outbounds,
		createdAt:  createdAt,
		logFactory: logFactory,
		logger:     logFactory.Logger(),
		hookServer: server,
		done:       make(chan struct{}),
	}, nil
}

func (b *Box) PreStart() error {
	err := b.preStart()
	if err != nil {
		// TODO: remove catch error
		defer func() {
			v := recover()
			if v != nil {
				log.Error(E.Cause(err, "origin error"))
				debug.PrintStack()
				panic("panic on early close: " + fmt.Sprint(v))
			}
		}()
		b.Close()
		return err
	}
	b.logger.Info("sing-box pre-started (", F.Seconds(time.Since(b.createdAt).Seconds()), "s)")
	return nil
}

func (b *Box) Start() error {
	err := b.start()
	if err != nil {
		// TODO: remove catch error
		defer func() {
			v := recover()
			if v != nil {
				log.Error(E.Cause(err, "origin error"))
				debug.PrintStack()
				panic("panic on early close: " + fmt.Sprint(v))
			}
		}()
		b.Close()
		return err
	}
	os.Setenv("CORE_RUNNING", "SING")
	log.Info("sing-box core version: ", constant.Version)
	b.logger.Info("sing-box started (", F.Seconds(time.Since(b.createdAt).Seconds()), "s)")
	return nil
}

func (b *Box) preStart() error {
	err := b.startOutbounds()
	if err != nil {
		return err
	}
	return b.router.Start()
}

func (b *Box) start() error {
	err := b.preStart()
	if err != nil {
		return err
	}
	for i, in := range b.inbounds {
		var tag string
		if in.Tag() == "" {
			tag = F.ToString(i)
		} else {
			tag = in.Tag()
		}
		b.logger.Trace("initializing inbound/", in.Type(), "[", tag, "]")
		err = in.Start()
		if err != nil {
			return E.Cause(err, "initialize inbound/", in.Type(), "[", tag, "]")
		}
	}
	return nil
}

func (b *Box) postStart() error {
	for serviceName, service := range b.outbounds {
		if lateService, isLateService := service.(adapter.PostStarter); isLateService {
			b.logger.Trace("post-starting ", service)
			err := lateService.PostStart()
			if err != nil {
				return E.Cause(err, "post-start ", serviceName)
			}
		}
	}
	return nil
}

func (b *Box) Close() error {
	select {
	case <-b.done:
		return os.ErrClosed
	default:
		close(b.done)
	}
	var errors error
	for i, in := range b.inbounds {
		b.logger.Trace("closing inbound/", in.Type(), "[", i, "]")
		errors = E.Append(errors, in.Close(), func(err error) error {
			return E.Cause(err, "close inbound/", in.Type(), "[", i, "]")
		})
	}
	for i, out := range b.outbounds {
		b.logger.Trace("closing outbound/", out.Type(), "[", i, "]")
		errors = E.Append(errors, common.Close(out), func(err error) error {
			return E.Cause(err, "close outbound/", out.Type(), "[", i, "]")
		})
	}
	b.logger.Trace("closing router")
	if err := common.Close(b.router); err != nil {
		errors = E.Append(errors, err, func(err error) error {
			return E.Cause(err, "close router")
		})
	}
	b.logger.Trace("closing log factory")
	if err := common.Close(b.logFactory); err != nil {
		errors = E.Append(errors, err, func(err error) error {
			return E.Cause(err, "close log factory")
		})
	}
	os.Setenv("SING_RUNNING", "")
	return errors
}

func (b *Box) Router() adapter.Router {
	return b.router
}

func (b *Box) Protocols() []string {
	return []string{
		"v2ray",
		"vless",
		"trojan",
		"shadowsocks",
		"hysteria",
		"tuic",
	}
}
