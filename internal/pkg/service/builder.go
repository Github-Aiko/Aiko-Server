package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Github-Aiko/Aiko-Server/internal/pkg/api"
	log "github.com/sirupsen/logrus"
	cProtocol "github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/task"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/inbound"
	"github.com/xtls/xray-core/features/stats"
	"github.com/xtls/xray-core/proxy"
)

type Config struct {
	SysInterval time.Duration
	Cert        *CertConfig
}

type Builder struct {
	instance                *core.Instance
	config                  *Config
	nodeInfo                *api.NodeInfo
	inboundTag              string
	userList                *[]api.UserInfo
	getUserList             func() (*[]api.UserInfo, error)
	reportUserTraffic       func([]*api.UserTraffic) error
	nodeInfoMonitorPeriodic *task.Periodic
	userReportPeriodic      *task.Periodic
}

// New return a builder service with default parameters.
func New(inboundTag string, instance *core.Instance, config *Config, nodeInfo *api.NodeInfo,
	getUserList func() (*[]api.UserInfo, error), reportUserTraffic func([]*api.UserTraffic) error,
) *Builder {
	builder := &Builder{
		inboundTag:        inboundTag,
		instance:          instance,
		config:            config,
		nodeInfo:          nodeInfo,
		getUserList:       getUserList,
		reportUserTraffic: reportUserTraffic,
	}
	return builder
}

// addUsers
func (b *Builder) addUsers(users []*cProtocol.User, tag string) error {
	inboundManager := b.instance.GetFeature(inbound.ManagerType()).(inbound.Manager)
	handler, err := inboundManager.GetHandler(context.Background(), tag)
	if err != nil {
		return fmt.Errorf("no such inbound tag: %s", err)
	}
	inboundInstance, ok := handler.(proxy.GetInbound)
	if !ok {
		return fmt.Errorf("handler %s is not implement proxy.GetInbound", tag)
	}

	userManager, ok := inboundInstance.GetInbound().(proxy.UserManager)
	if !ok {
		return fmt.Errorf("handler %s is not implement proxy.UserManager", err)
	}
	for _, item := range users {
		mUser, err := item.ToMemoryUser()
		if err != nil {
			return err
		}
		err = userManager.AddUser(context.Background(), mUser)
		if err != nil {
			return err
		}
	}
	return nil
}

// addNewUser
func (b *Builder) addNewUser(userInfo []api.UserInfo) (err error) {
	users := make([]*cProtocol.User, 0)
	users = buildUser(b.inboundTag, userInfo)

	err = b.addUsers(users, b.inboundTag)
	if err != nil {
		return err
	}
	log.Infof("Added %d new users", len(userInfo))
	return nil
}

// Start implement the Start() function of the service interface
func (b *Builder) Start() error {
	// Update user
	userList, err := b.getUserList()
	if err != nil {
		return err
	}
	err = b.addNewUser(*userList)
	if err != nil {
		return err
	}

	b.userList = userList

	b.nodeInfoMonitorPeriodic = &task.Periodic{
		Interval: b.config.SysInterval,
		Execute:  b.nodeInfoMonitor,
	}
	b.userReportPeriodic = &task.Periodic{
		Interval: b.config.SysInterval,
		Execute:  b.userInfoMonitor,
	}
	log.Infoln("Start monitor node status")
	err = b.nodeInfoMonitorPeriodic.Start()
	if err != nil {
		return fmt.Errorf("node info periodic, start erorr:%s", err)
	}
	log.Infoln("Start report node status")
	err = b.userReportPeriodic.Start()
	if err != nil {
		return fmt.Errorf("user report periodic, start erorr:%s", err)
	}
	return nil
}

// Close implement the Close() function of the service interface
func (b *Builder) Close() error {
	if b.nodeInfoMonitorPeriodic != nil {
		err := b.nodeInfoMonitorPeriodic.Close()
		if err != nil {
			return fmt.Errorf("node info periodic close failed: %s", err)
		}
	}

	if b.nodeInfoMonitorPeriodic != nil {
		err := b.userReportPeriodic.Close()
		if err != nil {
			return fmt.Errorf("user report periodic close failed: %s", err)
		}
	}
	return nil
}

