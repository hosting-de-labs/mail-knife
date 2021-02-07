package internal

type Session interface {
	Close() error
	Send(string) error
}
