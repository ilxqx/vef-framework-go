package log

import (
	"os"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"go.uber.org/zap"
)

var (
	logger = newLogger()
)

func Named(name string) log.Logger {
	return logger.Named(name)
}

func newLogger() *zapLogger {
	// level is the default log level
	level := zap.InfoLevel
	// levelString gets log level from environment
	levelString := strings.ToLower(os.Getenv(constants.EnvLogLevel))
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
		logger: newZapLogger(level).WithOptions(zap.AddCallerSkip(1)),
	}
}
