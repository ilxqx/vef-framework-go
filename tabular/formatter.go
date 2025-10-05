package tabular

import (
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/cast"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/null"
)

// defaultFormatter is the built-in formatter that handles common Go types.
type defaultFormatter struct {
	format string
}

// Format implements the Formatter interface for common Go types.
func (f *defaultFormatter) Format(value any) (string, error) {
	if value == nil {
		return constants.Empty, nil
	}

	// Handle null types first
	switch v := value.(type) {
	case null.String:
		return v.ValueOrZero(), nil
	case null.Int:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Int16:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Int32:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Float:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Bool:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Byte:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.DateTime:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Date:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Time:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()

	case null.Decimal:
		if !v.Valid {
			return constants.Empty, nil
		}

		value = v.ValueOrZero()
	}

	// Handle pointer types
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return constants.Empty, nil
		}

		value = rv.Elem().Interface()
	}

	if f.format != constants.Empty {
		switch v := value.(type) {
		case float32, float64:
			return fmt.Sprintf(f.format, v), nil
		case time.Time:
			return v.Format(f.format), nil
		case datetime.DateTime:
			return v.Format(f.format), nil
		case datetime.Date:
			return v.Format(f.format), nil
		case datetime.Time:
			return v.Format(f.format), nil
		}
	} else {
		switch v := value.(type) {
		case time.Time:
			return v.Format(time.DateTime), nil
		}
	}

	return cast.ToStringE(value)
}

// NewDefaultFormatter creates a default formatter with optional format template.
func NewDefaultFormatter(format string) *defaultFormatter {
	return &defaultFormatter{format: format}
}
