package log

import (
	"context"
	"log/slog"
	"strings"

	"github.com/spf13/cast"

	"github.com/ilxqx/vef-framework-go/constants"
	logPkg "github.com/ilxqx/vef-framework-go/log"
)

type sLogHandler struct {
	logger      logPkg.Logger // logger is the underlying logger instance
	attrs       []slog.Attr   // attrs contains the log attributes
	levelFilter logPkg.Level  // levelFilter indicates the minimum log level
}

func (s sLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	zapLevel := logPkg.LevelInfo
	switch {
	case level >= slog.LevelError:
		zapLevel = logPkg.LevelError
	case level >= slog.LevelWarn:
		zapLevel = logPkg.LevelWarn
	case level >= slog.LevelInfo:
		zapLevel = logPkg.LevelInfo
	case level >= slog.LevelDebug:
		zapLevel = logPkg.LevelDebug
	}

	return s.logger.Enabled(zapLevel) && zapLevel >= s.levelFilter
}

func (s sLogHandler) Handle(ctx context.Context, record slog.Record) error {
	fields := make([]string, 0, record.NumAttrs()+len(s.attrs))

	record.Attrs(func(attr slog.Attr) bool {
		if field := formatAttr(attr); field != constants.Empty {
			fields = append(fields, field)
		}

		return true
	})

	// fieldsValue joins all fields with separator
	fieldsValue := strings.Join(fields, " | ")
	if len(fields) > 0 {
		// fieldsValue adds prefix separator if fields exist
		fieldsValue = " | " + fieldsValue
	}

	level := record.Level
	switch level {
	case slog.LevelDebug:
		s.logger.Debug(record.Message + fieldsValue)
	case slog.LevelInfo:
		s.logger.Info(record.Message + fieldsValue)
	case slog.LevelWarn:
		s.logger.Warn(record.Message + fieldsValue)
	case slog.LevelError:
		s.logger.Error(record.Message + fieldsValue)
	}

	return nil
}

func (s sLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handler := &sLogHandler{
		logger: s.logger,
		attrs:  append(s.attrs, attrs...),
	}

	return handler
}

func (s sLogHandler) WithGroup(name string) slog.Handler {
	handler := &sLogHandler{
		logger: s.logger.Named(name),
		attrs:  s.attrs,
	}

	return handler
}

func formatAttr(attr slog.Attr) string {
	switch attr.Value.Kind() {
	case slog.KindString:
		return attr.Key + constants.ColonSpace + attr.Value.String()
	case slog.KindInt64:
		return attr.Key + constants.ColonSpace + cast.ToString(attr.Value.Int64())
	case slog.KindUint64:
		return attr.Key + constants.ColonSpace + cast.ToString(attr.Value.Uint64())
	case slog.KindFloat64:
		return attr.Key + constants.ColonSpace + cast.ToString(attr.Value.Float64())
	case slog.KindBool:
		return attr.Key + constants.ColonSpace + cast.ToString(attr.Value.Bool())
	case slog.KindDuration:
		return attr.Key + constants.ColonSpace + cast.ToString(attr.Value.Duration())
	case slog.KindTime:
		return attr.Key + constants.ColonSpace + cast.ToString(attr.Value.Time())
	case slog.KindAny:
		return attr.Key + constants.ColonSpace + cast.ToString(attr.Value.Any())
	default:
		return constants.Empty
	}
}

func NewSLogHandler(name string, callerSkip int, levelFilter ...logPkg.Level) slog.Handler {
	level := logPkg.LevelInfo
	if len(levelFilter) > 0 {
		level = levelFilter[0]
	}

	return &sLogHandler{
		logger:      Named(name).WithCallerSkip(callerSkip),
		levelFilter: level,
	}
}

func NewSLogger(name string, callerSkip int, levelFilter ...logPkg.Level) *slog.Logger {
	return slog.New(NewSLogHandler(name, callerSkip, levelFilter...))
}
