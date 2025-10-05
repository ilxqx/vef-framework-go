package crypto

import "errors"

var (
	// Common.
	ErrAtLeastOneKeyRequired        = errors.New("at least one of privateKey or publicKey must be provided")
	ErrPublicKeyRequiredForEncrypt  = errors.New("public key is required for encryption")
	ErrPrivateKeyRequiredForDecrypt = errors.New("private key is required for decryption")
	ErrFailedDecodePEMBlock         = errors.New("failed to decode PEM block")
	ErrUnsupportedPEMType           = errors.New("unsupported PEM type")

	// AES.
	ErrInvalidAESKeySize            = errors.New("invalid AES key size")
	ErrInvalidIVSizeCBC             = errors.New("invalid IV size for CBC mode")
	ErrCiphertextNotMultipleOfBlock = errors.New("ciphertext is not a multiple of the block size")
	ErrCiphertextTooShort           = errors.New("ciphertext too short")
	ErrDataEmpty                    = errors.New("data is empty")
	ErrInvalidPadding               = errors.New("invalid padding")

	// RSA.
	ErrNotRSAPrivateKey = errors.New("not an RSA private key")
	ErrNotRSAPublicKey  = errors.New("not an RSA public key")

	// SM4.
	ErrInvalidSM4KeySize = errors.New("invalid SM4 key size")
)
