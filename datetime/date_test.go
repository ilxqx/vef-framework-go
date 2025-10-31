package datetime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateOf(t *testing.T) {
	now := testTime(2023, 12, 25, 14, 30, 45)
	date := DateOf(now)

	unwrapped := date.Unwrap()
	assert.Equal(t, 2023, unwrapped.Year(), "Year should be preserved")
	assert.Equal(t, int(time.December), int(unwrapped.Month()), "Month should be preserved")
	assert.Equal(t, 25, unwrapped.Day(), "Day should be preserved")
	assert.Equal(t, 0, unwrapped.Hour(), "Hour should be zeroed")
	assert.Equal(t, 0, unwrapped.Minute(), "Minute should be zeroed")
	assert.Equal(t, 0, unwrapped.Second(), "Second should be zeroed")
	assert.Equal(t, 0, unwrapped.Nanosecond(), "Nanosecond should be zeroed")
}

func TestNowDate(t *testing.T) {
	before := time.Now()
	date := NowDate()

	unwrapped := date.Unwrap()
	assert.Equal(t, before.Year(), unwrapped.Year(), "Year should match current")
	assert.Equal(t, int(before.Month()), int(unwrapped.Month()), "Month should match current")
	assert.Equal(t, before.Day(), unwrapped.Day(), "Day should match current")
	assert.Equal(t, 0, unwrapped.Hour(), "Hour should be zero")
	assert.Equal(t, 0, unwrapped.Minute(), "Minute should be zero")
	assert.Equal(t, 0, unwrapped.Second(), "Second should be zero")
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
			"ValidDate",
			"2023-12-25",
			nil,
			false,
			testTime(2023, 12, 25, 0, 0, 0),
		},
		{
			"ValidDateWithCustomPattern",
			"25/12/2023",
			[]string{"02/01/2006"},
			false,
			testTime(2023, 12, 25, 0, 0, 0),
		},
		{
			"InvalidDate",
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
				assert.Error(t, err, "Should return error for invalid input")
			} else {
				assert.NoError(t, err, "Should parse valid date")

				if !tt.expected.IsZero() {
					unwrapped := date.Unwrap()
					assert.Equal(t, tt.expected.Year(), unwrapped.Year(), "Year should match")
					assert.Equal(t, int(tt.expected.Month()), int(unwrapped.Month()), "Month should match")
					assert.Equal(t, tt.expected.Day(), unwrapped.Day(), "Day should match")
					assert.Equal(t, 0, unwrapped.Hour(), "Hour should be zero")
					assert.Equal(t, 0, unwrapped.Minute(), "Minute should be zero")
					assert.Equal(t, 0, unwrapped.Second(), "Second should be zero")
				}
			}
		})
	}
}

func TestDateString(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	expected := "2023-12-25"
	assert.Equal(t, expected, date.String(), "String representation should match format")
}

func TestDateEqual(t *testing.T) {
	date1 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date2 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date3 := Date(testTime(2023, 12, 26, 0, 0, 0))

	assert.True(t, date1.Equal(date2), "Equal dates should be equal")
	assert.False(t, date1.Equal(date3), "Different dates should not be equal")
}

func TestDateBefore(t *testing.T) {
	date1 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date2 := Date(testTime(2023, 12, 26, 0, 0, 0))

	assert.True(t, date1.Before(date2), "Earlier date should be before later")
	assert.False(t, date2.Before(date1), "Later date should not be before earlier")
}

func TestDateAfter(t *testing.T) {
	date1 := Date(testTime(2023, 12, 25, 0, 0, 0))
	date2 := Date(testTime(2023, 12, 26, 0, 0, 0))

	assert.False(t, date1.After(date2), "Earlier date should not be after later")
	assert.True(t, date2.After(date1), "Later date should be after earlier")
}

