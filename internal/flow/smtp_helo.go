package flow

import (
	"net/textproto"
	"strings"
	"time"

	"github.com/hosting-de-labs/mail-knife/internal"
)

var (
	_ internal.Flow = SMTPHelo{}
)

type SMTPHelo struct{}

func (s SMTPHelo) Run(r *textproto.Reader, w *textproto.Writer) error {
	bannerFound := false
	for !bannerFound {
		l, err := r.ReadLine()
		if err != nil {
			return err
		}

		if strings.HasPrefix(l, "220 ") {
			bannerFound = true
		}
	}

	time.Sleep(500 * time.Millisecond)
	err := w.PrintfLine("EHLO host.name")
	if err != nil {
		return err
	}

	return nil
}
