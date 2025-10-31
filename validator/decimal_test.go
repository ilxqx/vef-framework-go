package validator

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimalMinValidation(t *testing.T) {
	type testStruct struct {
		Value decimal.Decimal `validate:"dec_min=10.5" label:"最小值"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"validMinimum", "10.5", false},
		{"validAboveMinimum", "20.0", false},
		{"invalidBelowMinimum", "5.0", true},
		{"invalidZero", "0", true},
		{"validExactMinimum", "10.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, _ := decimal.NewFromString(tt.value)
			s := testStruct{Value: value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error for value: %s", tt.value)
				assert.Contains(t, err.Error(), "最小值", "Error message should contain label")
			} else {
				assert.NoError(t, err, "Should not return validation error for value: %s", tt.value)
			}
		})
	}
}

func TestDecimalMaxValidation(t *testing.T) {
	type testStruct struct {
		Value decimal.Decimal `validate:"dec_max=100" label:"最大值"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"validMaximum", "100", false},
		{"validBelowMaximum", "50.5", false},
		{"invalidAboveMaximum", "150.0", true},
		{"validExactMaximum", "100.00", false},
		{"validZero", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, _ := decimal.NewFromString(tt.value)
			s := testStruct{Value: value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error for value: %s", tt.value)
				assert.Contains(t, err.Error(), "最大值", "Error message should contain label")
			} else {
				assert.NoError(t, err, "Should not return validation error for value: %s", tt.value)
			}
		})
	}
}

func TestDecimalRangeValidation(t *testing.T) {
	type testStruct struct {
		Value decimal.Decimal `validate:"dec_min=1,dec_max=50" label:"范围值"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"validInRange", "25.5", false},
		{"validMinimumBoundary", "1", false},
		{"validMaximumBoundary", "50", false},
		{"invalidBelowRange", "0.5", true},
		{"invalidAboveRange", "51", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, _ := decimal.NewFromString(tt.value)
			s := testStruct{Value: value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error for value: %s", tt.value)
				assert.Contains(t, err.Error(), "范围值", "Error message should contain label")
			} else {
				assert.NoError(t, err, "Should not return validation error for value: %s", tt.value)
			}
		})
	}
}
