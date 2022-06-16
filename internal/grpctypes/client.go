package grpctypes

import (
	pb "github.com/ezzer17/backconnectd/proto"
	"github.com/google/uuid"
)

type Client struct {
	stream pb.Backconnectd_ConnectSessionClient
}

func NewClient(stream pb.Backconnectd_ConnectSessionClient) *Client {
	return &Client{stream}
}

func (c *Client) Send(data []byte) error {
	return c.stream.Send(&pb.RawData{
		Message: &pb.RawData_Data{
			Data: data,
		},
	})
}

func (c *Client) Recieve() ([]byte, error) {
	data, err := c.stream.Recv()
	if err != nil {
		return nil, err
	}
	return data.GetData(), nil
}

func (c *Client) SendID(id uuid.UUID) error {
	return c.stream.Send(&pb.RawData{
		Message: &pb.RawData_Id{
			Id: id.String(),
		},
	})
}
