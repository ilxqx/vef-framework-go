package log

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/ilxqx/vef-framework-go/constants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newZapLogger(level zapcore.Level) *zap.SugaredLogger {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "message",
			StacktraceKey: zapcore.OmitKey,
			CallerKey:     zapcore.OmitKey,
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalColorLevelEncoder,
			EncodeTime: func() zapcore.TimeEncoder {
				layout := time.DateOnly + "T" + time.TimeOnly + ".000"
				return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
					enc.AppendString(
						color.CyanString(t.Format(layout)),
					)
				}
			}(),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeName: func(name string, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(constants.LeftBracket + color.HiGreenString(name) + constants.RightBracket)
			},
		},
		DisableStacktrace: true,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	// Build creates the logger without caller information
	logger, err := config.Build(zap.WithCaller(false))
	if err != nil {
		// Panic if logger creation fails
		panic(
			fmt.Errorf("failed to build zap logger: %w", err),
		)
	}

	// Sugar returns a sugared logger for easier usage
	return logger.Sugar()
}
