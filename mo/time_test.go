package mo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateTimeOf(t *testing.T) {
	now := time.Now()
	dt := DateTimeOf(now)

	if time.Time(dt) != now {
		t.Error("DateTimeOf should preserve the original time")
	}
}

func TestDateOf(t *testing.T) {
	now := time.Date(2023, 12, 25, 14, 30, 45, 123456789, time.UTC)
	date := DateOf(now)

	unwrapped := date.Unwrap()
	if unwrapped.Year() != 2023 || unwrapped.Month() != 12 || unwrapped.Day() != 25 {
		t.Error("DateOf should preserve date components")
	}
	if unwrapped.Hour() != 0 || unwrapped.Minute() != 0 || unwrapped.Second() != 0 || unwrapped.Nanosecond() != 0 {
		t.Error("DateOf should zero out time components")
	}
}

func TestTimeOf(t *testing.T) {
	now := time.Date(2023, 12, 25, 14, 30, 45, 123456789, time.UTC)
	timeOnly := TimeOf(now)

	unwrapped := timeOnly.Unwrap()
	if unwrapped.Year() != 1970 || unwrapped.Month() != 1 || unwrapped.Day() != 1 {
		t.Error("TimeOf should use epoch date (1970-01-01)")
	}
	if unwrapped.Hour() != 14 || unwrapped.Minute() != 30 || unwrapped.Second() != 45 || unwrapped.Nanosecond() != 123456789 {
		t.Error("TimeOf should preserve time components")
	}
}

func TestDateTimeNow(t *testing.T) {
	before := time.Now()
	dt := DateTimeNow()
	after := time.Now()

	unwrapped := dt.Unwrap()
	if unwrapped.Before(before) || unwrapped.After(after) {
		t.Error("DateTimeNow should return current time")
	}
}

func TestDateNow(t *testing.T) {
	before := time.Now()
	date := DateNow()
	_ = time.Now() // after variable not needed

	unwrapped := date.Unwrap()
	if unwrapped.Year() != before.Year() || unwrapped.Month() != before.Month() || unwrapped.Day() != before.Day() {
		t.Error("DateNow should return current date")
	}
	if unwrapped.Hour() != 0 || unwrapped.Minute() != 0 || unwrapped.Second() != 0 {
		t.Error("DateNow should have zero time components")
	}
}

func TestTimeNow(t *testing.T) {
	before := time.Now()
	timeOnly := TimeNow()
	_ = time.Now() // after variable not used

	unwrapped := timeOnly.Unwrap()
	if unwrapped.Year() != 1970 || unwrapped.Month() != 1 || unwrapped.Day() != 1 {
		t.Error("TimeNow should use epoch date")
	}

	// Time components should be close to current time
	if unwrapped.Hour() < before.Hour()-1 || unwrapped.Hour() > before.Hour()+1 {
		t.Error("TimeNow should return approximately current time")
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pattern   []string
		shouldErr bool
	}{
		{
			"Valid datetime",
			"2023-12-25 14:30:45",
			nil,
			false,
		},
		{
			"Valid datetime with custom pattern",
			"25/12/2023 14:30:45",
			[]string{"02/01/2006 15:04:05"},
			false,
		},
		{
			"Invalid datetime",
			"invalid",
			nil,
			true,
		},
		{
			"ISO format (fallback to cast)",
			"2023-12-25T14:30:45Z",
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, err := ParseDateTime(tt.input, tt.pattern...)
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Basic sanity check
				unwrapped := dt.Unwrap()
				if unwrapped.IsZero() {
					t.Error("Parsed datetime should not be zero")
				}
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pattern   []string
		shouldErr bool
	}{
		{
			"Valid date",
			"2023-12-25",
			nil,
			false,
		},
		{
			"Valid date with custom pattern",
			"25/12/2023",
			[]string{"02/01/2006"},
			false,
		},
		{
			"Invalid date",
			"invalid",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := ParseDate(tt.input, tt.pattern...)
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check that time components are zeroed
				unwrapped := date.Unwrap()
				if unwrapped.Hour() != 0 || unwrapped.Minute() != 0 || unwrapped.Second() != 0 {
					t.Error("Parsed date should have zero time components")
				}
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pattern   []string
		shouldErr bool
	}{
		{
			"Valid time",
			"14:30:45",
			nil,
			false,
		},
		{
			"Valid time with custom pattern",
			"2:30:45 PM",
			[]string{"3:04:05 PM"},
			false,
		},
		{
			"Invalid time",
			"invalid",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeOnly, err := ParseTime(tt.input, tt.pattern...)
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check that date is epoch
				unwrapped := timeOnly.Unwrap()
				if unwrapped.Year() != 1970 || unwrapped.Month() != 1 || unwrapped.Day() != 1 {
					t.Error("Parsed time should use epoch date")
				}
			}
		})
	}
}

