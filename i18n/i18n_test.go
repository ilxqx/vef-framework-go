package i18n

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetLanguage tests the SetLanguage function.
func TestSetLanguage(t *testing.T) {
	originalTranslator := translator

	t.Run("SetToChinese", func(t *testing.T) {
		err := SetLanguage("zh-CN")
		require.NoError(t, err)

		msg := T("validator_phone_number")
		t.Logf("Chinese message: %s", msg)
		assert.NotEqual(t, "validator_phone_number", msg, "Translation should succeed")
		assert.Contains(t, msg, "格式", "Should contain Chinese characters")
	})

	t.Run("SetToEnglish", func(t *testing.T) {
		err := SetLanguage("en")
		require.NoError(t, err)

		msg := T("validator_phone_number")
		t.Logf("English message: %s", msg)
		assert.NotEqual(t, "validator_phone_number", msg, "Translation should succeed")
		assert.Contains(t, msg, "format", "Should contain English text")
	})

	t.Run("SetToEmptyStringUsesDefault", func(t *testing.T) {
		originalEnv := os.Getenv("VEF_I18N_LANGUAGE")

		os.Unsetenv("VEF_I18N_LANGUAGE")

		defer func() {
			if originalEnv != "" {
				os.Setenv("VEF_I18N_LANGUAGE", originalEnv)
			}
		}()

		err := SetLanguage("")
		require.NoError(t, err)

		msg := T("ok")
		t.Logf("Default language message: %s", msg)
		assert.NotEqual(t, "ok", msg, "Translation should succeed")
		assert.Contains(t, msg, "成功", "Should use zh-CN as default")
	})

	t.Run("SetToUnsupportedLanguage", func(t *testing.T) {
		err := SetLanguage("fr")
		assert.Error(t, err, "Should return error for unsupported language")
		assert.Contains(t, err.Error(), "unsupported language code", "Error should mention unsupported language")
	})

	translator = originalTranslator
}

// TestGetSupportedLanguages tests the GetSupportedLanguages function.
func TestGetSupportedLanguages(t *testing.T) {
	langs := GetSupportedLanguages()

	assert.NotEmpty(t, langs, "Should return non-empty list")
	assert.Contains(t, langs, "zh-CN", "Should contain zh-CN")
	assert.Contains(t, langs, "en", "Should contain en")

	langs[0] = "modified"
	newLangs := GetSupportedLanguages()
	assert.NotEqual(t, "modified", newLangs[0], "Should return a copy, not the original slice")
}

// TestIsLanguageSupported tests the IsLanguageSupported function.
func TestIsLanguageSupported(t *testing.T) {
	tests := []struct {
		name     string
		langCode string
		want     bool
	}{
		{"Chinese", "zh-CN", true},
		{"English", "en", true},
		{"French", "fr", false},
		{"German", "de", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLanguageSupported(tt.langCode)
			assert.Equal(t, tt.want, got, "IsLanguageSupported(%q) = %v, want %v", tt.langCode, got, tt.want)
		})
	}
}

// TestTranslator tests the global T and TE functions.
func TestTranslator(t *testing.T) {
	err := SetLanguage("zh-CN")
	require.NoError(t, err)

	t.Run("TFunctionWithValidMessageID", func(t *testing.T) {
		msg := T("ok")
		assert.NotEmpty(t, msg, "Should return non-empty message")
		assert.NotEqual(t, "ok", msg, "Should translate the message")
		assert.Contains(t, msg, "成功", "Should contain Chinese translation")
	})

	t.Run("TFunctionWithInvalidMessageID", func(t *testing.T) {
		msg := T("nonexistent.message.key")
		assert.Equal(t, "nonexistent.message.key", msg, "Should return message ID as fallback")
	})

	t.Run("TEFunctionWithValidMessageID", func(t *testing.T) {
		msg, err := Te("ok")
		assert.NoError(t, err, "Should not return error for valid message")
		assert.NotEmpty(t, msg, "Should return non-empty message")
		assert.Contains(t, msg, "成功", "Should contain Chinese translation")
	})

	t.Run("TEFunctionWithInvalidMessageID", func(t *testing.T) {
		msg, err := Te("nonexistent.message.key")
		assert.Error(t, err, "Should return error for nonexistent message")
		assert.Empty(t, msg, "Should return empty message on error")
	})

	t.Run("TEFunctionWithEmptyMessageID", func(t *testing.T) {
		msg, err := Te("")
		assert.Error(t, err, "Should return error for empty message ID")
		assert.Empty(t, msg, "Should return empty message on error")
	})

	t.Cleanup(func() {
		_ = SetLanguage("")
	})
}
