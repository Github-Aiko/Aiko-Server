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
			exec.RunCommandStd("journalctl", "-u", "Aiko-Server.service", "-e", "--no-pager", "-f")
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
		fmt.Println(Err("AIKO-Server's launch failed"))
		return
	}
	if r {
		fmt.Println(Ok("AIKO-Server is running, no need to start again, if you need to restart, please choose to restart"))
	}
	_, err = exec.RunCommandByShell("systemctl start Aiko-Server.service")
	if err != nil {
		fmt.Println(Err("exec start cmd error: ", err))
		fmt.Println(Err("AIKO-Server launch failed"))
		return
	}
	time.Sleep(time.Second * 3)
	r, err = checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("AIKO-Server launch failed"))
	}
	if !r {
		fmt.Println(Err("Aiko-Server may start failure, please use Aiko-Server Log later to view log information"))
		return
	}
	fmt.Println(Ok("Aiko-Server starts successfully, please use Aiko-Server Log to view the running log"))
}

func stopHandle(_ *cobra.Command, _ []string) {
	_, err := exec.RunCommandByShell("systemctl stop Aiko-Server.service")
	if err != nil {
		fmt.Println(Err("exec stop cmd error: ", err))
		fmt.Println(Err("Aiko-Server stops failing"))
		return
	}
	time.Sleep(2 * time.Second)
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error:", err))
		fmt.Println(Err("Aiko-Server stops failing"))
		return
	}
	if r {
		fmt.Println(Err("Aiko-Server stops failing，It may be because the stop time exceeds two seconds, please check the log information later"))
		return
	}
	fmt.Println(Ok("Aiko-Server Stop success"))
}

func restartHandle(_ *cobra.Command, _ []string) {
	_, err := exec.RunCommandByShell("systemctl restart Aiko-Server.service")
	if err != nil {
		fmt.Println(Err("exec restart cmd error: ", err))
		fmt.Println(Err("Aiko-Server restart failed"))
		return
	}
	r, err := checkRunning()
	if err != nil {
		fmt.Println(Err("check status error: ", err))
		fmt.Println(Err("Aiko-Server restart failed"))
		return
	}
	if !r {
		fmt.Println(Err("Aiko-Server may start failure, please use Aiko-Server Log later to view log information"))
		return
	}
	fmt.Println(Ok("Aiko-Server restarted successfully"))
}
