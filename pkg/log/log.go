package log

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate sh -c "go run ./generator >zap_sugar_generated.go"

// alias
type (
	Logger        = zap.Logger
	SugaredLogger = zap.SugaredLogger
)

var (
	_globalMu sync.RWMutex
	_globalL  = zap.NewNop()
	_globalS  = _globalL.Sugar()
)

func SetLogger(l *Logger) {
	_globalMu.Lock()
	_globalL = l
	_globalS = _globalL.Sugar()
	_globalMu.Unlock()
}

func GetLogger() *Logger {
	_globalMu.RLock()
	l := _globalL
	_globalMu.RUnlock()
	return l
}

func NewLogger(lv zapcore.Level) *Logger {
	var zapConfig zap.Config

	switch lv {
	case zapcore.DebugLevel:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(lv)
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1), // AddCallerSkip because we made a layer of wrapper
	)
	if err != nil {
		Panic(err)
	}

	return logger
}
