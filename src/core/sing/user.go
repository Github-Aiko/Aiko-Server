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
	case "vmess", "vless":
		if p.NodeInfo.Type == "vless" {
			us := make([]option.VLESSUser, len(p.Users))
			for i := range p.Users {
				us[i] = option.VLESSUser{
					Name: p.Users[i].Uuid,
					Flow: p.VAllss.Flow,
					UUID: p.Users[i].Uuid,
				}
			}
			err = b.inbounds[p.Tag].(*inbound.VLESS).AddUsers(us)
		} else {
			us := make([]option.VMessUser, len(p.Users))
			for i := range p.Users {
				us[i] = option.VMessUser{
					Name: p.Users[i].Uuid,
					UUID: p.Users[i].Uuid,
				}
			}
			err = b.inbounds[p.Tag].(*inbound.VMess).AddUsers(us)
		}
	case "shadowsocks":
		us := make([]option.ShadowsocksUser, len(p.Users))
		for i := range p.Users {
			var password = p.Users[i].Uuid
			switch p.Shadowsocks.Cipher {
			case "2022-blake3-aes-128-gcm":
				password = base64.StdEncoding.EncodeToString([]byte(password[:16]))
			case "2022-blake3-aes-256-gcm":
				password = base64.StdEncoding.EncodeToString([]byte(password[:32]))
			}
			us[i] = option.ShadowsocksUser{
				Name:     p.Users[i].Uuid,
				Password: password,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.ShadowsocksMulti).AddUsers(us)
	case "trojan":
		us := make([]option.TrojanUser, len(p.Users))
		for i := range p.Users {
			us[i] = option.TrojanUser{
				Name:     p.Users[i].Uuid,
				Password: p.Users[i].Uuid,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.Trojan).AddUsers(us)
	case "hysteria":
		us := make([]option.HysteriaUser, len(p.Users))
		for i := range p.Users {
			us[i] = option.HysteriaUser{
				Name:       p.Users[i].Uuid,
				AuthString: p.Users[i].Uuid,
			}
		}
		err = b.inbounds[p.Tag].(*inbound.Hysteria).AddUsers(us)
	}
	if err != nil {
		return 0, err
	}
	return len(p.Users), err
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
