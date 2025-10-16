package golater

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
}

type ZapLogger struct {
	*zap.Logger
}

func NewLogger(level, output string) (*ZapLogger, error) {
	var lvl zapcore.Level
	if err := lvl.Set(level); err != nil {
		lvl = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.Encoding = "json"
	// cfg.OutputPaths = []string{output}

	l, err := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &ZapLogger{l}, nil
}

func (z *ZapLogger) With(fields ...zap.Field) Logger {
	return &ZapLogger{z.Logger.With(fields...)}
}

func ReplaceGlobals(l Logger) {
	if zl, ok := l.(*ZapLogger); ok {
		zap.ReplaceGlobals(zl.Logger)
	}
}
