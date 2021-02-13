package internal

type Flow interface {
	Run(c *Conn, args []string) error
}
