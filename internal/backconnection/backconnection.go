package backconnection

import (
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/ezzer17/backconnectd/internal/sessioninfo"
	"github.com/ezzer17/backconnectd/pkg/channelconn"
)

const checkInterval = 500 * time.Millisecond
const channelSize = 1024

type sessionCtrlMsg int

type sessionUpdateNotification int

const (
	sessionListUpdated sessionUpdateNotification = iota
)

type Session interface {
	ID() uuid.UUID
	RemoteAddr() string
}

type BackconnectSession struct {
	sessioninfo.SessionInfo
	*channelconn.Connection
}

// func BPrintf(msg string, args ...interface{}) []byte {
// 	return []byte(fmt.Sprintf(msg, args...))
// }

// func NewAdmin(conn net.Conn, closeCb func(Session)) AdminSession {
// 	sess := new(conn, closeCb)
// 	adminsess := AdminSession{session: sess}
// 	return adminsess
// }

func NewBackconnect(conn net.Conn) *BackconnectSession {
	return &BackconnectSession{
		SessionInfo: sessioninfo.SessionInfo{
			SID:         uuid.New(),
			SInitTime:   time.Now(),
			SRemoteAddr: conn.RemoteAddr().String(),
		},
		Connection: channelconn.New(conn),
	}
}

func (session *BackconnectSession) Run() {
	session.Connection.Run()
}

func (session *BackconnectSession) Kill() {
	session.Connection.Stop()
}

// func (session *session) NotifyOfSessionListUpdate() {
// 	select {
// 	case session.updateNotifications <- sessionListUpdated:
// 	default:
// 	}
// }

// func (adminSession *AdminSession) GetObjectFromUser(adminSessions *storage.ConcurrentSlice) (interface{}, error) {
// 	var err error
// 	for {
// 		err = adminSession.printSessions(adminSessions)
// 		if err != nil {
// 			return nil, err
// 		}

// 		select {
// 		case data, more := <-adminSession.readch:
// 			if !more {
// 				return nil, errors.New("connection channel closed")
// 			}
// 			idx, err := strconv.Atoi(strings.TrimSpace(string(data)))
// 			if err != nil {
// 				err = adminSession.error(fmt.Sprintf("Could not parse int: %s", err))
// 				if err != nil {
// 					return nil, err
// 				}
// 				continue
// 			}
// 			sess, ok := adminSessions.Get(idx)
// 			if !ok {
// 				err = adminSession.error(fmt.Sprintf("No adminSession with index %d!", idx))
// 				if err != nil {
// 					return nil, err
// 				}
// 				continue
// 			} else {
// 				return sess, nil
// 			}

// 		case <-adminSession.updateNotifications:
// 			err = adminSession.adminSessionListUpdate()
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 	}
// }

// func (adminSession *AdminSession) error(msg interface{}) error {
// 	_, err := adminSession.conn.Write(BPrintf("Error: %s\n", msg))
// 	return err
// }

// func (adminSession *AdminSession) String() string {
// 	return fmt.Sprintf("Admin session from %s, running for %s", adminSession.conn.RemoteAddr(), time.Since(adminSession.InitTime()))
// }

// func (adminSession *AdminSession) noSessionsAvailable() error {
// 	_, err := adminSession.conn.Write(BPrintf("No adminSessions avaiable!\nNew adminSessions will be displayed automatically.\n>"))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (adminSession *AdminSession) adminSessionListUpdate() error {
// 	_, err := adminSession.conn.Write(BPrintf("\n\nUPDATE\n"))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (adminSession *AdminSession) printSessions(adminSessions *storage.ConcurrentSlice) error {
// 	var err error
// 	if adminSessions.Len() == 0 {
// 		err = adminSession.noSessionsAvailable()
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		_, err = adminSession.conn.Write(BPrintf("Available adminSessions:\n"))
// 		if err != nil {
// 			return err
// 		}
// 		items := adminSessions.CurrentItems()
// 		for idx, sess := range items {
// 			_, err = adminSession.conn.Write(BPrintf("%d: %s\n", idx, sess.(*BackconnectSession)))
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		_, err = adminSession.conn.Write(BPrintf("Type adminSession idx to connect to : > "))
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (adminSession *AdminSession) Greet() error {
// 	_, err := adminSession.conn.Write(BPrintf("Backconnectd at %s greets you\n", adminSession.conn.LocalAddr()))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
