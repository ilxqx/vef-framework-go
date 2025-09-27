package validator

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type DecimalMinTestStruct struct {
	Value decimal.Decimal `validate:"dec_min=10.5" label:"最小值"`
}

type DecimalMaxTestStruct struct {
	Value decimal.Decimal `validate:"dec_max=100" label:"最大值"`
}

type DecimalRangeTestStruct struct {
	Value decimal.Decimal `validate:"dec_min=1,dec_max=50" label:"范围值"`
}

func TestDecimalMinValidation(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"Valid minimum", "10.5", false},
		{"Valid above minimum", "20.0", false},
		{"Invalid below minimum", "5.0", true},
		{"Invalid zero", "0", true},
		{"Valid exact minimum", "10.5", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, _ := decimal.NewFromString(tc.value)
			testStruct := DecimalMinTestStruct{Value: value}

			err := Validate(&testStruct)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "最小值")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecimalMaxValidation(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"Valid maximum", "100", false},
		{"Valid below maximum", "50.5", false},
		{"Invalid above maximum", "150.0", true},
		{"Valid exact maximum", "100.00", false},
		{"Valid zero", "0", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, _ := decimal.NewFromString(tc.value)
			testStruct := DecimalMaxTestStruct{Value: value}

			err := Validate(&testStruct)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "最大值")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecimalRangeValidation(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"Valid in range", "25.5", false},
		{"Valid minimum boundary", "1", false},
		{"Valid maximum boundary", "50", false},
		{"Invalid below range", "0.5", true},
		{"Invalid above range", "51", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, _ := decimal.NewFromString(tc.value)
			testStruct := DecimalRangeTestStruct{Value: value}

			err := Validate(&testStruct)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "范围值")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
