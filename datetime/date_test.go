package datetime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateOf(t *testing.T) {
	now := testTime(2023, 12, 25, 14, 30, 45)
	date := DateOf(now)

	unwrapped := date.Unwrap()
	assertIntEqual(t, 2023, unwrapped.Year(), "DateOf should preserve year")
	assertIntEqual(t, int(time.December), int(unwrapped.Month()), "DateOf should preserve month")
	assertIntEqual(t, 25, unwrapped.Day(), "DateOf should preserve day")
	assertIntEqual(t, 0, unwrapped.Hour(), "DateOf should zero out hour")
	assertIntEqual(t, 0, unwrapped.Minute(), "DateOf should zero out minute")
	assertIntEqual(t, 0, unwrapped.Second(), "DateOf should zero out second")
	assertIntEqual(t, 0, unwrapped.Nanosecond(), "DateOf should zero out nanosecond")
}

func TestNowDate(t *testing.T) {
	before := time.Now()
	date := NowDate()
	_ = time.Now() // after variable not needed

	unwrapped := date.Unwrap()
	assertIntEqual(t, before.Year(), unwrapped.Year(), "NowDate should return current year")
	assertIntEqual(t, int(before.Month()), int(unwrapped.Month()), "NowDate should return current month")
	assertIntEqual(t, before.Day(), unwrapped.Day(), "NowDate should return current day")
	assertIntEqual(t, 0, unwrapped.Hour(), "NowDate should have zero hour")
	assertIntEqual(t, 0, unwrapped.Minute(), "NowDate should have zero minute")
	assertIntEqual(t, 0, unwrapped.Second(), "NowDate should have zero second")
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pattern   []string
		shouldErr bool
		expected  time.Time
	}{
		{
			"Valid date",
			"2023-12-25",
			nil,
			false,
			testTime(2023, 12, 25, 0, 0, 0),
		},
		{
			"Valid date with custom pattern",
			"25/12/2023",
			[]string{"02/01/2006"},
			false,
			testTime(2023, 12, 25, 0, 0, 0),
		},
		{
			"Invalid date",
			"invalid",
			nil,
			true,
			time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := ParseDate(tt.input, tt.pattern...)
			if tt.shouldErr {
				assertError(t, err, "Expected ParseDate error")
			} else {
				assertNoError(t, err, "Unexpected ParseDate error")

				if !tt.expected.IsZero() {
					// Check that time components are zeroed
					unwrapped := date.Unwrap()
					assertIntEqual(t, tt.expected.Year(), unwrapped.Year(), "Year should match")
					assertIntEqual(t, int(tt.expected.Month()), int(unwrapped.Month()), "Month should match")
					assertIntEqual(t, tt.expected.Day(), unwrapped.Day(), "Day should match")
					assertIntEqual(t, 0, unwrapped.Hour(), "Hour should be zero")
					assertIntEqual(t, 0, unwrapped.Minute(), "Minute should be zero")
					assertIntEqual(t, 0, unwrapped.Second(), "Second should be zero")
				}
			}
		})
	}
}

func TestDateString(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	expected := "2023-12-25"
	assertStringEqual(t, expected, date.String(), "Date string representation")
}

func TestDateEqual(t *testing.T) {
	date1 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date2 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date3 := Date(testTime(2023, 12, 26, 0, 0, 0))

	assertBoolEqual(t, true, date1.Equal(date2), "Equal dates should be equal")
	assertBoolEqual(t, false, date1.Equal(date3), "Different dates should not be equal")
}

func TestDateBefore(t *testing.T) {
	date1 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date2 := Date(testTime(2023, 12, 26, 0, 0, 0))

	assertBoolEqual(t, true, date1.Before(date2), "Earlier date should be before later")
	assertBoolEqual(t, false, date2.Before(date1), "Later date should not be before earlier")
}

func TestDateAfter(t *testing.T) {
	date1 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date2 := Date(testTime(2023, 12, 26, 0, 0, 0))

	assertBoolEqual(t, false, date1.After(date2), "Earlier date should not be after later")
	assertBoolEqual(t, true, date2.After(date1), "Later date should be after earlier")
}

func TestDateBetween(t *testing.T) {
	start := Date(testTime(2023, 12, 25, 0, 0, 0))
	middle := Date(testTime(2023, 12, 26, 0, 0, 0))
	end := Date(testTime(2023, 12, 27, 0, 0, 0))

	assertBoolEqual(t, true, middle.Between(start, end), "Middle date should be between start and end")
	assertBoolEqual(t, false, start.Between(middle, end), "Start date should not be between middle and end")
}

func TestDateAddDays(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	result := date.AddDays(5)

	expected := testTime(2023, 12, 30, 0, 0, 0)
	assertTimeEqual(t, expected, result.Unwrap(), "AddDays should add days correctly")
}

func TestDateAddMonths(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	result := date.AddMonths(2)

	expected := testTime(2024, 2, 25, 0, 0, 0)
	assertTimeEqual(t, expected, result.Unwrap(), "AddMonths should add months correctly")
}

func TestDateAddYears(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	result := date.AddYears(1)

	expected := testTime(2024, 12, 25, 0, 0, 0)
	assertTimeEqual(t, expected, result.Unwrap(), "AddYears should add years correctly")
}

func TestDateComponents(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))

	assertIntEqual(t, 2023, date.Year(), "Year")
	assertIntEqual(t, int(time.December), int(date.Month()), "Month")
	assertIntEqual(t, 25, date.Day(), "Day")
	assertIntEqual(t, int(time.Monday), int(date.Weekday()), "Weekday")
	assertIntEqual(t, 359, date.YearDay(), "YearDay") // 25th December is 359th day of year
}

