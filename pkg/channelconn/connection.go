package channelconn

import (
	"io"
	"net"
	"time"

	"github.com/ezzer17/backconnectd/pkg/asyncrw"
)

const checkInterval = 2 * 500 * time.Millisecond
const channelSize = 1024

type Connection struct {
	conn      net.Conn
	readerctl chan struct{}
	writerctl chan struct{}
	readch    chan []byte
	writech   chan []byte
	closech   chan struct{}
}

func (c *Connection) close() {
	c.conn.Close()
	close(c.readch)
	close(c.writech)
	close(c.readerctl)
	close(c.writerctl)
}

func (c *Connection) reader() {
	data := make([]byte, 1024)
	err := c.conn.SetReadDeadline(time.Now())
	if err != nil {
		return
	}
	for {
		select {
		case <-c.readerctl:
			return
		default:
		}
		n, err := c.conn.Read(data)
		if err == io.EOF { // TODO: handle errors
			return
		}
		if n > 0 {
			sendme := make([]byte, n)
			copy(sendme, data[:n])
			c.readch <- sendme
		}
		err = c.conn.SetReadDeadline(time.Now().Add(checkInterval))
		if err != nil {
			return
		}

	}
}

func (c *Connection) writer() {
	err := c.conn.SetWriteDeadline(time.Now())
	if err != nil {
		return
	}
	for {
		select {
		case <-c.writerctl:
			return
		case data := <-c.writech:
			if len(data) > 0 {
				for {
					n, err := c.conn.Write(data)

					if err == io.EOF { // TODO: handle errors
						return
					}
					if len(data) == n {
						break
					}
					if n > 0 {
						data = data[n:]
					}
					err = c.conn.SetWriteDeadline(time.Now().Add(checkInterval))
					if err != nil {
						return
					}
				}
			}
		}
	}
}

func New(conn net.Conn) *Connection {
	return &Connection{
		readerctl: make(chan struct{}),
		writerctl: make(chan struct{}),
		readch:    make(chan []byte, channelSize),
		writech:   make(chan []byte, channelSize),
		conn:      conn,
	}
}

func (c *Connection) Run() {
	go c.writer()
	c.reader()
	c.close()
}

func (c *Connection) Stop() {
	c.writerctl <- struct{}{}
	c.readerctl <- struct{}{}
}

func (c *Connection) Connect(a asyncrw.AsyncSenderReciever) error {
	return asyncrw.AsyncSendRecieve(a, c.writech, c.readch)
}

func (c *Connection) Reader() <-chan []byte {
	return c.readch
}

func (c *Connection) Writer() chan<- []byte {
	return c.writech
}
