package log

import (
	"go.uber.org/zap"

	"github.com/ilxqx/vef-framework-go/log"
)

type zapLogger struct {
	logger *zap.SugaredLogger
}

func (l *zapLogger) Named(name string) log.Logger {
	return &zapLogger{
		logger: l.logger.Named(name),
	}
}

func (l *zapLogger) WithCallerSkip(skip int) log.Logger {
	return &zapLogger{
		logger: l.logger.WithOptions(zap.AddCallerSkip(skip)),
	}
}

func (l *zapLogger) Enabled(level log.Level) bool {
	switch level {
	case log.LevelDebug:
		return l.logger.Level().Enabled(zap.DebugLevel)
	case log.LevelInfo:
		return l.logger.Level().Enabled(zap.InfoLevel)
	case log.LevelWarn:
		return l.logger.Level().Enabled(zap.WarnLevel)
	case log.LevelError:
		return l.logger.Level().Enabled(zap.ErrorLevel)
	case log.LevelPanic:
		return l.logger.Level().Enabled(zap.PanicLevel)
	}

	return false
}

func (l *zapLogger) Sync() {
	if err := l.logger.Sync(); err != nil {
		l.Errorf("error occurred while flushing logger: %v", err)
	}
}

func (l *zapLogger) Debug(message string) {
	l.logger.Debug(message)
}

func (l *zapLogger) Debugf(template string, args ...any) {
	l.logger.Debugf(template, args...)
}

func (l *zapLogger) Info(message string) {
	l.logger.Info(message)
}

func (l *zapLogger) Infof(template string, args ...any) {
	l.logger.Infof(template, args...)
}

func (l *zapLogger) Warn(message string) {
	l.logger.Warn(message)
}

func (l *zapLogger) Warnf(template string, args ...any) {
	l.logger.Warnf(template, args...)
}

func (l *zapLogger) Error(message string) {
	l.logger.Error(message)
}

func (l *zapLogger) Errorf(template string, args ...any) {
	l.logger.Errorf(template, args...)
}

func (l *zapLogger) Panic(message string) {
	l.logger.Panic(message)
}

func (l *zapLogger) Panicf(template string, args ...any) {
	l.logger.Panicf(template, args...)
}