func TestDateIsZero(t *testing.T) {
	zeroDate := Date{}
	nonZeroDate := Date(testTime(2023, 12, 25, 0, 0, 0))

	assertBoolEqual(t, true, zeroDate.IsZero(), "Zero date should be zero")
	assertBoolEqual(t, false, nonZeroDate.IsZero(), "Non-zero date should not be zero")
}

func TestDateBeginOfMethods(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0)) // Monday

	// BeginOfDay (should return same date)
	beginDay := date.BeginOfDay()
	assertTimeEqual(t, date.Unwrap(), beginDay.Unwrap(), "BeginOfDay should return same date")

	// BeginOfWeek (Sunday is start of week)
	beginWeek := date.BeginOfWeek()
	expected := testTime(2023, 12, 24, 0, 0, 0) // Sunday
	assertTimeEqual(t, expected, beginWeek.Unwrap(), "BeginOfWeek")

	// BeginOfMonth
	beginMonth := date.BeginOfMonth()
	expected = testTime(2023, 12, 1, 0, 0, 0)
	assertTimeEqual(t, expected, beginMonth.Unwrap(), "BeginOfMonth")

	// BeginOfQuarter
	beginQuarter := date.BeginOfQuarter()
	expected = testTime(2023, 10, 1, 0, 0, 0) // Q4 starts in October
	assertTimeEqual(t, expected, beginQuarter.Unwrap(), "BeginOfQuarter")

	// BeginOfYear
	beginYear := date.BeginOfYear()
	expected = testTime(2023, 1, 1, 0, 0, 0)
	assertTimeEqual(t, expected, beginYear.Unwrap(), "BeginOfYear")
}

func TestDateEndOfMethods(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0)) // Monday

	// EndOfDay (should return same date)
	endDay := date.EndOfDay()
	assertTimeEqual(t, date.Unwrap(), endDay.Unwrap(), "EndOfDay should return same date")

	// EndOfWeek (Saturday is end of week)
	endWeek := date.EndOfWeek()
	expected := testTime(2023, 12, 30, 0, 0, 0) // Saturday
	assertTimeEqual(t, expected, endWeek.Unwrap(), "EndOfWeek")

	// EndOfMonth
	endMonth := date.EndOfMonth()
	expected = testTime(2023, 12, 31, 0, 0, 0)
	assertTimeEqual(t, expected, endMonth.Unwrap(), "EndOfMonth")

	// EndOfQuarter
	endQuarter := date.EndOfQuarter()
	expected = testTime(2023, 12, 31, 0, 0, 0) // Q4 ends on Dec 31
	assertTimeEqual(t, expected, endQuarter.Unwrap(), "EndOfQuarter")

	// EndOfYear
	endYear := date.EndOfYear()
	expected = testTime(2023, 12, 31, 0, 0, 0)
	assertTimeEqual(t, expected, endYear.Unwrap(), "EndOfYear")
}

func TestDateWeekdayMethods(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0)) // Monday

	// Test all weekday methods
	monday := date.Monday()
	assertTimeEqual(t, testTime(2023, 12, 25, 0, 0, 0), monday.Unwrap(), "Monday")

	tuesday := date.Tuesday()
	assertTimeEqual(t, testTime(2023, 12, 26, 0, 0, 0), tuesday.Unwrap(), "Tuesday")

	sunday := date.Sunday()
	assertTimeEqual(t, testTime(2023, 12, 24, 0, 0, 0), sunday.Unwrap(), "Sunday")

	saturday := date.Saturday()
	assertTimeEqual(t, testTime(2023, 12, 30, 0, 0, 0), saturday.Unwrap(), "Saturday")
}

func TestDateMarshalJSON(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	data, err := date.MarshalJSON()
	assertNoError(t, err, "MarshalJSON")

	expected := `"2023-12-25"`
	assertStringEqual(t, expected, string(data), "JSON marshaling")
}

func TestDateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"Valid date", `"2023-12-25"`, false},
		{"Null value", `null`, false},
		{"Invalid format", `"invalid"`, true},
		{"Wrong length", `"2023-12-25 14:30:45"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var date Date

			err := date.UnmarshalJSON([]byte(tt.input))
			if tt.shouldErr {
				assertError(t, err, "Expected UnmarshalJSON error")
			} else {
				assertNoError(t, err, "Unexpected UnmarshalJSON error")
			}
		})
	}
}

func TestDateValue(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	value, err := date.Value()
	assertNoError(t, err, "Value")

	expected := "2023-12-25"
	if str, ok := value.(string); !ok || str != expected {
		t.Errorf("Expected %s, got %v", expected, value)
	}
}

func TestDateScan(t *testing.T) {
	tests := []struct {
		name   string
		src    any
		hasErr bool
	}{
		{"String", "2023-12-25", false},
		{"[]byte", []byte("2023-12-25"), false},
		{"time.Time", testTime(2023, 12, 25, 14, 30, 45), false},
		{"nil *string", (*string)(nil), false},
		{"invalid string", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var date Date

			err := date.Scan(tt.src)
			if tt.hasErr {
				assertError(t, err, "Expected Scan error")
			} else {
				assertNoError(t, err, "Unexpected Scan error")
			}
		})
	}
}

func TestDateJSONRoundTrip(t *testing.T) {
	original := Date(testTime(2023, 12, 25, 0, 0, 0))

	// Marshal
	data, err := json.Marshal(original)
	assertNoError(t, err, "Marshal")

	// Unmarshal
	var result Date

	err = json.Unmarshal(data, &result)
	assertNoError(t, err, "Unmarshal")

	// Compare strings to avoid precision issues
	assertStringEqual(t, original.String(), result.String(), "Round trip")
}
