package security

import (
	"crypto/rsa"
	"fmt"

	"github.com/ilxqx/vef-framework-go/crypto"
)

// RSAPasswordDecryptor implements PasswordDecryptor using RSA encryption.
// It uses RSA-OAEP with SHA-256 for decryption.
// The encrypted password should be base64-encoded.
type RSAPasswordDecryptor struct {
	cipher crypto.Cipher
}

// NewRSAPasswordDecryptor creates a new RSA password decryptor with the given private key.
// The privateKey should be in PKCS#1 or PKCS#8 PEM format.
func NewRSAPasswordDecryptor(privateKey *rsa.PrivateKey) (PasswordDecryptor, error) {
	if privateKey == nil {
		return nil, ErrPrivateKeyNil
	}

	// Use crypto package's RSA cipher with OAEP mode (default)
	cipher, err := crypto.NewRSA(privateKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFailed, err)
	}

	return &RSAPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewRSAPasswordDecryptorFromPEM creates a new RSA password decryptor from PEM-encoded private key.
// The PEM key can be in PKCS#1 (-----BEGIN RSA PRIVATE KEY-----) or
// PKCS#8 (-----BEGIN PRIVATE KEY-----) format.
func NewRSAPasswordDecryptorFromPEM(pemKey []byte) (PasswordDecryptor, error) {
	// Use crypto package's RSA cipher from PEM
	cipher, err := crypto.NewRSAFromPEM(pemKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromPEMFailed, err)
	}

	return &RSAPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewRSAPasswordDecryptorFromHex creates a new RSA password decryptor from hex-encoded private key.
// The private key can be in PKCS#1 or PKCS#8 DER format.
func NewRSAPasswordDecryptorFromHex(privateKeyHex string) (PasswordDecryptor, error) {
	// Use crypto package's RSA cipher from hex
	cipher, err := crypto.NewRSAFromHex(privateKeyHex, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromHexFailed, err)
	}

	return &RSAPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewRSAPasswordDecryptorFromBase64 creates a new RSA password decryptor from base64-encoded private key.
// The private key can be in PKCS#1 or PKCS#8 DER format.
func NewRSAPasswordDecryptorFromBase64(privateKeyBase64 string) (PasswordDecryptor, error) {
	// Use crypto package's RSA cipher from base64
	cipher, err := crypto.NewRSAFromBase64(privateKeyBase64, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromBase64Failed, err)
	}

	return &RSAPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// Decrypt decrypts the base64-encoded RSA-encrypted password using OAEP with SHA-256.
// The encrypted password is expected to be in the format: base64(RSA-OAEP(plaintext)).
func (d *RSAPasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
