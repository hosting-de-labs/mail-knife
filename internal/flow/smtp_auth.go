package flow

import (
	"fmt"
	"net"
	"net/smtp"

	"github.com/hosting-de-labs/mail-knife/internal"
)

var (
	_ internal.Flow = SMTPAuth{}
)

type SMTPAuth struct{}

func (s SMTPAuth) Run(addr string, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("insufficient arguments: username and password needed")
	}

	hostname, _, _ := net.SplitHostPort(addr)
	if len(hostname) == 0 {
		return fmt.Errorf("run: len(hostname) == 0")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {

	}

	c, err := smtp.NewClient(conn, hostname)
	if err != nil {
		return fmt.Errorf("run: smtp.NewClient: %s", err)
	}

	err = c.Hello("host.name")
	if err != nil {
		return fmt.Errorf("run: hello: %s", err)
	}

	auth := smtp.PlainAuth("", args[0], args[1], hostname)
	err = c.Auth(auth)
	if err != nil {
		return fmt.Errorf("run: auth: %s", err)
	}

	return nil
}
