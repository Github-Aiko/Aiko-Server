package server

import (
	"fmt"
	"sync"

	"github.com/Github-Aiko/Aiko-Server/internal/pkg/api"
	_ "github.com/Github-Aiko/Aiko-Server/internal/pkg/dep"
	"github.com/Github-Aiko/Aiko-Server/internal/pkg/dispatcher"
	"github.com/Github-Aiko/Aiko-Server/internal/pkg/service"
	log "github.com/sirupsen/logrus"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/app/stats"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

type Config struct {
	LogLevel string
}

type Server struct {
	access        sync.Mutex
	instance      *core.Instance
	service       service.Service
	config        *Config
	apiConfig     *api.Config
	serviceConfig *service.Config
	Running       bool
}

func New(config *Config, apiConfig *api.Config, serviceConfig *service.Config) *Server {
	return &Server{config: config, apiConfig: apiConfig, serviceConfig: serviceConfig}
}

func (s *Server) Start() {
	s.access.Lock()
	defer s.access.Unlock()
	log.Infoln("server start")
	apiClient := api.New(s.apiConfig)
	nodeInfo, err := apiClient.GetNodeInfo()
	if err != nil {
		panic(fmt.Errorf("failed to get node inf :%s", err))
	}

	inBoundConfig, err := service.InboundBuilder(s.serviceConfig, nodeInfo)
	if err != nil {
		panic(fmt.Errorf("failed to build inbound config: %s", err))
	}

	outBoundConfig, err := service.OutboundBuilder(nodeInfo)
	if err != nil {
		panic(fmt.Errorf("failed to build outbound config: %s", err))
	}

	instance, err := s.loadCore(inBoundConfig, outBoundConfig)
	if err != nil {
		panic(err)
	}

	if err := instance.Start(); err != nil {
		panic(fmt.Errorf("failed to start instance: %s", err))
	}

	buildService := service.New(inBoundConfig.Tag, instance, s.serviceConfig, nodeInfo,
		apiClient.GetUserList, apiClient.ReportUserTraffic)
	s.service = buildService
	if err := s.service.Start(); err != nil {
		panic(fmt.Errorf("failed to start build service: %s", err))
	}
	s.Running = true
	s.instance = instance
	log.Infoln("server is running")
}

func (s *Server) loadCore(inboundConfig *core.InboundHandlerConfig, outboundConfig *core.OutboundHandlerConfig) (*core.Instance, error) {
	//Log Config
	logConfig := &conf.LogConfig{}
	logConfig.LogLevel = s.config.LogLevel
	pbLogConfig := logConfig.Build()

	//DNS Config
	dnsConfig := &conf.DNSConfig{}
	pbDnsConfig, _ := dnsConfig.Build()

	//Routing config
	routerConfig := &conf.RouterConfig{}
	pbRouteConfig, err := routerConfig.Build()

	//InboundConfig
	inboundConfigs := make([]*core.InboundHandlerConfig, 1)
	inboundConfigs[0] = inboundConfig

	//OutBound config
	outBoundConfigs := make([]*core.OutboundHandlerConfig, 1)
	outBoundConfigs[0] = outboundConfig

	//PolicyConfig
	policyConfig := &conf.PolicyConfig{}
	pbPolicy := &conf.Policy{
		StatsUserUplink:   true,
		StatsUserDownlink: true,
		Handshake:         &defaultConnectionConfig.Handshake,
		ConnectionIdle:    &defaultConnectionConfig.ConnIdle,
		UplinkOnly:        &defaultConnectionConfig.UplinkOnly,
		DownlinkOnly:      &defaultConnectionConfig.DownlinkOnly,
		BufferSize:        &defaultConnectionConfig.BufferSize,
	}
	policyConfig.Levels = map[uint32]*conf.Policy{0: pbPolicy}
	pbPolicyConfig, _ := policyConfig.Build()
	pbCoreConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(pbLogConfig),
			serial.ToTypedMessage(pbPolicyConfig),
			serial.ToTypedMessage(pbDnsConfig),
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(pbRouteConfig),
		},
		Outbound: outBoundConfigs,
		Inbound:  inboundConfigs,
	}
	instance, err := core.New(pbCoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %s", err)
	}
	return instance, nil
}

func (s *Server) Close() {
	s.access.Lock()
	defer s.access.Unlock()
	err := s.service.Close()
	if err != nil {
		log.Fatal("server close failed: %s", err)
	}
	log.Infoln("server close")
}
