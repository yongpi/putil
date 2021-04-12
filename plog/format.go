package plog

import (
	"fmt"
	"time"
)

var LevelFormats = map[LevelType]string{
	FATAL: "F",
	ERROR: "E",
	WARN:  "W",
	INFO:  "I",
	DEBUG: "D",
}

const (
	DefaultTimeFormat = "2006-01-02 15:04:05.000000000"
)

type Formatter interface {
	Format(*Entry) ([]byte, error)
}

type DefaultFormatter struct {
}

func (f *DefaultFormatter) Format(entry *Entry) ([]byte, error) {
	if entry.Time == nil {
		now := time.Now()
		entry.Time = &now
	}

	if entry.Buffer == nil {
		entry.Buffer = bufferPool.GetBuffer()
	}

	ls := LevelFormats[entry.Level]
	ts := entry.Time.Format(DefaultTimeFormat)
	entry.Buffer.WriteString(fmt.Sprintf("[%s %s", ls, ts))
	if entry.CallFrame != nil {
		entry.Buffer.WriteString(fmt.Sprintf(" %s.%d", entry.CallFrame.File, entry.CallFrame.Line))
	}
	entry.Buffer.WriteString(fmt.Sprintf("]"))
	entry.Buffer.WriteString(fmt.Sprintf(" %s", entry.Msg))
	if entry.Err != nil {
		entry.Buffer.WriteString(fmt.Sprintf(", err = %s", entry.Err.Error()))
	}
	entry.Buffer.WriteByte('\n')

	return entry.Buffer.Bytes(), nil
}
