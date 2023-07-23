package cmd

import (
	"fmt"
	"github.com/Github-Aiko/Aiko-Server/common/exec"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var targetVersion string

var (
	updateCommand = cobra.Command{
		Use:   "update",
		Short: "Update AikoR version",
		Run: func(_ *cobra.Command, _ []string) {
			exec.RunCommandStd("bash",
				"<(curl -Ls https://raw.githubusercontent.com/Github-Aiko/Aiko-Server-script/master/install.sh)",
				targetVersion)
		},
		Args: cobra.NoArgs,
	}
	uninstallCommand = cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall AikoR",
		Run:   uninstallHandle,
	}
)

func init() {
	updateCommand.PersistentFlags().StringVar(&targetVersion, "version", "", "update target version")
	command.AddCommand(&updateCommand)
	command.AddCommand(&uninstallCommand)
}

func uninstallHandle(_ *cobra.Command, _ []string) {
	var yes string
	fmt.Println(Warn("Are you sure you want to uninstall AikoR?(Y/n)"))
	fmt.Scan(&yes)
	if strings.ToLower(yes) != "y" {
		fmt.Println("Uninstallation canceled")
	}
	_, err := exec.RunCommandByShell("systemctl stop AikoR && systemctl disable AikoR")
	if err != nil {
		fmt.Println(Err("exec cmd error: ", err))
		fmt.Println(Err("Uninstallation failed"))
		return
	}
	_ = os.RemoveAll("/etc/systemd/system/AikoR.service")
	_ = os.RemoveAll("/etc/AikoR/")
	_ = os.RemoveAll("/usr/local/AikoR/")
	_ = os.RemoveAll("/bin/AikoR")
	_, err = exec.RunCommandByShell("systemctl daemon-reload && systemctl reset-failed")
	if err != nil {
		fmt.Println(Err("exec cmd error: ", err))
		fmt.Println(Err("Uninstallation failed"))
		return
	}
	fmt.Println(Ok("Uninstallation successful"))
}
