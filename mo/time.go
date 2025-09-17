package mo

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/utils/v2"
	"github.com/spf13/cast"
)

var (
	dateTimePatternLength = len(time.DateTime)
	datePatternLength     = len(time.DateOnly)
	timePatternLength     = len(time.TimeOnly)
	zeroDate              = DateOf(time.Time{})
	zeroTime              = TimeOf(time.Time{})
	zeroDateTime          = DateTimeOf(time.Time{})
)

// scanTimeValue is a generic helper function for scanning time values.
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
		return fmt.Errorf("failed to scan %s value: %v", typeName, src)
	}
}

// assignValue assigns the value to the destination pointer using reflection.
func assignValue(dest, value any) error {
	switch d := dest.(type) {
	case *DateTime:
		*d = value.(DateTime)
	case *Date:
		*d = value.(Date)
	case *Time:
		*d = value.(Time)
	default:
		return fmt.Errorf("unsupported destination type: %T", dest)
	}
	return nil
}

// DateTime represents a date and time value with database and JSON support.
// It uses the standard time.DateTime format (2006-01-02 15:04:05).
type DateTime time.Time

func (dt DateTime) Unwrap() time.Time {
	return time.Time(dt)
}

func (dt *DateTime) Scan(src any) error {
	return scanTimeValue(src, func(s string) (any, error) {
		return ParseDateTime(s)
	}, func(t time.Time) any {
		return DateTime(t)
	}, "datetime", dt)
}

func (dt DateTime) Value() (driver.Value, error) {
	return dt.String(), nil
}

func (dt DateTime) String() string {
	return time.Time(dt).Format(time.DateTime)
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	bs := make([]byte, 0, dateTimePatternLength+2)
	bs = append(bs, JSONQuote)
	bs = time.Time(dt).AppendFormat(bs, time.DateTime)
	bs = append(bs, JSONQuote)

	return bs, nil
}

func (dt *DateTime) UnmarshalJSON(bs []byte) error {
	value := utils.UnsafeString(bs)
	if value == JSONNull {
		return nil
	}

	if len(bs) != dateTimePatternLength+2 || bs[0] != JSONQuote || bs[len(bs)-1] != JSONQuote {
		return errors.New("invalid datetime format")
	}

	parsed, err := ParseDateTime(value[1 : dateTimePatternLength+1])
	if err != nil {
		return err
	}

	*dt = parsed
	return nil
}

func (dt DateTime) Equal(other DateTime) bool {
	return dt.Unwrap().Equal(other.Unwrap())
}

func (dt DateTime) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

func (dt *DateTime) UnmarshalText(text []byte) error {
	parsed, err := ParseDateTime(string(text))
	if err != nil {
		return err
	}

	*dt = parsed
	return nil
}

// Date represents a date value (without time) with database and JSON support.
// It uses the standard time.DateOnly format (2006-01-02).
type Date time.Time

func (d Date) Unwrap() time.Time {
	return time.Time(d)
}

func (d *Date) Scan(src any) error {
	return scanTimeValue(src, func(s string) (any, error) {
		return ParseDate(s)
	}, func(t time.Time) any {
		return DateOf(t)
	}, "date", d)
}

func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d Date) String() string {
	return time.Time(d).Format(time.DateOnly)
}

func (d Date) MarshalJSON() ([]byte, error) {
	bs := make([]byte, 0, datePatternLength+2)
	bs = append(bs, JSONQuote)
	bs = time.Time(d).AppendFormat(bs, time.DateOnly)
	bs = append(bs, JSONQuote)

	return bs, nil
}

