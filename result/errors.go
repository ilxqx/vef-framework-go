package result

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/i18n"
)

var (
	ErrTokenExpired     = ErrWithCodeAndStatus(ErrCodeTokenExpired, fiber.StatusUnauthorized, i18n.T(ErrMessageTokenExpired)) // ErrTokenExpired is the error for expired token
	ErrTokenInvalid     = ErrWithCodeAndStatus(ErrCodeTokenInvalid, fiber.StatusUnauthorized, i18n.T(ErrMessageTokenInvalid)) // ErrTokenInvalid is the error for invalid token
	ErrTokenNotValidYet = ErrWithCodeAndStatus(
		ErrCodeTokenNotValidYet,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageTokenNotValidYet),
	) // ErrTokenNotValidYet is the error for not valid yet token
	ErrTokenInvalidIssuer = ErrWithCodeAndStatus(
		ErrCodeTokenInvalidIssuer,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageTokenInvalidIssuer),
	) // ErrTokenInvalidIssuer is the error for invalid issuer token
	ErrTokenInvalidAudience = ErrWithCodeAndStatus(
		ErrCodeTokenInvalidAudience,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageTokenInvalidAudience),
	) // ErrTokenInvalidAudience is the error for invalid audience token
	ErrTokenMissingSubject = ErrWithCodeAndStatus(
		ErrCodeTokenMissingSubject,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageTokenMissingSubject),
	) // ErrTokenMissingSubject is the error for missing subject token
	ErrTokenMissingTokenType = ErrWithCodeAndStatus(
		ErrCodeTokenMissingTokenType,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageTokenMissingTokenType),
	) // ErrTokenMissingTokenType is the error for missing token type token
	ErrAppIdRequired     = ErrWithCodeAndStatus(ErrCodeAppIdRequired, fiber.StatusUnauthorized, i18n.T(ErrMessageAppIdRequired)) // ErrAppIdRequired is the error for missing app id
	ErrTimestampRequired = ErrWithCodeAndStatus(
		ErrCodeTimestampRequired,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageTimestampRequired),
	) // ErrTimestampRequired is the error for missing app timestamp
	ErrSignatureRequired = ErrWithCodeAndStatus(
		ErrCodeSignatureRequired,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageSignatureRequired),
	) // ErrSignatureRequired is the error for missing app signature
	ErrTimestampInvalid = ErrWithCodeAndStatus(
		ErrCodeTimestampInvalid,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageTimestampInvalid),
	) // ErrTimestampInvalid is the error for invalid app timestamp
	ErrSignatureExpired = ErrWithCodeAndStatus(
		ErrCodeSignatureExpired,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageSignatureExpired),
	) // ErrSignatureExpired is the error for expired app signature
	ErrSignatureInvalid = ErrWithCodeAndStatus(
		ErrCodeSignatureInvalid,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageSignatureInvalid),
	) // ErrSignatureInvalid is the error for invalid app signature
	ErrExternalAppNotFound = ErrWithCodeAndStatus(
		ErrCodeExternalAppNotFound,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageExternalAppNotFound),
	) // ErrExternalAppNotFound is the error for app not found
	ErrExternalAppDisabled = ErrWithCodeAndStatus(
		ErrCodeExternalAppDisabled,
		fiber.StatusUnauthorized,
		i18n.T(ErrMessageExternalAppDisabled),
	) // ErrExternalAppDisabled is the error for app disabled
	ErrIpNotAllowed   = ErrWithCodeAndStatus(ErrCodeIpNotAllowed, fiber.StatusUnauthorized, i18n.T(ErrMessageIpNotAllowed)) // ErrIpNotAllowed is the error for ip not allowed
	ErrUnknown        = ErrWithCodeAndStatus(ErrCodeUnknown, fiber.StatusInternalServerError, i18n.T(ErrMessageUnknown))    // ErrUnknown is the error for unknown error
	ErrRecordNotFound = ErrWithCode(
		ErrCodeRecordNotFound,
		i18n.T(ErrMessageRecordNotFound),
	) // ErrRecordNotFound is the error for record not found
	ErrRecordAlreadyExists = ErrWithCode(
		ErrCodeRecordAlreadyExists,
		i18n.T(ErrMessageRecordAlreadyExists),
	) // ErrRecordAlreadyExists is the error for record already exists
)
