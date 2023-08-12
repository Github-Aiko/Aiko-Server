package panel

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Github-Aiko/Aiko-Server/src/common/crypt"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
)

type CommonNodeRsp struct {
	Host       string     `json:"host"`
	ServerPort int        `json:"server_port"`
	ServerName string     `json:"server_name"`
	Routes     []Route    `json:"routes"`
	BaseConfig BaseConfig `json:"base_config"`
}

type Route struct {
	Id          int         `json:"id"`
	Match       interface{} `json:"match"`
	Action      string      `json:"action"`
	ActionValue string      `json:"action_value"`
}
type BaseConfig struct {
	PushInterval any `json:"push_interval"`
	PullInterval any `json:"pull_interval"`
}

type V2rayNodeRsp struct {
	Tls             int             `json:"tls"`
	Network         string          `json:"network"`
	NetworkSettings json.RawMessage `json:"networkSettings"`
	ServerName      string          `json:"server_name"`
}

type ShadowsocksNodeRsp struct {
	Cipher    string `json:"cipher"`
	ServerKey string `json:"server_key"`
}

type HysteriaNodeRsp struct {
	UpMbps   int    `json:"up_mbps"`
	DownMbps int    `json:"down_mbps"`
	Obfs     string `json:"obfs"`
}

type NodeInfo struct {
	Id              int
	Type            string
	Rules           Rules
	Host            string
	Port            int
	Network         string
	ExtraConfig     V2rayExtraConfig
	NetworkSettings json.RawMessage
	Tls             bool
	ServerName      string
	UpMbps          int
	DownMbps        int
	ServerKey       string
	Cipher          string
	HyObfs          string

	PushInterval time.Duration
	PullInterval time.Duration
}

type Rules struct {
	Regexp   []string
	Protocol []string
}

type V2rayExtraConfig struct {
	EnableVless   string         `json:"EnableVless"`
	VlessFlow     string         `json:"VlessFlow"`
	EnableReality string         `json:"EnableReality"`
	RealityConfig *RealityConfig `json:"RealityConfig"`
}

type RealityConfig struct {
	Dest         string   `yaml:"Dest" json:"Dest"`
	Xver         string   `yaml:"Xver" json:"Xver"`
	ServerNames  []string `yaml:"ServerNames" json:"ServerNames"`
	PrivateKey   string   `yaml:"PrivateKey" json:"PrivateKey"`
	MinClientVer string   `yaml:"MinClientVer" json:"MinClientVer"`
	MaxClientVer string   `yaml:"MaxClientVer" json:"MaxClientVer"`
	MaxTimeDiff  string   `yaml:"MaxTimeDiff" json:"MaxTimeDiff"`
	ShortIds     []string `yaml:"ShortIds" json:"ShortIds"`
}

type XrayDNSConfig struct {
	Servers []interface{} `json:"servers"`
	Tag     string        `json:"tag"`
}

type SingDNSConfig struct {
	Servers []map[string]string      `json:"servers"`
	Rules   []map[string]interface{} `json:"rules"`
}

var initFinish bool

