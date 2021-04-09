package plog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type BufferPool struct {
	pool sync.Pool
}

type EntryPool struct {
	pool sync.Pool
}

func (bp *BufferPool) GetBuffer() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func (bp *BufferPool) PutBuffer(buffer *bytes.Buffer) {
	buffer.Reset()
	bp.pool.Put(buffer)
}

func (ep *EntryPool) GetEntry() *Entry {
	return ep.pool.Get().(*Entry)
}

func (ep *EntryPool) PutEntry(entry *Entry) {
	entry.Release()
	ep.pool.Put(entry)
}

var (
	bufferPool      *BufferPool
	entryPool       *EntryPool
	packageName     string
	packageSkip     int
	packageInitOnce sync.Once
	root            *Logger
)

const (
	minSkip = 4
	maxSkip = 25
)

func init() {
	bufferPool = &BufferPool{sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}}

	entryPool = &EntryPool{sync.Pool{
		New: func() interface{} {
			return new(Entry)
		},
	}}
	root = NewLogger(INFO)
}

type Logger struct {
	sync.Mutex
	Out        io.Writer
	Level      LevelType
	IsCaller   bool
	Format     Formatter
	Hooks      Hooks
	AutoCaller bool
}

func NewLogger(levelType LevelType) *Logger {
	return &Logger{
		Out:        os.Stderr,
		Level:      levelType,
		Format:     new(DefaultFormatter),
		AutoCaller: true,
	}
}

func (logger *Logger) newEntry() *Entry {
	entry := entryPool.GetEntry()
	entry.logger = logger
	return entry
}

func (logger *Logger) WithContext(ctx context.Context) *Entry {
	entry := logger.newEntry()
	defer entryPool.PutEntry(entry)
	return entry.WithContext(ctx)
}

func (logger *Logger) WithError(err error) *Entry {
	entry := logger.newEntry()
	defer entryPool.PutEntry(entry)
	return entry.WithError(err)
}

func (logger *Logger) WithTime(time time.Time) *Entry {
	entry := logger.newEntry()
	defer entryPool.PutEntry(entry)
	return entry.WithTime(time)
}

func (logger *Logger) logf(levelType LevelType, format string, args ...interface{}) {
	if levelType > logger.Level {
		return
	}
	entry := logger.newEntry()
	entry.Logf(levelType, format, args...)
	entryPool.PutEntry(entry)
}

func (logger *Logger) Info(msg string) {
	logger.logf(INFO, msg)
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.logf(INFO, format, args...)
}

func (logger *Logger) Warn(msg string) {
	logger.logf(WARN, msg)
}

func (logger *Logger) Warnf(format string, args ...interface{}) {
	logger.logf(WARN, format, args...)
}

func (logger *Logger) Error(msg string) {
	logger.logf(ERROR, msg)
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.logf(ERROR, format, args...)
}

func (logger *Logger) Debug(msg string) {
	logger.logf(DEBUG, msg)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.logf(DEBUG, format, args...)
}

func (logger *Logger) Fatal(msg string) {
	logger.logf(FATAL, msg)
	os.Exit(1)

}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.logf(FATAL, format, args...)
	os.Exit(1)
}

type Entry struct {
	logger    *Logger
	Level     LevelType
	Context   context.Context
	Err       error
	CallFrame *runtime.Frame
	Time      *time.Time
	Msg       string
	Buffer    *bytes.Buffer
}

func (e *Entry) WithContext(ctx context.Context) *Entry {
	return &Entry{logger: e.logger, Context: ctx, Err: e.Err, Time: e.Time}
}

func (e *Entry) WithError(err error) *Entry {
	return &Entry{logger: e.logger, Context: e.Context, Err: err, Time: e.Time}
}

func (e *Entry) WithTime(time time.Time) *Entry {
	return &Entry{logger: e.logger, Context: e.Context, Err: e.Err, Time: &time}
}

func (e *Entry) Info(msg string) {
	e.Logf(INFO, msg)
}

func (e *Entry) Infof(format string, args ...interface{}) {
	e.Logf(INFO, format, args...)
}

