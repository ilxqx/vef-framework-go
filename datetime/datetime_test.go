package datetime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateTimeOf(t *testing.T) {
	now := testTime(2023, 12, 25, 14, 30, 45)
	dt := Of(now)

	assertTimeEqual(t, now, dt.Unwrap(), "Of should preserve the original time")
}

func TestNow(t *testing.T) {
	before := time.Now()
	dt := Now()
	after := time.Now()

	unwrapped := dt.Unwrap()
	if unwrapped.Before(before) || unwrapped.After(after) {
		t.Error("Now should return current time")
	}
}

func TestFromUnix(t *testing.T) {
	timestamp := int64(1703514645) // 2023-12-25 14:30:45 UTC
	dt := FromUnix(timestamp, 0)

	expected := testTime(2023, 12, 25, 22, 30, 45) // Converts to local time
	assertTimeEqual(t, expected, dt.Unwrap(), "FromUnix should create correct datetime")
}

func TestFromUnixMilli(t *testing.T) {
	timestamp := int64(1703514645000) // 2023-12-25 14:30:45 UTC in milliseconds
	dt := FromUnixMilli(timestamp)

	expected := testTime(2023, 12, 25, 22, 30, 45) // Converts to local time
	assertTimeEqual(t, expected, dt.Unwrap(), "FromUnixMilli should create correct datetime")
}

func TestFromUnixMicro(t *testing.T) {
	timestamp := int64(1703514645000000) // 2023-12-25 14:30:45 UTC in microseconds
	dt := FromUnixMicro(timestamp)

	expected := testTime(2023, 12, 25, 22, 30, 45) // Converts to local time
	assertTimeEqual(t, expected, dt.Unwrap(), "FromUnixMicro should create correct datetime")
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pattern   []string
		shouldErr bool
		expected  time.Time
	}{
		{
			"Valid datetime",
			"2023-12-25 14:30:45",
			nil,
			false,
			testTime(2023, 12, 25, 14, 30, 45),
		},
		{
			"Valid datetime with custom pattern",
			"25/12/2023 14:30:45",
			[]string{"02/01/2006 15:04:05"},
			false,
			testTime(2023, 12, 25, 14, 30, 45),
		},
		{
			"Invalid datetime",
			"invalid",
			nil,
			true,
			time.Time{},
		},
		{
			"ISO format (fallback to cast)",
			"2023-12-25T14:30:45Z",
			nil,
			false,
			testTimeUTC(2023, 12, 25, 14, 30, 45), // Z indicates UTC time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, err := Parse(tt.input, tt.pattern...)
			if tt.shouldErr {
				assertError(t, err, "Expected Parse error")
			} else {
				assertNoError(t, err, "Unexpected Parse error")

				if !tt.expected.IsZero() {
					assertTimeEqual(t, tt.expected, dt.Unwrap(), "Parsed datetime should match expected")
				}
			}
		})
	}
}

func TestDateTimeString(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	expected := "2023-12-25 14:30:45"
	assertStringEqual(t, expected, dt.String(), "String representation")
}

func TestDateTimeEqual(t *testing.T) {
	dt1 := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	dt2 := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	dt3 := DateTime(testTime(2023, 12, 25, 14, 30, 46))

	assertBoolEqual(t, true, dt1.Equal(dt2), "Equal datetimes should be equal")
	assertBoolEqual(t, false, dt1.Equal(dt3), "Different datetimes should not be equal")
}

func TestDateTimeBefore(t *testing.T) {
	dt1 := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	dt2 := DateTime(testTime(2023, 12, 25, 14, 30, 46))

	assertBoolEqual(t, true, dt1.Before(dt2), "Earlier datetime should be before later")
	assertBoolEqual(t, false, dt2.Before(dt1), "Later datetime should not be before earlier")
}

func TestDateTimeAfter(t *testing.T) {
	dt1 := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	dt2 := DateTime(testTime(2023, 12, 25, 14, 30, 46))

	assertBoolEqual(t, false, dt1.After(dt2), "Earlier datetime should not be after later")
	assertBoolEqual(t, true, dt2.After(dt1), "Later datetime should be after earlier")
}

func TestDateTimeBetween(t *testing.T) {
	start := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	middle := DateTime(testTime(2023, 12, 25, 14, 30, 46))
	end := DateTime(testTime(2023, 12, 25, 14, 30, 47))

	assertBoolEqual(t, true, middle.Between(start, end), "Middle datetime should be between start and end")
	assertBoolEqual(t, false, start.Between(middle, end), "Start datetime should not be between middle and end")
}

func TestDateTimeAdd(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	duration := 2 * time.Hour
	result := dt.Add(duration)

	expected := testTime(2023, 12, 25, 16, 30, 45)
	assertTimeEqual(t, expected, result.Unwrap(), "Add should add duration correctly")
}

func TestDateTimeAddDate(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	result := dt.AddDate(1, 2, 3)

	expected := testTime(2025, 2, 28, 14, 30, 45)
	assertTimeEqual(t, expected, result.Unwrap(), "AddDate should add years, months, days correctly")
}

