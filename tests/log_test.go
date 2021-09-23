package tests

import (
	"errors"
	"testing"

	"github.com/yongpi/putil/plog"
)

func TestLog(t *testing.T) {
	plog.WithError(errors.New("error is new")).Error("print error")
	plog.Error("root print error")
	plog.Info("root info")
	plog.Warn("root warn")
	plog.Fatal("root fatal")
	print("fatal")
}
