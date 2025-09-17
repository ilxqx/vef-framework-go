package log

import (
	"os"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"go.uber.org/zap"
)

var (
	logger = newLogger() // logger is the default logger instance
)

func Named(name string) log.Logger {
	return logger.Named(name)
}

func newLogger() *zapLogger {
	level := zap.InfoLevel                                           // level is the default log level
	levelString := strings.ToLower(os.Getenv(constants.EnvLogLevel)) // levelString gets log level from environment
	switch levelString {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	return &zapLogger{
		logger: newZapLogger(level).WithOptions(zap.AddCallerSkip(1)), // logger with caller skip for wrapper
	}
}
