package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/hosting-de-labs/mail-knife/internal"
	"github.com/hosting-de-labs/mail-knife/internal/flow"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if flag.NArg() < 2 {
		os.Exit(2)
	}

	host := args[0]
	port := args[1]

	if len(host) == 0 || len(port) == 0 {
		os.Exit(3)
	}

	// prompt
	sigHandler(exitHandler)
	app := internal.NewApp(exitHandler)

	app.Flows = []internal.Flow{
		//flow.SMTPHelo{},
		flow.SMTPAuth{},
	}

	err := app.Run(fmt.Sprintf("%s:%s", host, port), args[2:])
	if err != nil {
		fmt.Printf("error running app: %s\n", err)
	}
}

func exitHandler() {
	fmt.Printf("exiting...\n")
	os.Exit(0)
}

func sigHandler(exitHandler func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		exitHandler()
	}()
}
