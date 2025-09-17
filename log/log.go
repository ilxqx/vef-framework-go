package log

// A Level is a logging priority. Higher levels are more important.
type Level int8

const (
	// LevelDebug logs are typically voluminous, and are usually disabled in
	// production.
	LevelDebug Level = iota + 1
	// LevelInfo is the default logging priority.
	LevelInfo
	// LevelWarn logs are more important than Info, but don't need individual
	// human review.
	LevelWarn
	// LevelError logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	LevelError
	// LevelPanic logs a message, then panics.
	LevelPanic
)

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

type LoggerConfigurable[T any] interface {
	WithLogger(logger Logger) T
}
