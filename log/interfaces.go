package log

type Logger interface {
	Named(name string) Logger
	WithCallerSkip(skip int) Logger
	Enabled(level Level) bool
	Sync()
	Debug(message string)
	Debugf(template string, args ...any)
	Info(message string)
	Infof(template string, args ...any)
	Warn(message string)
	Warnf(template string, args ...any)
	Error(message string)
	Errorf(template string, args ...any)
	Panic(message string)
	Panicf(template string, args ...any)
}
