package conf

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type Conf struct {
	CoreConfig  CoreConfig    `yaml:"CoreConfig"`
	NodesConfig []*NodeConfig `yaml:"Nodes"`
}

func New() *Conf {
	return &Conf{
		CoreConfig: CoreConfig{
			Type: "xray",
			XrayConfig: &XrayConfig{
				LogConfig:          NewLogConfig(),
				AssetPath:          "/etc/Aiko-Server/",
				DnsConfigPath:      "",
				InboundConfigPath:  "",
				OutboundConfigPath: "",
				RouteConfigPath:    "",
				ConnectionConfig:   NewConnectionConfig(),
			},
		},
		NodesConfig: []*NodeConfig{},
	}
}

func (p *Conf) LoadFromPath(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open config file error: %s", err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read file error: %s", err)
	}
	err = yaml.Unmarshal(content, p)
	if err != nil {
		return fmt.Errorf("decode config error: %s", err)
	}
	old := &OldConfig{}
	err = yaml.Unmarshal(content, old)
	if err == nil {
		migrateOldConfig(p, old)
	}
	return nil
}

func (p *Conf) Watch(filePath string, reload func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("new watcher error: %s", err)
	}
	go func() {
		var pre time.Time
		defer watcher.Close()
		for {
			select {
			case e := <-watcher.Events:
				if e.Has(fsnotify.Chmod) {
					continue
				}
				if pre.Add(1 * time.Second).After(time.Now()) {
					continue
				}
				time.Sleep(2 * time.Second)
				pre = time.Now()
				log.Println("config dir changed, reloading...")
				*p = *New()
				err := p.LoadFromPath(filePath)
				if err != nil {
					log.Printf("reload config error: %s", err)
				}
				reload()
				log.Println("reload config success")
			case err := <-watcher.Errors:
				if err != nil {
					log.Printf("File watcher error: %s", err)
				}
			}
		}
	}()
	err = watcher.Add(path.Dir(filePath))
	if err != nil {
		return fmt.Errorf("watch file error: %s", err)
	}
	return nil
}
