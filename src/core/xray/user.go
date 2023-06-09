package xray

import (
	"context"
	"fmt"

	"github.com/Github-Aiko/Aiko-Server/src/common/builder"
	vCore "github.com/Github-Aiko/Aiko-Server/src/core"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/proxy"
)

func (c *Core) GetUserManager(tag string) (proxy.UserManager, error) {
	handler, err := c.ihm.GetHandler(context.Background(), tag)
	if err != nil {
		return nil, fmt.Errorf("no such inbound tag: %s", err)
	}
	inboundInstance, ok := handler.(proxy.GetInbound)
	if !ok {
		return nil, fmt.Errorf("handler %s is not implement proxy.GetInbound", tag)
	}
	userManager, ok := inboundInstance.GetInbound().(proxy.UserManager)
	if !ok {
		return nil, fmt.Errorf("handler %s is not implement proxy.UserManager", tag)
	}
	return userManager, nil
}

func (c *Core) DelUsers(users []string, tag string) error {
	userManager, err := c.GetUserManager(tag)
	if err != nil {
		return fmt.Errorf("get user manager error: %s", err)
	}
	for _, email := range users {
		err = userManager.RemoveUser(context.Background(), email)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Core) GetUserTraffic(tag, uuid string, reset bool) (up int64, down int64) {
	upName := "user>>>" + builder.BuildUserTag(tag, uuid) + ">>>traffic>>>uplink"
	downName := "user>>>" + builder.BuildUserTag(tag, uuid) + ">>>traffic>>>downlink"
	upCounter := c.shm.GetCounter(upName)
	downCounter := c.shm.GetCounter(downName)
	if reset {
		if upCounter != nil {
			up = upCounter.Set(0)
		}
		if downCounter != nil {
			down = downCounter.Set(0)
		}
	} else {
		if upCounter != nil {
			up = upCounter.Value()
		}
		if downCounter != nil {
			down = downCounter.Value()
		}
	}
	return up, down
}

func (c *Core) AddUsers(p *vCore.AddUsersParams) (added int, err error) {
	users := make([]*protocol.User, 0, len(p.UserInfo))
	switch p.NodeInfo.Type {
	case "v2ray":
		if p.Config.XrayOptions.EnableXtls {
			users = builder.BuildVlessUsers(p.Tag, p.UserInfo, true)
		} else {
			users = builder.BuildVmessUsers(p.Tag, p.UserInfo)
		}
	case "trojan":
		users = builder.BuildTrojanUsers(p.Tag, p.UserInfo)
	case "shadowsocks":
		users = builder.BuildSSUsers(p.Tag,
			p.UserInfo,
			p.NodeInfo.Cipher,
			p.NodeInfo.ServerKey)
	default:
		return 0, fmt.Errorf("unsupported node type: %s", p.NodeInfo.Type)
	}
	man, err := c.GetUserManager(p.Tag)
	if err != nil {
		return 0, fmt.Errorf("get user manager error: %s", err)
	}
	for _, u := range users {
		mUser, err := u.ToMemoryUser()
		if err != nil {
			return 0, err
		}
		err = man.AddUser(context.Background(), mUser)
		if err != nil {
			return 0, err
		}
	}
	return len(users), nil
}
