package service

import (
	"encoding/json"
	"fmt"

	"github.com/Github-Aiko/Aiko-Server/internal/pkg/api"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

// OutboundBuilder build freedom outbund config for addoutbound
func OutboundBuilder(nodeInfo *api.NodeInfo) (*core.OutboundHandlerConfig, error) {
	outboundDetourConfig := &conf.OutboundDetourConfig{}
	outboundDetourConfig.Protocol = "freedom"
	outboundDetourConfig.Tag = fmt.Sprintf("%s_%d", protocol, nodeInfo.ServerPort)
	// Freedom Protocol setting
	var domainStrategy = "Asis"
	proxySetting := &conf.FreedomConfig{
		DomainStrategy: domainStrategy,
	}
	var setting json.RawMessage
	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, fmt.Errorf("marshal proxy %s config fialed: %s", protocol, err)
	}
	outboundDetourConfig.Settings = &setting

	return outboundDetourConfig.Build()
}
