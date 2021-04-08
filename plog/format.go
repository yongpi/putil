package plog

type Formatter interface {
	Format(*Entry) ([]byte, error)
}
