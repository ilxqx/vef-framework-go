package crypto

import (
	"crypto/cipher"
	"fmt"

	"github.com/tjfoc/gmsm/sm4"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

// Sm4Mode defines the SM4 encryption mode.
type Sm4Mode string

const (
	// Sm4ModeCBC uses SM4-CBC mode with PKCS7 padding.
	Sm4ModeCBC Sm4Mode = "CBC"
	// Sm4ModeECB uses SM4-ECB mode with PKCS7 padding.
	Sm4ModeECB Sm4Mode = "ECB"
)

// Sm4Cipher implements Cipher interface using SM4 encryption (国密算法).
type Sm4Cipher struct {
	key  []byte
	iv   []byte // IV for CBC mode, not used in ECB mode
	mode Sm4Mode
}

// NewSM4 creates a new SM4 cipher with the given key, IV, and optional mode.
// Key must be 16 bytes (128 bits).
// For CBC mode: IV must be 16 bytes.
// For ECB mode: IV is not used.
// If mode is not specified, defaults to Sm4ModeCBC.
func NewSM4(key, iv []byte, mode ...Sm4Mode) (Cipher, error) {
	if len(key) != sm4.BlockSize {
		return nil, fmt.Errorf("%w: %d bytes (must be %d)", ErrInvalidSM4KeySize, len(key), sm4.BlockSize)
	}

	// Default to CBC mode if not specified
	selectedMode := Sm4ModeCBC
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	if selectedMode == Sm4ModeCBC {
		if len(iv) != sm4.BlockSize {
			return nil, fmt.Errorf("%w: %d bytes (must be %d)", ErrInvalidIVSizeCBC, len(iv), sm4.BlockSize)
		}
	}

	return &Sm4Cipher{
		key:  key,
		iv:   iv,
		mode: selectedMode,
	}, nil
}

// NewSM4FromHex creates a new SM4 cipher from hex-encoded key and IV strings.
// If mode is not specified, defaults to Sm4ModeCBC.
func NewSM4FromHex(keyHex, ivHex string, mode ...Sm4Mode) (Cipher, error) {
	key, err := encoding.FromHex(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from hex: %w", err)
	}

	// Default to CBC mode if not specified
	selectedMode := Sm4ModeCBC
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	var iv []byte
	if selectedMode == Sm4ModeCBC {
		iv, err = encoding.FromHex(ivHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode IV from hex: %w", err)
		}
	}

	return NewSM4(key, iv, selectedMode)
}

// NewSM4FromBase64 creates a new SM4 cipher from base64-encoded key and IV strings.
// If mode is not specified, defaults to Sm4ModeCBC.
func NewSM4FromBase64(keyBase64, ivBase64 string, mode ...Sm4Mode) (Cipher, error) {
	key, err := encoding.FromBase64(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from base64: %w", err)
	}

	// Default to CBC mode if not specified
	selectedMode := Sm4ModeCBC
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	var iv []byte
	if selectedMode == Sm4ModeCBC {
		iv, err = encoding.FromBase64(ivBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode IV from base64: %w", err)
		}
	}

	return NewSM4(key, iv, selectedMode)
}

// Encrypt encrypts the plaintext using SM4 and returns base64-encoded ciphertext.
func (s *Sm4Cipher) Encrypt(plaintext string) (string, error) {
	if s.mode == Sm4ModeECB {
		return s.encryptECB(plaintext)
	}

	return s.encryptCBC(plaintext)
}

// Decrypt decrypts the base64-encoded ciphertext using SM4 and returns plaintext.
func (s *Sm4Cipher) Decrypt(ciphertext string) (string, error) {
	if s.mode == Sm4ModeECB {
		return s.decryptECB(ciphertext)
	}

	return s.decryptCBC(ciphertext)
}

// encryptECB encrypts plaintext using SM4-ECB mode with PKCS7 padding.
func (s *Sm4Cipher) encryptECB(plaintext string) (string, error) {
	// PKCS7 padding
	paddedData := pkcs7Padding([]byte(plaintext), sm4.BlockSize)

	ciphertext, err := sm4.Sm4Ecb(s.key, paddedData, true)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encoding.ToBase64(ciphertext), nil
}

// decryptECB decrypts ciphertext using SM4-ECB mode and removes PKCS7 padding.
func (s *Sm4Cipher) decryptECB(ciphertext string) (string, error) {
	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := sm4.Sm4Ecb(s.key, encryptedData, false)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt: %w", err)
	}

	// Remove PKCS7 padding
	unpaddedData, err := pkcs7Unpadding(plaintext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(unpaddedData), nil
}

// encryptCBC encrypts plaintext using SM4-CBC mode with PKCS7 padding.
func (s *Sm4Cipher) encryptCBC(plaintext string) (string, error) {
	block, err := sm4.NewCipher(s.key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create SM4 cipher: %w", err)
	}

	// PKCS7 padding
	paddedData := pkcs7Padding([]byte(plaintext), sm4.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, s.iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return encoding.ToBase64(ciphertext), nil
}

// decryptCBC decrypts ciphertext using SM4-CBC mode and removes PKCS7 padding.
func (s *Sm4Cipher) decryptCBC(ciphertext string) (string, error) {
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

	// Remove PKCS7 padding
	unpaddedData, err := pkcs7Unpadding(plaintext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to remove padding: %w", err)
	}

	return string(unpaddedData), nil
}