func TestDateBetween(t *testing.T) {
	start := Date(testTime(2023, 12, 25, 0, 0, 0))
	middle := Date(testTime(2023, 12, 26, 0, 0, 0))
	end := Date(testTime(2023, 12, 27, 0, 0, 0))

	assert.True(t, middle.Between(start, end), "Middle date should be between start and end")
	assert.False(t, start.Between(middle, end), "Start date should not be between middle and end")
}

func TestDateAddDays(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	result := date.AddDays(5)

	expected := testTime(2023, 12, 30, 0, 0, 0)
	assert.True(t, expected.Equal(result.Unwrap()), "AddDays should add days correctly")
}

func TestDateAddMonths(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	result := date.AddMonths(2)

	expected := testTime(2024, 2, 25, 0, 0, 0)
	assert.True(t, expected.Equal(result.Unwrap()), "AddMonths should add months correctly")
}

func TestDateAddYears(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	result := date.AddYears(1)

	expected := testTime(2024, 12, 25, 0, 0, 0)
	assert.True(t, expected.Equal(result.Unwrap()), "AddYears should add years correctly")
}

func TestDateComponents(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))

	assert.Equal(t, 2023, date.Year(), "Year should match")
	assert.Equal(t, int(time.December), int(date.Month()), "Month should match")
	assert.Equal(t, 25, date.Day(), "Day should match")
	assert.Equal(t, int(time.Monday), int(date.Weekday()), "Weekday should match")
	assert.Equal(t, 359, date.YearDay(), "YearDay should match")
}

func TestDateIsZero(t *testing.T) {
	zeroDate := Date{}
	nonZeroDate := Date(testTime(2023, 12, 25, 0, 0, 0))

	assert.True(t, zeroDate.IsZero(), "Zero date should be zero")
	assert.False(t, nonZeroDate.IsZero(), "Non-zero date should not be zero")
}

func TestDateBeginOfMethods(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))

	t.Run("BeginOfDay", func(t *testing.T) {
		beginDay := date.BeginOfDay()
		assert.True(t, date.Unwrap().Equal(beginDay.Unwrap()), "BeginOfDay should return same date")
	})

	t.Run("BeginOfWeek", func(t *testing.T) {
		beginWeek := date.BeginOfWeek()
		expected := testTime(2023, 12, 24, 0, 0, 0)
		assert.True(t, expected.Equal(beginWeek.Unwrap()), "BeginOfWeek should return Sunday")
	})

	t.Run("BeginOfMonth", func(t *testing.T) {
		beginMonth := date.BeginOfMonth()
		expected := testTime(2023, 12, 1, 0, 0, 0)
		assert.True(t, expected.Equal(beginMonth.Unwrap()), "BeginOfMonth should return first day")
	})

	t.Run("BeginOfQuarter", func(t *testing.T) {
		beginQuarter := date.BeginOfQuarter()
		expected := testTime(2023, 10, 1, 0, 0, 0)
		assert.True(t, expected.Equal(beginQuarter.Unwrap()), "BeginOfQuarter should return Q4 start")
	})

	t.Run("BeginOfYear", func(t *testing.T) {
		beginYear := date.BeginOfYear()
		expected := testTime(2023, 1, 1, 0, 0, 0)
		assert.True(t, expected.Equal(beginYear.Unwrap()), "BeginOfYear should return Jan 1")
	})
}