func (e *Entry) Warn(msg string) {
	e.Logf(WARN, msg)
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	e.Logf(WARN, format, args...)
}

func (e *Entry) Error(msg string) {
	e.Logf(ERROR, msg)
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.Logf(ERROR, format, args...)
}

func (e *Entry) Debug(msg string) {
	e.Logf(DEBUG, msg)
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.Logf(DEBUG, format, args...)
}

func (e *Entry) Fatal(msg string) {
	e.Logf(FATAL, msg)
	os.Exit(1)
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.Logf(FATAL, format, args...)
	os.Exit(1)
}

func (e *Entry) Release() {
	e.logger = nil
	e.Context = nil
	e.Time = nil
	e.CallFrame = nil
	e.Err = nil
	e.Buffer = nil
}

func (e *Entry) Logf(levelType LevelType, format string, args ...interface{}) {
	if e.logger.Level < levelType {
		return
	}

	newEntry := *e
	if newEntry.Time == nil {
		now := time.Now()
		newEntry.Time = &now
	}
	newEntry.Level = levelType

	newEntry.logger.Lock()
	isCaller := newEntry.logger.IsCaller
	autoCaller := newEntry.logger.AutoCaller
	newEntry.logger.Unlock()
	if autoCaller && levelType <= ERROR {
		isCaller = true
	}
	if isCaller {
		newEntry.CallFrame = getCaller()
	}
	newEntry.Msg = fmt.Sprintf(format, args...)

	// 执行钩子
	newEntry.logger.Hooks.HookOn(&newEntry)

	buffer := bufferPool.GetBuffer()
	defer func() {
		bufferPool.PutBuffer(buffer)
	}()
	newEntry.Buffer = buffer

	// 写日志
	newEntry.write()
}

func (e *Entry) write() {
	lc, err := e.logger.Format.Format(e)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "plog：logger format 出现错误， err = %v", err)
		return
	}
	e.logger.Lock()
	defer e.logger.Unlock()
	if _, err := e.logger.Out.Write(lc); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "plog：logger io write 出现错误， err = %v", err)
		return
	}
}

func getCaller() *runtime.Frame {
	packageInitOnce.Do(func() {
		pcs := make([]uintptr, maxSkip)
		runtime.Callers(0, pcs)
		for i := 0; i < maxSkip; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				packageName = getPackageName(funcName)
				break
			}
		}
		packageSkip = minSkip
	})

	pcs := make([]uintptr, maxSkip)
	n := runtime.Callers(packageSkip, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if getPackageName(frame.Function) != packageName {
			return &frame
		}
	}

	return nil
}

func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

func WithContext(ctx context.Context) *Entry {
	entry := root.newEntry()
	defer entryPool.PutEntry(entry)
	return entry.WithContext(ctx)
}

func WithError(err error) *Entry {
	entry := root.newEntry()
	defer entryPool.PutEntry(entry)
	return entry.WithError(err)
}

func WithTime(time time.Time) *Entry {
	entry := root.newEntry()
	defer entryPool.PutEntry(entry)
	return entry.WithTime(time)
}

func Info(msg string) {
	root.logf(INFO, msg)
}

func Infof(format string, args ...interface{}) {
	root.logf(INFO, format, args...)
}

func Warn(msg string) {
	root.logf(WARN, msg)
}

func Warnf(format string, args ...interface{}) {
	root.logf(WARN, format, args...)
}

func Error(msg string) {
	root.logf(ERROR, msg)
}

func Errorf(format string, args ...interface{}) {
	root.logf(ERROR, format, args...)
}

func Debug(msg string) {
	root.logf(DEBUG, msg)
}

func Debugf(format string, args ...interface{}) {
	root.logf(DEBUG, format, args...)
}

func Fatal(msg string) {
	root.logf(FATAL, msg)
	os.Exit(1)

}

func Fatalf(format string, args ...interface{}) {
	root.logf(FATAL, format, args...)
	os.Exit(1)
}

func InjectHook(levelType LevelType, hook Hook) {
	root.Hooks[levelType] = append(root.Hooks[levelType], hook)
}
