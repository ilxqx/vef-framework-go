package datetime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeOf(t *testing.T) {
	now := testTime(2023, 12, 25, 14, 30, 45)
	timeOnly := TimeOf(now)

	unwrapped := timeOnly.Unwrap()
	assertIntEqual(t, 1970, unwrapped.Year(), "TimeOf should use epoch year")
	assertIntEqual(t, int(time.January), int(unwrapped.Month()), "TimeOf should use epoch month")
	assertIntEqual(t, 1, unwrapped.Day(), "TimeOf should use epoch day")
	assertIntEqual(t, 14, unwrapped.Hour(), "TimeOf should preserve hour")
	assertIntEqual(t, 30, unwrapped.Minute(), "TimeOf should preserve minute")
	assertIntEqual(t, 45, unwrapped.Second(), "TimeOf should preserve second")
}

func TestNowTime(t *testing.T) {
	before := time.Now()
	timeOnly := NowTime()
	_ = time.Now() // after variable not used

	unwrapped := timeOnly.Unwrap()
	assertIntEqual(t, 1970, unwrapped.Year(), "NowTime should use epoch year")
	assertIntEqual(t, int(time.January), int(unwrapped.Month()), "NowTime should use epoch month")
	assertIntEqual(t, 1, unwrapped.Day(), "NowTime should use epoch day")

	// Time components should be close to current time
	if unwrapped.Hour() < before.Hour()-1 || unwrapped.Hour() > before.Hour()+1 {
		t.Error("NowTime should return approximately current time")
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pattern   []string
		shouldErr bool
		expected  time.Time
	}{
		{
			"Valid time",
			"14:30:45",
			nil,
			false,
			time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local),
		},
		{
			"Valid time with custom pattern",
			"2:30:45 PM",
			[]string{"3:04:05 PM"},
			false,
			time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local),
		},
		{
			"Invalid time",
			"invalid",
			nil,
			true,
			time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeOnly, err := ParseTime(tt.input, tt.pattern...)
			if tt.shouldErr {
				assertError(t, err, "Expected ParseTime error")
			} else {
				assertNoError(t, err, "Unexpected ParseTime error")

				if !tt.expected.IsZero() {
					// Check that date is epoch
					unwrapped := timeOnly.Unwrap()
					assertIntEqual(t, 1970, unwrapped.Year(), "Year should be epoch")
					assertIntEqual(t, int(time.January), int(unwrapped.Month()), "Month should be epoch")
					assertIntEqual(t, 1, unwrapped.Day(), "Day should be epoch")
					assertIntEqual(t, tt.expected.Hour(), unwrapped.Hour(), "Hour should match")
					assertIntEqual(t, tt.expected.Minute(), unwrapped.Minute(), "Minute should match")
					assertIntEqual(t, tt.expected.Second(), unwrapped.Second(), "Second should match")
				}
			}
		})
	}
}

func TestTimeString(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	expected := "14:30:45"
	assertStringEqual(t, expected, timeOnly.String(), "Time string representation")
}

func TestTimeEqual(t *testing.T) {
	time1 := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	time2 := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	time3 := Time(time.Date(1970, 1, 1, 14, 30, 46, 0, time.Local))

	assertBoolEqual(t, true, time1.Equal(time2), "Equal times should be equal")
	assertBoolEqual(t, false, time1.Equal(time3), "Different times should not be equal")
}

func TestTimeBefore(t *testing.T) {
	time1 := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	time2 := Time(time.Date(1970, 1, 1, 14, 30, 46, 0, time.Local))

	assertBoolEqual(t, true, time1.Before(time2), "Earlier time should be before later")
	assertBoolEqual(t, false, time2.Before(time1), "Later time should not be before earlier")
}

func TestTimeAfter(t *testing.T) {
	time1 := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	time2 := Time(time.Date(1970, 1, 1, 14, 30, 46, 0, time.Local))

	assertBoolEqual(t, false, time1.After(time2), "Earlier time should not be after later")
	assertBoolEqual(t, true, time2.After(time1), "Later time should be after earlier")
}

func TestTimeBetween(t *testing.T) {
	start := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	middle := Time(time.Date(1970, 1, 1, 14, 30, 46, 0, time.Local))
	end := Time(time.Date(1970, 1, 1, 14, 30, 47, 0, time.Local))

	assertBoolEqual(t, true, middle.Between(start, end), "Middle time should be between start and end")
	assertBoolEqual(t, false, start.Between(middle, end), "Start time should not be between middle and end")
}

func TestTimeAdd(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	duration := 2 * time.Hour
	result := timeOnly.Add(duration)

	expected := time.Date(1970, 1, 1, 16, 30, 45, 0, time.Local)
	assertTimeEqual(t, expected, result.Unwrap(), "Add should add duration correctly")
}

func TestTimeAddHours(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	result := timeOnly.AddHours(3)

	expected := time.Date(1970, 1, 1, 17, 30, 45, 0, time.Local)
	assertTimeEqual(t, expected, result.Unwrap(), "AddHours should add hours correctly")
}

func TestTimeAddMinutes(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	result := timeOnly.AddMinutes(15)

	expected := time.Date(1970, 1, 1, 14, 45, 45, 0, time.Local)
	assertTimeEqual(t, expected, result.Unwrap(), "AddMinutes should add minutes correctly")
}

