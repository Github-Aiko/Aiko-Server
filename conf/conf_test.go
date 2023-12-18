package conf

import (
	"testing"
)

func TestConf_LoadFromPath(t *testing.T) {
	c := New()
	t.Log(c.LoadFromPath("../release/aiko.json"), c.NodeConfig)
}

func TestConf_Watch(t *testing.T) {
	c := New()
	t.Log(c.Watch("./1.json", "", "", func() {}))
	select {}
}
