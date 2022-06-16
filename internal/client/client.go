package client

import (
	"encoding/json"
	"log"
	"net"

	"github.com/google/uuid"

	"github.com/ezzer17/backconnectd/internal/rpc"
	"github.com/ezzer17/backconnectd/pkg/channelconn"
)

type Client struct {
	conn    *channelconn.Connection
	evtChan chan *rpc.Event
	stop    chan struct{}
	errors  chan error
	raw     bool
}

type ConnectionClosed struct{}

func (cc ConnectionClosed) Error() string {
	return "connection closed"
}

func (c *Client) ChanelEvents(userevts chan<- *rpc.Event, usererrors chan<- error) {
	for {
		select {
		case evt, more := <-c.evtChan:
			if !more {
				return
			}
			userevts <- evt
		case err := <-c.errors:
			usererrors <- err
			return

		}
	}
}

func (c *Client) ChanelRawData(userread chan<- []byte, userwrite <-chan []byte) {
	c.stop <- struct{}{}
loop:
	for {
		select {
		case data, more := <-c.conn.Reader():
			if !more {
				close(userread)
				break loop
			}
			userread <- data
		case data, more := <-userwrite:
			if !more {
				break loop
			}
			c.conn.Writer() <- data
		}
	}
	go c.RPCReadLoop()
}

func Connect(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := Client{
		channelconn.New(conn),
		make(chan *rpc.Event),
		make(chan struct{}),
		make(chan error),
		false,
	}
	go c.conn.Run()
	go c.RPCReadLoop()
	return &c, nil
}

func (c *Client) RPCReadLoop() {
	reader := rpc.Reader()
	for {
		select {
		case buffer, more := <-c.conn.Reader():
			if !more {
				c.errors <- &ConnectionClosed{}
				return
			}
			for _, msg := range reader(buffer) {
				var evt rpc.Event
				err := json.Unmarshal(msg, &evt)
				if err != nil {
					c.errors <- err
				} else {
					c.evtChan <- &evt
				}

			}
		case <-c.stop:
			return
		}
	}
}

func (c *Client) SessionConnect(id uuid.UUID) error {
	return c.send(&rpc.Request{
		ID:    id,
		AType: rpc.ActionConnect,
	})
}

func (c *Client) SessionKill(id uuid.UUID) error {
	return c.send(&rpc.Request{
		ID:    id,
		AType: rpc.ActionKill,
	})
}

func (c *Client) send(req *rpc.Request) error {
	message, err := json.Marshal(req)
	if err != nil {
		return err
	}
	message = append(message, 0)
	c.conn.Writer() <- message
	return nil
}
