package conf

import (
	"log"
	"testing"
)

func TestConf_LoadFromPath(t *testing.T) {
	c := New()
	t.Log(c.LoadFromPath("../Aiko-Server/aiko.yml.example"))
}

func TestConf_Watch(t *testing.T) {
	c := New()
	log.Println(c.Watch("../Aiko-Server/aiko.yml.example", "../Aiko-Server/aiko.yml", func() {
		log.Println(1)
	}))
	select {}
}
