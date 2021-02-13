package flow

import (
	"strings"
	"time"

	"github.com/hosting-de-labs/mail-knife/internal"
)

var (
	_ internal.Flow = SMTPHelo{}
)

type SMTPHelo struct{}

func (s SMTPHelo) Run(c *internal.Conn, _ []string) error {
	bannerFound := false
	for !bannerFound {
		l, err := c.Reader.ReadLine()
		if err != nil {
			return err
		}

		if strings.HasPrefix(l, "220 ") {
			bannerFound = true
		}
	}

	time.Sleep(500 * time.Millisecond)
	err := c.Writer.PrintfLine("EHLO host.name")
	if err != nil {
		return err
	}

	return nil
}
