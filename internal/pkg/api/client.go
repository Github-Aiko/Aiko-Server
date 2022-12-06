package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// Config  api config
type Config struct {
	APIHost  string
	NodeID   int
	Token    string
	NodeType string
	Timeout  int
}

// Client APIClient create a api client to the panel.
type Client struct {
	client *resty.Client
	config *Config
}

// New creat a api instance
func New(apiConfig *Config) *Client {
	client := resty.New()
	client.SetRetryCount(3)
	if apiConfig.Timeout > 0 {
		client.SetTimeout(time.Duration(apiConfig.Timeout) * time.Second)
	} else {
		client.SetTimeout(5 * time.Second)
	}
	client.OnError(func(req *resty.Request, err error) {
		if v, ok := err.(*resty.ResponseError); ok {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			log.Errorln(v.Err)
		}
	})
	client.SetBaseURL(apiConfig.APIHost)
	// Create Key for each requests
	client.SetQueryParam("key", apiConfig.Token)
	client.SetQueryParams(map[string]string{
		"node_id": strconv.Itoa(apiConfig.NodeID),
		"token":   apiConfig.Token,
	})

	apiClient := &Client{
		client: client,
		config: apiConfig,
	}
	return apiClient
}

// Describe return a description of the client
func (c *Client) Describe() *ClientInfo {
	return &ClientInfo{APIHost: c.config.APIHost, NodeID: c.config.NodeID, Token: c.config.Token, NodeType: c.config.NodeType}
}

// Debug set the client debug for client
func (c *Client) Debug() {
	c.client.SetDebug(true)
}

func (c *Client) assembleURL(path string) string {
	return c.config.APIHost + path
}

// GetNodeInfo will pull NodeInfo Config from sspanel
func (c *Client) GetNodeInfo() (nodeInfo *NodeInfo, err error) {
	var path string
	switch c.config.NodeType {
	case "Trojan":
		path = "/api/v1/server/trojan/config"
	case "V2ray":
		path = "/api/v1/server/vmess/config"
	case "Shadowsocks":
		path = "/api/v1/server/shadowsocks/config"
	default:
		return nil, fmt.Errorf("node type %s not support", c.config.NodeType)
	}
	res, err := c.client.R().
		ForceContentType("application/json").
		SetQueryParam("node_id", strconv.Itoa(c.config.NodeID)).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return nil, fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}

	var repNodeInfo RepNodeInfo
	if err := json.Unmarshal(res.Body(), &repNodeInfo); err != nil {
		return nil, fmt.Errorf("parse node info failed: %s", err)
	}

	if len(repNodeInfo.Message) > 0 {
		return nil, fmt.Errorf("api error, message: %s", repNodeInfo.Message)
	}
	return repNodeInfo.Data, nil
}

// GetUserList will pull user form sspanel
func (c *Client) GetUserList() (UserList *[]UserInfo, err error) {
	var path string
	switch c.config.NodeType {
	case "Trojan":
		path = "/api/v1/server/trojan/users"
	case "V2ray":
		path = "/api/v1/server/vmess/users"
	case "Shadowsocks":
		path = "/api/v1/server/shadowsocks/users"
	default:
		return nil, fmt.Errorf("node type %s not support", c.config.NodeType)
	}
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.config.NodeID)).
		ForceContentType("application/json").
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return nil, fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}

	var repUserList RepUserList
	if err := json.Unmarshal(res.Body(), &repUserList); err != nil {
		return nil, fmt.Errorf("parse node info failed: %s", err)
	}

	if len(repUserList.Message) > 0 {
		return nil, fmt.Errorf("api error, message: %s", repUserList.Message)
	}

	return repUserList.Data, nil
}

// ReportUserTraffic reports the user traffic
func (c *Client) ReportUserTraffic(userTraffic []*UserTraffic) error {
	var path string
	switch c.config.NodeType {
	case "Trojan":
		path = "/api/v1/server/trojan/submit"
	case "V2ray":
		path = "/api/v1/server/vmess/submit"
	case "Shadowsocks":
		path = "/api/v1/server/shadowsocks/submit"
	default:
		return fmt.Errorf("node type %s not support", c.config.NodeType)
	}

	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.config.NodeID)).
		SetBody(userTraffic).
		ForceContentType("application/json").
		Post(path)
	if err != nil {
		return fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}

	var repUserTraffic RepUserTraffic
	if err := json.Unmarshal(res.Body(), &repUserTraffic); err != nil {
		return fmt.Errorf("parse node info failed: %s", err)
	}
	if len(repUserTraffic.Message) > 0 {
		return fmt.Errorf("api error, message: %s", repUserTraffic.Message)
	}
	return nil
}
