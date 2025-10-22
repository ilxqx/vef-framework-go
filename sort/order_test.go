package sort

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderDirectionString(t *testing.T) {
	tests := []struct {
		name     string
		od       OrderDirection
		expected string
	}{
		{"Ascending order", OrderAsc, "ASC"},
		{"Descending order", OrderDesc, "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.od.String())
		})
	}
}

func TestOrderDirectionMarshalText(t *testing.T) {
	tests := []struct {
		name     string
		od       OrderDirection
		expected string
	}{
		{"Ascending order", OrderAsc, "asc"},
		{"Descending order", OrderDesc, "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := tt.od.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(text))
		})
	}
}

func TestOrderDirectionUnmarshalText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  OrderDirection
		shouldErr bool
	}{
		{"Lowercase asc", "asc", OrderAsc, false},
		{"Uppercase ASC", "ASC", OrderAsc, false},
		{"Mixed case Asc", "Asc", OrderAsc, false},
		{"Lowercase desc", "desc", OrderDesc, false},
		{"Uppercase DESC", "DESC", OrderDesc, false},
		{"Mixed case Desc", "Desc", OrderDesc, false},
		{"With leading space", " asc", OrderAsc, false},
		{"With trailing space", "desc ", OrderDesc, false},
		{"With both spaces", " DESC ", OrderDesc, false},
		{"Invalid value", "invalid", OrderAsc, true},
		{"Empty string", "", OrderAsc, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var od OrderDirection

			err := od.UnmarshalText([]byte(tt.input))

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, od)
			}
		})
	}
}

func TestOrderDirectionMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		od       OrderDirection
		expected string
	}{
		{"Ascending order", OrderAsc, `"asc"`},
		{"Descending order", OrderDesc, `"desc"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.od)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestOrderDirectionUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  OrderDirection
		shouldErr bool
	}{
		{"Lowercase asc", `"asc"`, OrderAsc, false},
		{"Uppercase ASC", `"ASC"`, OrderAsc, false},
		{"Mixed case Asc", `"Asc"`, OrderAsc, false},
		{"Lowercase desc", `"desc"`, OrderDesc, false},
		{"Uppercase DESC", `"DESC"`, OrderDesc, false},
		{"Mixed case Desc", `"Desc"`, OrderDesc, false},
		{"With spaces", `" desc "`, OrderDesc, false},
		{"Invalid value", `"invalid"`, OrderAsc, true},
		{"Not a string", `123`, OrderAsc, true},
		{"Boolean value", `true`, OrderAsc, true},
		{"Null value", `null`, OrderAsc, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var od OrderDirection

			err := json.Unmarshal([]byte(tt.input), &od)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, od)
			}
		})
	}
}

func TestOrderDirectionJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input OrderDirection
	}{
		{"Ascending order", OrderAsc},
		{"Descending order", OrderDesc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.input)
			require.NoError(t, err)

			// Unmarshal
			var result OrderDirection

			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Compare
			assert.Equal(t, tt.input, result)
		})
	}
}

func TestOrderDirectionInStruct(t *testing.T) {
	type testStruct struct {
		Direction OrderDirection `json:"direction"`
		Column    string         `json:"column"`
	}

	tests := []struct {
		name  string
		input testStruct
	}{
		{"With ascending", testStruct{Direction: OrderAsc, Column: "name"}},
		{"With descending", testStruct{Direction: OrderDesc, Column: "created_at"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.input)
			require.NoError(t, err)

			// Unmarshal
			var result testStruct

			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Compare
			assert.Equal(t, tt.input.Direction, result.Direction)
			assert.Equal(t, tt.input.Column, result.Column)
		})
	}
}

func TestNullsOrderString(t *testing.T) {
	tests := []struct {
		name     string
		no       NullsOrder
		expected string
	}{
		{"Default nulls order", NullsDefault, ""},
		{"Nulls first", NullsFirst, "NULLS FIRST"},
		{"Nulls last", NullsLast, "NULLS LAST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.no.String())
		})
	}
}

func TestOrderSpecIsValid(t *testing.T) {
	tests := []struct {
		name     string
		spec     OrderSpec
		expected bool
	}{
		{
			"Valid with column",
			OrderSpec{Column: "name", Direction: OrderAsc},
			true,
		},
		{
			"Valid with column and nulls order",
			OrderSpec{Column: "age", Direction: OrderDesc, NullsOrder: NullsLast},
			true,
		},
		{
			"Invalid without column",
			OrderSpec{Direction: OrderAsc},
			false,
		},
		{
			"Invalid with empty column",
			OrderSpec{Column: "", Direction: OrderDesc},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.spec.IsValid())
		})
	}
}
