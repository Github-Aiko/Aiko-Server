package panel

import (
	"log"
	"testing"

	"github.com/Github-Aiko/Aiko-Server/src/conf"
)

var client *Client

func init() {
	c, err := New(&conf.ApiConfig{
		APIHost:      "http://127.0.0.1",
		NodeID:       1,
		Key:          "token",
		NodeType:     "hysteria",
		Timeout:      0,
		RuleListPath: "",
	})
	if err != nil {
		log.Panic(err)
	}
	client = c
}

func TestClient_GetNodeInfo(t *testing.T) {
	log.Println(client.GetNodeInfo())
}

func TestClient_ReportUserTraffic(t *testing.T) {
	log.Println(client.ReportUserTraffic([]UserTraffic{
		{
			UID:      10372,
			Upload:   1000,
			Download: 1000,
		},
	}))
}
