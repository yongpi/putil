package plog

type LevelType int

const (
	Fatal LevelType = iota
	ERROR
	WARN
	INFO
	DEBUG
)

type Log interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})
}