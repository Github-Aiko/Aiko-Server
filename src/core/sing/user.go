package sing

import (
	"encoding/base64"
	"errors"

	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/src/common/counter"
	"github.com/Github-Aiko/Aiko-Server/src/core"
	"github.com/inazumav/sing-box/inbound"
	"github.com/inazumav/sing-box/option"
)

func (b *Box) AddUsers(p *core.AddUsersParams) (added int, err error) {
	switch p.NodeInfo.Type {
	case "v2ray":
		us := make([]option.VMessUser, len(p.UserInfo))
		for i := range p.UserInfo {
			us[i] = option.VMessUser{
				Name: p.UserInfo[i].Uuid,
				UUID: p.UserInfo[i].Uuid,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.VMess).AddUsers(us)

	case "vless":
		us := make([]option.VLESSUser, len(p.UserInfo))
		for i := range p.UserInfo {
			us[i] = option.VLESSUser{
				Name: p.UserInfo[i].Uuid,
				UUID: p.UserInfo[i].Uuid,
				Flow: p.NodeInfo.ExtraConfig.VlessFlow,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.VLESS).AddUsers(us)
	case "shadowsocks":
		us := make([]option.ShadowsocksUser, len(p.UserInfo))
		for i := range p.UserInfo {
			var password = p.UserInfo[i].Uuid
			switch p.NodeInfo.Cipher {
			case "2022-blake3-aes-128-gcm":
				password = base64.StdEncoding.EncodeToString([]byte(password[:16]))
			case "2022-blake3-aes-256-gcm":
				password = base64.StdEncoding.EncodeToString([]byte(password[:32]))
			}
			us[i] = option.ShadowsocksUser{
				Name:     p.UserInfo[i].Uuid,
				Password: password,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.ShadowsocksMulti).AddUsers(us)
	case "trojan":
		us := make([]option.TrojanUser, len(p.UserInfo))
		for i := range p.UserInfo {
			us[i] = option.TrojanUser{
				Name:     p.UserInfo[i].Uuid,
				Password: p.UserInfo[i].Uuid,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.Trojan).AddUsers(us)
	case "hysteria":
		us := make([]option.HysteriaUser, len(p.UserInfo))
		for i := range p.UserInfo {
			us[i] = option.HysteriaUser{
				Name:       p.UserInfo[i].Uuid,
				AuthString: p.UserInfo[i].Uuid,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.Hysteria).AddUsers(us)
	case "tuic":
		us := make([]option.TUICUser, len(p.UserInfo))
		for i := range p.UserInfo {
			us[i] = option.TUICUser{
				Name:     p.UserInfo[i].Uuid,
				UUID:     p.UserInfo[i].Uuid,
				Password: "tuic",
			}
		}
		err = b.inbounds[p.Tag].(*inbound.TUIC).AddUsers(us)
	}
	if err != nil {
		return 0, err
	}
	return len(p.UserInfo), err
}

func (b *Box) GetUserTraffic(tag, uuid string, reset bool) (up int64, down int64) {
	if v, ok := b.hookServer.counter.Load(tag); ok {
		c := v.(*counter.TrafficCounter)
		up = c.GetUpCount(uuid)
		down = c.GetDownCount(uuid)
		if reset {
			c.Reset(uuid)
		}
		return
	}
	return 0, 0
}

type UserDeleter interface {
	DelUsers(uuid []string) error
}

func (b *Box) DelUsers(users []panel.UserInfo, tag string) error {
	var del UserDeleter
	if i, ok := b.inbounds[tag]; ok {
		switch i.Type() {
		case "vmess":
			del = i.(*inbound.VMess)
		case "vless":
			del = i.(*inbound.VLESS)
		case "shadowsocks":
			del = i.(*inbound.ShadowsocksMulti)
		case "tuic":
			del = i.(*inbound.TUIC)
		case "trojan":
			del = i.(*inbound.Trojan)
		case "hysteria":
			del = i.(*inbound.Hysteria)
		}
	} else {
		return errors.New("the inbound not found")
	}
	uuids := make([]string, len(users))
	for i := range users {
		uuids[i] = users[i].Uuid
	}
	err := del.DelUsers(uuids)
	if err != nil {
		return err
	}
	return nil
}
