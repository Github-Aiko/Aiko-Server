package limiter

import (
	"errors"
	"regexp"
	"sync"
	"time"

	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/common/format"
	"github.com/Github-Aiko/Aiko-Server/conf"
	"github.com/juju/ratelimit"
	log "github.com/sirupsen/logrus"
	"github.com/xtls/xray-core/common/task"
)

var limitLock sync.RWMutex
var limiter map[string]*Limiter

func Init() {
	limiter = map[string]*Limiter{}
	c := task.Periodic{
		Interval: time.Minute * 2,
		Execute:  ClearOnlineIP,
	}
	go func() {
		log.WithField("Type", "Limiter").
			Debug("ClearOnlineIP started")
		time.Sleep(time.Minute * 2)
		_ = c.Start()
	}()
}

type Limiter struct {
	DomainRules   []*regexp.Regexp
	ProtocolRules []string
	SpeedLimit    int
	UserOnlineIP  *sync.Map      // Key: Name, value: {Key: Ip, value: Uid}
	UUIDtoUID     map[string]int // Key: UUID, value: UID
	UserLimitInfo *sync.Map      // Key: Uid value: UserLimitInfo
	ConnLimiter   *ConnLimiter   // Key: Uid value: ConnLimiter
	SpeedLimiter  *sync.Map      // key: Uid, value: *ratelimit.Bucket
}

type UserLimitInfo struct {
	UID               int
	SpeedLimit        int
	DynamicSpeedLimit int
	ExpireTime        int64
}

func AddLimiter(tag string, l *conf.LimitConfig, users []panel.UserInfo) *Limiter {
	info := &Limiter{
		SpeedLimit:    l.SpeedLimit,
		UserOnlineIP:  new(sync.Map),
		UserLimitInfo: new(sync.Map),
		ConnLimiter:   NewConnLimiter(l.ConnLimit, l.IPLimit, l.EnableRealtime),
		SpeedLimiter:  new(sync.Map),
	}
	uuidmap := make(map[string]int)
	for i := range users {
		uuidmap[users[i].Uuid] = users[i].Id
		if users[i].SpeedLimit != 0 {
			userLimit := &UserLimitInfo{
				UID:        users[i].Id,
				SpeedLimit: users[i].SpeedLimit,
			}
			info.UserLimitInfo.Store(format.UserTag(tag, users[i].Uuid), userLimit)
		}
	}
	info.UUIDtoUID = uuidmap
	limitLock.Lock()
	limiter[tag] = info
	limitLock.Unlock()
	return info
}

func GetLimiter(tag string) (info *Limiter, err error) {
	limitLock.RLock()
	info, ok := limiter[tag]
	limitLock.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	return info, nil
}

func DeleteLimiter(tag string) {
	limitLock.Lock()
	delete(limiter, tag)
	limitLock.Unlock()
}

func (l *Limiter) UpdateUser(tag string, added []panel.UserInfo, deleted []panel.UserInfo) {
	for i := range deleted {
		l.UserLimitInfo.Delete(format.UserTag(tag, deleted[i].Uuid))
		delete(l.UUIDtoUID, deleted[i].Uuid)
	}
	for i := range added {
		if added[i].SpeedLimit != 0 {
			userLimit := &UserLimitInfo{
				UID:        added[i].Id,
				SpeedLimit: added[i].SpeedLimit,
				ExpireTime: 0,
			}
			l.UserLimitInfo.Store(format.UserTag(tag, added[i].Uuid), userLimit)
		}
		l.UUIDtoUID[added[i].Uuid] = added[i].Id
	}
}

func (l *Limiter) UpdateDynamicSpeedLimit(tag, uuid string, limit int, expire time.Time) error {
	if v, ok := l.UserLimitInfo.Load(format.UserTag(tag, uuid)); ok {
		info := v.(*UserLimitInfo)
		info.DynamicSpeedLimit = limit
		info.ExpireTime = expire.Unix()
	} else {
		return errors.New("not found")
	}
	return nil
}

func (l *Limiter) CheckLimit(email string, ip string, isTcp bool) (Bucket *ratelimit.Bucket, Reject bool) {
	// ip and conn limiter
	if l.ConnLimiter.AddConnCount(email, ip, isTcp) {
		return nil, true
	}
	// check and gen speed limit Bucket
	nodeLimit := l.SpeedLimit
	userLimit := 0
	if v, ok := l.UserLimitInfo.Load(email); ok {
		u := v.(*UserLimitInfo)
		if u.ExpireTime < time.Now().Unix() && u.ExpireTime != 0 {
			if u.SpeedLimit != 0 {
				userLimit = u.SpeedLimit
				u.DynamicSpeedLimit = 0
				u.ExpireTime = 0
			} else {
				l.UserLimitInfo.Delete(email)
			}
		} else {
			userLimit = determineSpeedLimit(u.SpeedLimit, u.DynamicSpeedLimit)
		}
	}

	// Store online user for device limit
	ipMap := new(sync.Map)
	uid := l.UUIDtoUID[email]
	ipMap.Store(ip, uid)
	// If any device is online
	if v, ok := l.UserOnlineIP.LoadOrStore(email, ipMap); ok {
		ipMap := v.(*sync.Map)
		// If this is a new ip
		if _, ok := ipMap.LoadOrStore(ip, uid); !ok {
			counter := 0
			ipMap.Range(func(key, value interface{}) bool {
				counter++
				return true
			})
		}
	}

	limit := int64(determineSpeedLimit(nodeLimit, userLimit)) * 1000000 / 8 // If you need the Speed limit
	if limit > 0 {
		Bucket = ratelimit.NewBucketWithQuantum(time.Second, limit, limit) // Byte/s
		if v, ok := l.SpeedLimiter.LoadOrStore(email, Bucket); ok {
			return v.(*ratelimit.Bucket), false
		} else {
			l.SpeedLimiter.Store(email, Bucket)
			return Bucket, false
		}
	} else {
		return nil, false
	}
}

func (l *Limiter) GetOnlineDevice() (*[]panel.OnlineUser, error) {
	var onlineUser []panel.OnlineUser

	l.UserOnlineIP.Range(func(key, value interface{}) bool {
		email := key.(string)
		ipMap := value.(*sync.Map)
		ipMap.Range(func(key, value interface{}) bool {
			uid := value.(int)
			ip := key.(string)
			onlineUser = append(onlineUser, panel.OnlineUser{UID: uid, IP: ip})
			return true
		})
		l.UserOnlineIP.Delete(email) // Reset online device
		return true
	})

	return &onlineUser, nil
}

type UserIpList struct {
	Uid    int      `json:"Uid"`
	IpList []string `json:"Ips"`
}

func determineDeviceLimit(nodeLimit, userLimit int) (limit int) {
	if nodeLimit == 0 || userLimit == 0 {
		if nodeLimit > userLimit {
			return nodeLimit
		} else if nodeLimit < userLimit {
			return userLimit
		} else {
			return 0
		}
	} else {
		if nodeLimit > userLimit {
			return userLimit
		} else if nodeLimit < userLimit {
			return nodeLimit
		} else {
			return nodeLimit
		}
	}
}
