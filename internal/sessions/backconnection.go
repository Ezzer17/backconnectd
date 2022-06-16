package sessions

import (
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/ezzer17/backconnectd/internal/sessioninfo"
	"github.com/ezzer17/backconnectd/pkg/asyncrw"
	"github.com/ezzer17/backconnectd/pkg/channelconn"
)

type BackconnectSession struct {
	sessioninfo.SessionInfo
	connection *channelconn.Connection
}

func NewBackconnection(conn net.Conn) *BackconnectSession {
	return &BackconnectSession{
		SessionInfo: sessioninfo.SessionInfo{
			SID:         uuid.New(),
			SInitTime:   time.Now(),
			SRemoteAddr: conn.RemoteAddr().String(),
		},
		connection: channelconn.New(conn),
	}
}

func (session *BackconnectSession) Run() {
	session.connection.Run()
}

func (session *BackconnectSession) Kill() {
	session.connection.Stop()
}

func (session *BackconnectSession) Connect(a asyncrw.AsyncSenderReciever) error {
	return session.connection.Connect(a)
}
