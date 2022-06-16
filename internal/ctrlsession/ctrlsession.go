package ctrlsession

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/ezzer17/backconnectd/internal/rpc"
	"github.com/ezzer17/backconnectd/pkg/channelconn"
)

type UInterface interface {
	ReqChan() <-chan *rpc.Request
	EvtChan() chan<- *rpc.Event
}

type CliCtrlSession struct {
	id     uuid.UUID
	reqs   chan *rpc.Request
	evts   chan *rpc.Event
	errors chan error
	pairs  chan *channelconn.Connection
	conn   *channelconn.Connection
}

type SessionStopped struct{}

func (ss SessionStopped) Error() string {
	return "session stopped"
}

const checkInterval = 500 * time.Millisecond

func StartCli(conn net.Conn) *CliCtrlSession {
	int := CliCtrlSession{
		uuid.New(),
		make(chan *rpc.Request),
		make(chan *rpc.Event),
		make(chan error),
		make(chan *channelconn.Connection),
		channelconn.New(conn),
	}
	go int.run()
	return &int
}

func (ci *CliCtrlSession) run() {
	go ci.conn.Run()
	ci.processor()
	ci.Stop()
}

func (ci *CliCtrlSession) ReqChan() <-chan *rpc.Request {
	return ci.reqs
}

func (ci *CliCtrlSession) Accept() (*rpc.Request, error) {
	//TODO: errors
	req, more := <-ci.reqs
	if !more {
		return nil, SessionStopped{}
	}
	return req, nil

}
func (ci *CliCtrlSession) Pair(other *channelconn.Connection) {
	ci.pairs <- other
}

func (ci *CliCtrlSession) processor() {
	reader := rpc.Reader()
	for {
		select {
		case buffer, more := <-ci.conn.Reader():
			if !more {
				return
			}
			for _, msg := range reader(buffer) {
				var req rpc.Request
				err := json.Unmarshal(msg, &req)
				if err != nil {
					ci.errors <- fmt.Errorf("unmarshal error: %s", err)
				} else {
					ci.reqs <- &req
				}
			}

		case evt, more := <-ci.evts:
			if !more {
				return
			}
			message, err := json.Marshal(evt)
			if err != nil {
				ci.errors <- fmt.Errorf("marshal error: %s", err)
			}
			message = append(message, 0)
			ci.conn.Writer() <- message
		case other := <-ci.pairs:
			ci.conn.Pair(other)
		}
	}
}

func (ci *CliCtrlSession) SendEvent(evt *rpc.Event) {
	ci.evts <- evt
}

func (ci *CliCtrlSession) Stop() {
	close(ci.evts)
	close(ci.reqs)
	close(ci.errors)
}

func (ci *CliCtrlSession) ID() uuid.UUID {
	return ci.id
}
