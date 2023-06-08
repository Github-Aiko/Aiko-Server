package cmd

import (
	"fmt"
	"strings"

	vCore "github.com/Github-Aiko/Aiko-Server/src/core"
	"github.com/spf13/cobra"
)

var (
	version  = "TempVersion" //use ldflags replace
	codename = "Aiko-Server"
	intro    = "A V2board backend based on multi core"
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
	fmt.Printf("%s %s (%s) \n", codename, version, intro)
	fmt.Printf("Supported cores: %s\n", strings.Join(vCore.RegisteredCore(), ", "))
	// Warning
	fmt.Println(Warn("This version need V2board version >= 1.7.0."))
}
