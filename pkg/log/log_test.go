package log

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestLog(t *testing.T) {
	SetLogger(NewLogger(zapcore.DebugLevel))
	Debug("debug")
	Info("info")
}
