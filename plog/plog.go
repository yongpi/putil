package plog

type LevelType int

const (
	FATAL LevelType = iota
	ERROR
	WARN
	INFO
	DEBUG
)

var LevelTypeMap = map[string]LevelType{
	"fatal": FATAL,
	"error": ERROR,
	"warn":  WARN,
	"info":  INFO,
	"debug": DEBUG,
}

func LevelTypeFromString(levelTypeStr string) LevelType {
	levelType, ok := LevelTypeMap[levelTypeStr]
	if !ok {
		levelType = INFO
	}
	return levelType
}

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
