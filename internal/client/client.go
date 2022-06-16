package client

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"google.golang.org/grpc"

	"github.com/ezzer17/backconnectd/internal/grpctypes"
	"github.com/ezzer17/backconnectd/pkg/asyncrw"
	pb "github.com/ezzer17/backconnectd/proto"
)

type Client struct {
	grpcconn pb.BackconnectdClient
	evts     chan *pb.SessionEvent
	errors   chan error
}

type ConnectionClosed struct{}

func (cc ConnectionClosed) Error() string {
	return "connection closed"
}

func (c *Client) ChanelEvents(userevts chan<- *pb.SessionEvent, usererrors chan<- error) {
	for {
		select {
		case evt, more := <-c.evts:
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

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		pb.NewBackconnectdClient(conn),
		make(chan *pb.SessionEvent),
		make(chan error),
	}
}

func (c *Client) Subscribe() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := c.grpcconn.Subscribe(ctx, &pb.SubscribeRequest{})
	if err != nil {
		c.errors <- err
	}
	for {
		evt, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			c.errors <- err
		}
		c.evts <- evt
	}
}

func (c *Client) SessionConnect(id uuid.UUID, send chan<- []byte, recieve <-chan []byte) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := c.grpcconn.ConnectSession(ctx)
	if err != nil {
		return err
	}
	clistream := grpctypes.NewClient(stream)
	if err := clistream.SendID(id); err != nil {
		return err
	}
	return asyncrw.AsyncSendRecieve(clistream, send, recieve)
}

func (c *Client) SessionKill(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := c.grpcconn.KillSession(ctx, &pb.SessionKillRequest{
		Id: id.String(),
	})
	return err
}
