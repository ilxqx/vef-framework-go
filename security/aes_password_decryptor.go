package security

import (
	"crypto/aes"
	"fmt"

	"github.com/ilxqx/vef-framework-go/crypto"
)

type AesPasswordDecryptor struct {
	cipher crypto.Cipher
}

func NewAesPasswordDecryptor(key, iv []byte) (PasswordDecryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("%w: %d (must be 16, 24, or 32 bytes)", ErrInvalidAESKeyLength, len(key))
	}

	if iv == nil {
		if len(key) >= 16 {
			iv = key[:16]
		} else {
			return nil, fmt.Errorf("%w: key too short", ErrCannotDeriveIV)
		}
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("%w: %d (must be %d bytes)", ErrInvalidIVLength, len(iv), aes.BlockSize)
	}

	cipher, err := crypto.NewAES(key, iv, crypto.AesModeCBC)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAESCipherFailed, err)
	}

	return &AesPasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewAesPasswordDecryptorFromHex(keyHex, ivHex string) (PasswordDecryptor, error) {
	cipher, err := crypto.NewAESFromHex(keyHex, ivHex, crypto.AesModeCBC)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAESCipherFailed, err)
	}

	return &AesPasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewAesPasswordDecryptorFromBase64(keyBase64, ivBase64 string) (PasswordDecryptor, error) {
	cipher, err := crypto.NewAESFromBase64(keyBase64, ivBase64, crypto.AesModeCBC)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAESCipherFailed, err)
	}

	return &AesPasswordDecryptor{
		cipher: cipher,
	}, nil
}

func (d *AesPasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