func TestDateTimeString(t *testing.T) {
	dt := DateTime(time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC))
	expected := "2023-12-25 14:30:45"
	if dt.String() != expected {
		t.Errorf("Expected %s, got %s", expected, dt.String())
	}
}

func TestDateString(t *testing.T) {
	date := Date(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))
	expected := "2023-12-25"
	if date.String() != expected {
		t.Errorf("Expected %s, got %s", expected, date.String())
	}
}

func TestTimeString(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.UTC))
	expected := "14:30:45"
	if timeOnly.String() != expected {
		t.Errorf("Expected %s, got %s", expected, timeOnly.String())
	}
}

func TestDateTimeMarshalJSON(t *testing.T) {
	dt := DateTime(time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC))
	data, err := dt.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `"2023-12-25 14:30:45"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestDateTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			"Valid datetime",
			`"2023-12-25 14:30:45"`,
			false,
		},
		{
			"Null value",
			`null`,
			false,
		},
		{
			"Invalid format",
			`"invalid"`,
			true,
		},
		{
			"Wrong length",
			`"2023-12-25"`,
			true,
		},
		{
			"Missing quotes",
			`2023-12-25 14:30:45`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dt DateTime
			err := dt.UnmarshalJSON([]byte(tt.input))
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDateMarshalJSON(t *testing.T) {
	date := Date(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))
	data, err := date.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `"2023-12-25"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestDateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			"Valid date",
			`"2023-12-25"`,
			false,
		},
		{
			"Null value",
			`null`,
			false,
		},
		{
			"Invalid format",
			`"invalid"`,
			true,
		},
		{
			"Wrong length",
			`"2023-12-25 14:30:45"`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var date Date
			err := date.UnmarshalJSON([]byte(tt.input))
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTimeMarshalJSON(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.UTC))
	data, err := timeOnly.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `"14:30:45"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			"Valid time",
			`"14:30:45"`,
			false,
		},
		{
			"Null value",
			`null`,
			false,
		},
		{
			"Invalid format",
			`"invalid"`,
			true,
		},
		{
			"Wrong length",
			`"14:30"`,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var timeOnly Time
			err := timeOnly.UnmarshalJSON([]byte(tt.input))
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDateTimeValue(t *testing.T) {
	dt := DateTime(time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC))
	value, err := dt.Value()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "2023-12-25 14:30:45"
	if str, ok := value.(string); !ok || str != expected {
		t.Errorf("Expected %s, got %v", expected, value)
	}
}

func TestDateValue(t *testing.T) {
	date := Date(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))
	value, err := date.Value()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "2023-12-25"
	if str, ok := value.(string); !ok || str != expected {
		t.Errorf("Expected %s, got %v", expected, value)
	}
}

func TestTimeValue(t *testing.T) {
	timeOnly := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.UTC))
	value, err := timeOnly.Value()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "14:30:45"
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
		{"*string", func() *string { s := "2023-12-25 14:30:45"; return &s }(), false},
		{"*[]byte", func() *[]byte { b := []byte("2023-12-25 14:30:45"); return &b }(), false},
		{"time.Time", time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC), false},
		{"*time.Time", func() *time.Time { t := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC); return &t }(), false},
		{"nil", nil, true},
		{"nil *string", (*string)(nil), false},
		{"nil *[]byte", (*[]byte)(nil), false},
		{"nil *time.Time", (*time.Time)(nil), false},
		{"int (via cast)", 1703517045, true}, // Unix timestamp - may not parse correctly as datetime string
		{"invalid string", "invalid", true},
		{"unsupported type", complex(1, 2), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dt DateTime
			err := dt.Scan(tt.src)
			if tt.hasErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
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
		{"time.Time", time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC), false},
		{"nil", nil, true},
		{"invalid string", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var date Date
			err := date.Scan(tt.src)
			if tt.hasErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
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
		{"time.Time", time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC), false},
		{"nil", nil, true},
		{"invalid string", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var timeOnly Time
			err := timeOnly.Scan(tt.src)
			if tt.hasErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTimeJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original DateTime
	}{
		{
			"Normal datetime",
			DateTime(time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)),
		},
		{
			"Epoch datetime",
			DateTime(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.original)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			// Unmarshal
			var result DateTime
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			// Compare strings instead of time.Time to avoid precision issues
			if result.String() != tt.original.String() {
				t.Errorf("Round trip failed: %v != %v", result, tt.original)
			}
		})
	}
}

func TestDateJSONRoundTrip(t *testing.T) {
	original := Date(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var result Date
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Compare strings instead of time.Time to avoid precision issues
	if result.String() != original.String() {
		t.Errorf("Round trip failed: %v != %v", result, original)
	}
}

func TestTimeJSONRoundTripTime(t *testing.T) {
	original := Time(time.Date(1970, 1, 1, 14, 30, 45, 0, time.UTC))

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var result Time
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Compare strings instead of time.Time to avoid precision issues
	if result.String() != original.String() {
		t.Errorf("Round trip failed: %v != %v", result, original)
	}
}