func TestDateTimeAddDays(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	result := dt.AddDays(5)

	expected := testTime(2023, 12, 30, 14, 30, 45)
	assertTimeEqual(t, expected, result.Unwrap(), "AddDays should add days correctly")
}

func TestDateTimeAddMonths(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	result := dt.AddMonths(2)

	expected := testTime(2024, 2, 25, 14, 30, 45)
	assertTimeEqual(t, expected, result.Unwrap(), "AddMonths should add months correctly")
}

func TestDateTimeAddYears(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	result := dt.AddYears(1)

	expected := testTime(2024, 12, 25, 14, 30, 45)
	assertTimeEqual(t, expected, result.Unwrap(), "AddYears should add years correctly")
}

func TestDateTimeAddHours(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	result := dt.AddHours(3)

	expected := testTime(2023, 12, 25, 17, 30, 45)
	assertTimeEqual(t, expected, result.Unwrap(), "AddHours should add hours correctly")
}

func TestDateTimeAddMinutes(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	result := dt.AddMinutes(15)

	expected := testTime(2023, 12, 25, 14, 45, 45)
	assertTimeEqual(t, expected, result.Unwrap(), "AddMinutes should add minutes correctly")
}

func TestDateTimeAddSeconds(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	result := dt.AddSeconds(30)

	expected := testTime(2023, 12, 25, 14, 31, 15)
	assertTimeEqual(t, expected, result.Unwrap(), "AddSeconds should add seconds correctly")
}

func TestDateTimeComponents(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))

	assertIntEqual(t, 2023, dt.Year(), "Year")
	assertIntEqual(t, int(time.December), int(dt.Month()), "Month")
	assertIntEqual(t, 25, dt.Day(), "Day")
	assertIntEqual(t, 14, dt.Hour(), "Hour")
	assertIntEqual(t, 30, dt.Minute(), "Minute")
	assertIntEqual(t, 45, dt.Second(), "Second")
	assertIntEqual(t, 0, dt.Nanosecond(), "Nanosecond")
	assertIntEqual(t, int(time.Monday), int(dt.Weekday()), "Weekday")
	assertIntEqual(t, 359, dt.YearDay(), "YearDay") // 25th December is 359th day of year
}

func TestDateTimeUnixMethods(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))

	expectedUnix := int64(1703485845)
	assertIntEqual(t, int(expectedUnix), int(dt.Unix()), "Unix timestamp")
	assertIntEqual(t, int(expectedUnix*1000), int(dt.UnixMilli()), "Unix milliseconds")
	assertIntEqual(t, int(expectedUnix*1000000), int(dt.UnixMicro()), "Unix microseconds")
	assertIntEqual(t, int(expectedUnix*1000000000), int(dt.UnixNano()), "Unix nanoseconds")
}

func TestDateTimeIsZero(t *testing.T) {
	zeroTime := DateTime{}
	nonZeroTime := DateTime(testTime(2023, 12, 25, 14, 30, 45))

	assertBoolEqual(t, true, zeroTime.IsZero(), "Zero datetime should be zero")
	assertBoolEqual(t, false, nonZeroTime.IsZero(), "Non-zero datetime should not be zero")
}

func TestDateTimeBeginOfMethods(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45)) // Monday

	// BeginOfMinute
	beginMinute := dt.BeginOfMinute()
	expected := testTime(2023, 12, 25, 14, 30, 0)
	assertTimeEqual(t, expected, beginMinute.Unwrap(), "BeginOfMinute")

	// BeginOfHour
	beginHour := dt.BeginOfHour()
	expected = testTime(2023, 12, 25, 14, 0, 0)
	assertTimeEqual(t, expected, beginHour.Unwrap(), "BeginOfHour")

	// BeginOfDay
	beginDay := dt.BeginOfDay()
	expected = testTime(2023, 12, 25, 0, 0, 0)
	assertTimeEqual(t, expected, beginDay.Unwrap(), "BeginOfDay")

	// BeginOfWeek (Sunday is start of week)
	beginWeek := dt.BeginOfWeek()
	expected = testTime(2023, 12, 24, 0, 0, 0) // Sunday
	assertTimeEqual(t, expected, beginWeek.Unwrap(), "BeginOfWeek")

	// BeginOfMonth
	beginMonth := dt.BeginOfMonth()
	expected = testTime(2023, 12, 1, 0, 0, 0)
	assertTimeEqual(t, expected, beginMonth.Unwrap(), "BeginOfMonth")

	// BeginOfQuarter
	beginQuarter := dt.BeginOfQuarter()
	expected = testTime(2023, 10, 1, 0, 0, 0) // Q4 starts in October
	assertTimeEqual(t, expected, beginQuarter.Unwrap(), "BeginOfQuarter")

	// BeginOfYear
	beginYear := dt.BeginOfYear()
	expected = testTime(2023, 1, 1, 0, 0, 0)
	assertTimeEqual(t, expected, beginYear.Unwrap(), "BeginOfYear")
}