func TestTimeAddSeconds(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	result := timeOnly.AddSeconds(30)

	expected := time.Date(1970, 1, 1, 14, 31, 15, 0, time.Local)
	assertTimeEqual(t, expected, result.Unwrap(), "AddSeconds should add seconds correctly")
}

func TestTimeAddNanoseconds(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	result := timeOnly.AddNanoseconds(500000000) // 0.5 seconds

	expected := time.Date(1970, 1, 1, 14, 30, 45, 500000000, time.Local)
	assertTimeEqual(t, expected, result.Unwrap(), "AddNanoseconds should add nanoseconds correctly")
}

func TestTimeAddMicroseconds(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	result := timeOnly.AddMicroseconds(500000) // 0.5 seconds

	expected := time.Date(1970, 1, 1, 14, 30, 45, 500000000, time.Local)
	assertTimeEqual(t, expected, result.Unwrap(), "AddMicroseconds should add microseconds correctly")
}

func TestTimeAddMilliseconds(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	result := timeOnly.AddMilliseconds(500) // 0.5 seconds

	expected := time.Date(1970, 1, 1, 14, 30, 45, 500000000, time.Local)
	assertTimeEqual(t, expected, result.Unwrap(), "AddMilliseconds should add milliseconds correctly")
}

func TestTimeComponents(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 123456789, time.Local))

	assertIntEqual(t, 14, timeOnly.Hour(), "Hour")
	assertIntEqual(t, 30, timeOnly.Minute(), "Minute")
	assertIntEqual(t, 45, timeOnly.Second(), "Second")
	assertIntEqual(t, 123456789, timeOnly.Nanosecond(), "Nanosecond")
}

func TestTimeIsZero(t *testing.T) {
	zeroTime := Time{}
	nonZeroTime := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))

	assertBoolEqual(t, true, zeroTime.IsZero(), "Zero time should be zero")
	assertBoolEqual(t, false, nonZeroTime.IsZero(), "Non-zero time should not be zero")
}

func TestTimeBeginOfMethods(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 123456789, time.Local))

	// BeginOfMinute
	beginMinute := timeOnly.BeginOfMinute()
	expected := time.Date(1970, 1, 1, 14, 30, 0, 0, time.Local)
	assertTimeEqual(t, expected, beginMinute.Unwrap(), "BeginOfMinute")

	// BeginOfHour
	beginHour := timeOnly.BeginOfHour()
	expected = time.Date(1970, 1, 1, 14, 0, 0, 0, time.Local)
	assertTimeEqual(t, expected, beginHour.Unwrap(), "BeginOfHour")
}

func TestTimeEndOfMethods(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 123456789, time.Local))

	// EndOfMinute
	endMinute := timeOnly.EndOfMinute()
	expected := time.Date(1970, 1, 1, 14, 30, 59, 999999999, time.Local)
	assertTimeEqual(t, expected, endMinute.Unwrap(), "EndOfMinute")

	// EndOfHour
	endHour := timeOnly.EndOfHour()
	expected = time.Date(1970, 1, 1, 14, 59, 59, 999999999, time.Local)
	assertTimeEqual(t, expected, endHour.Unwrap(), "EndOfHour")
}

func TestTimeMarshalJSON(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	data, err := timeOnly.MarshalJSON()
	assertNoError(t, err, "MarshalJSON")

	expected := `"14:30:45"`
	assertStringEqual(t, expected, string(data), "JSON marshaling")
}

func TestTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"Valid time", `"14:30:45"`, false},
		{"Null value", `null`, false},
		{"Invalid format", `"invalid"`, true},
		{"Wrong length", `"14:30"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var timeOnly Time

			err := timeOnly.UnmarshalJSON([]byte(tt.input))
			if tt.shouldErr {
				assertError(t, err, "Expected UnmarshalJSON error")
			} else {
				assertNoError(t, err, "Unexpected UnmarshalJSON error")
			}
		})
	}
}

func TestTimeValue(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))
	value, err := timeOnly.Value()
	assertNoError(t, err, "Value")

	expected := "14:30:45"
	if str, ok := value.(string); !ok || str != expected {
		t.Errorf("Expected %s, got %v", expected, value)
	}
}

func TestTimeScan(t *testing.T) {
	tests := []struct {
		name   string
		src    any
		hasErr bool
	}{
		{"String", "14:30:45", false},
		{"[]byte", []byte("14:30:45"), false},
		{"time.Time", testTime(2023, 12, 25, 14, 30, 45), false},
		{"nil *string", (*string)(nil), false},
		{"invalid string", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var timeOnly Time

			err := timeOnly.Scan(tt.src)
			if tt.hasErr {
				assertError(t, err, "Expected Scan error")
			} else {
				assertNoError(t, err, "Unexpected Scan error")
			}
		})
	}
}

func TestTimeJSONRoundTrip(t *testing.T) {
	original := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.Local))

	// Marshal
	data, err := json.Marshal(original)
	assertNoError(t, err, "Marshal")

	// Unmarshal
	var result Time

	err = json.Unmarshal(data, &result)
	assertNoError(t, err, "Unmarshal")

	// Compare strings to avoid precision issues
	assertStringEqual(t, original.String(), result.String(), "Round trip")
}
