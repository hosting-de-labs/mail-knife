package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/hosting-de-labs/mail-knife/internal/flow"

	"github.com/hosting-de-labs/mail-knife/internal"
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

	app.Run(fmt.Sprintf("%s:%s", host, port))
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
