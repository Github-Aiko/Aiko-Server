package cmd

import (
	log "github.com/sirupsen/logrus"

	_ "github.com/Github-Aiko/Aiko-Server/src/core/imports"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use: "AikoR",
}

func Run() {
	err := command.Execute()
	if err != nil {
		log.WithField("err", err).Error("Execute command failed")
	}
}
