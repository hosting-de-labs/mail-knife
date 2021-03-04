package flow

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/hosting-de-labs/mail-knife/internal"
)

var (
	_ internal.Flow = SMTPHelo{}
)

type SMTPHelo struct{}

func (s SMTPHelo) Run(addr string, _ []string) error {
	c, err := internal.Dial(addr, func() {})
	if err != nil {
		return err
	}

	msg := "220 "
	_, err = c.WaitMessage(msg, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed waiting for message %q: %s", msg, err)
	}

	time.Sleep(500 * time.Millisecond)
	smtpC := smtp.Client{Text: c.Text}
	err = smtpC.Hello("host.name")
	if err != nil {
		return err
	}

	return nil
}
