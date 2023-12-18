package cmd

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/Github-Aiko/Aiko-Server/conf"
	vCore "github.com/Github-Aiko/Aiko-Server/core"
	_ "github.com/Github-Aiko/Aiko-Server/core/imports"
	"github.com/Github-Aiko/Aiko-Server/limiter"
	"github.com/Github-Aiko/Aiko-Server/node"
	"github.com/spf13/cobra"
)

var (
	config string
	watch  bool
)

func init() {
	command.PersistentFlags().
		StringVarP(&config, "config", "c",
			"/etc/Aiko-Server/aiko.json", "config file path")
	command.PersistentFlags().
		BoolVarP(&watch, "watch", "w",
			true, "watch file path change")
}

var command = &cobra.Command{
	Use:   "Aiko-Server",
	Short: "Run Aiko-Server server",
	Run:   serverHandle,
	Args:  cobra.NoArgs,
}

func Run() {
	err := command.Execute()
	if err != nil {
		log.WithField("err", err).Error("Execute command failed")
	}
}

func serverHandle(_ *cobra.Command, _ []string) {
	showVersion()
	c := conf.New()
	err := c.LoadFromPath(config)
	if err != nil {
		log.WithField("err", err).Error("Load config file failed")
		return
	}
	switch c.LogConfig.Level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
	if c.LogConfig.Output != "" {
		w := &lumberjack.Logger{
			Filename:   c.LogConfig.Output,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
		log.SetOutput(w)
	}
	limiter.Init()
	log.Info("Start Aiko-Server...")
	vc, err := vCore.NewCore(c.CoresConfig)
	if err != nil {
		log.WithField("err", err).Error("new core failed")
		return
	}
	err = vc.Start()
	if err != nil {
		log.WithField("err", err).Error("Start core failed")
		return
	}
	defer vc.Close()
	log.Info("Core ", vc.Type(), " started")
	nodes := node.New()
	err = nodes.Start(c.NodeConfig, vc)
	if err != nil {
		log.WithField("err", err).Error("Run nodes failed")
		return
	}
	log.Info("Nodes started")
	xdns := os.Getenv("XRAY_DNS_PATH")
	sdns := os.Getenv("SING_DNS_PATH")
	if watch {
		err = c.Watch(config, xdns, sdns, func() {
			nodes.Close()
			err = vc.Close()
			if err != nil {
				log.WithField("err", err).Error("Restart node failed")
				return
			}
			vc, err = vCore.NewCore(c.CoresConfig)
			if err != nil {
				log.WithField("err", err).Error("New core failed")
				return
			}
			err = vc.Start()
			if err != nil {
				log.WithField("err", err).Error("Start core failed")
				return
			}
			log.Info("Core ", vc.Type(), " restarted")
			err = nodes.Start(c.NodeConfig, vc)
			if err != nil {
				log.WithField("err", err).Error("Run nodes failed")
				return
			}
			log.Info("Nodes restarted")
			runtime.GC()
		})
		if err != nil {
			log.WithField("err", err).Error("start watch failed")
			return
		}
	}
	// clear memory
	runtime.GC()
	// wait exit signal
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		<-osSignals
	}
}
