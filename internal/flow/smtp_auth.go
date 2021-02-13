package flow

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/hosting-de-labs/mail-knife/internal"
)

var (
	_ internal.Flow = SMTPAuth{}
)

type SMTPAuth struct{}

func (s SMTPAuth) Run(c *internal.Conn, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("run: insufficient arguments: username and password needed")
	}

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

	authMethods := make(map[string]interface{})
	for len(authMethods) == 0 {
		l, err := c.Reader.ReadLine()
		if err != nil {
			return err
		}

		if strings.HasPrefix(l, "250-AUTH") {
			if strings.Contains(l, "LOGIN") {
				authMethods["login"] = true
			}

			if strings.Contains(l, "PLAIN") {
				authMethods["plain"] = true
			}
		}
	}

	if _, ok := authMethods["plain"]; ok {
		err = c.Writer.PrintfLine("AUTH PLAIN")
		if err != nil {
			return err
		}

		credString := fmt.Sprintf("%s:%s", args[0], args[1])
		credEnc := base64.StdEncoding.EncodeToString([]byte(credString))

		err = c.Writer.PrintfLine(credEnc)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported authentication methods")
	}

	return nil
}