func TestDateEndOfMethods(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))

	t.Run("EndOfDay", func(t *testing.T) {
		endDay := date.EndOfDay()
		assert.True(t, date.Unwrap().Equal(endDay.Unwrap()), "EndOfDay should return same date")
	})

	t.Run("EndOfWeek", func(t *testing.T) {
		endWeek := date.EndOfWeek()
		expected := testTime(2023, 12, 30, 0, 0, 0)
		assert.True(t, expected.Equal(endWeek.Unwrap()), "EndOfWeek should return Saturday")
	})

	t.Run("EndOfMonth", func(t *testing.T) {
		endMonth := date.EndOfMonth()
		expected := testTime(2023, 12, 31, 0, 0, 0)
		assert.True(t, expected.Equal(endMonth.Unwrap()), "EndOfMonth should return last day")
	})

	t.Run("EndOfQuarter", func(t *testing.T) {
		endQuarter := date.EndOfQuarter()
		expected := testTime(2023, 12, 31, 0, 0, 0)
		assert.True(t, expected.Equal(endQuarter.Unwrap()), "EndOfQuarter should return Q4 end")
	})

	t.Run("EndOfYear", func(t *testing.T) {
		endYear := date.EndOfYear()
		expected := testTime(2023, 12, 31, 0, 0, 0)
		assert.True(t, expected.Equal(endYear.Unwrap()), "EndOfYear should return Dec 31")
	})
}

func TestDateWeekdayMethods(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))

	t.Run("Monday", func(t *testing.T) {
		monday := date.Monday()
		expected := testTime(2023, 12, 25, 0, 0, 0)
		assert.True(t, expected.Equal(monday.Unwrap()), "Monday should return Dec 25")
	})

	t.Run("Tuesday", func(t *testing.T) {
		tuesday := date.Tuesday()
		expected := testTime(2023, 12, 26, 0, 0, 0)
		assert.True(t, expected.Equal(tuesday.Unwrap()), "Tuesday should return Dec 26")
	})

	t.Run("Sunday", func(t *testing.T) {
		sunday := date.Sunday()
		expected := testTime(2023, 12, 24, 0, 0, 0)
		assert.True(t, expected.Equal(sunday.Unwrap()), "Sunday should return Dec 24")
	})

	t.Run("Saturday", func(t *testing.T) {
		saturday := date.Saturday()
		expected := testTime(2023, 12, 30, 0, 0, 0)
		assert.True(t, expected.Equal(saturday.Unwrap()), "Saturday should return Dec 30")
	})
}

func TestDateMarshalJSON(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	data, err := date.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should succeed")

	expected := `"2023-12-25"`
	assert.Equal(t, expected, string(data), "JSON should match expected format")
}

func TestDateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"ValidDate", `"2023-12-25"`, false},
		{"NullValue", `null`, false},
		{"InvalidFormat", `"invalid"`, true},
		{"WrongLength", `"2023-12-25 14:30:45"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var date Date

			err := date.UnmarshalJSON([]byte(tt.input))

			if tt.shouldErr {
				assert.Error(t, err, "Should return error for invalid input")
			} else {
				assert.NoError(t, err, "Should unmarshal valid input")
			}
		})
	}
}

func TestDateValue(t *testing.T) {
	date := Date(testTime(2023, 12, 25, 0, 0, 0))
	value, err := date.Value()
	assert.NoError(t, err, "Value should succeed")

	expected := "2023-12-25"
	str, ok := value.(string)
	assert.True(t, ok, "Value should be string")
	assert.Equal(t, expected, str, "Value should match expected format")
}

func TestDateScan(t *testing.T) {
	tests := []struct {
		name   string
		src    any
		hasErr bool
	}{
		{"String", "2023-12-25", false},
		{"ByteSlice", []byte("2023-12-25"), false},
		{"TimeTime", testTime(2023, 12, 25, 14, 30, 45), false},
		{"NilPointer", (*string)(nil), false},
		{"InvalidString", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var date Date

			err := date.Scan(tt.src)

			if tt.hasErr {
				assert.Error(t, err, "Should return error for invalid input")
			} else {
				assert.NoError(t, err, "Should scan valid input")
			}
		})
	}
}

func TestDateJSONRoundTrip(t *testing.T) {
	original := Date(testTime(2023, 12, 25, 0, 0, 0))

	data, err := json.Marshal(original)
	assert.NoError(t, err, "Marshal should succeed")

	var result Date

	err = json.Unmarshal(data, &result)
	assert.NoError(t, err, "Unmarshal should succeed")

	assert.Equal(t, original.String(), result.String(), "Round trip should preserve value")
}
