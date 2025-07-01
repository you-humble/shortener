package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) *Logger {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		Fatal("logger.New", Error(err))
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	l, err := cfg.Build()
	if err != nil {
		Fatal("logger.New", Error(err))
	}

	instance = l

	return l
}

func NewFile(level string) *Logger {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		Fatal("logger.New", Error(err))
	}

	file, err := os.OpenFile("./log/info.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		Fatal("logger.New", Error(err))
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(cfg)

	syncer := zapcore.AddSync(file)
	core := zapcore.NewCore(encoder, syncer, lvl)

	l := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	instance = l

	return l
}

func L() *zap.Logger {
	globalMu.RLock()
	l := instance
	globalMu.RUnlock()
	return l
}
