package internal

import (
	"fmt"
	"net/textproto"
	"os"
	"time"

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
		},
	}
)

type App struct {
	ExitHandler func()
	LineEnding  string

	conn   *textproto.Conn
	prompt prompt.Prompt
}

func NewApp(exitHandler func()) App {
	app := App{
		ExitHandler: exitHandler,
		LineEnding:  LineEndingLf,
	}

	return app
}

func (a *App) Run(connectAddr string) error {
	tmpConn, err := textproto.Dial("tcp", connectAddr)
	if err != nil {
		return err
	}
	a.conn = tmpConn

	a.prompt = *prompt.New(
		executorFunc(a.conn, a.LineEnding, a.exitHandler),
		completerFunc(),
		prompt.OptionTitle("mk: interactive tcp client (like telnet command) on steroids"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionInputTextColor(prompt.Yellow),
	)

	// start prompt
	go a.prompt.Run()

	time.Sleep(10 * time.Second)

	return nil
}

func (a *App) exitHandler() {
	err := a.conn.Close()
	if err != nil {
		fmt.Printf("exit: failed to close connection: %s\n", err)
	}

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

func executorFunc(conn *textproto.Conn, lineEnding string, exitHandler func()) func(cmd string) {
	return func(cmd string) {
		if cmd == "exit" {
			exitHandler()
			return
		}

		err := conn.PrintfLine(cmd + lineEnding)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}
