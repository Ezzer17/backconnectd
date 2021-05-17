package session

import (
	"errors"
	"fmt"
	"github.com/ezzer17/backconnectd/storage"
	"github.com/google/uuid"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
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
	RemoteAddr() net.Addr
}

type session struct {
	ctrl                chan sessionCtrlMsg
	readch              chan []byte
	updateNotifications chan sessionUpdateNotification
	conn                net.Conn
	initTime            time.Time
	id                  uuid.UUID
	closeCb             func(Session)
}

type AdminSession struct {
	session
}

type BackconnectSession struct {
	session
}

func BPrintf(msg string, args ...interface{}) []byte {
	return []byte(fmt.Sprintf(msg, args...))
}

func (session *session) close() {
	session.conn.Close()
	close(session.readch)
	close(session.ctrl)
	session.closeCb(session)
}

func (session *session) Run() {
	data := make([]byte, 1024)
	err := session.conn.SetReadDeadline(time.Now())
	if err != nil {
		session.close()
		return
	}
	for {
		select {
		case <-session.ctrl:
			session.close()
			return
		default:
		}
		n, err := session.conn.Read(data)
		if err == io.EOF { // TODO: handle errors
			session.close()
			return
		}
		if n > 0 {
			sendme := make([]byte, n)
			copy(sendme, data[:n])
			session.readch <- sendme
		}
		err = session.conn.SetReadDeadline(time.Now().Add(checkInterval))
		if err != nil {
			session.close()
			return
		}

	}
}

func new(conn net.Conn, closeCb func(Session)) session {
	ctrl := make(chan sessionCtrlMsg, 1)
	readch := make(chan []byte, channelSize)
	updatech := make(chan sessionUpdateNotification, 1)
	id := uuid.New()
	return session{ctrl, readch, updatech, conn, time.Now(), id, closeCb}
}

func NewAdmin(conn net.Conn, closeCb func(Session)) AdminSession {
	sess := new(conn, closeCb)
	adminsess := AdminSession{session: sess}
	return adminsess
}

func NewBackconnect(conn net.Conn, closeCb func(Session)) BackconnectSession {
	sess := new(conn, closeCb)
	return BackconnectSession{session: sess}
}

func (session *session) NotifyOfSessionListUpdate() {
	select {
	case session.updateNotifications <- sessionListUpdated:
	default:
	}
}

func (session *session) RemoteAddr() net.Addr {
	return session.conn.RemoteAddr()
}

func (session *session) ID() uuid.UUID {
	return session.id
}

func (adminSession *AdminSession) String() string {
	return fmt.Sprintf("Admin session from %s, running for %s", adminSession.conn.RemoteAddr(), time.Since(adminSession.initTime))
}

func (session *BackconnectSession) String() string {
	return fmt.Sprintf("Backconnect session from %s, running for %s", session.conn.RemoteAddr(), time.Since(session.initTime))
}

func (adminSession *AdminSession) ConnectTo(backconnectSession *BackconnectSession) error {
	for {
		select {
		case data, more := <-adminSession.readch:
			if !more {
				return errors.New("admin connection closed")
			}
			_, err := backconnectSession.conn.Write(data)
			if err != nil {
				return err
			}
		case data, more := <-backconnectSession.readch:
			if !more {
				return errors.New("backconnection closed")
			}
			_, err := adminSession.conn.Write(data)
			if err != nil {
				return err
			}

		}
	}
}

func (adminSession *AdminSession) GetObjectFromUser(adminSessions *storage.ConcurrentSlice) (interface{}, error) {
	var err error
	for {
		err = adminSession.printSessions(adminSessions)
		if err != nil {
			return nil, err
		}

		select {
		case data, more := <-adminSession.readch:
			if !more {
				return nil, errors.New("connection channel closed")
			}
			idx, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				err = adminSession.error(fmt.Sprintf("Could not parse int: %s", err))
				if err != nil {
					return nil, err
				}
				continue
			}
			sess, ok := adminSessions.Get(idx)
			if !ok {
				err = adminSession.error(fmt.Sprintf("No adminSession with index %d!", idx))
				if err != nil {
					return nil, err
				}
				continue
			} else {
				return sess, nil
			}

		case <-adminSession.updateNotifications:
			err = adminSession.adminSessionListUpdate()
			if err != nil {
				return nil, err
			}
		}
	}
}

func (adminSession *AdminSession) error(msg interface{}) error {
	_, err := adminSession.conn.Write(BPrintf("Error: %s\n", msg))
	return err
}

func (adminSession *AdminSession) noSessionsAvailable() error {
	_, err := adminSession.conn.Write(BPrintf("No adminSessions avaiable!\nNew adminSessions will be displayed automatically.\n>"))
	if err != nil {
		return err
	}
	return nil
}

func (adminSession *AdminSession) adminSessionListUpdate() error {
	_, err := adminSession.conn.Write(BPrintf("\n\nUPDATE\n"))
	if err != nil {
		return err
	}
	return nil
}

func (adminSession *AdminSession) printSessions(adminSessions *storage.ConcurrentSlice) error {
	var err error
	if adminSessions.Len() == 0 {
		err = adminSession.noSessionsAvailable()
		if err != nil {
			return err
		}
	} else {
		_, err = adminSession.conn.Write(BPrintf("Available adminSessions:\n"))
		if err != nil {
			return err
		}
		items := adminSessions.CurrentItems()
		for idx, sess := range items {
			_, err = adminSession.conn.Write(BPrintf("%d: %s\n", idx, sess.(*BackconnectSession)))
			if err != nil {
				return err
			}
		}
		_, err = adminSession.conn.Write(BPrintf("Type adminSession idx to connect to : > "))
		if err != nil {
			return err
		}
	}
	return nil
}

func (adminSession *AdminSession) Greet() error {
	_, err := adminSession.conn.Write(BPrintf("Backconnectd at %s greets you\n", adminSession.conn.LocalAddr()))
	if err != nil {
		return err
	}
	return nil
}
