package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
)

const (
	LevelDebug = zap.DebugLevel
	LevelInfo  = zap.InfoLevel
	LevelWarn  = zap.WarnLevel
	LevelError = zap.ErrorLevel
	LevelFatal = zap.FatalLevel
)

var (
	String   = zap.String
	Int      = zap.Int
	Duration = zap.Duration
	Error    = zap.Error
)

type (
	Logger = zap.Logger
	Field  = zap.Field
)

var (
	globalMu sync.RWMutex = sync.RWMutex{}
	instance *Logger      = zap.NewNop()
)

func Fatal(msg string, fields ...Field) {
	L().Fatal(msg, fields...)
	os.Exit(1)
}

func ErrorS(val string) Field {
	return String("error", val)
}
