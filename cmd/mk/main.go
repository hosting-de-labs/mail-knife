package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/hosting-de-labs/mail-knife/internal/flow"

	"github.com/hosting-de-labs/mail-knife/internal"
)

type flagFlows []string

func (ff *flagFlows) String() string {
	out := ""
	for _, f := range *ff {
		out = fmt.Sprintf("-flow %s ", f)
	}

	return out
}

func (ff *flagFlows) Set(value string) error {
	*ff = append(*ff, value)
	return nil
}

var (
	flowFlags flagFlows
)

func main() {
	flag.Var(&flowFlags, "flows", "Define flows to run")
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
	app.Flows = parseFlowFlags(&flowFlags)

	err := app.Run(fmt.Sprintf("%s:%s", host, port), args[2:])
	if err != nil {
		fmt.Printf("error running app: %s\n", err)
	}
}

func parseFlowFlags(ff *flagFlows) []internal.Flow {
	var out []internal.Flow

	for _, f := range *ff {
		switch f {
		case "smtp-helo":
			out = append(out, flow.SMTPHelo{})
		case "smtp-auth":
			out = append(out, flow.SMTPAuth{})
		default:
			log.Printf("failed parsing flow %s", f)
		}
	}

	return out
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