func (c *Client) GetNodeInfo() (node *NodeInfo, err error) {
	const path = "/api/v1/server/UniProxy/config"
	r, err := c.client.
		R().
		SetHeader("If-None-Match", c.nodeEtag).
		Get(path)
	if err = c.checkResponse(r, path, err); err != nil {
		return
	}
	if r.StatusCode() == 304 {
		return nil, nil
	}
	// parse common params
	node = &NodeInfo{
		Id:   c.NodeId,
		Type: c.NodeType,
	}
	common := CommonNodeRsp{}

	var env = os.Getenv("CORE_RUNNING")
	XrayDnsPath := os.Getenv("XRAY_DNS_PATH")
	SingDnsPath := os.Getenv("SING_DNS_PATH")
	type dnsConfigStruct struct {
		Config  interface{}
		DnsPath string
	}
	var dnsConfigMap = map[string]dnsConfigStruct{
		"SING": {Config: SingDNSConfig{
			Servers: []map[string]string{
				{
					"tag":     "default",
					"address": "https://8.8.8.8/dns-query",
					"detour":  "direct",
				},
			},
		}, DnsPath: SingDnsPath},
		"XRAY": {Config: XrayDNSConfig{
			Servers: []interface{}{
				"1.1.1.1",
				"localhost"},
			Tag: "dns_inbound",
		}, DnsPath: XrayDnsPath},
	}
	var dnsConfigObj, _ = dnsConfigMap[env]
	var isDnsConfigUpdating bool
	err = json.Unmarshal(r.Body(), &common)
	if err != nil {
		return nil, fmt.Errorf("decode common params error: %s", err)
	}
	for i := range common.Routes { // parse rules from routes
		var matchs []string
		if _, ok := common.Routes[i].Match.(string); ok {
			matchs = strings.Split(common.Routes[i].Match.(string), ",")
		} else if _, ok = common.Routes[i].Match.([]string); ok {
			matchs = common.Routes[i].Match.([]string)
		} else {
			temp := common.Routes[i].Match.([]interface{})
			matchs = make([]string, len(temp))
			for i := range temp {
				matchs[i] = temp[i].(string)
			}
		}
		switch common.Routes[i].Action {
		case "block":
			for _, v := range matchs {
				if strings.HasPrefix(v, "protocol:") {
					// protocol
					node.Rules.Protocol = append(node.Rules.Protocol, strings.TrimPrefix(v, "protocol:"))
				} else {
					// domain
					node.Rules.Regexp = append(node.Rules.Regexp, strings.TrimPrefix(v, "regexp:"))
				}
			}
		case "dns":
			if matchs[0] != "main" {
				switch env {
				case "SING":
					dnsConfig, _ := dnsConfigObj.Config.(SingDNSConfig)
					updateSingDnsConfig(matchs, common.Routes[i], &dnsConfig)
					dnsConfigObj.Config = dnsConfig
				case "XRAY":
					dnsConfig, _ := dnsConfigObj.Config.(XrayDNSConfig)
					updateXrayDnsConfig(matchs, common.Routes[i], &dnsConfig)
					dnsConfigObj.Config = dnsConfig
				}
				isDnsConfigUpdating = true
			} else {
				dns := []byte(strings.Join(matchs[1:], ""))
				switch env {
				case "SING":
					go saveDnsConfig(dns, SingDnsPath)
				case "XRAY":
					go saveDnsConfig(dns, XrayDnsPath)
				}
				isDnsConfigUpdating = false
			}
		}
	}
	if isDnsConfigUpdating {
		dnsConfigJSON, err := json.MarshalIndent(dnsConfigObj.Config, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling dnsConfig to JSON:", err)
		} else {
			switch env {
			case "SING":
				go saveDnsConfig(dnsConfigJSON, SingDnsPath)
			case "XRAY":
				go saveDnsConfig(dnsConfigJSON, XrayDnsPath)
			}
		}
	}
	node.ServerName = common.ServerName
	node.Host = common.Host
	node.Port = common.ServerPort
	node.PullInterval = intervalToTime(common.BaseConfig.PullInterval)
	node.PushInterval = intervalToTime(common.BaseConfig.PushInterval)
	// parse protocol params
	switch c.NodeType {
	case "v2ray", "vless":
		rsp := V2rayNodeRsp{}
		err = json.Unmarshal(r.Body(), &rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		node.Network = rsp.Network
		node.NetworkSettings = rsp.NetworkSettings
		node.ServerName = rsp.ServerName
		if rsp.Tls == 1 {
			node.Tls = true
		}
		err = json.Unmarshal(rsp.NetworkSettings, &node.ExtraConfig)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray extra error: %s", err)
		}
		if node.ExtraConfig.EnableVless == "true" {
			node.Type = "vless"
		}
		if node.ExtraConfig.EnableReality == "true" {
			if node.ExtraConfig.RealityConfig == nil {
				node.ExtraConfig.EnableReality = "false"
			} else {
				key := crypt.GenX25519Private([]byte(strconv.Itoa(c.NodeId) + c.NodeType + c.Token +
					node.ExtraConfig.RealityConfig.PrivateKey))
				node.ExtraConfig.RealityConfig.PrivateKey = base64.RawURLEncoding.EncodeToString(key)
			}
		}
	case "shadowsocks":
		rsp := ShadowsocksNodeRsp{}
		err = json.Unmarshal(r.Body(), &rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		node.ServerKey = rsp.ServerKey
		node.Cipher = rsp.Cipher
	case "trojan":
		node.Tls = true
	case "hysteria":
		rsp := HysteriaNodeRsp{}
		err = json.Unmarshal(r.Body(), &rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		node.DownMbps = rsp.DownMbps
		node.UpMbps = rsp.UpMbps
		node.HyObfs = rsp.Obfs
	case "tuic":
		//NONE
	}
	c.nodeEtag = r.Header().Get("ETag")
	return
}

func intervalToTime(i interface{}) time.Duration {
	switch reflect.TypeOf(i).Kind() {
	case reflect.Int:
		return time.Duration(i.(int)) * time.Second
	case reflect.String:
		i, _ := strconv.Atoi(i.(string))
		return time.Duration(i) * time.Second
	case reflect.Float64:
		return time.Duration(i.(float64)) * time.Second
	default:
		return time.Duration(reflect.ValueOf(i).Int()) * time.Second
	}
}

func saveDnsConfig(dns []byte, dnsPath string) {
	if !initFinish {
		time.Sleep(5 * time.Second)
	}
	if dnsPath == "" {
		return
	}
	currentData, err := os.ReadFile(dnsPath)
	if err != nil {
		log.WithField("err", err).Error("Failed to read DNS_PATH")
		return
	}
	if !bytes.Equal(currentData, dns) {
		if err = os.Truncate(dnsPath, 0); err != nil {
			log.WithField("err", err).Error("Failed to clear XRAY DNS PATH file")
		}
		if err = os.WriteFile(dnsPath, dns, 0644); err != nil {
			log.WithField("err", err).Error("Failed to write DNS to XRAY DNS PATH file")
		}
	}
	initFinish = true
}

func updateXrayDnsConfig(matchs []string, common Route, dnsConfig *XrayDNSConfig) {
	var domains []string
	for _, v := range matchs {
		domains = append(domains, v)
	}
	dnsConfig.Servers = append(dnsConfig.Servers,
		map[string]interface{}{
			"address": common.ActionValue,
			"domains": domains,
		},
	)
}

func updateSingDnsConfig(matchs []string, common Route, dnsConfig *SingDNSConfig) {
	dnsConfig.Servers = append(dnsConfig.Servers,
		map[string]string{
			"tag":              strconv.Itoa(common.Id),
			"address":          common.ActionValue,
			"address_resolver": "default",
			"detour":           "direct",
		},
	)
	rule := map[string]interface{}{
		"server":        strconv.Itoa(common.Id),
		"disable_cache": true,
	}

	for _, ruleType := range []string{"domain_suffix", "domain_keyword", "domain_regex", "geosite"} {
		var domains []string
		for _, v := range matchs {
			split := strings.SplitN(v, ":", 2)
			prefix := strings.ToLower(split[0])
			if prefix == ruleType || (prefix == "domain" && ruleType == "domain_suffix") {
				if len(split) > 1 {
					domains = append(domains, split[1])
				}
				if len(domains) > 0 {
					rule[ruleType] = domains
				}
			}
		}
	}
	dnsConfig.Rules = append(dnsConfig.Rules, rule)
}
