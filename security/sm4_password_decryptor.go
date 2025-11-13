package security

import (
	"fmt"

	"github.com/tjfoc/gmsm/sm4"

	"github.com/ilxqx/vef-framework-go/crypto"
)

type Sm4PasswordDecryptor struct {
	cipher crypto.Cipher
}

func NewSm4PasswordDecryptor(key, iv []byte, mode ...crypto.Sm4Mode) (PasswordDecryptor, error) {
	if len(key) != sm4.BlockSize {
		return nil, fmt.Errorf("%w: %d (must be %d bytes)", ErrInvalidSM4KeyLength, len(key), sm4.BlockSize)
	}

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

	cipher, err := crypto.NewSM4(key, iv, mode...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM4CipherFailed, err)
	}

	return &Sm4PasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewSm4PasswordDecryptorFromHex(keyHex, ivHex string, mode ...crypto.Sm4Mode) (PasswordDecryptor, error) {
	cipher, err := crypto.NewSM4FromHex(keyHex, ivHex, mode...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM4CipherFromHexFailed, err)
	}

	return &Sm4PasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewSm4PasswordDecryptorFromBase64(keyBase64, ivBase64 string, mode ...crypto.Sm4Mode) (PasswordDecryptor, error) {
	cipher, err := crypto.NewSM4FromBase64(keyBase64, ivBase64, mode...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM4CipherFromBase64Failed, err)
	}

	return &Sm4PasswordDecryptor{
		cipher: cipher,
	}, nil
}

func (d *Sm4PasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
