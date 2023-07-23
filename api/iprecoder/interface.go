package iprecoder

import (
	"github.com/Github-Aiko/Aiko-Server/src/limiter"
)

type IpRecorder interface {
	SyncOnlineIp(Ips []limiter.UserIpList) ([]limiter.UserIpList, error)
}
