package security

import (
	"crypto/aes"
	"fmt"

	"github.com/ilxqx/vef-framework-go/crypto"
)

// AESPasswordDecryptor implements PasswordDecryptor using AES encryption.
// It supports AES-128, AES-192, and AES-256 based on the key length.
// The encrypted password should be base64-encoded.
type AESPasswordDecryptor struct {
	cipher crypto.Cipher
}

// NewAESPasswordDecryptor creates a new AES password decryptor.
// The key length must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256 respectively.
// The iv (initialization vector) must be 16 bytes for AES block size.
// If iv is nil, it will use the first 16 bytes of the key as IV (not recommended for production).
func NewAESPasswordDecryptor(key []byte, iv []byte) (PasswordDecryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid AES key length: %d (must be 16, 24, or 32 bytes)", len(key))
	}

	// If no IV provided, derive from key (simple fallback, not cryptographically ideal)
	if iv == nil {
		if len(key) >= 16 {
			iv = key[:16]
		} else {
			return nil, fmt.Errorf("cannot derive IV: key too short")
		}
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("invalid IV length: %d (must be %d bytes)", len(iv), aes.BlockSize)
	}

	// Use crypto package's AES cipher with CBC mode
	cipher, err := crypto.NewAES(key, iv, crypto.AESModeCBC)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	return &AESPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewAESPasswordDecryptorFromHex creates a new AES password decryptor from hex-encoded key and IV.
func NewAESPasswordDecryptorFromHex(keyHex, ivHex string) (PasswordDecryptor, error) {
	// Use crypto package's AES cipher from hex
	cipher, err := crypto.NewAESFromHex(keyHex, ivHex, crypto.AESModeCBC)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher from hex: %w", err)
	}

	return &AESPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewAESPasswordDecryptorFromBase64 creates a new AES password decryptor from base64-encoded key and IV.
func NewAESPasswordDecryptorFromBase64(keyBase64, ivBase64 string) (PasswordDecryptor, error) {
	// Use crypto package's AES cipher from base64
	cipher, err := crypto.NewAESFromBase64(keyBase64, ivBase64, crypto.AESModeCBC)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher from base64: %w", err)
	}

	return &AESPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// Decrypt decrypts the base64-encoded AES-encrypted password.
// The encrypted password is expected to be in the format: base64(AES-CBC(plaintext))
func (d *AESPasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
