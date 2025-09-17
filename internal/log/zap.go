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
		Level:       zap.NewAtomicLevelAt(level), // Level sets the minimum log level
		Development: false,                       // Development is false for production logging
		Encoding:    "console",                   // Encoding uses console format for human readability
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "time",                           // TimeKey is the key for timestamp field
			LevelKey:      "level",                          // LevelKey is the key for log level field
			NameKey:       "logger",                         // NameKey is the key for logger name field
			FunctionKey:   zapcore.OmitKey,                  // FunctionKey is omitted to reduce noise
			MessageKey:    "message",                        // MessageKey is the key for log message field
			StacktraceKey: zapcore.OmitKey,                  // StacktraceKey is omitted by default
			CallerKey:     zapcore.OmitKey,                  // CallerKey is omitted by default
			LineEnding:    zapcore.DefaultLineEnding,        // LineEnding uses default line ending
			EncodeLevel:   zapcore.CapitalColorLevelEncoder, // EncodeLevel uses colored level encoding
			EncodeTime: func() zapcore.TimeEncoder { // EncodeTime creates a custom time encoder
				layout := time.DateOnly + "T" + time.TimeOnly + ".000" // layout defines the time format
				return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
					enc.AppendString(
						color.CyanString(t.Format(layout)), // color.CyanString makes time cyan colored
					)
				}
			}(),
			EncodeDuration: zapcore.StringDurationEncoder, // EncodeDuration uses string format for durations
			EncodeName: func(name string, enc zapcore.PrimitiveArrayEncoder) { // EncodeName creates a custom name encoder
				enc.AppendString(constants.LeftBracket + color.HiGreenString(name) + constants.RightBracket) // color.HiGreenString makes logger name bright green
			},
		},
		DisableStacktrace: true,               // DisableStacktrace disables stack traces by default
		OutputPaths:       []string{"stdout"}, // OutputPaths sends logs to stdout
		ErrorOutputPaths:  []string{"stderr"}, // ErrorOutputPaths sends errors to stderr
	}

	// Build creates the logger with caller information
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
