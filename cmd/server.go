package cmd

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
	"github.com/Github-Aiko/Aiko-Server/src/core"
	"github.com/Github-Aiko/Aiko-Server/src/limiter"
	"github.com/Github-Aiko/Aiko-Server/src/node"
	"github.com/spf13/cobra"
)

var (
	config string
	watch  bool
)

var serverCommand = cobra.Command{
	Use:   "server",
	Short: "Run Aiko-Server server",
	Run:   serverHandle,
	Args:  cobra.NoArgs,
}

func init() {
	serverCommand.PersistentFlags().
		StringVarP(&config, "config", "c",
			"/etc/Aiko-Server/aiko.yml", "config file path")
	serverCommand.PersistentFlags().
		BoolVarP(&watch, "watch", "w",
			true, "watch file path change")
	command.AddCommand(&serverCommand)
}

func serverHandle(_ *cobra.Command, _ []string) {
	c := conf.New()
	err := c.LoadFromPath(config)
	if err != nil {
		log.Fatalf("can't unmarshal config file: %s \n", err)
	}
	limiter.Init()
	log.Println("Start Aiko-Server...")
	x := core.New(c)
	err = x.Start()
	if err != nil {
		log.Fatalf("Start xray-core error: %s", err)
	}
	defer x.Close()
	nodes := node.New()
	err = nodes.Start(c.NodesConfig, x)
	if err != nil {
		log.Fatalf("Run nodes error: %s", err)
		return
	}
	if watch {
		err = c.Watch(config, func() {
			nodes.Close()
			err = x.Restart(c)
			if err != nil {
				log.Fatalf("Failed to restart xray-core: %s", err)
			}
			err = nodes.Start(c.NodesConfig, x)
			if err != nil {
				log.Fatalf("run nodes error: %s", err)
			}
			runtime.GC()
		})
		if err != nil {
			log.Fatalf("watch config file error: %s", err)
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
