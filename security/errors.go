package security

import "errors"

var (
	ErrInvalidAESKeyLength   = errors.New("invalid AES key length")
	ErrCannotDeriveIV        = errors.New("cannot derive IV")
	ErrInvalidIVLength       = errors.New("invalid IV length")
	ErrCreateAESCipherFailed = errors.New("failed to create AES cipher")

	ErrPrivateKeyNil                   = errors.New("private key cannot be nil")
	ErrCreateRSACipherFailed           = errors.New("failed to create RSA cipher")
	ErrCreateRSACipherFromPEMFailed    = errors.New("failed to create RSA cipher from PEM")
	ErrCreateRSACipherFromHexFailed    = errors.New("failed to create RSA cipher from hex")
	ErrCreateRSACipherFromBase64Failed = errors.New("failed to create RSA cipher from base64")

	ErrCreateSM2CipherFailed        = errors.New("failed to create SM2 cipher")
	ErrCreateSM2CipherFromPEMFailed = errors.New("failed to create SM2 cipher from PEM")
	ErrCreateSM2CipherFromHexFailed = errors.New("failed to create SM2 cipher from hex")

	ErrInvalidSM4KeyLength             = errors.New("invalid SM4 key length")
	ErrCreateSM4CipherFailed           = errors.New("failed to create SM4 cipher")
	ErrCreateSM4CipherFromHexFailed    = errors.New("failed to create SM4 cipher from hex")
	ErrCreateSM4CipherFromBase64Failed = errors.New("failed to create SM4 cipher from base64")

	ErrDecodeJWTSecretFailed = errors.New("failed to decode jwt secret")

	ErrUserDetailsNotStruct        = errors.New("user details type must be a struct or struct pointer")
	ErrExternalAppDetailsNotStruct = errors.New("external app details type must be a struct or struct pointer")

	ErrQueryNotQueryBuilder = errors.New("query does not implement QueryBuilder interface")
	ErrQueryModelNotSet     = errors.New("query must call Model() before applying data permission")
)
