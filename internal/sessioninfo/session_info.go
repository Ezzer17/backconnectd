package sessioninfo

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/ezzer17/backconnectd/proto"
)

type SessionInfo struct {
	SID         uuid.UUID `json:"id"`
	SInitTime   time.Time `json:"init_time"`
	SRemoteAddr string    `json:"remote_addr"`
}

func (si *SessionInfo) RemoteAddr() string {
	return si.SRemoteAddr
}

func (si *SessionInfo) ID() uuid.UUID {
	return si.SID
}

func (si *SessionInfo) InitTime() time.Time {
	return si.SInitTime
}
func (si *SessionInfo) ToPb() *pb.SessionInfo {
	return &pb.SessionInfo{
		Id:            si.SID.String(),
		InitTime:      timestamppb.New(si.SInitTime),
		RemoteAddress: si.SRemoteAddr,
	}
}
func (si *SessionInfo) String() string {
	return fmt.Sprintf("Session from %s, running for %ds", si.RemoteAddr(), int(time.Since(si.InitTime()).Seconds()))
}
