package cmd

import (
	"fmt"
	"strings"

	vCore "github.com/Github-Aiko/Aiko-Server/core"
	"github.com/spf13/cobra"
)

var (
	version  = "TempVersion" //use ldflags replace
	codename = "Aiko-Server"
	intro    = "A AikoBackend backend based on multi core"
)

var versionCommand = cobra.Command{
	Use:   "version",
	Short: "Print version info",
	Run: func(_ *cobra.Command, _ []string) {
		showVersion()
	},
}

func init() {
	command.AddCommand(&versionCommand)
}

func showVersion() {
	fmt.Println(` 
			____                          
    		/ /\ \     _   _   _    __     
		  / /  \ \   | | | | / / /     \  
		 / /____\ \  | | | |/ / |   _   | 
		/ /______\ \ | | | |\ \ |  (_)  | 
	   /_/        \_\|_| |_| \_\ \ ___ /  `)

	fmt.Printf("%s %s (%s) \n", codename, version, intro)
	fmt.Printf("Supported cores: %s\n", strings.Join(vCore.RegisteredCore(), ", "))
	// Warning
	fmt.Println(Warn("This Backend Support Only AikoPanelv2."))
	fmt.Println(Warn("The version have many changed for config, please check your config file"))
}
