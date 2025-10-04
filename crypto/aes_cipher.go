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

// AESMode defines the AES encryption mode
type AESMode string

const (
	// AESModeCBC uses AES-CBC mode with PKCS7 padding
	AESModeCBC AESMode = "CBC"
	// AESModeGCM uses AES-GCM mode (authenticated encryption)
	AESModeGCM AESMode = "GCM"
)

// AESCipher implements Cipher interface using AES encryption
type AESCipher struct {
	key  []byte
	iv   []byte // IV for CBC mode, not used in GCM mode
	mode AESMode
}

// NewAES creates a new AES cipher with the given key, IV, and optional mode.
// For CBC mode: key must be 16, 24, or 32 bytes (AES-128, AES-192, AES-256), IV must be 16 bytes.
// For GCM mode: key must be 16, 24, or 32 bytes, IV is not used (GCM generates random nonce).
// If mode is not specified, defaults to AESModeGCM.
func NewAES(key, iv []byte, mode ...AESMode) (Cipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid AES key size: %d bytes (must be 16, 24, or 32)", len(key))
	}

	// Default to GCM mode if not specified
	selectedMode := AESModeGCM
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	if selectedMode == AESModeCBC {
		if len(iv) != aes.BlockSize {
			return nil, fmt.Errorf("invalid IV size for CBC mode: %d bytes (must be %d)", len(iv), aes.BlockSize)
		}
	}

	return &AESCipher{
		key:  key,
		iv:   iv,
		mode: selectedMode,
	}, nil
}

// NewAESFromHex creates a new AES cipher from hex-encoded key and IV strings.
// If mode is not specified, defaults to AESModeGCM.
func NewAESFromHex(keyHex, ivHex string, mode ...AESMode) (Cipher, error) {
	key, err := encoding.FromHex(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from hex: %w", err)
	}

	// Default to GCM mode if not specified
	selectedMode := AESModeGCM
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	var iv []byte
	if selectedMode == AESModeCBC {
		iv, err = encoding.FromHex(ivHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode IV from hex: %w", err)
		}
	}

	return NewAES(key, iv, selectedMode)
}

// NewAESFromBase64 creates a new AES cipher from base64-encoded key and IV strings.
// If mode is not specified, defaults to AESModeGCM.
func NewAESFromBase64(keyBase64, ivBase64 string, mode ...AESMode) (Cipher, error) {
	key, err := encoding.FromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from base64: %w", err)
	}

	// Default to GCM mode if not specified
	selectedMode := AESModeGCM
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	var iv []byte
	if selectedMode == AESModeCBC {
		iv, err = encoding.FromBase64(ivBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode IV from base64: %w", err)
		}
	}

	return NewAES(key, iv, selectedMode)
}

// Encrypt encrypts the plaintext using AES and returns base64-encoded ciphertext.
func (a *AESCipher) Encrypt(plaintext string) (string, error) {
	if a.mode == AESModeGCM {
		return a.encryptGCM(plaintext)
	}
	return a.encryptCBC(plaintext)
}

// Decrypt decrypts the base64-encoded ciphertext using AES and returns plaintext.
func (a *AESCipher) Decrypt(ciphertext string) (string, error) {
	if a.mode == AESModeGCM {
		return a.decryptGCM(ciphertext)
	}
	return a.decryptCBC(ciphertext)
}

// encryptCBC encrypts plaintext using AES-CBC mode with PKCS7 padding
func (a *AESCipher) encryptCBC(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// PKCS7 padding
	paddedData := pkcs7Padding([]byte(plaintext), aes.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, a.iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return encoding.ToBase64(ciphertext), nil
}

// decryptCBC decrypts ciphertext using AES-CBC mode and removes PKCS7 padding
func (a *AESCipher) decryptCBC(ciphertext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(encryptedData)%aes.BlockSize != 0 {
		return constants.Empty, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(encryptedData))
	mode := cipher.NewCBCDecrypter(block, a.iv)
	mode.CryptBlocks(plaintext, encryptedData)

	// Remove PKCS7 padding
	unpaddedData, err := pkcs7Unpadding(plaintext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(unpaddedData), nil
}

// encryptGCM encrypts plaintext using AES-GCM mode (authenticated encryption)
func (a *AESCipher) encryptGCM(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return constants.Empty, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate (nonce is prepended to ciphertext)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return encoding.ToBase64(ciphertext), nil
}

// decryptGCM decrypts ciphertext using AES-GCM mode
func (a *AESCipher) decryptGCM(ciphertext string) (string, error) {
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
		return constants.Empty, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt and verify: %w", err)
	}

	return string(plaintext), nil
}

// pkcs7Padding adds PKCS7 padding to the data
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpadding removes PKCS7 padding from the data
func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding")
	}

	// Verify padding
	for i := range padding {
		if data[length-1-i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:length-padding], nil
}
