package security

import "errors"

var (
	// ErrInvalidAESKeyLength indicates the AES key length is invalid.
	ErrInvalidAESKeyLength = errors.New("invalid AES key length")
	// ErrCannotDeriveIV indicates IV cannot be derived from the key due to insufficient length.
	ErrCannotDeriveIV = errors.New("cannot derive IV")
	// ErrInvalidIVLength indicates the IV length is invalid.
	ErrInvalidIVLength = errors.New("invalid IV length")
	// ErrCreateAESCipherFailed indicates AES cipher creation failed.
	ErrCreateAESCipherFailed = errors.New("failed to create AES cipher")

	// ErrPrivateKeyNil indicates the provided private key is nil.
	ErrPrivateKeyNil = errors.New("private key cannot be nil")
	// ErrCreateRSACipherFailed indicates RSA cipher creation failed.
	ErrCreateRSACipherFailed = errors.New("failed to create RSA cipher")
	// ErrCreateRSACipherFromPEMFailed indicates RSA cipher creation from PEM failed.
	ErrCreateRSACipherFromPEMFailed = errors.New("failed to create RSA cipher from PEM")
	// ErrCreateRSACipherFromHexFailed indicates RSA cipher creation from hex failed.
	ErrCreateRSACipherFromHexFailed = errors.New("failed to create RSA cipher from hex")
	// ErrCreateRSACipherFromBase64Failed indicates RSA cipher creation from base64 failed.
	ErrCreateRSACipherFromBase64Failed = errors.New("failed to create RSA cipher from base64")

	// ErrCreateSM2CipherFailed indicates SM2 cipher creation failed.
	ErrCreateSM2CipherFailed = errors.New("failed to create SM2 cipher")
	// ErrCreateSM2CipherFromPEMFailed indicates SM2 cipher creation from PEM failed.
	ErrCreateSM2CipherFromPEMFailed = errors.New("failed to create SM2 cipher from PEM")
	// ErrCreateSM2CipherFromHexFailed indicates SM2 cipher creation from hex failed.
	ErrCreateSM2CipherFromHexFailed = errors.New("failed to create SM2 cipher from hex")

	// ErrInvalidSM4KeyLength indicates the SM4 key length is invalid.
	ErrInvalidSM4KeyLength = errors.New("invalid SM4 key length")
	// ErrCreateSM4CipherFailed indicates SM4 cipher creation failed.
	ErrCreateSM4CipherFailed = errors.New("failed to create SM4 cipher")
	// ErrCreateSM4CipherFromHexFailed indicates SM4 cipher creation from hex failed.
	ErrCreateSM4CipherFromHexFailed = errors.New("failed to create SM4 cipher from hex")
	// ErrCreateSM4CipherFromBase64Failed indicates SM4 cipher creation from base64 failed.
	ErrCreateSM4CipherFromBase64Failed = errors.New("failed to create SM4 cipher from base64")

	// ErrDecodeJwtSecretFailed indicates decoding Jwt secret failed.
	ErrDecodeJwtSecretFailed = errors.New("failed to decode jwt secret")

	// ErrUserDetailsTypeMustBeStruct indicates user details type must be a struct.
	ErrUserDetailsTypeMustBeStruct = errors.New("user details type must be a struct")
	// ErrExternalAppDetailsTypeMustBeStruct indicates external app details type must be a struct.
	ErrExternalAppDetailsTypeMustBeStruct = errors.New("external app details type must be a struct")

	// ErrQueryNotQueryBuilder is returned when the query does not implement QueryBuilder interface.
	ErrQueryNotQueryBuilder = errors.New("query does not implement QueryBuilder interface")
	// ErrQueryModelNotSet is returned when the query does not have a model set before applying data permission.
	ErrQueryModelNotSet = errors.New("query must call Model() before applying data permission")
)
