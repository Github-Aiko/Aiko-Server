package node

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"

	vCore "github.com/Github-Aiko/Aiko-Server/src/core"

	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/src/common/builder"
	"github.com/Github-Aiko/Aiko-Server/src/limiter"
	"github.com/Github-Aiko/Aiko-Server/src/node/lego"
	"github.com/xtls/xray-core/common/task"
)

func (c *Controller) initTask() {
	// fetch node info task
	c.nodeInfoMonitorPeriodic = &task.Periodic{
		Interval: c.nodeInfo.BaseConfig.PullInterval.(time.Duration),
		Execute:  c.nodeInfoMonitor,
	}
	// fetch user list task
	c.userReportPeriodic = &task.Periodic{
		Interval: c.nodeInfo.BaseConfig.PushInterval.(time.Duration),
		Execute:  c.reportUserTraffic,
	}
	log.Printf("[%s: %d] Start monitor node status", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	// delay to start nodeInfoMonitor
	go func() {
		time.Sleep(c.nodeInfo.BaseConfig.PullInterval.(time.Duration))
		_ = c.nodeInfoMonitorPeriodic.Start()
	}()
	log.Printf("[%s: %d] Start report node status", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	// delay to start userReport
	go func() {
		time.Sleep(c.nodeInfo.BaseConfig.PushInterval.(time.Duration))
		_ = c.userReportPeriodic.Start()
	}()
	if c.nodeInfo.Tls != 0 && c.CertConfig.CertMode != "none" &&
		(c.CertConfig.CertMode == "dns" || c.CertConfig.CertMode == "http") {
		c.renewCertPeriodic = &task.Periodic{
			Interval: time.Hour * 24,
			Execute:  c.reportUserTraffic,
		}
		log.Printf("[%s: %d] Start renew cert", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
		// delay to start renewCert
		go func() {
			_ = c.renewCertPeriodic.Start()
		}()
	}
}

func (c *Controller) nodeInfoMonitor() (err error) {
	// First fetch Node Info
	newNodeInfo, err := c.apiClient.GetNodeInfo()
	if err != nil {
		log.Print(err)
		return nil
	}
	var nodeInfoChanged = false
	// If nodeInfo changed
	if newNodeInfo != nil {
		// Remove old tag
		oldTag := c.Tag
		err := c.server.DelNode(oldTag)
		if err != nil {
			log.Print(err)
			return nil
		}
		// Remove Old limiter
		limiter.DeleteLimiter(oldTag)
		// Add new tag
		c.nodeInfo = newNodeInfo
		c.Tag = c.buildNodeTag()
		err = c.server.AddNode(c.Tag, newNodeInfo, c.ControllerConfig)
		if err != nil {
			log.Print(err)
			return nil
		}
		nodeInfoChanged = true
	}
	// Update User
	newUserInfo, err := c.apiClient.GetUserList()
	if err != nil {
		log.Print(err)
		return nil
	}
	if nodeInfoChanged {
		c.userList = newUserInfo
		// Add new Limiter
		l := limiter.AddLimiter(c.Tag, &c.LimitConfig, newUserInfo)
		_, err = c.server.AddUsers(&vCore.AddUsersParams{
			Tag:    c.Tag,
			Config: c.ControllerConfig,
		})
		if err != nil {
			log.Print(err)
			return nil
		}
		err = l.UpdateRule(newNodeInfo.Rules)
		if err != nil {
			log.Printf("Update Rule error: %s", err)
		}
		// Check interval
		if c.nodeInfoMonitorPeriodic.Interval != newNodeInfo.BaseConfig.PullInterval.(time.Duration) &&
			newNodeInfo.BaseConfig.PullInterval.(time.Duration) != 0 {
			c.nodeInfoMonitorPeriodic.Interval = newNodeInfo.BaseConfig.PullInterval.(time.Duration)
			_ = c.nodeInfoMonitorPeriodic.Close()
			go func() {
				time.Sleep(c.nodeInfoMonitorPeriodic.Interval)
				_ = c.nodeInfoMonitorPeriodic.Start()
			}()
		}
		if c.userReportPeriodic.Interval != newNodeInfo.BaseConfig.PushInterval.(time.Duration) &&
			newNodeInfo.BaseConfig.PushInterval.(time.Duration) != 0 {
			c.userReportPeriodic.Interval = newNodeInfo.BaseConfig.PullInterval.(time.Duration)
			_ = c.userReportPeriodic.Close()
			go func() {
				time.Sleep(c.userReportPeriodic.Interval)
				_ = c.userReportPeriodic.Start()
			}()
		}
	} else {
		deleted, added := compareUserList(c.userList, newUserInfo)
		if len(deleted) > 0 {
			deletedEmail := make([]string, len(deleted))
			for i := range deleted {
				deletedEmail[i] = fmt.Sprintf("%s|%s|%d",
					c.Tag,
					(deleted)[i].Uuid,
					(deleted)[i].Id)
			}
			err := c.server.DelUsers(deletedEmail, c.Tag)
			if err != nil {
				log.Print(err)
			}
		}
		if len(added) > 0 {
			_, err := c.server.AddUsers(&vCore.AddUsersParams{
				Tag:      c.Tag,
				Config:   c.ControllerConfig,
				UserInfo: added,
				NodeInfo: c.nodeInfo,
			})
			if err != nil {
				log.Print(err)
			}
		}
		if len(added) > 0 || len(deleted) > 0 {
			// Update Limiter
			err = limiter.UpdateLimiter(c.Tag, added, deleted)
			if err != nil {
				log.Print("update limiter:", err)
			}
		}
		log.Printf("[%s: %d] %d user deleted, %d user added", c.nodeInfo.NodeType, c.nodeInfo.NodeId,
			len(deleted), len(added))
		c.userList = newUserInfo
	}
	return nil
}

func compareUserList(old, new []panel.UserInfo) (deleted, added []panel.UserInfo) {
	tmp := map[string]struct{}{}
	tmp2 := map[string]struct{}{}
	for i := range old {
		tmp[old[i].Uuid+strconv.Itoa(old[i].SpeedLimit)] = struct{}{}
	}
	l := len(tmp)
	for i := range new {
		e := new[i].Uuid + strconv.Itoa(new[i].SpeedLimit)
		tmp[e] = struct{}{}
		tmp2[e] = struct{}{}
		if l != len(tmp) {
			added = append(added, new[i])
			l++
		}
	}
	tmp = nil
	l = len(tmp2)
	for i := range old {
		tmp2[old[i].Uuid+strconv.Itoa(old[i].SpeedLimit)] = struct{}{}
		if l != len(tmp2) {
			deleted = append(deleted, old[i])
			l++
		}
	}
	return deleted, added
}

func (c *Controller) reportUserTraffic() (err error) {
	// Get User traffic
	userTraffic := make([]panel.UserTraffic, 0)
	for i := range c.userList {
		up, down := c.server.GetUserTraffic(builder.BuildUserTag(c.Tag, &c.userList[i]), true)
		if up > 0 || down > 0 {
			if c.LimitConfig.EnableDynamicSpeedLimit {
				c.userList[i].Traffic += up + down
			}
			userTraffic = append(userTraffic, panel.UserTraffic{
				UID:      (c.userList)[i].Id,
				Upload:   up,
				Download: down})
		}
	}
	if len(userTraffic) > 0 && !c.DisableUploadTraffic {
		err = c.apiClient.ReportUserTraffic(userTraffic)
		if err != nil {
			log.Printf("Report user traffic faild: %s", err)
		} else {
			log.Printf("[%s: %d] Report %d online users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(userTraffic))
		}
	}
	userTraffic = nil
	runtime.GC()
	return nil
}

func (c *Controller) RenewCert() {
	l, err := lego.New(c.CertConfig)
	if err != nil {
		log.Print(err)
		return
	}
	err = l.RenewCert()
	if err != nil {
		log.Print(err)
		return
	}
}
