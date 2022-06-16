package rpc

import (
	"bytes"

	"github.com/ezzer17/backconnectd/internal/sessioninfo"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/event"
)

type EventType int

const (
	Delete EventType = iota
	Add
	Close
)

type ActionType int

const (
	ActionConnect ActionType = iota
	ActionKill
)

type ServerEvent struct{
	event interface{}
}

func (e *ServerEvent) MarshalJSON() []byte {
	var 
	switch event.(type) {
	case NewSessionEvent:
	case DeadSessionEvent:
	case CloseEvent:
	}

}

type NewSessionEvent struct {
	Sess sessioninfo.SessionInfo `json:"session_info"`
}

func (ne *NewSessionEvent) SessionInfo() sessioninfo.SessionInfo {
	return ne.Sess
}

type DeadSessionEvent struct {
	ID uuid.UUID `json:"id"`
}

func (de *DeadSessionEvent) SessionID() uuid.UUID {
	return de.ID
}

type CloseEvent struct{}

func (ce *CloseEvent) Data() []byte {
	return []byte{}
}

type Request struct {
	ID    uuid.UUID  `json:"id,omitempty"`
	AType ActionType `json:"action_type"`
}

type Gen func([]byte) [][]byte

func Reader() Gen {
	buffer := []byte{}
	return func(more []byte) [][]byte {
		messages := [][]byte{}
		for _, m := range bytes.Split(append(buffer, more...), []byte{0}) {
			if len(m) > 0 {
				messages = append(messages, m)
			}
		}
		if more[len(more)-1] == 0 {
			buffer = []byte{}
			return messages
		}
		buffer = messages[len(messages)-1]
		return messages[:len(messages)-1]

	}

}
