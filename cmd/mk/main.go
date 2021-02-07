package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

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

	// SMTP Example
	c := internal.NewClient(internal.LineEndingCrLf)
	defer c.Close()

	tmpSess, err := c.Connect(fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		panic(err)
	}
	sess = tmpSess

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("shell: %s\n", err)
		}

		send(strings.TrimSpace(line))
	}
}

func send(msg string) {
	err := sess.Send(msg)
	if err != nil {
		panic(err)
	}
}
