package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/i18n"
)

// TestValidatorI18nMessages tests that custom validation rules use i18n translations
// in both Chinese and English.
func TestValidatorI18nMessages(t *testing.T) {
	type TestStruct struct {
		PhoneNumber string `validate:"phone_number" label:"手机号"`
		Action      string `validate:"alphanum_us" label:"操作"`
		Resource    string `validate:"alphanum_us_slash" label:"资源"`
		FileName    string `validate:"alphanum_us_dot" label:"文件名"`
	}

	data := TestStruct{
		PhoneNumber: "invalid",
		Action:      "get/user",        // contains slash - invalid for alphanum_us
		Resource:    "sys.user",        // contains dot - invalid for alphanum_us_slash
		FileName:    "app/config.yaml", // contains slash - invalid for alphanum_us_dot
	}

	// Test with Chinese (default)
	t.Run("Chinese messages", func(t *testing.T) {
		err := i18n.SetLanguage("zh-CN")
		require.NoError(t, err, "Failed to set language to zh-CN")

		err = Validate(&data)
		require.Error(t, err, "Expected validation errors")

		errMsg := err.Error()
		t.Logf("Chinese validation error: %s", errMsg)

		// Chinese messages should contain Chinese characters
		assert.NotEmpty(t, errMsg, "Error message should not be empty")
		// Verify it contains Chinese labels
		assert.True(t, strings.Contains(errMsg, "手机号") ||
			strings.Contains(errMsg, "操作") ||
			strings.Contains(errMsg, "资源") ||
			strings.Contains(errMsg, "文件名"),
			"Error message should contain Chinese labels")
	})

	// Test with English
	t.Run("English messages", func(t *testing.T) {
		err := i18n.SetLanguage("en")
		require.NoError(t, err, "Failed to set language to en")

		err = Validate(&data)
		require.Error(t, err, "Expected validation errors")

		errMsg := err.Error()
		t.Logf("English validation error: %s", errMsg)

		// English messages should be in English (check for English keywords)
		assert.NotEmpty(t, errMsg, "Error message should not be empty")
		assert.True(t,
			strings.Contains(errMsg, "format is invalid") ||
				strings.Contains(errMsg, "can only contain") ||
				strings.Contains(errMsg, "invalid"),
			"Error message should contain English text")
	})

	// Restore to default language
	t.Cleanup(func() {
		_ = i18n.SetLanguage("")
	})
}

// TestValidatorI18nPhoneNumber tests phone number validation with i18n support in both languages.
func TestValidatorI18nPhoneNumber(t *testing.T) {
	type TestStruct struct {
		PhoneNumber string `validate:"phone_number" label:"手机号" label_i18n:"Phone Number"`
	}

	tests := []struct {
		name        string
		phoneNumber string
		wantErr     bool
	}{
		{"Valid phone", "13800138000", false},
		{"Invalid phone", "invalid", true},
		{"Empty phone", "", true},
	}

	languages := []struct {
		code string
		name string
	}{
		{"zh-CN", "Chinese"},
		{"en", "English"},
	}

	for _, lang := range languages {
		t.Run(lang.name, func(t *testing.T) {
			err := i18n.SetLanguage(lang.code)
			require.NoError(t, err, "Failed to set language to %s", lang.code)

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					data := TestStruct{PhoneNumber: tt.phoneNumber}
					err := Validate(&data)

					if tt.wantErr {
						assert.Error(t, err)

						if err != nil {
							t.Logf("[%s] Error message: %s", lang.code, err.Error())
						}
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})
	}

	// Restore to default language
	t.Cleanup(func() {
		_ = i18n.SetLanguage("")
	})
}

// TestValidatorI18nAlphanumRules tests alphanum validation rules with i18n support in both languages.
func TestValidatorI18nAlphanumRules(t *testing.T) {
	type TestStruct struct {
		Action   string `validate:"alphanum_us" label:"操作" label_i18n:"Action"`
		Resource string `validate:"alphanum_us_slash" label:"资源" label_i18n:"Resource"`
		FileName string `validate:"alphanum_us_dot" label:"文件名" label_i18n:"File Name"`
	}

	tests := []struct {
		name     string
		data     TestStruct
		wantErr  bool
		errCount int // expected number of validation errors
	}{
		{
			name: "All valid",
			data: TestStruct{
				Action:   "get_user_info",
				Resource: "sys/user",
				FileName: "config.yaml",
			},
			wantErr:  false,
			errCount: 0,
		},
		{
			name: "Action with slash - invalid",
			data: TestStruct{
				Action:   "get/user",
				Resource: "sys/user",
				FileName: "config.yaml",
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "Resource with dot - invalid",
			data: TestStruct{
				Action:   "get_user",
				Resource: "sys.user",
				FileName: "config.yaml",
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "FileName with slash - invalid",
			data: TestStruct{
				Action:   "get_user",
				Resource: "sys/user",
				FileName: "config/app.yaml",
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "Multiple invalid fields",
			data: TestStruct{
				Action:   "get-user",        // dash invalid
				Resource: "sys.user",        // dot invalid
				FileName: "config/app.yaml", // slash invalid
			},
			wantErr:  true,
			errCount: 3,
		},
	}

	languages := []struct {
		code string
		name string
	}{
		{"zh-CN", "Chinese"},
		{"en", "English"},
	}

	for _, lang := range languages {
		t.Run(lang.name, func(t *testing.T) {
			err := i18n.SetLanguage(lang.code)
			require.NoError(t, err, "Failed to set language to %s", lang.code)

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					err := Validate(&tt.data)

					if tt.wantErr {
						assert.Error(t, err)

						if err != nil {
							t.Logf("[%s] Error message: %s", lang.code, err.Error())
						}
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})
	}

	// Restore to default language
	t.Cleanup(func() {
		_ = i18n.SetLanguage("")
	})
}
