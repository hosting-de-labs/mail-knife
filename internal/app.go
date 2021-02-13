package internal

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

const (
	LineEndingLf   string = "\n"
	LineEndingCrLf string = "\r\n"
)

var (
	menu = map[string][]prompt.Suggest{
		"__global": {
			{Text: "exit", Description: "exit application"},
		},
		"smtp": {
			{Text: "EHLO", Description: "Send EHLO to server"},
			{Text: "EHLO host.name", Description: "Send EHLO with test host fqdn"},
			{Text: "AUTH", Description: "Authenticate using given authentication scheme"},
			{Text: "AUTH LOGIN", Description: "Authenticate using login method"},
			{Text: "AUTH PLAIN", Description: "Authenticate using plain method"},
			{Text: "MAIL FROM: ", Description: "Send 'MAIL FROM: '"},
			{Text: "RCPT TO: ", Description: "Send 'RCPT TO: '"},
			{Text: "DATA", Description: "Send 'DATA'"},
			{Text: "QUIT", Description: "Send 'QUIT'"},
		},
	}
)

type App struct {
	ExitHandler func()
	Flows       []Flow

	conn       *textproto.Conn
	connClosed chan bool
	prompt     prompt.Prompt
}

func NewApp(exitHandler func()) App {
	app := App{
		ExitHandler: exitHandler,
		Flows:       []Flow{},
	}

	return app
}

func (a *App) Run(connectAddr string, args []string) error {
	tmpConn, err := textproto.Dial("tcp", connectAddr)
	if err != nil {
		return err
	}
	a.conn = tmpConn

	a.prompt = *prompt.New(
		executorFunc(a.conn, a.exitHandler),
		completerFunc(),
		prompt.OptionTitle("mk: interactive tcp client (like telnet command) on steroids"),
		prompt.OptionPrefix(""),
		prompt.OptionInputTextColor(prompt.Yellow),
	)

	// create reader / writer to intercept protocol messages
	r, w := logInterceptor(&a.conn.Reader, &a.conn.Writer, a.ExitHandler)

	// run flows if any
	a.runFlows(r, w, args)

	// start connection reader
	go connReader(r)

	// start prompt
	go a.prompt.Run()

	<-a.connClosed

	return nil
}

func (a *App) runFlows(r *textproto.Reader, w *textproto.Writer, args []string) {
	if len(a.Flows) > 0 {
		// run flows
		wgFlows := &sync.WaitGroup{}
		wgFlows.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for _, f := range a.Flows {
				err := f.Run(r, w, args)
				if err != nil {
					fmt.Printf("Error on running flow %T: %s\n", f, err)
				}
			}
		}(wgFlows)

		wgFlows.Wait()
	}
}

func (a *App) exitHandler() {
	a.conn.Close()

	os.Exit(0)
}

func completerFunc() func(document prompt.Document) []prompt.Suggest {
	return func(doc prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(
			append(menu["__global"], menu["smtp"]...),
			doc.Text,
			true,
		)
	}
}

func executorFunc(conn *textproto.Conn, exitHandler func()) func(cmd string) {
	return func(cmd string) {
		cmd = strings.TrimSpace(cmd)

		if cmd == "exit" {
			exitHandler()
			return
		}

		err := conn.PrintfLine(cmd)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
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
