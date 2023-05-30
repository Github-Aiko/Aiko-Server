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
			downloadIPLocation(c.CountryRestrictionConfig)
		}
	} else {
		log.Println("CountryRestriction is disabled")
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
