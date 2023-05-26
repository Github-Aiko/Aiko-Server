package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

var (
	startCommand = cobra.Command{
		Use:   "start",
		Short: "Start Aiko-Server service",
		Run:   startHandle,
	}
	stopCommand = cobra.Command{
		Use:   "stop",
		Short: "Stop Aiko-Server service",
		Run:   stopHandle,
	}
	restartCommand = cobra.Command{
		Use:   "restart",
		Short: "Restart Aiko-Server service",
		Run:   restartHandle,
	}
	logCommand = cobra.Command{
		Use:   "log",
		Short: "Output Aiko-Server log",
		Run: func(_ *cobra.Command, _ []string) {
			execCommandStd("journalctl", "-u", "Aiko-Server.service", "-e", "--no-pager", "-f")
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
		fmt.Println(Err("Failed to start Aiko-Server"))
		return
	}
	if r {
		fmt.Println(Ok("Aiko-Server is already running, no need to start again, please choose restart if you need to restart"))
	}
	_, err = execCommand("systemctl start Aiko-Server.service")
	if err != nil {
		fmt.Println(Err("exec start cmd error: ", err))
		fmt.Println(Err("Failed to start Aiko-Server"))
		return
	}
	time.Sleep(time.Second * 3)
	r, err = checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("Failed to start Aiko-Server"))
	}
	if !r {
		fmt.Println(Err("Aiko-Server may have failed to start, please use Aiko-Server log to view the log information later"))
		return
	}
	fmt.Println(Ok("Aiko-Server started successfully, please use Aiko-Server log to view the running log"))
}

func stopHandle(_ *cobra.Command, _ []string) {
	_, err := execCommand("systemctl stop Aiko-Server.service")
	if err != nil {
		fmt.Println(Err("exec stop cmd error: ", err))
		fmt.Println(Err("Failed to stop Aiko-Server"))
		return
	}
	time.Sleep(2 * time.Second)
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error:", err))
		fmt.Println(Err("Failed to stop Aiko-Server"))
		return
	}
	if r {
		fmt.Println(Err("Failed to stop Aiko-Server, it may be because the stop time exceeded two seconds, please check the log information later"))
		return
	}
	fmt.Println(Ok("Aiko-Server stopped successfully"))
}

func restartHandle(_ *cobra.Command, _ []string) {
	_, err := execCommand("systemctl restart Aiko-Server.service")
	if err != nil {
		fmt.Println(Err("exec restart cmd error: ", err))
		fmt.Println(Err("Failed to restart Aiko-Server"))
		return
	}
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("Failed to restart Aiko-Server"))
		return
	}
	if !r {
		fmt.Println(Err("Aiko-Server may have failed to start, please use Aiko-Server log to view the log information later"))
		return
	}
	fmt.Println(Ok("Aiko-Server restarted successfully"))
}
