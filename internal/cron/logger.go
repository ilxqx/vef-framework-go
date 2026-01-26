package cron

// cronLogger implements gocron.Logger interface to integrate with the framework's logging system.
type cronLogger struct{}

func (l *cronLogger) Debug(msg string, args ...any) {
	logger.Debugf(msg, args...)
}

func (l *cronLogger) Error(msg string, args ...any) {
	logger.Errorf(msg, args...)
}

func (l *cronLogger) Info(msg string, args ...any) {
	logger.Infof(msg, args...)
}

func (l *cronLogger) Warn(msg string, args ...any) {
	logger.Warnf(msg, args...)
}

func newCronLogger() *cronLogger {
	return &cronLogger{}
}
