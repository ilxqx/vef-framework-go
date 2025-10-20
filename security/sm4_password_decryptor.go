package security

import (
	"fmt"

	"github.com/tjfoc/gmsm/sm4"

	"github.com/ilxqx/vef-framework-go/crypto"
)

// Sm4PasswordDecryptor implements PasswordDecryptor using SM4 encryption (国密算法).
// It supports SM4-CBC and SM4-ECB modes.
// The encrypted password should be base64-encoded.
type Sm4PasswordDecryptor struct {
	cipher crypto.Cipher
}

// NewSm4PasswordDecryptor creates a new SM4 password decryptor.
// The key length must be 16 bytes (128 bits).
// The iv (initialization vector) must be 16 bytes for CBC mode.
// If iv is nil, it will use the first 16 bytes of the key as IV (not recommended for production).
// If mode is not specified, defaults to SM4ModeCBC.
func NewSm4PasswordDecryptor(key, iv []byte, mode ...crypto.Sm4Mode) (PasswordDecryptor, error) {
	if len(key) != sm4.BlockSize {
		return nil, fmt.Errorf("%w: %d (must be %d bytes)", ErrInvalidSM4KeyLength, len(key), sm4.BlockSize)
	}

	// If no IV provided, derive from key (simple fallback, not cryptographically ideal)
	if iv == nil {
		if len(key) >= sm4.BlockSize {
			iv = key[:sm4.BlockSize]
		} else {
			return nil, fmt.Errorf("%w: key too short", ErrCannotDeriveIV)
		}
	}

	if len(iv) != sm4.BlockSize {
		return nil, fmt.Errorf("%w: %d (must be %d bytes)", ErrInvalidIVLength, len(iv), sm4.BlockSize)
	}

	// Use crypto package's SM4 cipher with CBC mode (default)
	cipher, err := crypto.NewSM4(key, iv, mode...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM4CipherFailed, err)
	}

	return &Sm4PasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewSm4PasswordDecryptorFromHex creates a new SM4 password decryptor from hex-encoded key and IV.
// If mode is not specified, defaults to SM4ModeCBC.
func NewSm4PasswordDecryptorFromHex(keyHex, ivHex string, mode ...crypto.Sm4Mode) (PasswordDecryptor, error) {
	// Use crypto package's SM4 cipher from hex
	cipher, err := crypto.NewSM4FromHex(keyHex, ivHex, mode...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM4CipherFromHexFailed, err)
	}

	return &Sm4PasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewSm4PasswordDecryptorFromBase64 creates a new SM4 password decryptor from base64-encoded key and IV.
// If mode is not specified, defaults to SM4ModeCBC.
func NewSm4PasswordDecryptorFromBase64(keyBase64, ivBase64 string, mode ...crypto.Sm4Mode) (PasswordDecryptor, error) {
	// Use crypto package's SM4 cipher from base64
	cipher, err := crypto.NewSM4FromBase64(keyBase64, ivBase64, mode...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM4CipherFromBase64Failed, err)
	}

	return &Sm4PasswordDecryptor{
		cipher: cipher,
	}, nil
}

// Decrypt decrypts the base64-encoded SM4-encrypted password.
// The encrypted password is expected to be in the format: base64(SM4-CBC/ECB(plaintext)).
func (d *Sm4PasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
