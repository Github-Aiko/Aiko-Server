package xray

import (
	"fmt"

	"github.com/AikoPanel/Xray-core/common/net"
	"github.com/AikoPanel/Xray-core/core"
	"github.com/AikoPanel/Xray-core/infra/conf"
	conf2 "github.com/Github-Aiko/Aiko-Server/src/conf"
	"github.com/goccy/go-json"
)

// BuildOutbound build freedom outbund config for addoutbound
func buildOutbound(config *conf2.ControllerConfig, tag string) (*core.OutboundHandlerConfig, error) {
	outboundDetourConfig := &conf.OutboundDetourConfig{}
	outboundDetourConfig.Protocol = "freedom"
	outboundDetourConfig.Tag = tag

	// Build Send IP address
	if config.SendIP != "" {
		ipAddress := net.ParseAddress(config.SendIP)
		outboundDetourConfig.SendThrough = &conf.Address{Address: ipAddress}
	}

	// Freedom Protocol setting
	var domainStrategy = "Asis"
	if config.XrayOptions.EnableDNS {
		if config.XrayOptions.DNSType != "" {
			domainStrategy = config.XrayOptions.DNSType
		} else {
			domainStrategy = "UseIP"
		}
	}
	proxySetting := &conf.FreedomConfig{
		DomainStrategy: domainStrategy,
	}
	var setting json.RawMessage
	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, fmt.Errorf("marshal proxy config error: %s", err)
	}
	outboundDetourConfig.Settings = &setting
	return outboundDetourConfig.Build()
}
