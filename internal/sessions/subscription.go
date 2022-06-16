package sessions

import (
	"github.com/google/uuid"

	pb "github.com/ezzer17/backconnectd/proto"
)

type Subscription struct {
	id     uuid.UUID
	evts   chan *pb.SessionEvent
	stream pb.Backconnectd_SubscribeServer
}

func NewSubscription(stream pb.Backconnectd_SubscribeServer) *Subscription {
	sub := Subscription{
		id:     uuid.New(),
		evts:   make(chan *pb.SessionEvent),
		stream: stream,
	}
	return &sub
}

func (s *Subscription) Run() {
	for evt := range s.evts {
		if err := s.stream.Send(evt); err != nil {
			// Handle errors
			return
		}
	}
}
func (ci *Subscription) SendEvent(evt *pb.SessionEvent) {
	ci.evts <- evt
}

func (ci *Subscription) ID() uuid.UUID {
	return ci.id
}
