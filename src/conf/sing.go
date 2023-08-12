package conf

type SingConfig struct {
	LogConfig     SingLogConfig `yaml:"Log"`
	OriginalPath  string        `yaml:"OriginalPath"`
	DnsConfigPath string        `yaml:"DnsConfigPath"`
}

type SingLogConfig struct {
	Disabled  bool   `yaml:"Disable"`
	Level     string `yaml:"Level"`
	Output    string `yaml:"Output"`
	Timestamp bool   `yaml:"Timestamp"`
}

func NewSingConfig() *SingConfig {
	return &SingConfig{
		LogConfig: SingLogConfig{
			Level:     "error",
			Timestamp: true,
		},
	}
}
