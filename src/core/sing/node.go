package sing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/inazumav/sing-box/inbound"
	F "github.com/sagernet/sing/common/format"

	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/src/conf"
	"github.com/goccy/go-json"
	"github.com/inazumav/sing-box/option"
)

type WsNetworkConfig struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

func getInboundOptions(tag string, info *panel.NodeInfo, c *conf.Options) (option.Inbound, error) {
	addr, err := netip.ParseAddr(c.ListenIP)
	if err != nil {
		return option.Inbound{}, fmt.Errorf("the listen ip not vail")
	}
	ds, _ := getDomainStrategy(c.SingOptions)
	listen := option.ListenOptions{
		Listen:        (*option.ListenAddress)(&addr),
		ListenPort:    uint16(info.Port),
		ProxyProtocol: c.SingOptions.EnableProxyProtocol,
		TCPFastOpen:   c.SingOptions.TCPFastOpen,
		TCPMultiPath:  c.SingOptions.TCPMultiPath,
		InboundOptions: option.InboundOptions{
			SniffEnabled:             c.SingOptions.SniffEnabled,
			SniffOverrideDestination: c.SingOptions.SniffOverrideDestination,
			SniffOverrideFallback:    c.SingOptions.SniffOverrideFallback,
			DomainStrategy:           option.DomainStrategy(ds),
		},
	}
	var tls option.InboundTLSOptions
	if info.Tls || info.Type == "hysteria" || info.Type == "tuic" {
		if c.CertConfig == nil {
			return option.Inbound{}, fmt.Errorf("the CertConfig is not vail")
		}
		tls.Enabled = true
		tls.Insecure = true
		tls.ServerName = info.ServerName
		switch c.CertConfig.CertMode {
		case "none", "":
			break // disable
		case "reality":
			if c.CertConfig.RealityConfig == nil {
				return option.Inbound{}, fmt.Errorf("RealityConfig is not valid")
			}
			rc := c.CertConfig.RealityConfig
			tls.ServerName = rc.ServerNames[0]
			if len(rc.ShortIds) == 0 {
				rc.ShortIds = []string{""}
			}
			dest, _ := strconv.Atoi(rc.Dest)
			mtd, _ := strconv.Atoi(strconv.FormatUint(rc.MaxTimeDiff, 10))
			tls.Reality = &option.InboundRealityOptions{
				Enabled:           true,
				ShortID:           rc.ShortIds,
				PrivateKey:        rc.PrivateKey,
				MaxTimeDifference: option.Duration(time.Duration(mtd) * time.Second),
				Handshake: option.InboundRealityHandshakeOptions{
					ServerOptions: option.ServerOptions{
						Server:     rc.ServerNames[0],
						ServerPort: uint16(dest),
					},
				},
			}

		case "remote":
			if info.ExtraConfig.EnableReality == "true" {
				if c.CertConfig.RealityConfig == nil {
					return option.Inbound{}, fmt.Errorf("RealityConfig is not valid")
				}
				rc := info.ExtraConfig.RealityConfig
				if len(rc.ShortIds) == 0 {
					rc.ShortIds = []string{""}
				}
				dest, _ := strconv.Atoi(rc.Dest)
				mtd, _ := strconv.Atoi(rc.MaxTimeDiff)
				tls.Reality = &option.InboundRealityOptions{
					Enabled:           true,
					ShortID:           rc.ShortIds,
					PrivateKey:        rc.PrivateKey,
					MaxTimeDifference: option.Duration(time.Duration(mtd) * time.Second),
					Handshake: option.InboundRealityHandshakeOptions{
						ServerOptions: option.ServerOptions{
							Server:     rc.ServerNames[0],
							ServerPort: uint16(dest),
						},
					},
				}
			}
		default:
			tls.CertificatePath = c.CertConfig.CertFile
			tls.KeyPath = c.CertConfig.KeyFile
		}
	}
	in := option.Inbound{
		Tag: tag,
	}
	if info.Type == "hysteria" && c.SingOptions.EnableTUIC {
		info.Type = "tuic"
	}
	if info.Type == "v2ray" && c.SingOptions.EnableVLESS {
		info.Type = "vless"
	}

	switch info.Type {
	case "v2ray", "vless":
		t := option.V2RayTransportOptions{
			Type: info.Network,
		}
		switch info.Network {
		case "tcp":
			t.Type = ""
		case "ws":
			network := WsNetworkConfig{}
			err := json.Unmarshal(info.NetworkSettings, &network)
			if err != nil {
				return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
			}
			var u *url.URL
			u, err = url.Parse(network.Path)
			if err != nil {
				return option.Inbound{}, fmt.Errorf("parse path error: %s", err)
			}
			ed, _ := strconv.Atoi(u.Query().Get("ed"))
			h := make(map[string]option.Listable[string], len(network.Headers))
			for k, v := range network.Headers {
				h[k] = option.Listable[string]{
					v,
				}
			}
			t.WebsocketOptions = option.V2RayWebsocketOptions{
				Path:                u.Path,
				EarlyDataHeaderName: "Sec-WebSocket-Protocol",
				MaxEarlyData:        uint32(ed),
				Headers:             h,
			}
		case "grpc":
			err := json.Unmarshal(info.NetworkSettings, &t.GRPCOptions)
			if err != nil {
				return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
			}
		}
		if info.ExtraConfig.EnableVless == "true" || info.Type == "vless" {
			in.Type = "vless"
			in.VLESSOptions = option.VLESSInboundOptions{
				ListenOptions: listen,
				TLS:           &tls,
				Transport:     &t,
			}
		} else {
			in.Type = "vmess"
			in.VMessOptions = option.VMessInboundOptions{
				ListenOptions: listen,
				TLS:           &tls,
				Transport:     &t,
			}
		}
	case "shadowsocks":
		in.Type = "shadowsocks"
		var keyLength int
		switch info.Cipher {
		case "2022-blake3-aes-128-gcm":
			keyLength = 16
		case "2022-blake3-aes-256-gcm":
			keyLength = 32
		default:
			keyLength = 16
		}
		in.ShadowsocksOptions = option.ShadowsocksInboundOptions{
			ListenOptions: listen,
			Method:        info.Cipher,
		}
		p := make([]byte, keyLength)
		_, _ = rand.Read(p)
		randomPasswd := string(p)
		if strings.Contains(info.Cipher, "2022") {
			fmt.Println(info.ServerKey)
			in.ShadowsocksOptions.Password = info.ServerKey
			randomPasswd = base64.StdEncoding.EncodeToString([]byte(randomPasswd))
		}
		in.ShadowsocksOptions.Users = []option.ShadowsocksUser{{
			Password: base64.StdEncoding.EncodeToString([]byte(randomPasswd)),
		}}
	case "trojan":
		in.Type = "trojan"
		t := option.V2RayTransportOptions{
			Type: info.Network,
		}
		switch info.Network {
		case "tcp":
			t.Type = ""
		case "grpc":
			err := json.Unmarshal(info.NetworkSettings, &t.GRPCOptions)
			if err != nil {
				return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
			}
		}
		randomPasswd := uuid.New().String()
		in.TrojanOptions = option.TrojanInboundOptions{
			ListenOptions: listen,
			Users: []option.TrojanUser{{
				Name:     randomPasswd,
				Password: randomPasswd,
			}},
			TLS:       &tls,
			Transport: &t,
		}
		if c.SingOptions.FallBackConfigs != nil {
			// fallback handling
			fallback := c.SingOptions.FallBackConfigs.FallBack
			fallbackPort, err := strconv.Atoi(fallback.ServerPort)
			if err == nil {
				in.TrojanOptions.Fallback = &option.ServerOptions{
					Server:     fallback.Server,
					ServerPort: uint16(fallbackPort),
				}
			}
			fallbackForALPNMap := c.SingOptions.FallBackConfigs.FallBackForALPN
			fallbackForALPN := make(map[string]*option.ServerOptions, len(fallbackForALPNMap))
			if err := processFallback(c, fallbackForALPN); err == nil {
				in.TrojanOptions.FallbackForALPN = fallbackForALPN
			}
		}
	case "hysteria":
		in.Type = "hysteria"
		randomPasswd := uuid.New().String()
		in.HysteriaOptions = option.HysteriaInboundOptions{
			ListenOptions: listen,
			UpMbps:        info.UpMbps,
			DownMbps:      info.DownMbps,
			Obfs:          info.HyObfs,
			Users: []option.HysteriaUser{{
				Name:       randomPasswd,
				AuthString: randomPasswd,
			}},
			DisableMTUDiscovery: true,
			TLS:                 &tls,
		}
	case "tuic":
		in.Type = "tuic"
		randomPasswd := uuid.New().String()
		tls.ALPN = c.SingOptions.TuicConfig.Alpn
		in.TUICOptions = option.TUICInboundOptions{
			ListenOptions: listen,
			Users: []option.TUICUser{
				{
					Name:     randomPasswd,
					UUID:     randomPasswd,
					Password: "tuic",
				},
			},
			CongestionControl: c.SingOptions.TuicConfig.CongestionControl,
			TLS:               &tls,
		}
	}
	return in, nil
}

func (b *Box) AddNode(tag string, info *panel.NodeInfo, config *conf.Options) error {
	c, err := getInboundOptions(tag, info, config)
	if err != nil {
		return err
	}
	in, err := inbound.New(
		context.Background(),
		b.router,
		b.logFactory.NewLogger(F.ToString("inbound/", c.Type, "[", tag, "]")),
		c,
		nil,
	)
	if err != nil {
		return err
	}
	b.inbounds[tag] = in
	err = in.Start()
	if err != nil {
		return fmt.Errorf("start inbound error: %s", err)
	}
	err = b.router.AddInbound(in)
	if err != nil {
		return fmt.Errorf("add inbound error: %s", err)
	}
	return nil
}

func (b *Box) DelNode(tag string) error {
	err := b.inbounds[tag].Close()
	if err != nil {
		return fmt.Errorf("close inbound error: %s", err)
	}
	err = b.router.DelInbound(tag)
	if err != nil {
		return fmt.Errorf("delete inbound error: %s", err)
	}
	return nil
}
