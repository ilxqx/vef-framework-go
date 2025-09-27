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
