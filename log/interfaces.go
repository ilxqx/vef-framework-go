package log

// Logger defines the core logging interface for structured logging across the framework.
type Logger interface {
	// Named creates a child logger with the given namespace.
	Named(name string) Logger
	// WithCallerSkip adjusts the number of stack frames to skip when reporting caller location.
	WithCallerSkip(skip int) Logger
	// Enabled checks whether the given log level is enabled.
	Enabled(level Level) bool
	// Sync flushes any buffered log entries.
	Sync()
	// Debug logs a message at Debug level.
	Debug(message string)
	// Debugf logs a formatted message at Debug level.
	Debugf(template string, args ...any)
	// Info logs a message at Info level.
	Info(message string)
	// Infof logs a formatted message at Info level.
	Infof(template string, args ...any)
	// Warn logs a message at Warn level.
	Warn(message string)
	// Warnf logs a formatted message at Warn level.
	Warnf(template string, args ...any)
	// Error logs a message at Error level.
	Error(message string)
	// Errorf logs a formatted message at Error level.
	Errorf(template string, args ...any)
	// Panic logs a message at Panic level and then panics.
	Panic(message string)
	// Panicf logs a formatted message at Panic level and then panics.
	Panicf(template string, args ...any)
}

// LoggerConfigurable defines an interface for components that can be configured with a logger.
type LoggerConfigurable[T any] interface {
	// WithLogger sets the logger for the component and returns the configured instance.
	WithLogger(logger Logger) T
}
