package grpctypes

import (
	pb "github.com/ezzer17/backconnectd/proto"
	"github.com/google/uuid"
)

type Server struct {
	stream pb.Backconnectd_ConnectSessionServer
}

func NewServer(stream pb.Backconnectd_ConnectSessionServer) *Server {
	return &Server{stream}
}

func (c *Server) Send(data []byte) error {
	return c.stream.Send(&pb.RawData{
		Message: &pb.RawData_Data{
			Data: data,
		},
	})
}

func (c *Server) Recieve() ([]byte, error) {
	data, err := c.stream.Recv()
	if err != nil {
		return nil, err
	}
	return data.GetData(), nil
}

func (c *Server) GetID() (uuid.UUID, error) {
	data, err := c.stream.Recv()
	if err != nil {
		return uuid.UUID{}, err
	}
	id, err := uuid.Parse(data.GetId())
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}
