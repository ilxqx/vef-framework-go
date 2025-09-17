package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumnWithAlias(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		alias    []string
		expected string
	}{
		{
			name:     "column without alias",
			column:   "name",
			alias:    []string{},
			expected: "name",
		},
		{
			name:     "column with alias",
			column:   "name",
			alias:    []string{"u"},
			expected: "u.name",
		},
		{
			name:     "column with empty alias",
			column:   "age",
			alias:    []string{""},
			expected: "age",
		},
		{
			name:     "column with multiple alias values",
			column:   "email",
			alias:    []string{"user", "profile"},
			expected: "user.email",
		},
		{
			name:     "empty column with alias",
			column:   "",
			alias:    []string{"t"},
			expected: "t.",
		},
		{
			name:     "empty column without alias",
			column:   "",
			alias:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColumnWithAlias(tt.column, tt.alias...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
