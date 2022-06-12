package session

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type SessionInfo struct {
	SID         uuid.UUID `json:"id"`
	SInitTime   time.Time `json:"init_time"`
	SRemoteAddr net.Addr  `json:"remote_addr"`
}

func Test() *SessionInfo {
	a, err := net.ResolveTCPAddr("tcp", "95.31.20.108:9")
	if err != nil {
		panic(err)
	}
	return &SessionInfo{
		uuid.New(),
		time.Now(),
		a,
	}
}

func (si *SessionInfo) RemoteAddr() net.Addr {
	return si.SRemoteAddr
}

func (si *SessionInfo) ID() uuid.UUID {
	return si.SID
}

func (si *SessionInfo) InitTime() time.Time {
	return si.SInitTime
}