func (d *Date) UnmarshalJSON(bs []byte) error {
	value := utils.UnsafeString(bs)
	if value == JSONNull {
		return nil
	}

	if len(bs) != datePatternLength+2 || bs[0] != JSONQuote || bs[len(bs)-1] != JSONQuote {
		return errors.New("invalid date format")
	}

	parsed, err := ParseDate(value[1 : datePatternLength+1])
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

func (d Date) Equal(other Date) bool {
	return d.Unwrap().Equal(other.Unwrap())
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(text []byte) error {
	parsed, err := ParseDate(string(text))
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// Time represents a time value (without date) with database and JSON support.
// It uses the standard time.TimeOnly format (15:04:05).
type Time time.Time

func (t Time) Unwrap() time.Time {
	return time.Time(t)
}

func (t *Time) Scan(src any) error {
	return scanTimeValue(src, func(s string) (any, error) {
		return ParseTime(s)
	}, func(t time.Time) any {
		return TimeOf(t)
	}, "time", t)
}

func (t Time) Value() (driver.Value, error) {
	return t.String(), nil
}

func (t Time) String() string {
	return time.Time(t).Format(time.TimeOnly)
}

func (t Time) MarshalJSON() ([]byte, error) {
	bs := make([]byte, 0, timePatternLength+2)
	bs = append(bs, JSONQuote)
	bs = time.Time(t).AppendFormat(bs, time.TimeOnly)
	bs = append(bs, JSONQuote)

	return bs, nil
}

func (t *Time) UnmarshalJSON(bs []byte) error {
	value := utils.UnsafeString(bs)
	if value == JSONNull {
		return nil
	}

	if len(bs) != timePatternLength+2 || bs[0] != JSONQuote || bs[len(bs)-1] != JSONQuote {
		return errors.New("invalid time format")
	}

	parsed, err := ParseTime(value[1 : timePatternLength+1])
	if err != nil {
		return err
	}

	*t = parsed
	return nil
}

func (t Time) Equal(other Time) bool {
	return t.Unwrap().Equal(other.Unwrap())
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Time) UnmarshalText(text []byte) error {
	parsed, err := ParseTime(string(text))
	if err != nil {
		return err
	}

	*t = parsed
	return nil
}

// DateTimeNow returns the current date time in the local timezone.
func DateTimeNow() DateTime {
	now := time.Now().Local()
	return DateTime(now)
}

// DateNow returns the current date in the local timezone.
func DateNow() Date {
	now := time.Now().Local()
	return DateOf(now)
}

// TimeNow returns the current time in the local timezone.
func TimeNow() Time {
	now := time.Now().Local()
	return TimeOf(now)
}

// DateOf returns the date of the given time.
func DateOf(t time.Time) Date {
	return Date(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()))
}

// TimeOf returns the time of the given time.
func TimeOf(t time.Time) Time {
	return Time(time.Date(1970, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()))
}

// DateTimeOf returns the date time of the given time.
func DateTimeOf(t time.Time) DateTime {
	return DateTime(t)
}

// ParseDate parses a date string and returns a Date.
// First tries with the provided pattern, then falls back to cast.ToTime as a backup.
func ParseDate(value string, pattern ...string) (Date, error) {
	var layout = time.DateOnly
	if len(pattern) > 0 {
		layout = pattern[0]
	}

	// Primary: try with specified layout
	parsed, err := time.ParseInLocation(layout, value, time.Local)
	if err == nil {
		return Date(parsed), nil
	}

	// Fallback: try cast library for common time formats
	if castTime, castErr := cast.ToTimeE(value); castErr == nil {
		return DateOf(castTime), nil
	}

	// Return original error if both methods fail
	return zeroDate, err
}

// ParseTime parses a time string and returns a Time.
// First tries with the provided pattern, then falls back to cast.ToTime as a backup.
func ParseTime(value string, pattern ...string) (Time, error) {
	var layout = time.TimeOnly
	if len(pattern) > 0 {
		layout = pattern[0]
	}

	// Primary: try with specified layout
	parsed, err := time.ParseInLocation(layout, value, time.Local)
	if err == nil {
		return TimeOf(parsed), nil
	}

	// Fallback: try cast library for common time formats
	if castTime, castErr := cast.ToTimeE(value); castErr == nil {
		return TimeOf(castTime), nil
	}

	// Return original error if both methods fail
	return zeroTime, err
}

// ParseDateTime parses a date time string and returns a DateTime.
// First tries with the provided pattern, then falls back to cast.ToTime as a backup.
func ParseDateTime(value string, pattern ...string) (DateTime, error) {
	var layout = time.DateTime
	if len(pattern) > 0 {
		layout = pattern[0]
	}

	// Primary: try with specified layout
	parsed, err := time.ParseInLocation(layout, value, time.Local)
	if err == nil {
		return DateTime(parsed), nil
	}

	// Fallback: try cast library for common time formats
	if castTime, castErr := cast.ToTimeE(value); castErr == nil {
		return DateTime(castTime), nil
	}

	// Return original error if both methods fail
	return zeroDateTime, err
}
