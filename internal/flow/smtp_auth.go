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
		return fmt.Errorf("insufficient arguments: username and password needed")
	}

	helo := SMTPHelo{}
	err := helo.Run(c, []string{})
	if err != nil {
		return err
	}

	msg := "250-AUTH"
	l, err := c.WaitMessage(msg, 1*time.Second)
	if err != nil {
		return fmt.Errorf("failed waiting for message %q: %s", msg, err)
	}

	authMethods := parseAuthMethods(l)
	if _, ok := authMethods["plain"]; ok {
		err = c.PrintfLine("AUTH PLAIN")
		if err != nil {
			return err
		}

		credString := fmt.Sprintf("%s:%s", args[0], args[1])
		credEnc := base64.StdEncoding.EncodeToString([]byte(credString))

		err = c.PrintfLine(credEnc)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported authentication methods")
	}

	return nil
}

func parseAuthMethods(l string) map[string]bool {
	authMethods := make(map[string]bool)
	if strings.Contains(l, "LOGIN") {
		authMethods["login"] = true
	}

	if strings.Contains(l, "PLAIN") {
		authMethods["plain"] = true
	}
	return authMethods
}
