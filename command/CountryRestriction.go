package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func checkLinux() {
	// if the operating system is Linux then install CountryRestriction
	if osRelease, err := execCommand("uname -s"); err == nil && osRelease == "Linux" {
		log.Println("Your operating system is supported")
	} else {
		log.Println("Your operating system is not supported")
	}
}

func createTempFolder() {
	// create temp folder
	if _, err := execCommand("mkdir -p /etc/Aiko-Server/temp"); err != nil {
		log.Printf("Error creating temp folder: %s\n", err.Error())
		return
	}
}

func downloadIPLocation(LocationsList []string, IpOtherList []string) {
	var content []byte
	for _, LocationsList := range LocationsList {
		urlv4 := fmt.Sprintf("https://raw.githubusercontent.com/Github-Aiko/IPLocation/master/%s/ipv4.txt", strings.ToLower(LocationsList))
		urlv6 := fmt.Sprintf("https://raw.githubusercontent.com/Github-Aiko/IPLocation/master/%s/ipv6.txt", strings.ToLower(LocationsList))
		respv4, err := http.Get(urlv4)
		if err != nil {
			log.Printf("Error downloading content from %s: %s\n", urlv4, err.Error())
			continue
		}
		defer respv4.Body.Close()
		bodyv4, err := ioutil.ReadAll(respv4.Body)
		if err != nil {
			log.Printf("Error reading content from %s: %s\n", urlv4, err.Error())
			continue
		}
		content = append(content, bodyv4...)

		respv6, err := http.Get(urlv6)
		if err != nil {
			log.Printf("Error downloading content from %s: %s\n", urlv6, err.Error())
			continue
		}
		defer respv6.Body.Close()
		bodyv6, err := ioutil.ReadAll(respv6.Body)
		if err != nil {
			log.Printf("Error reading content from %s: %s\n", urlv6, err.Error())
			continue
		}
		content = append(content, bodyv6...)
	}
	err := ioutil.WriteFile("/etc/Aiko-Server/temp/output.txt", content, 0644)
	if err != nil {
		log.Printf("Error writing content to file: %s\n", err.Error())
		return
	}

	var content2 []byte
	// add IP addresses to the output.txt file
	for _, ip := range IpOtherList {
		content2 = append(content2, []byte(ip+"\n")...)
	}
	err = ioutil.WriteFile("/etc/Aiko-Server/temp/output.txt", content2, 0644)
	if err != nil {
		log.Printf("Error writing content to file: %s\n", err.Error())
		return
	}

	// read the content of output.txt
	data, err := ioutil.ReadFile("/etc/Aiko-Server/temp/output.txt")
	if err != nil {
		log.Printf("Error reading content from file: %s\n", err.Error())
		return
	}

	// split the content by newline
	lines := strings.Split(string(data), "\n")

	// create iptables command to add each IP address to the INPUT chain
	for _, line := range lines {
		if line != "" {
			cmd := exec.Command("iptables", "-A", "INPUT", "-s", line, "-j", "ACCEPT")
			err := cmd.Run()
			if err != nil {
				log.Printf("Error adding IP address %s to INPUT chain: %s\n", line, err.Error())
				continue
			}
		}
	}

	// block all input traffic other than the IP addresses in the INPUT chain
	_, err = execCommand("iptables -A INPUT -j DROP")
	if err != nil {
		log.Printf("Error blocking all input traffic: %s\n", err.Error())
		return
	}

	// delete temp folder
	_, err = execCommand("rm -rf /etc/Aiko-Server/temp")
	if err != nil {
		log.Printf("Error deleting temp folder: %s\n", err.Error())
		return
	}
}
