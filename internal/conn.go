package internal

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	_ io.Closer = &Conn{}
)

type Conn struct {
	Text *textproto.Conn

	messages chan string
	netConn  net.Conn
}

func Dial(addr string, exitHandler func()) (*Conn, error) {
	netConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %s", err)
	}

	rawConn := textproto.NewConn(netConn)

	// create reader / writer to intercept protocol messages
	r, w := logInterceptor(rawConn, exitHandler)

	logConn := textproto.NewConn(netConn)
	logConn.Reader = *r
	logConn.Writer = *w

	c := &Conn{
		messages: make(chan string, 64),
		netConn:  netConn,
		Text:     logConn,
	}

	// start connection reader
	go connReader(&logConn.Reader, c.messages)

	return c, nil
}

// Close implements io.Closer interface
func (c Conn) Close() error {
	return c.netConn.Close()
}

func (c Conn) WaitMessage(prefix string, timeout time.Duration) (string, error) {
	for {
		select {
		case l := <-c.messages:
			if strings.HasPrefix(l, prefix) {
				return l, nil
			}
		case <-time.After(timeout):
			return "", fmt.Errorf("wait message: timeout reached")
		}
	}
}

func (c Conn) PrintfLine(format string, args ...interface{}) error {
	return c.Text.PrintfLine(format, args...)
}

func connReader(r *textproto.Reader, lines chan<- string) {
	for {
		l, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				// close on EOF
				break
			}

			fmt.Printf("connReader: error occured: %s\n", err)
		}

		lines <- l
	}
}

func logInterceptor(c *textproto.Conn, exitHandler func()) (*textproto.Reader, *textproto.Writer) {
	done := make(chan bool)
	printMux := &sync.Mutex{}

	colorReceive := color.New(color.FgCyan)
	colorSend := color.New(color.FgYellow)

	receiverPipeR, receiverPipeW := io.Pipe()

	receiverTextR := textproto.NewReader(bufio.NewReader(receiverPipeR))
	receiverTextW := textproto.NewWriter(bufio.NewWriter(receiverPipeW))
	go func(r *textproto.Reader, w *textproto.Writer) {
		for {
			l, err := r.ReadLine()
			if err != nil {
				if err == io.EOF {
					done <- true
					break
				}

				fmt.Printf("receive-logger error read: %s\n", err)
			}

			err = w.PrintfLine(l)
			if err != nil {
				fmt.Printf("receive-logger error write: %s\n")
			}

			printMux.Lock()
			_, _ = colorReceive.Printf("<<< %s\n", l)
			printMux.Unlock()
		}
	}(&c.Reader, receiverTextW)

	senderPipeR, senderPipeW := io.Pipe()

	senderTextR := textproto.NewReader(bufio.NewReader(senderPipeR))
	senderTextW := textproto.NewWriter(bufio.NewWriter(senderPipeW))
	go func(r *textproto.Reader, w *textproto.Writer) {
		for {
			l, err := r.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}

				fmt.Printf("send-logger error read: %s\n", err)
			}

			err = w.PrintfLine(l)
			if err != nil {
				fmt.Printf("send-logger error write: %s\n")
			}

			printMux.Lock()
			_, _ = colorSend.Printf(">>> %s\n", l)
			printMux.Unlock()
		}
	}(senderTextR, &c.Writer)

	go func(chan<- bool) {
		<-done
		senderPipeW.Close()
		exitHandler()
	}(done)

	return receiverTextR, senderTextW
}