func TestDateTimeEndOfMethods(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45)) // Monday

	// EndOfMinute
	endMinute := dt.EndOfMinute()
	expected := time.Date(2023, 12, 25, 14, 30, 59, 999999999, time.Local)
	assertTimeEqual(t, expected, endMinute.Unwrap(), "EndOfMinute")

	// EndOfHour
	endHour := dt.EndOfHour()
	expected = time.Date(2023, 12, 25, 14, 59, 59, 999999999, time.Local)
	assertTimeEqual(t, expected, endHour.Unwrap(), "EndOfHour")

	// EndOfDay
	endDay := dt.EndOfDay()
	expected = time.Date(2023, 12, 25, 23, 59, 59, 999999999, time.Local)
	assertTimeEqual(t, expected, endDay.Unwrap(), "EndOfDay")

	// EndOfWeek (Saturday is end of week)
	endWeek := dt.EndOfWeek()
	expected = time.Date(2023, 12, 30, 23, 59, 59, 999999999, time.Local) // Saturday
	assertTimeEqual(t, expected, endWeek.Unwrap(), "EndOfWeek")

	// EndOfYear
	endYear := dt.EndOfYear()
	expected = time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.Local)
	assertTimeEqual(t, expected, endYear.Unwrap(), "EndOfYear")
}

func TestDateTimeWeekdayMethods(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45)) // Monday

	// Test all weekday methods
	monday := dt.Monday()
	assertTimeEqual(t, testTime(2023, 12, 25, 0, 0, 0), monday.Unwrap(), "Monday")

	tuesday := dt.Tuesday()
	assertTimeEqual(t, testTime(2023, 12, 26, 0, 0, 0), tuesday.Unwrap(), "Tuesday")

	sunday := dt.Sunday()
	assertTimeEqual(t, testTime(2023, 12, 24, 0, 0, 0), sunday.Unwrap(), "Sunday")
}

func TestDateTimeMarshalJSON(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	data, err := dt.MarshalJSON()
	assertNoError(t, err, "MarshalJSON")

	expected := `"2023-12-25 14:30:45"`
	assertStringEqual(t, expected, string(data), "JSON marshaling")
}

func TestDateTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"Valid datetime", `"2023-12-25 14:30:45"`, false},
		{"Null value", `null`, false},
		{"Invalid format", `"invalid"`, true},
		{"Wrong length", `"2023-12-25"`, true},
		{"Missing quotes", `2023-12-25 14:30:45`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dt DateTime

			err := dt.UnmarshalJSON([]byte(tt.input))
			if tt.shouldErr {
				assertError(t, err, "Expected UnmarshalJSON error")
			} else {
				assertNoError(t, err, "Unexpected UnmarshalJSON error")
			}
		})
	}
}

func TestDateTimeValue(t *testing.T) {
	dt := DateTime(testTime(2023, 12, 25, 14, 30, 45))
	value, err := dt.Value()
	assertNoError(t, err, "Value")

	expected := "2023-12-25 14:30:45"
	if str, ok := value.(string); !ok || str != expected {
		t.Errorf("Expected %s, got %v", expected, value)
	}
}

func TestDateTimeScan(t *testing.T) {
	tests := []struct {
		name   string
		src    any
		hasErr bool
	}{
		{"String", "2023-12-25 14:30:45", false},
		{"[]byte", []byte("2023-12-25 14:30:45"), false},
		{"*string", func() *string {
			s := "2023-12-25 14:30:45"

			return &s
		}(), false},
		{"*[]byte", func() *[]byte {
			b := []byte("2023-12-25 14:30:45")

			return &b
		}(), false},
		{"time.Time", testTime(2023, 12, 25, 14, 30, 45), false},
		{"*time.Time", func() *time.Time {
			t := testTime(2023, 12, 25, 14, 30, 45)

			return &t
		}(), false},
		{"nil *string", (*string)(nil), false},
		{"nil *[]byte", (*[]byte)(nil), false},
		{"nil *time.Time", (*time.Time)(nil), false},
		{"invalid string", "invalid", true},
		{"unsupported type", complex(1, 2), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dt DateTime

			err := dt.Scan(tt.src)
			if tt.hasErr {
				assertError(t, err, "Expected Scan error")
			} else {
				assertNoError(t, err, "Unexpected Scan error")
			}
		})
	}
}

func TestDateTimeJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original DateTime
	}{
		{"Normal datetime", DateTime(testTime(2023, 12, 25, 14, 30, 45))},
		{"Epoch datetime", DateTime(testTime(1970, 1, 1, 0, 0, 0))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.original)
			assertNoError(t, err, "Marshal")

			// Unmarshal
			var result DateTime

			err = json.Unmarshal(data, &result)
			assertNoError(t, err, "Unmarshal")

			// Compare strings to avoid precision issues
			assertStringEqual(t, tt.original.String(), result.String(), "Round trip")
		})
	}
}
