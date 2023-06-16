package node

import (
	"errors"
	"fmt"
	"log"

	"github.com/Github-Aiko/Aiko-Server/api/limit"
	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/src/common/task"
	"github.com/Github-Aiko/Aiko-Server/src/conf"
	vCore "github.com/Github-Aiko/Aiko-Server/src/core"
	"github.com/Github-Aiko/Aiko-Server/src/limiter"
)

type Controller struct {
	server                    vCore.Core
	apiClient                 *panel.Client
	nodeInfo                  *panel.NodeInfo
	Tag                       string
	userList                  []panel.UserInfo
	ipRecorder                limit.IpRecorder
	nodeInfoMonitorPeriodic   *task.Task
	userReportPeriodic        *task.Task
	renewCertPeriodic         *task.Task
	dynamicSpeedLimitPeriodic *task.Task
	onlineIpReportPeriodic    *task.Task
	*conf.ControllerConfig
}

// NewController return a Node controller with default parameters.
func NewController(server vCore.Core, api *panel.Client, config *conf.ControllerConfig) *Controller {
	controller := &Controller{
		server:           server,
		ControllerConfig: config,
		apiClient:        api,
	}
	return controller
}

// Start implement the Start() function of the service interface
func (c *Controller) Start() error {
	// First fetch Node Info
	var err error
	c.nodeInfo, err = c.apiClient.GetNodeInfo()
	if err != nil {
		return fmt.Errorf("get node info error: %s", err)
	}
	// Update user
	c.userList, err = c.apiClient.GetUserList()
	if err != nil {
		return fmt.Errorf("get user list error: %s", err)
	}
	if len(c.userList) == 0 {
		return errors.New("add users error: not have any user")
	}
	c.Tag = c.buildNodeTag()

	// add limiter
	l := limiter.AddLimiter(c.Tag, &c.LimitConfig, c.userList)
	// add rule limiter
	if !c.DisableGetRule {
		if err = l.UpdateRule(c.nodeInfo.Rules); err != nil {
			return fmt.Errorf("update rule error: %s", err)
		}
	}
	if c.nodeInfo.Tls {
		err = c.requestCert()
		if err != nil {
			return fmt.Errorf("request cert error: %s", err)
		}
	}
	// Add new tag
	err = c.server.AddNode(c.Tag, c.nodeInfo, c.ControllerConfig)
	if err != nil {
		return fmt.Errorf("add new node error: %s", err)
	}
	added, err := c.server.AddUsers(&vCore.AddUsersParams{
		Tag:      c.Tag,
		Config:   c.ControllerConfig,
		UserInfo: c.userList,
		NodeInfo: c.nodeInfo,
	})
	if err != nil {
		return fmt.Errorf("add users error: %s", err)
	}
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.Type, c.nodeInfo.Id, added)
	c.initTask()
	return nil
}

// Close implement the Close() function of the service interface
func (c *Controller) Close() error {
	limiter.DeleteLimiter(c.Tag)
	if c.nodeInfoMonitorPeriodic != nil {
		c.nodeInfoMonitorPeriodic.Close()
	}
	if c.userReportPeriodic != nil {
		c.userReportPeriodic.Close()
	}
	if c.renewCertPeriodic != nil {
		c.renewCertPeriodic.Close()
	}
	if c.dynamicSpeedLimitPeriodic != nil {
		c.dynamicSpeedLimitPeriodic.Close()
	}
	if c.onlineIpReportPeriodic != nil {
		c.onlineIpReportPeriodic.Close()
	}
	return nil
}

func (c *Controller) buildNodeTag() string {
	return fmt.Sprintf("%s-%s-%d", c.apiClient.APIHost, c.nodeInfo.Type, c.nodeInfo.Id)
}
