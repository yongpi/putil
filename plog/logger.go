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

func (ep *EntryPool) PutBuffer(entry *Entry) {
	entry.Release()
	ep.pool.Put(entry)
}

var (
	bufferPool      *BufferPool
	entryPool       *EntryPool
	packageName     string
	packageSkip     int
	packageInitOnce sync.Once
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
}

type Logger struct {
	sync.Mutex
	Out      io.Writer
	Level    LevelType
	IsCaller bool
	Format   Formatter
	Hooks    Hooks
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
	return &Entry{logger: e.logger, Context: e.Context, Err: e.Err, Time: time}
}

func (e *Entry) Info(msg string) {
	panic("implement me")
}

func (e *Entry) Infof(format string, args ...interface{}) {
	panic("implement me")
}

func (e *Entry) Warn(msg string) {
	panic("implement me")
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	panic("implement me")
}

func (e *Entry) Error(msg string) {
	panic("implement me")
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	panic("implement me")
}

func (e *Entry) Debug(msg string) {
	panic("implement me")
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	panic("implement me")
}

func (e *Entry) Fatal(msg string) {
	panic("implement me")
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	panic("implement me")
}

func (e *Entry) Release() {
	e.logger = nil
	e.Context = nil
	e.Time = nil
	e.CallFrame = nil
	e.Err = nil
	e.Buffer = nil
}

func (e *Entry) Log(levelType LevelType, msg string) {
	if e.logger.Level <= levelType {
		return
	}

	newEntry := *e
	if newEntry.Time == nil {
		now := time.Now()
		newEntry.Time = &now
	}

	newEntry.logger.Lock()
	isCaller := newEntry.logger.IsCaller
	newEntry.logger.Unlock()

	if isCaller {
		newEntry.CallFrame = getCaller()
	}
	newEntry.Msg = msg

	// 执行钩子
	newEntry.logger.Hooks.HookOn(&newEntry)

	buffer := bufferPool.GetBuffer()
	defer func() {
		bufferPool.PutBuffer(buffer)
	}()
	newEntry.Buffer = buffer

	// 写日志
	newEntry.write()

	entryPool.PutBuffer(&newEntry)

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
