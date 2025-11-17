package crypto

import (
	"crypto/cipher"
	"fmt"

	"github.com/tjfoc/gmsm/sm4"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

type Sm4Mode string

const (
	Sm4ModeCbc Sm4Mode = "CBC"
	Sm4ModeEcb Sm4Mode = "ECB"
)

type sm4Cipher struct {
	key  []byte
	iv   []byte
	mode Sm4Mode
}

type Sm4Option func(*sm4Cipher)

func WithSm4Iv(iv []byte) Sm4Option {
	return func(c *sm4Cipher) {
		c.iv = iv
	}
}

func WithSm4Mode(mode Sm4Mode) Sm4Option {
	return func(c *sm4Cipher) {
		c.mode = mode
	}
}

func NewSm4(key []byte, opts ...Sm4Option) (Cipher, error) {
	if len(key) != sm4.BlockSize {
		return nil, fmt.Errorf("%w: %d bytes (must be %d)", ErrInvalidSm4KeySize, len(key), sm4.BlockSize)
	}

	cipher := &sm4Cipher{
		key:  key,
		mode: Sm4ModeCbc,
	}

	for _, opt := range opts {
		opt(cipher)
	}

	if cipher.mode == Sm4ModeCbc {
		if len(cipher.iv) != sm4.BlockSize {
			return nil, fmt.Errorf("%w: %d bytes (must be %d)", ErrInvalidIvSizeCbc, len(cipher.iv), sm4.BlockSize)
		}
	}

	return cipher, nil
}

func NewSm4FromHex(keyHex string, opts ...Sm4Option) (Cipher, error) {
	key, err := encoding.FromHex(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from hex: %w", err)
	}

	return NewSm4(key, opts...)
}

func NewSm4FromBase64(keyBase64 string, opts ...Sm4Option) (Cipher, error) {
	key, err := encoding.FromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from base64: %w", err)
	}

	return NewSm4(key, opts...)
}

func (s *sm4Cipher) Encrypt(plaintext string) (string, error) {
	if s.mode == Sm4ModeEcb {
		return s.encryptEcb(plaintext)
	}

	return s.encryptCbc(plaintext)
}

func (s *sm4Cipher) Decrypt(ciphertext string) (string, error) {
	if s.mode == Sm4ModeEcb {
		return s.decryptEcb(ciphertext)
	}

	return s.decryptCbc(ciphertext)
}

func (s *sm4Cipher) encryptEcb(plaintext string) (string, error) {
	paddedData := pkcs7Padding([]byte(plaintext), sm4.BlockSize)

	ciphertext, err := sm4.Sm4Ecb(s.key, paddedData, true)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encoding.ToBase64(ciphertext), nil
}

func (s *sm4Cipher) decryptEcb(ciphertext string) (string, error) {
	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := sm4.Sm4Ecb(s.key, encryptedData, false)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt: %w", err)
	}

	unpaddedData, err := pkcs7Unpadding(plaintext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(unpaddedData), nil
}

func (s *sm4Cipher) encryptCbc(plaintext string) (string, error) {
	block, err := sm4.NewCipher(s.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create SM4 cipher: %w", err)
	}

	paddedData := pkcs7Padding([]byte(plaintext), sm4.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, s.iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return encoding.ToBase64(ciphertext), nil
}

func (s *sm4Cipher) decryptCbc(ciphertext string) (string, error) {
	block, err := sm4.NewCipher(s.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create SM4 cipher: %w", err)
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(encryptedData)%sm4.BlockSize != 0 {
		return constants.Empty, ErrCiphertextNotMultipleOfBlock
	}

	plaintext := make([]byte, len(encryptedData))
	mode := cipher.NewCBCDecrypter(block, s.iv)
	mode.CryptBlocks(plaintext, encryptedData)

	unpaddedData, err := pkcs7Unpadding(plaintext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(unpaddedData), nil
}
