package internal

import (
	"net/textproto"
)

type Flow interface {
	Run(addr string, args []string) error
}

type FlowLogger interface {
	SetProtocolLoggers(receive textproto.Writer, send textproto.Writer)
}
