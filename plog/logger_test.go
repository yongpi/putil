package plog

import (
	"context"
	"errors"
	"testing"
)

func TestNewLogger(t *testing.T) {
	ctx := context.Background()
	WithContext(ctx).WithError(errors.New("err is")).Error("test print err")
	Error("test root log")
	Info("info msg")
	Warn("warn msg")
}
