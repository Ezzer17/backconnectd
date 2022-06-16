package context

import (
	"fmt"
	"log"
	"net"

	"github.com/ezzer17/backconnectd/internal/backconnection"
	"github.com/ezzer17/backconnectd/internal/ctrlsession"
	"github.com/ezzer17/backconnectd/internal/rpc"
	"github.com/ezzer17/backconnectd/pkg/storage"
)

// ServerContext stores information about open sessions and loggers
type ServerContext struct {
	logger              *log.Logger
	backconnectSessions *storage.ConcurrentSlice
	adminSessions       *storage.ConcurrentSlice
}

// New returns an empty ServerContext
func New(logger *log.Logger) ServerContext {
	backconnectSessions := storage.New()
	adminSessions := storage.New()
	return ServerContext{logger, &backconnectSessions, &adminSessions}
}

func (ctx *ServerContext) broadcastEvent(evt *rpc.Event) {
	sessions := ctx.adminSessions.CurrentItems()
	for _, s := range sessions {
		s.(*ctrlsession.CliCtrlSession).SendEvent(evt)
	}
}

func (ctx *ServerContext) handleBackonnection(conn net.Conn) {
	ctx.logger.Printf("Recieved connection from %s", conn.RemoteAddr())
	sess := backconnection.NewBackconnect(conn)
	ctx.backconnectSessions.Append(sess)
	ctx.broadcastEvent(&rpc.Event{
		Sess:  sess.SessionInfo,
		EType: rpc.Add,
	})
	sess.Run()
	ctx.broadcastEvent(&rpc.Event{
		Sess:  sess.SessionInfo,
		EType: rpc.Delete,
	})
	ctx.backconnectSessions.DeleteByID(sess.ID())
	ctx.logger.Printf("Backconnect session closed (%s from %s)", sess.ID(), sess.RemoteAddr())
}

func (ctx *ServerContext) handleRequest(req *rpc.Request, ctrlSess *ctrlsession.CliCtrlSession) error {
	ctx.logger.Printf("%#v", req)
	switch req.AType {
	case rpc.ActionKill:
		sess, ok := ctx.backconnectSessions.GetByID(req.ID)
		if !ok {
			return fmt.Errorf("Session with id %s not found", req.ID)
		}
		sess.(*backconnection.BackconnectSession).Kill()
	case rpc.ActionConnect:
		sess, ok := ctx.backconnectSessions.GetByID(req.ID)
		if !ok {
			return fmt.Errorf("Session with id %s not found", req.ID)
		}
		ctrlSess.Pair(sess.(*backconnection.BackconnectSession).Connection)
	}
	return nil
}

func (ctx *ServerContext) handleAdminConnection(conn net.Conn) {
	// TODO: implement authentication
	adminSession := ctrlsession.StartCli(conn)
	ctx.adminSessions.Append(adminSession)
	for _, s := range ctx.backconnectSessions.CurrentItems() {
		adminSession.SendEvent(&rpc.Event{
			Sess:  s.(*backconnection.BackconnectSession).SessionInfo,
			EType: rpc.Add,
		})
	}
	for {
		req, err := adminSession.Accept()
		if _, ok := err.(ctrlsession.SessionStopped); ok {
			break
		}
		if err != nil {
			ctx.logger.Printf("Adminsession %s error: %s", adminSession.ID(), err)
		} else {
			if err := ctx.handleRequest(req, adminSession); err != nil {
				ctx.logger.Printf("Failed to handle admin request: %s", err)
			}
		}
	}
	ctx.adminSessions.DeleteByID(adminSession.ID())
	ctx.logger.Printf("Admin session %s closed", adminSession.ID())

}

// AdminLoop starts listener for admin connections
func (ctx *ServerContext) AdminLoop(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		ctx.logger.Fatalf("Cannot lsten on addr %s: %s", addr, err)
	}
	defer listener.Close()
	ctx.logger.Printf("Admin server listening on tcp://%s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			ctx.logger.Fatalf("Cannot accept connection: %s", err)
		}
		ctx.logger.Printf("New admin connection from %s", conn.RemoteAddr())
		go ctx.handleAdminConnection(conn)
	}
}

// BackconnectLoop starts listener for backconnections
func (ctx *ServerContext) BackconnectLoop(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		ctx.logger.Fatalf("Cannot lsten on addr %s: %s", addr, err)
	}
	defer listener.Close()
	ctx.logger.Printf("Backconnect server listening on tcp://%s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			ctx.logger.Fatalf("Cannot accept connection: %s", err)
		}
		// TODO: limit size of connection pool or be killed by OS
		go ctx.handleBackonnection(conn)
	}
}
