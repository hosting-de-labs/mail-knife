package internal

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"sync"
)

const (
	LineEndingLf   string = "\n"
	LineEndingCrLf string = "\r\n"
)

var (
	// compile-time interface implement check
	_ Session = &Connection{}
)

type Connection struct {
	Conn       *net.Conn
	LineEnding string

	done     chan bool
	readerWg *sync.WaitGroup
}

func NewConnection(rawCon *net.Conn) Connection {
	conn := Connection{
		Conn:       rawCon,
		LineEnding: LineEndingLf,

		done:     make(chan bool),
		readerWg: &sync.WaitGroup{},
	}

	return conn
}

func (conn *Connection) Close() error {
	conn.done <- true

	fmt.Println("Before Wait")

	//TODO: timeout
	conn.readerWg.Wait()

	return nil
}

func (conn *Connection) Send(msg string) error {
	msg = strings.TrimSpace(msg)
	_, err := fmt.Fprintf(*conn.Conn, "%s%s", msg, conn.LineEnding)
	if err != nil {
		return fmt.Errorf("send: %s", err)
	}
	return nil
}

func (conn *Connection) reader(wg *sync.WaitGroup) error {
	defer wg.Done()
	wg.Add(1)

	buf := bufio.NewReader(*conn.Conn)
	tr := textproto.NewReader(buf)

	for {
		select {
		case <-conn.done:
			break

		default:
		}

		line, err := tr.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}

		fmt.Println(line)
	}

	return nil
}
