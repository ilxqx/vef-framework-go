package datetime

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	"github.com/ilxqx/vef-framework-go/constants"
)

var (
	// Layout constants for different time formats.
	dateTimeLayout = time.DateTime // "2006-01-02 15:04:05"
	dateLayout     = time.DateOnly // "2006-01-02"
	timeLayout     = time.TimeOnly // "15:04:05"

	// Pattern length constants for efficient JSON processing.
	dateTimePatternLength = len(time.DateTime)
	datePatternLength     = len(time.DateOnly)
	timePatternLength     = len(time.TimeOnly)
)

// scanTimeValue is a generic helper function for scanning time values from database sources.
// It handles various input types including []byte, string, time.Time and their pointer variants.
func scanTimeValue(src any, parseString func(string) (any, error), convertTime func(time.Time) any, typeName string, dest any) error {
	switch value := src.(type) {
	case []byte:
		parsed, err := parseString(string(value))
		if err != nil {
			return err
		}

		return assignValue(dest, parsed)

	case *[]byte:
		if value == nil {
			return nil
		}

		parsed, err := parseString(string(*value))
		if err != nil {
			return err
		}

		return assignValue(dest, parsed)

	case string:
		parsed, err := parseString(value)
		if err != nil {
			return err
		}

		return assignValue(dest, parsed)

	case *string:
		if value == nil {
			return nil
		}

		parsed, err := parseString(*value)
		if err != nil {
			return err
		}

		return assignValue(dest, parsed)

	case time.Time:
		converted := convertTime(value)

		return assignValue(dest, converted)
	case *time.Time:
		if value == nil {
			return nil
		}

		converted := convertTime(*value)

		return assignValue(dest, converted)

	default:
		// Try using cast library as fallback for other types
		if str, err := cast.ToStringE(src); err == nil {
			parsed, err := parseString(str)
			if err != nil {
				return err
			}

			return assignValue(dest, parsed)
		}

		return fmt.Errorf("%w: %s value: %v", ErrFailedScan, typeName, src)
	}
}

// assignValue assigns the parsed value to the destination pointer using type assertion.
func assignValue(dest, value any) error {
	switch d := dest.(type) {
	case *DateTime:
		*d = value.(DateTime)
	case *Date:
		*d = value.(Date)
	case *Time:
		*d = value.(Time)
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedDestType, dest)
	}

	return nil
}

// parseTimeWithFallback provides a standardized way to parse time strings with fallback support.
// It first tries the provided layout, then falls back to the cast library for common formats.
func parseTimeWithFallback(value, layout string) (time.Time, error) {
	// Primary: try with specified layout (use Local timezone for DateTime parsing)
	parsed, err := time.ParseInLocation(layout, value, time.Local)
	if err == nil {
		return parsed, nil
	}

	// Fallback: try cast library for common time formats
	if castTime, castErr := cast.ToTimeE(value); castErr == nil {
		return castTime, nil
	}

	// Return original error if both methods fail
	return time.Time{}, err
}

// validateJSONFormat checks if the JSON bytes have the expected format for time types.
func validateJSONFormat(bs []byte, expectedLength int) error {
	if len(bs) != expectedLength+2 || bs[0] != constants.JSONQuote || bs[len(bs)-1] != constants.JSONQuote {
		return fmt.Errorf("%w: expected length %d with quotes", ErrInvalidJSONFormat, expectedLength)
	}

	return nil
}
