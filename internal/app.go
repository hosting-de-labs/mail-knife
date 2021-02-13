package internal

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
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

	conn       *Conn
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
	conn, err := Dial(connectAddr, a.ExitHandler)
	if err != nil {
		return err
	}
	a.conn = conn

	a.prompt = *prompt.New(
		executorFunc(a.conn, a.exitHandler),
		completerFunc(),
		prompt.OptionTitle("mk: interactive tcp client (like telnet command) on steroids"),
		prompt.OptionPrefix(""),
		prompt.OptionInputTextColor(prompt.Yellow),
	)

	// run flows if any
	a.runFlows(conn, args)

	// start prompt
	go a.prompt.Run()

	<-a.connClosed

	return nil
}

func (a *App) runFlows(c *Conn, args []string) {
	if len(a.Flows) > 0 {
		// run flows
		wgFlows := &sync.WaitGroup{}
		wgFlows.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for _, f := range a.Flows {
				err := f.Run(c, args)
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

func executorFunc(w *Conn, exitHandler func()) func(cmd string) {
	return func(cmd string) {
		cmd = strings.TrimSpace(cmd)

		if cmd == "exit" {
			exitHandler()
			return
		}

		err := w.PrintfLine(cmd)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}
