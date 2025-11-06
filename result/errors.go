package result

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/i18n"
)

var (
	// ErrTokenExpired is the error for expired token.
	ErrTokenExpired = Err(
		i18n.T(ErrMessageTokenExpired),
		WithCode(ErrCodeTokenExpired),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTokenInvalid is the error for invalid token.
	ErrTokenInvalid = Err(
		i18n.T(ErrMessageTokenInvalid),
		WithCode(ErrCodeTokenInvalid),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTokenNotValidYet is the error for not valid yet token.
	ErrTokenNotValidYet = Err(
		i18n.T(ErrMessageTokenNotValidYet),
		WithCode(ErrCodeTokenNotValidYet),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTokenInvalidIssuer is the error for invalid issuer token.
	ErrTokenInvalidIssuer = Err(
		i18n.T(ErrMessageTokenInvalidIssuer),
		WithCode(ErrCodeTokenInvalidIssuer),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTokenInvalidAudience is the error for invalid audience token.
	ErrTokenInvalidAudience = Err(
		i18n.T(ErrMessageTokenInvalidAudience),
		WithCode(ErrCodeTokenInvalidAudience),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTokenMissingSubject is the error for missing subject token.
	ErrTokenMissingSubject = Err(
		i18n.T(ErrMessageTokenMissingSubject),
		WithCode(ErrCodeTokenMissingSubject),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTokenMissingTokenType is the error for missing token type token.
	ErrTokenMissingTokenType = Err(
		i18n.T(ErrMessageTokenMissingTokenType),
		WithCode(ErrCodeTokenMissingTokenType),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrAppIdRequired is the error for missing app id.
	ErrAppIdRequired = Err(
		i18n.T(ErrMessageAppIdRequired),
		WithCode(ErrCodeAppIdRequired),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTimestampRequired is the error for missing app timestamp.
	ErrTimestampRequired = Err(
		i18n.T(ErrMessageTimestampRequired),
		WithCode(ErrCodeTimestampRequired),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrSignatureRequired is the error for missing app signature.
	ErrSignatureRequired = Err(
		i18n.T(ErrMessageSignatureRequired),
		WithCode(ErrCodeSignatureRequired),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrTimestampInvalid is the error for invalid app timestamp.
	ErrTimestampInvalid = Err(
		i18n.T(ErrMessageTimestampInvalid),
		WithCode(ErrCodeTimestampInvalid),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrSignatureExpired is the error for expired app signature.
	ErrSignatureExpired = Err(
		i18n.T(ErrMessageSignatureExpired),
		WithCode(ErrCodeSignatureExpired),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrSignatureInvalid is the error for invalid app signature.
	ErrSignatureInvalid = Err(
		i18n.T(ErrMessageSignatureInvalid),
		WithCode(ErrCodeSignatureInvalid),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrExternalAppNotFound is the error for app not found.
	ErrExternalAppNotFound = Err(
		i18n.T(ErrMessageExternalAppNotFound),
		WithCode(ErrCodeExternalAppNotFound),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrExternalAppDisabled is the error for app disabled.
	ErrExternalAppDisabled = Err(
		i18n.T(ErrMessageExternalAppDisabled),
		WithCode(ErrCodeExternalAppDisabled),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrIpNotAllowed is the error for ip not allowed.
	ErrIpNotAllowed = Err(
		i18n.T(ErrMessageIpNotAllowed),
		WithCode(ErrCodeIpNotAllowed),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrUnauthenticated is the error for unauthenticated.
	ErrUnauthenticated = Err(
		i18n.T(ErrMessageUnauthenticated),
		WithCode(ErrCodeUnauthenticated),
		WithStatus(fiber.StatusUnauthorized),
	)
	// ErrAccessDenied is the error for access denied.
	ErrAccessDenied = Err(
		i18n.T(ErrMessageAccessDenied),
		WithCode(ErrCodeAccessDenied),
		WithStatus(fiber.StatusForbidden),
	)
	// ErrUnknown is the error for unknown error.
	ErrUnknown = Err(
		i18n.T(ErrMessageUnknown),
		WithCode(ErrCodeUnknown),
		WithStatus(fiber.StatusInternalServerError),
	)
	// ErrRecordNotFound is the error for record not found.
	ErrRecordNotFound = Err(
		i18n.T(ErrMessageRecordNotFound),
		WithCode(ErrCodeRecordNotFound),
	)
	// ErrRecordAlreadyExists is the error for record already exists.
	ErrRecordAlreadyExists = Err(
		i18n.T(ErrMessageRecordAlreadyExists),
		WithCode(ErrCodeRecordAlreadyExists),
	)
	// ErrForeignKeyViolation is the error for foreign key constraint violation.
	ErrForeignKeyViolation = Err(
		i18n.T(ErrMessageForeignKeyViolation),
		WithCode(ErrCodeForeignKeyViolation),
	)
)
