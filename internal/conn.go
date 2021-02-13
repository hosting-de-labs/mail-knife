package internal

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"

	"github.com/fatih/color"
)

var (
	_ io.Closer = &Conn{}
)

type Conn struct {
	Reader *textproto.Reader
	Writer *textproto.Writer

	rawConn *textproto.Conn
}

func Dial(addr string, exitHandler func()) (*Conn, error) {
	rawConn, err := textproto.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %s", err)
	}

	// create reader / writer to intercept protocol messages
	r, w := logInterceptor(&rawConn.Reader, &rawConn.Writer, exitHandler)

	return &Conn{
		Reader:  r,
		Writer:  w,
		rawConn: rawConn,
	}, nil
}

// Close implements io.Closer interface
func (c Conn) Close() error {
	return c.rawConn.Close()
}

func logInterceptor(r *textproto.Reader, w *textproto.Writer, exitHandler func()) (*textproto.Reader, *textproto.Writer) {
	done := make(chan bool)

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

			_, _ = colorReceive.Printf("<<< %s\n", l)
			err = w.PrintfLine(l)
			if err != nil {
				fmt.Printf("receive-logger error write: %s\n")
			}
		}
	}(r, receiverTextW)

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

			_, _ = colorSend.Printf(">>> %s\n", l)
			err = w.PrintfLine(l)
			if err != nil {
				fmt.Printf("send-logger error write: %s\n")
			}
		}
	}(senderTextR, w)

	go func(chan<- bool) {
		<-done
		senderPipeW.Close()
		exitHandler()
	}(done)

	return receiverTextR, senderTextW
}

func connReader(r *textproto.Reader) {
	for {
		_, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				// close on EOF
				break
			}

			fmt.Printf("connReader: error occured: %s\n", err)
		}
	}
}
