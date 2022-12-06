package service

import (
	"fmt"

	"github.com/Github-Aiko/Aiko-Server/internal/pkg/api"
	cProtocol "github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/proxy/trojan"
)

func buildUser(tag string, userInfo []api.UserInfo) (users []*cProtocol.User) {
	users = make([]*cProtocol.User, len(userInfo))
	for i, user := range userInfo {
		trojanAccount := &trojan.Account{
			Password: user.UUID,
			Flow:     "xtls-rprx-direct",
		}
		email := buildUserEmail(tag, user.ID, user.UUID)
		users[i] = &cProtocol.User{
			Level:   0,
			Email:   email,
			Account: serial.ToTypedMessage(trojanAccount),
		}
	}
	return users
}

func buildUserEmail(tag string, id int, uuid string) string {
	return fmt.Sprintf("%s|%d|%s", tag, id, uuid)
}
