package xray

import (
	"github.com/AikoPanel/Xray-core/common/protocol"
	"github.com/AikoPanel/Xray-core/common/serial"
	"github.com/AikoPanel/Xray-core/proxy/trojan"
	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/src/common/format"
)

func buildTrojanUsers(tag string, userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = buildTrojanUser(tag, &(userInfo)[i])
	}
	return users
}

func buildTrojanUser(tag string, userInfo *panel.UserInfo) (user *protocol.User) {
	trojanAccount := &trojan.Account{
		Password: userInfo.Uuid,
	}
	return &protocol.User{
		Level:   0,
		Email:   format.UserTag(tag, userInfo.Uuid),
		Account: serial.ToTypedMessage(trojanAccount),
	}
}
