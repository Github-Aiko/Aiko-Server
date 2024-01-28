package xray

import (
	"fmt"
	"os"
	"sync"

	"github.com/Github-Aiko/Aiko-Server/conf"
	vCore "github.com/Github-Aiko/Aiko-Server/core"
	"github.com/Github-Aiko/Aiko-Server/core/xray/app/dispatcher"
	_ "github.com/Github-Aiko/Aiko-Server/core/xray/distro/all"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/app/stats"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/inbound"
	"github.com/xtls/xray-core/features/outbound"
	"github.com/xtls/xray-core/features/routing"
	statsFeature "github.com/xtls/xray-core/features/stats"
	coreConf "github.com/xtls/xray-core/infra/conf"
)

var _ vCore.Core = (*Xray)(nil)

func init() {
	vCore.RegisterCore("xray", New)
}

// Xray Structure
type Xray struct {
	access     sync.Mutex
	Server     *core.Instance
	ihm        inbound.Manager
	ohm        outbound.Manager
	shm        statsFeature.Manager
	dispatcher *dispatcher.DefaultDispatcher
}

func New(c *conf.CoreConfig) (vCore.Core, error) {
	return &Xray{Server: getCore(c.XrayConfig)}, nil
}

func parseConnectionConfig(c *conf.XrayConnectionConfig) (policy *coreConf.Policy) {
	policy = &coreConf.Policy{
		StatsUserUplink:   true,
		StatsUserDownlink: true,
		Handshake:         &c.Handshake,
		ConnectionIdle:    &c.ConnIdle,
		UplinkOnly:        &c.UplinkOnly,
		DownlinkOnly:      &c.DownlinkOnly,
		BufferSize:        &c.BufferSize,
	}
	return
}

func getCore(c *conf.XrayConfig) *core.Instance {
	os.Setenv("XRAY_LOCATION_ASSET", c.AssetPath)
	// Log Config
	coreLogConfig := &coreConf.LogConfig{
		LogLevel:  c.LogConfig.Level,
		AccessLog: c.LogConfig.AccessPath,
		ErrorLog:  c.LogConfig.ErrorPath,
	}
	// DNS config
	coreDnsConfig := &coreConf.DNSConfig{}
	os.Setenv("XRAY_DNS_PATH", "")
	if c.DnsConfigPath != "" {
		data, err := os.ReadFile(c.DnsConfigPath)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to read xray dns config file: %v", err))
			coreDnsConfig = &coreConf.DNSConfig{}
		} else {
			if err := json.Unmarshal(data, coreDnsConfig); err != nil {
				log.Error(fmt.Sprintf("Failed to unmarshal xray dns config: %v. Using default DNS options.", err))
				coreDnsConfig = &coreConf.DNSConfig{}
			}
		}
		os.Setenv("XRAY_DNS_PATH", c.DnsConfigPath)
	}
	dnsConfig, err := coreDnsConfig.Build()
	if err != nil {
		log.WithField("err", err).Panic("Failed to understand DNS config, Please check: https://xtls.github.io/config/dns.html for help")
	}
	// Routing config
	coreRouterConfig := &coreConf.RouterConfig{}
	if c.RouteConfigPath != "" {
		data, err := os.ReadFile(c.RouteConfigPath)
		if err != nil {
			log.WithField("err", err).Panic("Failed to read Routing config file")
		} else {
			if err = json.Unmarshal(data, coreRouterConfig); err != nil {
				log.WithField("err", err).Panic("Failed to unmarshal Routing config")
			}
		}
	}
	routeConfig, err := coreRouterConfig.Build()
	if err != nil {
		log.WithField("err", err).Panic("Failed to understand Routing config. Please check: https://xtls.github.io/config/routing.html for help")
	}
	// Custom Inbound config
	var coreCustomInboundConfig []coreConf.InboundDetourConfig
	if c.InboundConfigPath != "" {
		data, err := os.ReadFile(c.InboundConfigPath)
		if err != nil {
			log.WithField("err", err).Panic("Failed to read Custom Inbound config file")
		} else {
			if err = json.Unmarshal(data, &coreCustomInboundConfig); err != nil {
				log.WithField("err", err).Panic("Failed to unmarshal Custom Inbound config")
			}
		}
	}
	var inBoundConfig []*core.InboundHandlerConfig
	for _, config := range coreCustomInboundConfig {
		oc, err := config.Build()
		if err != nil {
			log.WithField("err", err).Panic("Failed to understand Inbound config. Please check: https://xtls.github.io/config/inbound.html for help")
		}
		inBoundConfig = append(inBoundConfig, oc)
	}
	// Custom Outbound config
	var coreCustomOutboundConfig []coreConf.OutboundDetourConfig
	if c.OutboundConfigPath != "" {
		data, err := os.ReadFile(c.OutboundConfigPath)
		if err != nil {
			log.WithField("err", err).Panic("Failed to read Custom Outbound config file")
		} else {
			if err = json.Unmarshal(data, &coreCustomOutboundConfig); err != nil {
				log.WithField("err", err).Panic("Failed to unmarshal Custom Outbound config")
			}
		}
	}
	var outBoundConfig []*core.OutboundHandlerConfig
	for _, config := range coreCustomOutboundConfig {
		oc, err := config.Build()
		if err != nil {
			log.WithField("err", err).Panic("Failed to understand Outbound config, Please check: https://xtls.github.io/config/outbound.html for help")
		}
		outBoundConfig = append(outBoundConfig, oc)
	}
	// Policy config
	levelPolicyConfig := parseConnectionConfig(c.ConnectionConfig)
	corePolicyConfig := &coreConf.PolicyConfig{}
	corePolicyConfig.Levels = map[uint32]*coreConf.Policy{0: levelPolicyConfig}
	policyConfig, _ := corePolicyConfig.Build()
	// Build Xray conf
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(coreLogConfig.Build()),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(policyConfig),
			serial.ToTypedMessage(dnsConfig),
			serial.ToTypedMessage(routeConfig),
		},
		Inbound:  inBoundConfig,
		Outbound: outBoundConfig,
	}
	server, err := core.New(config)
	if err != nil {
		log.WithField("err", err).Panic("failed to create instance")
	}
	log.Info("Xray Core Version: ", core.Version())
	return server
}

// Start the Xray
func (c *Xray) Start() error {
	c.access.Lock()
	defer c.access.Unlock()
	if err := c.Server.Start(); err != nil {
		return err
	}
	c.shm = c.Server.GetFeature(statsFeature.ManagerType()).(statsFeature.Manager)
	c.ihm = c.Server.GetFeature(inbound.ManagerType()).(inbound.Manager)
	c.ohm = c.Server.GetFeature(outbound.ManagerType()).(outbound.Manager)
	c.dispatcher = c.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	return nil
}

// Close  the core
func (c *Xray) Close() error {
	c.access.Lock()
	defer c.access.Unlock()
	c.ihm = nil
	c.ohm = nil
	c.shm = nil
	c.dispatcher = nil
	err := c.Server.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Xray) Protocols() []string {
	return []string{
		"vmess",
		"vless",
		"shadowsocks",
		"trojan",
	}
}

func (c *Xray) Type() string {
	return "xray"
}
