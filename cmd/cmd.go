package cmd

import (
	"log"

	_ "github.com/Github-Aiko/Aiko-Server/src/core/imports"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use: "Aiko-Server",
}

func Run() {
	err := command.Execute()
	if err != nil {
		log.Println("execute failed, error:", err)
	}
}
