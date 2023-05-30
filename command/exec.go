package command

import (
	"errors"
	"log"
	"os"
	"os/exec"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
)

func Run(c *conf.ApiConfig) {
	if c.CountryRestriction {
		log.Println("CountryRestriction is enabled")
		if _, err := os.Stat("/etc/Aiko-Server/CountryRestriction"); os.IsNotExist(err) {
			checkLinux()
			createTempFolder()
			downloadIPLocation(c.CountryRestrictionConfig.List, c.CountryRestrictionConfig.IpOtherList, c.CountryRestrictionConfig.UnlockPort)
			// touch file
			if _, err := execCommand("touch /etc/Aiko-Server/CountryRestriction"); err != nil {
				log.Printf("Error creating file: %s\n", err.Error())
				return
			}
		}
	} else {
		log.Println("CountryRestriction is disabled")
		os.RemoveAll("/etc/Aiko-Server/CountryRestriction")
		// reset iptables rules to default
		if _, err := execCommand("iptables -F"); err != nil {
			log.Printf("Error reset iptables rules: %s\n", err.Error())
			return
		}
	}
}

func execCommand(cmd string) (string, error) {
	e := exec.Command("bash", "-c", cmd)
	out, err := e.CombinedOutput()
	if errors.Unwrap(err) == exec.ErrNotFound {
		e = exec.Command("sh", "-c", cmd)
		out, err = e.CombinedOutput()
	}
	return string(out), err
}
