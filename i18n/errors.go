package i18n

import "errors"

var (
	// ErrUnsupportedLanguage is returned when an unsupported language code is provided.
	ErrUnsupportedLanguage = errors.New("unsupported language code")
	// ErrMessageIdEmpty is returned when a translation message ID is empty.
	ErrMessageIdEmpty = errors.New("messageId cannot be empty")
)
