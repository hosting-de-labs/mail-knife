package internal

import (
	"net/textproto"
)

type Flow interface {
	Run(r *textproto.Reader, w *textproto.Writer) error
}
