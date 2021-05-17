package context

import (
	"github.com/ezzer17/backconnectd/session"
	"github.com/ezzer17/backconnectd/storage"
	"log"
	"net"
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

func (ctx *ServerContext) notifyOfSessionListUpdate() {
	sessions := ctx.adminSessions.CurrentItems()
	for _, s := range sessions {
		s.(*session.AdminSession).NotifyOfSessionListUpdate()
	}
}

func (ctx *ServerContext) backconnectSessionCloseCallback(session session.Session) {
	ctx.notifyOfSessionListUpdate()
	ctx.backconnectSessions.DeleteByID(session.ID())
	ctx.logger.Printf("Backconnect session closed (%s from %s)", session.ID(), session.RemoteAddr())
}

func (ctx *ServerContext) adminSessionCloseCallback(session session.Session) {
	ctx.adminSessions.DeleteByID(session.ID())
	ctx.logger.Printf("Admin session closed (%s from %s)", session.ID(), session.RemoteAddr())
}

func (ctx *ServerContext) startBackconnectSession(conn net.Conn) session.BackconnectSession {
	sess := session.NewBackconnect(conn, ctx.backconnectSessionCloseCallback)
	ctx.backconnectSessions.Append(&sess)
	ctx.notifyOfSessionListUpdate()
	go sess.Run()
	return sess
}

func (ctx *ServerContext) startAdminSession(conn net.Conn) (session.AdminSession, error) {
	sess := session.NewAdmin(conn, ctx.adminSessionCloseCallback)
	err := sess.Greet()
	if err != nil {
		return sess, err
	}
	ctx.adminSessions.Append(&sess)
	go sess.Run()
	return sess, nil
}

func (ctx *ServerContext) handleBackonnection(conn net.Conn) {
	ctx.logger.Printf("Recieved connection from %s", conn.RemoteAddr())
	ctx.startBackconnectSession(conn)
}

func (ctx *ServerContext) handleAdminConnection(conn net.Conn) {
	// TODO: implement authentication
	ctx.logger.Printf("New admin connection from %s", conn.RemoteAddr())
	adminSession, err := ctx.startAdminSession(conn)
	if err != nil {
		ctx.logger.Printf("Unexpected error handling %s", conn.RemoteAddr())
		return
	}
	for {
		obj, err := adminSession.GetObjectFromUser(ctx.backconnectSessions)
		if err != nil {
			ctx.logger.Printf("End session from %s: %s", adminSession.RemoteAddr(), err)
			break
		}
		backconnectSession := obj.(*session.BackconnectSession)
		ctx.logger.Printf("Getting session from %s for admin on %s", backconnectSession.RemoteAddr(), adminSession.RemoteAddr())
		err = adminSession.ConnectTo(backconnectSession)
		ctx.logger.Printf("Disconnected: %s", err)
	}

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
