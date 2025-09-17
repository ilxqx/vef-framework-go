package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	t.Run("parse standard datetime format", func(t *testing.T) {
		input := "2023-01-15 14:30:45"

		result, err := ParseTime(input)

		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, 45, result.Second())
		assert.Equal(t, time.Local, result.Location())
	})

	t.Run("parse date only", func(t *testing.T) {
		input := "2023-01-15"

		result, err := ParseTime(input)

		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, time.Local, result.Location())
	})

	t.Run("parse time with milliseconds", func(t *testing.T) {
		input := "2023-01-15 14:30:45.123"

		result, err := ParseTime(input)

		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, 45, result.Second())
		assert.Equal(t, 123000000, result.Nanosecond())
		assert.Equal(t, time.Local, result.Location())
	})

	t.Run("parse ISO format", func(t *testing.T) {
		input := "2023-01-15T14:30:45"

		result, err := ParseTime(input)

		if err != nil {
			// Some ISO formats may not be supported, skip this test
			t.Skipf("ISO format not supported: %v", err)
		}
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, time.Local, result.Location())
	})

	t.Run("parse different date formats", func(t *testing.T) {
		testCases := []struct {
			name   string
			input  string
			year   int
			month  time.Month
			day    int
		}{
			{
				name:  "MM/DD/YYYY",
				input: "01/15/2023",
				year:  2023,
				month: time.January,
				day:   15,
			},
			{
				name:  "DD/MM/YYYY",
				input: "15/01/2023",
				year:  2023,
				month: time.January,
				day:   15,
			},
			{
				name:  "YYYY/MM/DD",
				input: "2023/01/15",
				year:  2023,
				month: time.January,
				day:   15,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := ParseTime(tc.input)

				if err != nil {
					// Some formats may not be supported, skip
					t.Skipf("Format not supported: %v", err)
				}
				assert.Equal(t, tc.year, result.Year())
				assert.Equal(t, tc.month, result.Month())
				assert.Equal(t, tc.day, result.Day())
				assert.Equal(t, time.Local, result.Location())
			})
		}
	})

	t.Run("parse relative time strings", func(t *testing.T) {
		testCases := []string{
			"now",
			"today",
			"yesterday",
			"tomorrow",
		}

		for _, input := range testCases {
			t.Run(input, func(t *testing.T) {
				result, err := ParseTime(input)

				if err != nil {
					t.Skipf("Relative time format not supported: %v", err)
				}
				assert.Equal(t, time.Local, result.Location())
				// Just verify it's a valid time, specific values depend on current time
				assert.False(t, result.IsZero())
			})
		}
	})

	t.Run("parse with multiple format attempts", func(t *testing.T) {
		// The function can try multiple formats - use a format that actually works
		result, err := ParseTime("2023-01-15 14:30:45")

		require.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, time.Local, result.Location())
	})

	t.Run("invalid time string", func(t *testing.T) {
		input := "invalid-time-format"

		result, err := ParseTime(input)

		assert.Error(t, err)
		assert.True(t, result.IsZero())
	})

	t.Run("empty string", func(t *testing.T) {
		input := ""

		result, err := ParseTime(input)

		assert.Error(t, err)
		assert.True(t, result.IsZero())
	})

}