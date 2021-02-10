package internal

import (
	"fmt"
	"sync"

	"github.com/c-bata/go-prompt"
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
	prompt      prompt.Prompt
	exitHandler func()
	session     Session
}

func NewApp(exitHandler func()) App {
	app := App{exitHandler: exitHandler}

	return app
}

func (a *App) Run(connectAddr string) {

	// SMTP Example
	c := NewClient(LineEndingCrLf)
	defer c.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	tmpSess, err := c.Connect(connectAddr)
	if err != nil {
		panic(err)
	}

	a.session = tmpSess

	a.prompt = *prompt.New(
		executorFunc(a.session, a.exitHandler),
		completerFunc(),
		prompt.OptionTitle("mk: interactive tcp client (like telnet command) on steroids"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionInputTextColor(prompt.Yellow),
	)

	go a.prompt.Run()

	wg.Wait()
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

func executorFunc(session Session, exitHandler func()) func(cmd string) {
	return func(cmd string) {
		if cmd == "exit" {
			exitHandler()
			return
		}

		err := session.Send(cmd)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}