// getTraffic
func (b *Builder) getTraffic(email string) (up int64, down int64, count int64) {
	upName := "user>>>" + email + ">>>traffic>>>uplink"
	downName := "user>>>" + email + ">>>traffic>>>downlink"
	countName := "user>>>" + email + ">>>request>>>count"
	statsManager := b.instance.GetFeature(stats.ManagerType()).(stats.Manager)
	upCounter := statsManager.GetCounter(upName)
	downCounter := statsManager.GetCounter(downName)
	countCounter := statsManager.GetCounter(countName)
	if upCounter != nil {
		up = upCounter.Value()
		upCounter.Set(0)
	}
	if downCounter != nil {
		down = downCounter.Value()
		downCounter.Set(0)
	}
	if countCounter != nil {
		count = countCounter.Value()
		countCounter.Set(0)
	}

	return up, down, count

}

// removeUsers
func (b *Builder) removeUsers(users []string, tag string) error {
	inboundManager := b.instance.GetFeature(inbound.ManagerType()).(inbound.Manager)
	handler, err := inboundManager.GetHandler(context.Background(), tag)
	if err != nil {
		return fmt.Errorf("no such inbound tag: %s", err)
	}
	inboundInstance, ok := handler.(proxy.GetInbound)
	if !ok {
		return fmt.Errorf("handler %s is not implement proxy.GetInbound", tag)
	}

	userManager, ok := inboundInstance.GetInbound().(proxy.UserManager)
	if !ok {
		return fmt.Errorf("handler %s is not implement proxy.UserManager", err)
	}
	for _, email := range users {
		err = userManager.RemoveUser(context.Background(), email)
		if err != nil {
			return err
		}
	}
	return nil
}

// nodeInfoMonitor
func (b *Builder) nodeInfoMonitor() (err error) {

	// Update User
	newUserList, err := b.getUserList()
	if err != nil {
		log.Errorln(err)
		return nil
	}

	deleted, added := compareUserList(b.userList, newUserList)
	if len(deleted) > 0 {
		deletedEmail := make([]string, len(deleted))
		for i, u := range deleted {
			deletedEmail[i] = buildUserEmail(b.inboundTag, u.ID, u.UUID)
		}
		err := b.removeUsers(deletedEmail, b.inboundTag)
		if err != nil {
			log.Errorln(err)
			return nil
		}
	}
	if len(added) > 0 {
		err = b.addNewUser(added)
		if err != nil {
			log.Errorln(err)
			return nil
		}

	}
	log.Infof("%d user deleted, %d user added", len(deleted), len(added))

	b.userList = newUserList
	return nil
}

// userInfoMonitor
func (b *Builder) userInfoMonitor() (err error) {
	// Get User traffic
	userTraffic := make([]*api.UserTraffic, 0)
	for _, user := range *b.userList {
		email := buildUserEmail(b.inboundTag, user.ID, user.UUID)
		up, down, count := b.getTraffic(email)
		if up > 0 || down > 0 || count > 0 {
			userTraffic = append(userTraffic, &api.UserTraffic{
				UID:      user.ID,
				Upload:   up,
				Download: down,
				Count:    count,
			})
		}
	}
	log.Infof("%d user traffic needs to be reported", len(userTraffic))
	if len(userTraffic) > 0 {
		err = b.reportUserTraffic(userTraffic)
		if err != nil {
			log.Errorln(err)
			return nil
		}
	}

	return nil
}

func compareUserList(old, new *[]api.UserInfo) (deleted, added []api.UserInfo) {
	msrc := make(map[api.UserInfo]byte) //按源数组建索引
	mall := make(map[api.UserInfo]byte) //源+目所有元素建索引

	var set []api.UserInfo //交集

	//1.源数组建立map
	for _, v := range *old {
		msrc[v] = 0
		mall[v] = 0
	}
	//2.目数组中，存不进去，即重复元素，所有存不进去的集合就是并集
	for _, v := range *new {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) { //长度变化，即可以存
			l = len(mall)
		} else { //存不了，进并集
			set = append(set, v)
		}
	}
	//3.遍历交集，在并集中找，找到就从并集中删，删完后就是补集（即并-交=所有变化的元素）
	for _, v := range set {
		delete(mall, v)
	}
	//4.此时，mall是补集，所有元素去源中找，找到就是删除的，找不到的必定能在目数组中找到，即新加的
	for v := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		} else {
			added = append(added, v)
		}
	}

	return deleted, added
}
