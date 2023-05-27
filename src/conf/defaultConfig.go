package conf

func getDefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		ListenIP:        "0.0.0.0",
		SendIP:          "0.0.0.0",
		DNSType:         "AsIs",
		DisableSniffing: true,
	}
}
