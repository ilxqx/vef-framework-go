package log

import (
	logPkg "github.com/ilxqx/vef-framework-go/log"
	"go.uber.org/zap"
)

type zapLogger struct {
	logger *zap.SugaredLogger
}

func (l *zapLogger) Named(name string) logPkg.Logger {
	return &zapLogger{
		logger: l.logger.Named(name),
	}
}

func (l *zapLogger) WithCallerSkip(skip int) logPkg.Logger {
	return &zapLogger{
		logger: l.logger.WithOptions(zap.AddCallerSkip(skip)),
	}
}

func (l *zapLogger) Enabled(level logPkg.Level) bool {
	switch level {
	case logPkg.LevelDebug:
		return l.logger.Level().Enabled(zap.DebugLevel)
	case logPkg.LevelInfo:
		return l.logger.Level().Enabled(zap.InfoLevel)
	case logPkg.LevelWarn:
		return l.logger.Level().Enabled(zap.WarnLevel)
	case logPkg.LevelError:
		return l.logger.Level().Enabled(zap.ErrorLevel)
	case logPkg.LevelPanic:
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
