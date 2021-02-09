package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/hosting-de-labs/mail-knife/internal"
)

var (
	sess internal.Session
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

func stdInReader(wg *sync.WaitGroup) {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("shell: %s\n", err)
		}

		msg := strings.TrimSpace(line)
		err = sess.Send(msg)
		if err != nil {
			panic(err)
		}
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
