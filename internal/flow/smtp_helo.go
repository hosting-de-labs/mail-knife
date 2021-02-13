package flow

import (
	"fmt"
	"time"

	"github.com/hosting-de-labs/mail-knife/internal"
)

var (
	_ internal.Flow = SMTPHelo{}
)

type SMTPHelo struct{}

func (s SMTPHelo) Run(c *internal.Conn, _ []string) error {
	_, err := c.WaitMessage("220 ", 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp-helo: %s", err)
	}

	time.Sleep(500 * time.Millisecond)
	err = c.PrintfLine("EHLO host.name")
	if err != nil {
		return err
	}

	return nil
}
