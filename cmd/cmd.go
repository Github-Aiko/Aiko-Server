package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	version  = "TempVersion" //use ldflags replace
	codename = "Aiko-Server"
	intro    = "A V2board backend based on Xray-core"
)

func showVersion() {
	fmt.Printf("%s %s (%s) \n", codename, version, intro)
	// Warning
	fmt.Println(Warn("This version Using Only AikoPanel."))
	fmt.Println(Warn("This version changed config file. Please check config file before running."))
}

var command = &cobra.Command{
	Use: "Aiko-Server",
	PreRun: func(_ *cobra.Command, _ []string) {
		showVersion()
	},
}

func Run() {
	err := command.Execute()
	if err != nil {
		log.Println("execute failed, error:", err)
	}
}
