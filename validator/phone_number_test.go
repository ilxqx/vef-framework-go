package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type PhoneTestStruct struct {
	PhoneNumber string `validate:"phone_number" label:"手机号"`
}

func TestPhoneNumberValidation(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		shouldErr bool
	}{
		// Valid phone numbers
		{"Valid phone - starts with 13", "13888888888", false},
		{"Valid phone - starts with 15", "15999999999", false},
		{"Valid phone - starts with 18", "18666666666", false},
		{"Valid phone - starts with 19", "19777777777", false},

		// Invalid phone numbers
		{"Invalid - starts with 12", "12888888888", true},
		{"Invalid - starts with 10", "10888888888", true},
		{"Invalid - too short", "1388888888", true},
		{"Invalid - too long", "138888888888", true},
		{"Invalid - contains letters", "13888888a88", true},
		{"Invalid - contains special chars", "1388888-888", true},
		{"Invalid - empty string", "", true},
		{"Invalid - starts with 0", "01888888888", true},
		{"Invalid - starts with 2", "21888888888", true},

		// Edge cases
		{"Valid - all same digits", "13333333333", false},
		{"Valid - starts with 14", "14888888888", false},
		{"Valid - starts with 16", "16888888888", false},
		{"Valid - starts with 17", "17888888888", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testStruct := PhoneTestStruct{PhoneNumber: tc.phone}

			err := Validate(&testStruct)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "手机号")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPhoneNumberValidationWithLabelCheck(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		shouldErr bool
	}{
		{"Valid 13x", "13012345678", false},
		{"Valid 19x", "19812345678", false},
		{"Invalid 12x", "12012345678", true},
		{"Invalid length", "130123456", true},
		{"Non-numeric", "abc12345678", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testStruct := PhoneTestStruct{PhoneNumber: tc.phone}
			err := Validate(&testStruct)

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "手机号")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
