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

// RSAMode defines the RSA encryption mode
type RSAMode string

const (
	// RSAModeOAEP uses RSA-OAEP mode with SHA-256 (recommended)
	RSAModeOAEP RSAMode = "OAEP"
	// RSAModePKCS1v15 uses RSA-PKCS1v15 mode (legacy, less secure)
	RSAModePKCS1v15 RSAMode = "PKCS1v15"
)

// RSACipher implements Cipher interface using RSA encryption
type RSACipher struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	mode       RSAMode
}

// NewRSA creates a new RSA cipher with the given private and public keys and optional mode.
// For encryption-only operations, privateKey can be nil.
// For decryption-only operations, publicKey can be nil.
// If mode is not specified, defaults to RSAModeOAEP (recommended).
func NewRSA(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, mode ...RSAMode) (Cipher, error) {
	if privateKey == nil && publicKey == nil {
		return nil, fmt.Errorf("at least one of privateKey or publicKey must be provided")
	}

	// Default to OAEP mode if not specified
	selectedMode := RSAModeOAEP
	if len(mode) > 0 {
		selectedMode = mode[0]
	}

	// If only private key is provided, derive public key
	if publicKey == nil && privateKey != nil {
		publicKey = &privateKey.PublicKey
	}

	return &RSACipher{
		privateKey: privateKey,
		publicKey:  publicKey,
		mode:       selectedMode,
	}, nil
}

// NewRSAFromPEM creates a new RSA cipher from PEM-encoded keys.
// Either privatePEM or publicPEM can be nil, but not both.
// If mode is not specified, defaults to RSAModeOAEP (recommended).
func NewRSAFromPEM(privatePEM, publicPEM []byte, mode ...RSAMode) (Cipher, error) {
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
// If mode is not specified, defaults to RSAModeOAEP (recommended).
func NewRSAFromHex(privateKeyHex, publicKeyHex string, mode ...RSAMode) (Cipher, error) {
	var (
		privateKey *rsa.PrivateKey
		publicKey  *rsa.PublicKey
	)

	if privateKeyHex != constants.Empty {
		keyBytes, err := encoding.FromHex(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key from hex: %w", err)
		}

		// Try PKCS1 format first
		if privateKey, err = x509.ParsePKCS1PrivateKey(keyBytes); err != nil {
			// Try PKCS8 format
			key, err2 := x509.ParsePKCS8PrivateKey(keyBytes)
			if err2 != nil {
				return nil, fmt.Errorf("failed to parse private key (tried PKCS1 and PKCS8): %w", err)
			}
			var ok bool
			privateKey, ok = key.(*rsa.PrivateKey)
			if !ok {
				return nil, fmt.Errorf("not an RSA private key")
			}
		}
	}

	if publicKeyHex != constants.Empty {
		keyBytes, err := encoding.FromHex(publicKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key from hex: %w", err)
		}

		// Try PKIX format first
		key, err := x509.ParsePKIXPublicKey(keyBytes)
		if err != nil {
			// Try PKCS1 format
			publicKey, err = x509.ParsePKCS1PublicKey(keyBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key (tried PKIX and PKCS1): %w", err)
			}
		} else {
			var ok bool
			publicKey, ok = key.(*rsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf("not an RSA public key")
			}
		}
	}

	return NewRSA(privateKey, publicKey, mode...)
}

// NewRSAFromBase64 creates a new RSA cipher from base64-encoded DER keys.
// Either privateKeyBase64 or publicKeyBase64 can be empty, but not both.
// If mode is not specified, defaults to RSAModeOAEP (recommended).
func NewRSAFromBase64(privateKeyBase64, publicKeyBase64 string, mode ...RSAMode) (Cipher, error) {
	var (
		privateKey *rsa.PrivateKey
		publicKey  *rsa.PublicKey
	)

	if privateKeyBase64 != constants.Empty {
		keyBytes, err := encoding.FromBase64(privateKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key from base64: %w", err)
		}

		// Try PKCS1 format first
		if privateKey, err = x509.ParsePKCS1PrivateKey(keyBytes); err != nil {
			// Try PKCS8 format
			key, err2 := x509.ParsePKCS8PrivateKey(keyBytes)
			if err2 != nil {
				return nil, fmt.Errorf("failed to parse private key (tried PKCS1 and PKCS8): %w", err)
			}
			var ok bool
			privateKey, ok = key.(*rsa.PrivateKey)
			if !ok {
				return nil, fmt.Errorf("not an RSA private key")
			}
		}
	}

	if publicKeyBase64 != constants.Empty {
		keyBytes, err := encoding.FromBase64(publicKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key from base64: %w", err)
		}

		// Try PKIX format first
		key, err := x509.ParsePKIXPublicKey(keyBytes)
		if err != nil {
			// Try PKCS1 format
			publicKey, err = x509.ParsePKCS1PublicKey(keyBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key (tried PKIX and PKCS1): %w", err)
			}
		} else {
			var ok bool
			publicKey, ok = key.(*rsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf("not an RSA public key")
			}
		}
	}

	return NewRSA(privateKey, publicKey, mode...)
}

// Encrypt encrypts the plaintext using RSA public key and returns base64-encoded ciphertext.
func (r *RSACipher) Encrypt(plaintext string) (string, error) {
	if r.publicKey == nil {
		return constants.Empty, fmt.Errorf("public key is required for encryption")
	}

	var (
		ciphertext []byte
		err        error
	)

	if r.mode == RSAModeOAEP {
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
func (r *RSACipher) Decrypt(ciphertext string) (string, error) {
	if r.privateKey == nil {
		return constants.Empty, fmt.Errorf("private key is required for decryption")
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	var plaintext []byte
	if r.mode == RSAModeOAEP {
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

// parseRSAPrivateKeyFromPEM parses RSA private key from PEM-encoded data
func parseRSAPrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
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
			return nil, fmt.Errorf("not an RSA private key")
		}
		return rsaKey, nil
	}

	return nil, fmt.Errorf("unsupported PEM type: %s", block.Type)
}

// parseRSAPublicKeyFromPEM parses RSA public key from PEM-encoded data
func parseRSAPublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKIX format
	if block.Type == "PUBLIC KEY" {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA public key")
		}
		return rsaKey, nil
	}

	// Try PKCS1 format
	if block.Type == "RSA PUBLIC KEY" {
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}

	return nil, fmt.Errorf("unsupported PEM type: %s", block.Type)
}
