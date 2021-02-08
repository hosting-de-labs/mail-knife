package internal

import (
	"net"
	"sync"
)

type Client struct {
	LineEnding string

	conns    []Connection
	connsMux *sync.Mutex
	connsWg  *sync.WaitGroup
}

func NewClient(lineEnding string) Client {
	return Client{
		LineEnding: lineEnding,
		connsMux:   &sync.Mutex{},
		connsWg:    &sync.WaitGroup{},
	}
}

func (c *Client) Close() error {
	c.connsMux.Lock()

	for _, conn := range c.conns {
		conn.Close()
	}

	// TODO: timeout
	c.connsWg.Wait()

	return nil
}

func (c *Client) Connect(addr string) (Session, error) {
	rawConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn := NewConnection(&rawConn)
	conn.LineEnding = c.LineEnding

	c.connsMux.Lock()
	c.conns = append(c.conns, conn)
	c.connsMux.Unlock()

	c.connsWg.Add(1)
	go conn.reader(c.connsWg)

	return &conn, err
}
