package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

// RsaMode defines the RSA encryption mode.
type RsaMode string

const (
	// RsaModeOAEP uses RSA-OAEP mode with SHA-256 (recommended).
	RsaModeOAEP RsaMode = "OAEP"
	// RsaModePKCS1v15 uses RSA-PKCS1v15 mode (legacy, less secure).
	RsaModePKCS1v15 RsaMode = "PKCS1v15"
)

// RsaCipher implements Cipher interface using RSA encryption.
type RsaCipher struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	mode       RsaMode
}

// NewRSA creates a new RSA cipher with the given private and public keys and optional mode.
// For encryption-only operations, privateKey can be nil.
// For decryption-only operations, publicKey can be nil.
// If mode is not specified, defaults to RsaModeOAEP (recommended).
func NewRSA(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, mode ...RsaMode) (Cipher, error) {
	if privateKey == nil && publicKey == nil {
		return nil, ErrAtLeastOneKeyRequired
	}

	// Default to OAEP mode if not specified
	selectedMode := RsaModeOAEP
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	// If only private key is provided, derive public key
	if publicKey == nil && privateKey != nil {
		publicKey = &privateKey.PublicKey
	}

	return &RsaCipher{
		privateKey: privateKey,
		publicKey:  publicKey,
		mode:       selectedMode,
	}, nil
}

// parseRSAKeysFromBytes tries to parse RSA private/public keys from DER bytes.
// It tries multiple formats for robustness.
func parseRSAKeysFromBytes(privateKeyBytes, publicKeyBytes []byte) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	var (
		privateKey *rsa.PrivateKey
		publicKey  *rsa.PublicKey
		err        error
	)

	if len(privateKeyBytes) > 0 {
		if privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBytes); err != nil {
			key, err2 := x509.ParsePKCS8PrivateKey(privateKeyBytes)
			if err2 != nil {
				return nil, nil, fmt.Errorf("failed to parse private key (tried PKCS1 and PKCS8): %w", err)
			}

			var ok bool

			privateKey, ok = key.(*rsa.PrivateKey)
			if !ok {
				return nil, nil, ErrNotRSAPrivateKey
			}
		}
	}

	if len(publicKeyBytes) > 0 {
		var key any
		if key, err = x509.ParsePKIXPublicKey(publicKeyBytes); err != nil {
			publicKey, err = x509.ParsePKCS1PublicKey(publicKeyBytes)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse public key (tried PKIX and PKCS1): %w", err)
			}
		} else {
			var ok bool

			publicKey, ok = key.(*rsa.PublicKey)
			if !ok {
				return nil, nil, ErrNotRSAPublicKey
			}
		}
	}

	return privateKey, publicKey, nil
}

// NewRSAFromPEM creates a new RSA cipher from PEM-encoded keys.
// Either privatePEM or publicPEM can be nil, but not both.
// If mode is not specified, defaults to RsaModeOAEP (recommended).
func NewRSAFromPEM(privatePEM, publicPEM []byte, mode ...RsaMode) (Cipher, error) {
	var (
		privateKey *rsa.PrivateKey
		publicKey  *rsa.PublicKey
		err        error
	)

	if privatePEM != nil {
		privateKey, err = parseRSAPrivateKeyFromPEM(privatePEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicPEM != nil {
		publicKey, err = parseRSAPublicKeyFromPEM(publicPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewRSA(privateKey, publicKey, mode...)
}

// NewRSAFromHex creates a new RSA cipher from hex-encoded DER keys.
// Either privateKeyHex or publicKeyHex can be empty, but not both.
// If mode is not specified, defaults to RsaModeOAEP (recommended).
func NewRSAFromHex(privateKeyHex, publicKeyHex string, mode ...RsaMode) (Cipher, error) {
	var (
		privateBytes []byte
		publicBytes  []byte
		err          error
	)

	if privateKeyHex != constants.Empty {
		if privateBytes, err = encoding.FromHex(privateKeyHex); err != nil {
			return nil, fmt.Errorf("failed to decode private key from hex: %w", err)
		}
	}

	if publicKeyHex != constants.Empty {
		if publicBytes, err = encoding.FromHex(publicKeyHex); err != nil {
			return nil, fmt.Errorf("failed to decode public key from hex: %w", err)
		}
	}

	privateKey, publicKey, err := parseRSAKeysFromBytes(privateBytes, publicBytes)
	if err != nil {
		return nil, err
	}

	return NewRSA(privateKey, publicKey, mode...)
}

// NewRSAFromBase64 creates a new RSA cipher from base64-encoded DER keys.
// Either privateKeyBase64 or publicKeyBase64 can be empty, but not both.
// If mode is not specified, defaults to RsaModeOAEP (recommended).
func NewRSAFromBase64(privateKeyBase64, publicKeyBase64 string, mode ...RsaMode) (Cipher, error) {
	var (
		privateBytes []byte
		publicBytes  []byte
		err          error
	)

	if privateKeyBase64 != constants.Empty {
		if privateBytes, err = encoding.FromBase64(privateKeyBase64); err != nil {
			return nil, fmt.Errorf("failed to decode private key from base64: %w", err)
		}
	}

	if publicKeyBase64 != constants.Empty {
		if publicBytes, err = encoding.FromBase64(publicKeyBase64); err != nil {
			return nil, fmt.Errorf("failed to decode public key from base64: %w", err)
		}
	}

	privateKey, publicKey, err := parseRSAKeysFromBytes(privateBytes, publicBytes)
	if err != nil {
		return nil, err
	}

	return NewRSA(privateKey, publicKey, mode...)
}

// Encrypt encrypts the plaintext using RSA public key and returns base64-encoded ciphertext.
func (r *RsaCipher) Encrypt(plaintext string) (string, error) {
	if r.publicKey == nil {
		return constants.Empty, ErrPublicKeyRequiredForEncrypt
	}

	var (
		ciphertext []byte
		err        error
	)

	if r.mode == RsaModeOAEP {
		hash := sha256.New()
		ciphertext, err = rsa.EncryptOAEP(hash, rand.Reader, r.publicKey, []byte(plaintext), nil)
	} else {
		ciphertext, err = rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, []byte(plaintext))
	}

	if err != nil {
		return constants.Empty, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encoding.ToBase64(ciphertext), nil
}

// Decrypt decrypts the base64-encoded ciphertext using RSA private key and returns plaintext.
func (r *RsaCipher) Decrypt(ciphertext string) (string, error) {
	if r.privateKey == nil {
		return constants.Empty, ErrPrivateKeyRequiredForDecrypt
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	var plaintext []byte
	if r.mode == RsaModeOAEP {
		hash := sha256.New()
		plaintext, err = rsa.DecryptOAEP(hash, rand.Reader, r.privateKey, encryptedData, nil)
	} else {
		plaintext, err = rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, encryptedData)
	}

	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// parseRSAPrivateKeyFromPEM parses RSA private key from PEM-encoded data.
func parseRSAPrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePEMBlock
	}

	// Try PKCS1 format first
	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	// Try PKCS8 format
	if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrNotRSAPrivateKey
		}

		return rsaKey, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedPEMType, block.Type)
}

// parseRSAPublicKeyFromPEM parses RSA public key from PEM-encoded data.
func parseRSAPublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePEMBlock
	}

	// Try PKIX format
	if block.Type == "PUBLIC KEY" {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, ErrNotRSAPublicKey
		}

		return rsaKey, nil
	}

	// Try PKCS1 format
	if block.Type == "RSA PUBLIC KEY" {
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedPEMType, block.Type)
}
