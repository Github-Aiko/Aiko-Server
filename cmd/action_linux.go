package cmd

import (
	"fmt"
	"github.com/Github-Aiko/Aiko-Server/src/common/exec"
	"github.com/spf13/cobra"
	"time"
)

var (
	startCommand = cobra.Command{
		Use:   "start",
		Short: "Start AikoR service",
		Run:   startHandle,
	}
	stopCommand = cobra.Command{
		Use:   "stop",
		Short: "Stop AikoR service",
		Run:   stopHandle,
	}
	restartCommand = cobra.Command{
		Use:   "restart",
		Short: "Restart AikoR service",
		Run:   restartHandle,
	}
	logCommand = cobra.Command{
		Use:   "log",
		Short: "Output AikoR log",
		Run: func(_ *cobra.Command, _ []string) {
			exec.RunCommandStd("journalctl", "-u", "AikoR.service", "-e", "--no-pager", "-f")
		},
	}
)

func init() {
	command.AddCommand(&startCommand)
	command.AddCommand(&stopCommand)
	command.AddCommand(&restartCommand)
	command.AddCommand(&logCommand)
}

func startHandle(_ *cobra.Command, _ []string) {
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("Failed to start AikoR"))
		return
	}
	if r {
		fmt.Println(Ok("AikoR is already running, no need to start again. If you want to restart, please use the restart command"))
	}
	_, err = exec.RunCommandByShell("systemctl start AikoR.service")
	if err != nil {
		fmt.Println(Err("exec start cmd error: ", err))
		fmt.Println(Err("Failed to start AikoR"))
		return
	}
	time.Sleep(time.Second * 3)
	r, err = checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("Failed to start AikoR"))
	}
	if !r {
		fmt.Println(Err("AikoR may have failed to start, please use the AikoR log command to view the log information later"))
		return
	}
	fmt.Println(Ok("AikoR started successfully, please use the AikoR log command to view the running log"))
}

func stopHandle(_ *cobra.Command, _ []string) {
	_, err := exec.RunCommandByShell("systemctl stop AikoR.service")
	if err != nil {
		fmt.Println(Err("exec stop cmd error: ", err))
		fmt.Println(Err("Failed to stop AikoR"))
		return
	}
	time.Sleep(2 * time.Second)
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error:", err))
		fmt.Println(Err("Failed to stop AikoR"))
		return
	}
	if r {
		fmt.Println(Err("Failed to stop AikoR, it may be because the stop time exceeded two seconds, please check the log information later"))
		return
	}
	fmt.Println(Ok("AikoR stopped successfully"))
}

func restartHandle(_ *cobra.Command, _ []string) {
	_, err := exec.RunCommandByShell("systemctl restart AikoR.service")
	if err != nil {
		fmt.Println(Err("exec restart cmd error: ", err))
		fmt.Println(Err("Failed to restart AikoR"))
		return
	}
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("Failed to restart AikoR"))
		return
	}
	if !r {
		fmt.Println(Err("AikoR may have failed to start, please use the AikoR log command to view the log information later"))
		return
	}
	fmt.Println(Ok("AikoR restarted successfully"))
}
