package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

type AesMode string

const (
	AesModeCbc AesMode = "CBC"
	AesModeGcm AesMode = "GCM"
)

type aesCipher struct {
	key  []byte
	iv   []byte
	mode AesMode
}

type AesOption func(*aesCipher)

func WithAesIv(iv []byte) AesOption {
	return func(c *aesCipher) {
		c.iv = iv
	}
}

func WithAesMode(mode AesMode) AesOption {
	return func(c *aesCipher) {
		c.mode = mode
	}
}

func NewAes(key []byte, opts ...AesOption) (Cipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("%w: %d bytes (must be 16, 24, or 32)", ErrInvalidAesKeySize, len(key))
	}

	cipher := &aesCipher{
		key:  key,
		mode: AesModeGcm,
	}

	for _, opt := range opts {
		opt(cipher)
	}

	if cipher.mode == AesModeCbc {
		if len(cipher.iv) != aes.BlockSize {
			return nil, fmt.Errorf("%w: %d bytes (must be %d)", ErrInvalidIvSizeCbc, len(cipher.iv), aes.BlockSize)
		}
	}

	return cipher, nil
}

func NewAesFromHex(keyHex string, opts ...AesOption) (Cipher, error) {
	key, err := encoding.FromHex(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from hex: %w", err)
	}

	return NewAes(key, opts...)
}

func NewAesFromBase64(keyBase64 string, opts ...AesOption) (Cipher, error) {
	key, err := encoding.FromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from base64: %w", err)
	}

	return NewAes(key, opts...)
}

func (a *aesCipher) Encrypt(plaintext string) (string, error) {
	if a.mode == AesModeGcm {
		return a.encryptGcm(plaintext)
	}

	return a.encryptCbc(plaintext)
}

func (a *aesCipher) Decrypt(ciphertext string) (string, error) {
	if a.mode == AesModeGcm {
		return a.decryptGcm(ciphertext)
	}

	return a.decryptCbc(ciphertext)
}

func (a *aesCipher) encryptCbc(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	paddedData := pkcs7Padding([]byte(plaintext), aes.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, a.iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return encoding.ToBase64(ciphertext), nil
}

func (a *aesCipher) decryptCbc(ciphertext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(encryptedData)%aes.BlockSize != 0 {
		return constants.Empty, ErrCiphertextNotMultipleOfBlock
	}

	plaintext := make([]byte, len(encryptedData))
	mode := cipher.NewCBCDecrypter(block, a.iv)
	mode.CryptBlocks(plaintext, encryptedData)

	unpaddedData, err := pkcs7Unpadding(plaintext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(unpaddedData), nil
}

func (a *aesCipher) encryptGcm(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return constants.Empty, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return encoding.ToBase64(ciphertext), nil
}

func (a *aesCipher) decryptGcm(ciphertext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create GCM: %w", err)
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return constants.Empty, ErrCiphertextTooShort
	}

	nonce, ciphertextBytes := encryptedData[:nonceSize], encryptedData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt and verify: %w", err)
	}

	return string(plaintext), nil
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize

	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}

	return append(data, padtext...)
}

func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, ErrDataEmpty
	}

	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, ErrInvalidPadding
	}

	for i := range padding {
		if data[length-1-i] != byte(padding) {
			return nil, ErrInvalidPadding
		}
	}

	return data[:length-padding], nil
}
