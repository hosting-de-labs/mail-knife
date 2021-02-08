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

	return io.EOF
}

func (conn *Connection) Send(msg string) error {
	msg = strings.TrimSpace(msg)
	_, err := fmt.Fprintf(*conn.Conn, "%s%s", msg, conn.LineEnding)
	if err != nil {
		return fmt.Errorf("send: %s", err)
	}
	return nil
}

func (conn *Connection) reader(wg *sync.WaitGroup) {
	defer wg.Done()

	buf := bufio.NewReader(*conn.Conn)
	tr := textproto.NewReader(buf)

	defer conn.Close()

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

			fmt.Println("reader closed")
			break
		}

		fmt.Println(line)
	}
}
