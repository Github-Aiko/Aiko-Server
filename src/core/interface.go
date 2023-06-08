package core

import (
	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/src/conf"
)

type AddUsersParams struct {
	Tag      string
	Config   *conf.ControllerConfig
	UserInfo []panel.UserInfo
	NodeInfo *panel.NodeInfo
}
type Core interface {
	Start() error
	Close() error
	AddNode(tag string, info *panel.NodeInfo, config *conf.ControllerConfig) error
	DelNode(tag string) error
	AddUsers(p *AddUsersParams) (added int, err error)
	GetUserTraffic(tag, uuid string, reset bool) (up int64, down int64)
	DelUsers(users []string, tag string) error
}
