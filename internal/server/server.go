package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/google/uuid"

	"github.com/ezzer17/backconnectd/internal/grpctypes"
	_sessions "github.com/ezzer17/backconnectd/internal/sessions"
	"github.com/ezzer17/backconnectd/pkg/storage"
	pb "github.com/ezzer17/backconnectd/proto"
)

// Server stores information about open sessions and loggers
type Server struct {
	pb.UnimplementedBackconnectdServer

	logger              *log.Logger
	backconnectSessions *storage.ConcurrentSlice
	adminSessions       *storage.ConcurrentSlice
}

// New returns an empty Server
func New(logger *log.Logger) Server {
	backconnectSessions := storage.New()
	adminSessions := storage.New()
	return Server{
		logger:              logger,
		backconnectSessions: &backconnectSessions,
		adminSessions:       &adminSessions,
	}
}

func (s *Server) broadcastEvent(evt *pb.SessionEvent) {
	sessions := s.adminSessions.CurrentItems()
	for _, s := range sessions {
		s.(*_sessions.Subscription).SendEvent(evt)
	}
}

func (s *Server) KillSession(ctx context.Context, req *pb.SessionKillRequest) (*pb.Response, error) {
	sessionID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}
	sess, ok := s.backconnectSessions.GetByID(sessionID)
	if !ok {
		return nil, fmt.Errorf("Session with id %s not found", sessionID)
	}
	sess.(*_sessions.BackconnectSession).Kill()
	return &pb.Response{Ok: true}, nil
}

func (s *Server) ConnectSession(stream pb.Backconnectd_ConnectSessionServer) error {
	srvstream := grpctypes.NewServer(stream)
	backconnectionID, err := srvstream.GetID()
	if err != nil {
		return err
	}
	bconn, ok := s.backconnectSessions.GetByID(backconnectionID)
	if !ok {
		return fmt.Errorf("session %s not found", backconnectionID)
	}
	s.logger.Printf("Connecting to session %s", backconnectionID)
	err = bconn.(*_sessions.BackconnectSession).Connect(srvstream)
	s.logger.Printf("Disconnected from %s: %s", backconnectionID, err)
	return nil

}

func (s *Server) Subscribe(req *pb.SubscribeRequest, stream pb.Backconnectd_SubscribeServer) error {
	sub := _sessions.NewSubscription(stream)
	s.adminSessions.Append(sub)
	go func() {
		for _, s := range s.backconnectSessions.CurrentItems() {
			sub.SendEvent(&pb.SessionEvent{
				Type:    pb.EventType_ADD,
				Session: s.(*_sessions.BackconnectSession).ToPb(),
			})
		}
	}()
	sub.Run()
	s.adminSessions.DeleteByID(sub.ID())
	return nil

}

func (s *Server) handleBackonnection(conn net.Conn) {
	s.logger.Printf("Recieved connection from %s", conn.RemoteAddr())
	sess := _sessions.NewBackconnection(conn)
	s.backconnectSessions.Append(sess)
	s.broadcastEvent(&pb.SessionEvent{
		Session: sess.ToPb(),
		Type:    pb.EventType_ADD,
	})
	sess.Run()
	s.broadcastEvent(&pb.SessionEvent{
		Session: sess.ToPb(),
		Type:    pb.EventType_DELETE,
	})
	s.backconnectSessions.DeleteByID(sess.ID())
	s.logger.Printf("Backconnect session closed (%s from %s)", sess.ID(), sess.RemoteAddr())
}

// AdminLoop starts listener for admin connections
func (s *Server) AdminLoop(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Fatalf("Cannot lsten on addr %s: %s", addr, err)
	}
	defer listener.Close()
	grpcserver := grpc.NewServer()
	pb.RegisterBackconnectdServer(grpcserver, s)
	s.logger.Printf("Admin server listening on tcp://%s", addr)
	if err := grpcserver.Serve(listener); err != nil {
		s.logger.Printf("failed to serve: %v", err)
	}
}

// BackconnectLoop starts listener for backconnections
func (s *Server) BackconnectLoop(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Fatalf("Cannot lsten on addr %s: %s", addr, err)
	}
	defer listener.Close()
	s.logger.Printf("Backconnect server listening on tcp://%s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Fatalf("Cannot accept connection: %s", err)
		}
		// TODO: limit size of connection pool or be killed by OS
		go s.handleBackonnection(conn)
	}
}
