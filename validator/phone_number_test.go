package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoneNumberValidation(t *testing.T) {
	type testStruct struct {
		PhoneNumber string `validate:"phone_number" label:"手机号"`
	}

	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"validPhoneStartsWith13", "13888888888", false},
		{"validPhoneStartsWith15", "15999999999", false},
		{"validPhoneStartsWith18", "18666666666", false},
		{"validPhoneStartsWith19", "19777777777", false},
		{"validPhoneStartsWith14", "14888888888", false},
		{"validPhoneStartsWith16", "16888888888", false},
		{"validPhoneStartsWith17", "17888888888", false},
		{"validPhoneAllSameDigits", "13333333333", false},
		{"invalidStartsWith12", "12888888888", true},
		{"invalidStartsWith10", "10888888888", true},
		{"invalidStartsWith0", "01888888888", true},
		{"invalidStartsWith2", "21888888888", true},
		{"invalidTooShort", "1388888888", true},
		{"invalidTooLong", "138888888888", true},
		{"invalidContainsLetters", "13888888a88", true},
		{"invalidContainsSpecialChars", "1388888-888", true},
		{"invalidEmptyString", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{PhoneNumber: tt.phone}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error for phone: %q", tt.phone)
				assert.Contains(t, err.Error(), "手机号", "Error message should contain label")
			} else {
				assert.NoError(t, err, "Should not return validation error for phone: %q", tt.phone)
			}
		})
	}
}
