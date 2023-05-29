package command

import (
	"log"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
)

func installCountryRestriction() {
	execCommandStd("bash",
		"<(curl -Ls https://raw.githubusercontent.com/Github-Aiko/Aiko-Server/master/command/shellscript/CountryRestriction.sh)")
}

func checkubuntu() {
	// nếu là ubuntu thì chạy hàm installCountryRestriction ở trên còn không thì trả về không hỗ trợ hệ điều hành
	if osRelease, err := execCommand("lsb_release -is"); err == nil && osRelease == "Ubuntu" {
		installCountryRestriction()
	} else {
		log.Println("Your operating system is not supported")
	}
}

func Run(c *conf.ApiConfig) {
	if c.CountryRestriction {
		log.Println("CountryRestriction is enabled")
		checkubuntu()
	} else {
		log.Println("CountryRestriction is disabled")
	}
}
