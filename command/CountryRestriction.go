package command

import (
	"log"
	"os"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
)

func installCountryRestriction() {
	execCommandStd("bash",
		"<(curl -Ls https://raw.githubusercontent.com/Github-Aiko/Aiko-Server/master/command/shellscript/CountryRestriction.sh)")
	// add file to check if CountryRestriction is installed
	execCommandStd("touch", "/etc/Aiko-Server/CountryRestriction")
}

func checkubuntu() {
	// iff is ubuntu or debian then install CountryRestriction
	if osRelease, err := execCommand("lsb_release -is"); err == nil && (osRelease == "Ubuntu" || osRelease == "Debian") {
		log.Println("Your operating system is supported")
	} else {
		log.Println("Your operating system is not supported")
	}
}

func Run(c *conf.ApiConfig) {
	if c.CountryRestriction {
		log.Println("CountryRestriction is enabled")
		if _, err := os.Stat("/etc/Aiko-Server/CountryRestriction"); os.IsNotExist(err) {
			checkubuntu()
			installCountryRestriction()
		}
	} else {
		log.Println("CountryRestriction is disabled")
	}
}
