package context

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"testing"
	"time"
)

const testbackconnectAddr = ":2234"
const testadminAddr = ":2235"

var testData []byte
var testData2 []byte

var globalTctx testingCtx

func init() {
	globalTctx = newTestingCtx()
	globalTctx.startTestLoops()
	testData = []byte("Hello!\n")
	testData2 = []byte("U shall never escape me, filthy creatures\n")
}

type testingCtx struct {
	ctx    *ServerContext
	writer *flushingWriter
}

func newTestingCtx() testingCtx {
	writer := &flushingWriter{""}
	logger := log.New(writer, "", 0)
	ctx := New(logger)
	return testingCtx{&ctx, writer}
}
func (tctx *testingCtx) startTestLoops() {
	go tctx.ctx.BackconnectLoop(testbackconnectAddr)
	go tctx.ctx.AdminLoop(testadminAddr)
	for len(tctx.writer.logOutput) == 0 {
	}
	for strings.Count(tctx.writer.logOutput, "server listening") < 2 {
	}
}

func (tctx *testingCtx) newAdminConnection() (net.Conn, error) {
	logString := "New admin connection"
	oldCount := strings.Count(tctx.writer.logOutput, logString)
	adminconn, err := net.Dial("tcp", testadminAddr)
	if err != nil {
		return nil, err
	}
	for strings.Count(tctx.writer.logOutput, logString) == oldCount {
	}
	return adminconn, nil
}

func (tctx *testingCtx) newBackConnection() (net.Conn, error) {
	logString := "Recieved connection"
	oldCount := strings.Count(tctx.writer.logOutput, logString)
	bconn, err := net.Dial("tcp", testbackconnectAddr)
	if err != nil {
		return nil, err
	}
	for strings.Count(tctx.writer.logOutput, logString) == oldCount {
	}
	return bconn, nil
}

func (tctx *testingCtx) reliablyCloseBackconn(backconn net.Conn) {
	logString := "Backconnect session closed "
	oldCount := strings.Count(tctx.writer.logOutput, logString)
	backconn.Close()
	for strings.Count(tctx.writer.logOutput, logString) == oldCount {
	}
}

func (tctx *testingCtx) reliablyCoupleSessions(adminconn net.Conn, sessionid string) error {
	logString := "Getting session from"
	oldCount := strings.Count(tctx.writer.logOutput, logString)
	_, err := adminconn.Write([]byte(sessionid))
	if err != nil {
		return err
	}
	for strings.Count(tctx.writer.logOutput, logString) == oldCount {
	}
	return nil
}

type flushingWriter struct {
	logOutput string
}

func (w *flushingWriter) Write(data []byte) (int, error) {
	w.logOutput += string(data)
	return len(data), nil
}

func readMenu(с net.Conn) ([]string, error) {
	var sessionids []string
	r := bufio.NewReader(с)
	data, err := r.ReadString('>')
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(data, "\n") {
		if strings.Count(line, "session from") != 0 {
			sessionids = append(sessionids, strings.Trim(strings.Split(line, " ")[0], ":"))
		}
	}
	return sessionids, nil
}

func TestOpenCloseBackConnection(t *testing.T) {
	backconn, err := globalTctx.newBackConnection()
	if err != nil {
		fmt.Printf(globalTctx.writer.logOutput)
		t.Errorf("%s", err)
	}

	adminconn, err := globalTctx.newAdminConnection()
	if err != nil {
		fmt.Printf(globalTctx.writer.logOutput)
		t.Errorf("%s", err)
	}
	sessionids, err := readMenu(adminconn)
	if err != nil {
		fmt.Printf(globalTctx.writer.logOutput)
		t.Errorf("%s", err)
	}
	if len(sessionids) != 1 {
		fmt.Printf(globalTctx.writer.logOutput)
		t.Errorf("Unexpected number of open sessions: %d", len(sessionids))
	}
	adminconn.Close()
	globalTctx.reliablyCloseBackconn(backconn)

	adminconn, err = globalTctx.newAdminConnection()
	if err != nil {
		fmt.Printf(globalTctx.writer.logOutput)
		t.Errorf("%s", err)
	}
	defer adminconn.Close()
	sessionids, err = readMenu(adminconn)
	if err != nil {
		fmt.Printf(globalTctx.writer.logOutput)
		t.Errorf("%s", err)
	}
	if len(sessionids) != 0 {
		fmt.Printf(globalTctx.writer.logOutput)
		t.Errorf("Unexpected number of open sessions: %d", len(sessionids))
	}
}

