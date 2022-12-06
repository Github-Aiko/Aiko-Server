package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Github-Aiko/Aiko-Server/internal/app/server"
	"github.com/Github-Aiko/Aiko-Server/internal/pkg/api"
	"github.com/Github-Aiko/Aiko-Server/internal/pkg/service"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/xtls/xray-core/core"
)

const (
	Name          = "Aiko-Server"
	Version       = "0.1"
	CopyRight     = "AikoCute"
	LogLevelDebug = "debug"
	LogLevelError = "error"
	LogLevelInfo  = "info"
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}
	cli.ErrWriter = ioutil.Discard

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version=%s xray.version=%s\n", Version, core.Version())
	}
}

func main() {
	var config server.Config
	var apiConfig api.Config
	var serviceConfig service.Config
	var certConfig service.CertConfig

	app := &cli.App{
		Name:      Name,
		Version:   Version,
		Copyright: CopyRight,
		Usage:     "Provide trojan service for the v2Board(XFLASH-PANDA)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "api",
				Usage:       "Server address",
				EnvVars:     []string{"AIKO_SERVER_API", "API"},
				Required:    true,
				Destination: &apiConfig.APIHost,
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "Token of server API",
				EnvVars:     []string{"AIKO_SERVER_TOKEN", "TOKEN"},
				Required:    true,
				Destination: &apiConfig.Token,
			},
			&cli.StringFlag{
				Name:        "nodetype",
				Usage:       "NodeType of server",
				EnvVars:     []string{"AIKO_SERVER_NDOETYPE", "NODETYPE"},
				Required:    true,
				Destination: &apiConfig.NodeType,
			},
			&cli.IntFlag{
				Name:        "node",
				Usage:       "Node ID",
				EnvVars:     []string{"AIKO_SERVER_NODE", "NODE"},
				Required:    true,
				Destination: &apiConfig.NodeID,
			},
			&cli.StringFlag{
				Name:        "cert_file",
				Usage:       "Cert file",
				EnvVars:     []string{"AIKO_SERVER_CERT_FILE", "CERT_FILE"},
				Value:       "/root/.cert/server.crt",
				Required:    false,
				DefaultText: "/root/.cert/server.crt",
				Destination: &certConfig.CertFile,
			},
			&cli.StringFlag{
				Name:        "key_file",
				Usage:       "Key file",
				EnvVars:     []string{"AIKO_SERVER_KEY_FILE", "KEY_FILE"},
				Value:       "/root/.cert/server.key",
				Required:    false,
				DefaultText: "/root/.cert/server.key",
				Destination: &certConfig.KeyFile,
			},
			&cli.DurationFlag{
				Name:        "sys_interval",
				Usage:       "API request cycle, unit: second",
				EnvVars:     []string{"AIKO_SERVER_SYS_INTERVAL", "SYS_INTERVAL"},
				Value:       time.Second * 60,
				DefaultText: "60",
				Required:    false,
				Destination: &serviceConfig.SysInterval,
			},
			&cli.StringFlag{
				Name:        "log_mode",
				Value:       LogLevelError,
				Usage:       "Log mode",
				EnvVars:     []string{"AIKO_SERVER_LOG_LEVEL", "LOG_LEVEL"},
				Destination: &config.LogLevel,
				Required:    false,
			},
		},
		Before: func(c *cli.Context) error {
			log.SetFormatter(&log.TextFormatter{})
			if config.LogLevel == LogLevelDebug {
				log.SetFormatter(&log.TextFormatter{
					FullTimestamp: true,
				})
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
			} else if config.LogLevel == LogLevelInfo {
				log.SetLevel(log.InfoLevel)
			} else if config.LogLevel == LogLevelError {
				log.SetLevel(log.ErrorLevel)
			} else {
				return fmt.Errorf("log mode %s not supported", config.LogLevel)
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			if config.LogLevel != LogLevelDebug {
				defer func() {
					if e := recover(); e != nil {
						panic(e)
					}
				}()
			}
			serviceConfig.Cert = &certConfig
			serv := server.New(&config, &apiConfig, &serviceConfig)
			serv.Start()
			defer serv.Close()
			runtime.GC()
			{
				osSignals := make(chan os.Signal, 1)
				signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
				<-osSignals
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
