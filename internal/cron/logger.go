package cron

import (
	ilog "github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/log"
)

// cronLogger implements gocron.Logger interface to integrate with the framework's logging system.
// It adapts the framework's logger to the gocron logger interface.
type cronLogger struct {
	logger log.Logger
}

func (l *cronLogger) Debug(msg string, args ...any) {
	l.logger.Debugf(msg, args...)
}

func (l *cronLogger) Error(msg string, args ...any) {
	l.logger.Errorf(msg, args...)
}

func (l *cronLogger) Info(msg string, args ...any) {
	l.logger.Infof(msg, args...)
}

func (l *cronLogger) Warn(msg string, args ...any) {
	l.logger.Warnf(msg, args...)
}

func newCronLogger() *cronLogger {
	return &cronLogger{
		logger: ilog.Named("cron"),
	}
}