func TestConnectionCoupling(t *testing.T) {

	backconn, err := globalTctx.newBackConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	defer backconn.Close()

	adminconn, err := globalTctx.newAdminConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	defer adminconn.Close()
	sessionids, err := readMenu(adminconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if len(sessionids) != 1 {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Unexpected number of open sessions: %d", len(sessionids))
	}
	_, err = adminconn.Write([]byte(sessionids[0]))
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	_, err = backconn.Write(testData)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	recvAdmin := make([]byte, 1024)
	l, err := adminconn.Read(recvAdmin)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if string(recvAdmin[:l]) != string(testData) {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Recieved unexpected data: %s", recvAdmin)
	}
	_, err = adminconn.Write(testData2)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	recvBack := make([]byte, 1024)
	l, err = backconn.Read(recvBack)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if string(recvBack[:l]) != string(testData2) {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Recieved unexpected data: %s", recvBack)
	}
}

func TestBackConnectionBuffering(t *testing.T) {
	backconn, err := globalTctx.newBackConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}

	adminconn, err := globalTctx.newAdminConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	defer adminconn.Close()
	sessionids, err := readMenu(adminconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if len(sessionids) != 1 {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Unexpected number of open sessions: %d", len(sessionids))
	}
	// Write to backconnection
	_, err = backconn.Write(testData)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	// Read written data after, data should be buffered
	err = globalTctx.reliablyCoupleSessions(adminconn, sessionids[0])
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	recvAdmin := make([]byte, 1024)
	l, err := adminconn.Read(recvAdmin)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if string(recvAdmin[:l]) != string(testData) {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Recieved unexpected data: %s", recvAdmin)
	}
	globalTctx.reliablyCloseBackconn(backconn)
}

func TestBackConnectionPersistance(t *testing.T) {
	backconn, err := globalTctx.newBackConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	defer backconn.Close()

	adminconn, err := globalTctx.newAdminConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	sessionids, err := readMenu(adminconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if len(sessionids) != 1 {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Unexpected number of open sessions: %d", len(sessionids))
	}
	err = globalTctx.reliablyCoupleSessions(adminconn, sessionids[0])
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	_, err = backconn.Write(testData)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	recvAdmin := make([]byte, 1024)
	l, err := adminconn.Read(recvAdmin)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if string(recvAdmin[:l]) != string(testData) {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Recieved unexpected data: %s", recvAdmin)
	}
	adminconn.Close()

	// Connect again and try to write to backconnection
	adminconn, err = globalTctx.newAdminConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	defer adminconn.Close()
	sessionids, err = readMenu(adminconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if len(sessionids) != 1 {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Unexpected number of open sessions: %d", len(sessionids))
	}
	err = globalTctx.reliablyCoupleSessions(adminconn, sessionids[0])
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}

	_, err = adminconn.Write(testData2)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}

	recvBconn := make([]byte, 1024)
	err = backconn.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	l, err = backconn.Read(recvBconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if string(recvBconn[:l]) != string(testData2) {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Recieved unexpected data: %s", recvBconn)
	}
}

func TestMenuUpdate(t *testing.T) {
	adminconn, err := globalTctx.newAdminConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	defer adminconn.Close()
	sessionids, err := readMenu(adminconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if len(sessionids) != 0 {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Unexpected number of open sessions: %d", len(sessionids))
	}

	backconnSkip, err := globalTctx.newBackConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	// This is the only way to reliably test that menu updates two times
	time.Sleep(time.Millisecond * 50)
	backconn, err := globalTctx.newBackConnection()
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}

	sessionids, err = readMenu(adminconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if len(sessionids) != 1 {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Unexpected number of open sessions: %d", len(sessionids))
	}

	sessionids, err = readMenu(adminconn)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if len(sessionids) != 2 {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Unexpected number of open sessions: %d", len(sessionids))
	}

	// Write to backconnection
	_, err = backconn.Write(testData)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	err = globalTctx.reliablyCoupleSessions(adminconn, sessionids[1])
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	recvAdmin := make([]byte, 1024)
	l, err := adminconn.Read(recvAdmin)
	if err != nil {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("%s", err)
	}
	if string(recvAdmin[:l]) != string(testData) {
		t.Errorf(globalTctx.writer.logOutput)
		t.Fatalf("Recieved unexpected data: %s", recvAdmin)
	}
	globalTctx.reliablyCloseBackconn(backconn)
	globalTctx.reliablyCloseBackconn(backconnSkip)
}
