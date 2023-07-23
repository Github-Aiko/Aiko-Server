package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Github-Aiko/Aiko-Server/src/common/crypt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/curve25519"
)

var x25519Command = cobra.Command{
	Use:   "x25519",
	Short: "Generate key pair for x25519 key exchange",
	Run: func(cmd *cobra.Command, args []string) {
		executeX25519()
	},
}

func init() {
	command.AddCommand(&x25519Command)
}

func executeX25519() {
	var output string
	var err error
	defer func() {
		fmt.Println(output)
	}()
	var privateKey []byte
	var publicKey []byte
	var yes, key string
	fmt.Println("Do you want to generate a key based on the node information?(Y/n)")
	fmt.Scan(&yes)
	if strings.ToLower(yes) == "y" {
		var temp string
		fmt.Println("Please enter the node ID:")
		fmt.Scan(&temp)
		key = temp
		fmt.Println("Please enter the node type:")
		fmt.Scan(&temp)
		key += strings.ToLower(temp)
		fmt.Println("Please enter the Token:")
		fmt.Scan(&temp)
		key += temp
		privateKey = crypt.GenX25519Private([]byte(key))
	} else {
		privateKey = make([]byte, curve25519.ScalarSize)
		if _, err = rand.Read(privateKey); err != nil {
			output = Err("read rand error: ", err)
			return
		}
	}
	if publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint); err != nil {
		output = Err("gen X25519 error: ", err)
		return
	}
	p := base64.RawURLEncoding.EncodeToString(privateKey)
	output = fmt.Sprint("Private key: ",
		p,
		"\nPublic key: ",
		base64.RawURLEncoding.EncodeToString(publicKey))
}
