package result

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/internal/i18n"
)

var (
	ErrTokenExpired          Error // ErrTokenExpired is the error for expired token
	ErrTokenInvalid          Error // ErrTokenInvalid is the error for invalid token
	ErrTokenNotValidYet      Error // ErrTokenNotValidYet is the error for not valid yet token
	ErrTokenInvalidIssuer    Error // ErrTokenInvalidIssuer is the error for invalid issuer token
	ErrTokenInvalidAudience  Error // ErrTokenInvalidAudience is the error for invalid audience token
	ErrTokenMissingSubject   Error // ErrTokenMissingSubject is the error for missing subject token
	ErrTokenMissingTokenType Error // ErrTokenMissingTokenType is the error for missing token type token
	ErrAppIdRequired         Error // ErrAppIdRequired is the error for missing app id
	ErrTimestampRequired     Error // ErrTimestampRequired is the error for missing app timestamp
	ErrSignatureRequired     Error // ErrSignatureRequired is the error for missing app signature
	ErrTimestampInvalid      Error // ErrTimestampInvalid is the error for invalid app timestamp
	ErrSignatureExpired      Error // ErrSignatureExpired is the error for expired app signature
	ErrSignatureInvalid      Error // ErrSignatureInvalid is the error for invalid app signature
	ErrExternalAppNotFound   Error // ErrExternalAppNotFound is the error for app not found
	ErrExternalAppDisabled   Error // ErrExternalAppDisabled is the error for app disabled
	ErrIpNotAllowed          Error // ErrIpNotAllowed is the error for ip not allowed
	ErrUnknown               Error // ErrUnknown is the error for unknown error
	ErrRecordNotFound        Error // ErrRecordNotFound is the error for record not found
	ErrRecordAlreadyExists   Error // ErrRecordAlreadyExists is the error for record already exists
)

func initErrors(translator i18n.Translator) (err error) {
	messageIds := []string{
		ErrMessageTokenExpired,
		ErrMessageTokenInvalid,
		ErrMessageTokenNotValidYet,
		ErrMessageTokenInvalidIssuer,
		ErrMessageTokenInvalidAudience,
		ErrMessageTokenMissingSubject,
		ErrMessageTokenMissingTokenType,
		ErrMessageAppIdRequired,
		ErrMessageTimestampRequired,
		ErrMessageSignatureRequired,
		ErrMessageTimestampInvalid,
		ErrMessageSignatureExpired,
		ErrMessageSignatureInvalid,
		ErrMessageExternalAppNotFound,
		ErrMessageExternalAppDisabled,
		ErrMessageIpNotAllowed,
		ErrMessageUnknown,
		ErrMessageRecordNotFound,
		ErrMessageRecordAlreadyExists,
	}

	for i := range len(messageIds) {
		if messageIds[i], err = translator.TE(messageIds[i]); err != nil {
			return err
		}
	}

	ErrTokenExpired = ErrWithCodeAndStatus(ErrCodeTokenExpired, fiber.StatusUnauthorized, messageIds[0])
	ErrTokenInvalid = ErrWithCodeAndStatus(ErrCodeTokenInvalid, fiber.StatusUnauthorized, messageIds[1])
	ErrTokenNotValidYet = ErrWithCodeAndStatus(ErrCodeTokenNotValidYet, fiber.StatusUnauthorized, messageIds[2])
	ErrTokenInvalidIssuer = ErrWithCodeAndStatus(ErrCodeTokenInvalidIssuer, fiber.StatusUnauthorized, messageIds[3])
	ErrTokenInvalidAudience = ErrWithCodeAndStatus(ErrCodeTokenInvalidAudience, fiber.StatusUnauthorized, messageIds[4])
	ErrTokenMissingSubject = ErrWithCodeAndStatus(ErrCodeTokenMissingSubject, fiber.StatusUnauthorized, messageIds[5])
	ErrTokenMissingTokenType = ErrWithCodeAndStatus(ErrCodeTokenMissingTokenType, fiber.StatusUnauthorized, messageIds[6])
	ErrAppIdRequired = ErrWithCodeAndStatus(ErrCodeAppIdRequired, fiber.StatusUnauthorized, messageIds[7])
	ErrTimestampRequired = ErrWithCodeAndStatus(ErrCodeTimestampRequired, fiber.StatusUnauthorized, messageIds[8])
	ErrSignatureRequired = ErrWithCodeAndStatus(ErrCodeSignatureRequired, fiber.StatusUnauthorized, messageIds[9])
	ErrTimestampInvalid = ErrWithCodeAndStatus(ErrCodeTimestampInvalid, fiber.StatusUnauthorized, messageIds[10])
	ErrSignatureExpired = ErrWithCodeAndStatus(ErrCodeSignatureExpired, fiber.StatusUnauthorized, messageIds[11])
	ErrSignatureInvalid = ErrWithCodeAndStatus(ErrCodeSignatureInvalid, fiber.StatusUnauthorized, messageIds[12])
	ErrExternalAppNotFound = ErrWithCodeAndStatus(ErrCodeExternalAppNotFound, fiber.StatusUnauthorized, messageIds[13])
	ErrExternalAppDisabled = ErrWithCodeAndStatus(ErrCodeExternalAppDisabled, fiber.StatusUnauthorized, messageIds[14])
	ErrIpNotAllowed = ErrWithCodeAndStatus(ErrCodeIpNotAllowed, fiber.StatusUnauthorized, messageIds[15])
	ErrUnknown = ErrWithCodeAndStatus(ErrCodeUnknown, fiber.StatusInternalServerError, messageIds[16])
	ErrRecordNotFound = ErrWithCode(ErrCodeRecordNotFound, messageIds[17])
	ErrRecordAlreadyExists = ErrWithCode(ErrCodeRecordAlreadyExists, messageIds[18])
	return nil
}
