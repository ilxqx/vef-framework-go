package security

import (
	"crypto/rsa"
	"fmt"

	"github.com/ilxqx/vef-framework-go/crypto"
)

// RsaPasswordDecryptor implements PasswordDecryptor using RSA encryption.
// It uses RSA-OAEP with SHA-256 for decryption.
// The encrypted password should be base64-encoded.
type RsaPasswordDecryptor struct {
	cipher crypto.Cipher
}

// NewRsaPasswordDecryptor creates a new RSA password decryptor with the given private key.
// The privateKey should be in PKCS#1 or PKCS#8 PEM format.
func NewRsaPasswordDecryptor(privateKey *rsa.PrivateKey) (PasswordDecryptor, error) {
	if privateKey == nil {
		return nil, ErrPrivateKeyNil
	}

	// Use crypto package's RSA cipher with OAEP mode (default)
	cipher, err := crypto.NewRSA(privateKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFailed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewRsaPasswordDecryptorFromPEM creates a new RSA password decryptor from PEM-encoded private key.
// The PEM key can be in PKCS#1 (-----BEGIN RSA PRIVATE KEY-----) or
// PKCS#8 (-----BEGIN PRIVATE KEY-----) format.
func NewRsaPasswordDecryptorFromPEM(pemKey []byte) (PasswordDecryptor, error) {
	// Use crypto package's RSA cipher from PEM
	cipher, err := crypto.NewRSAFromPEM(pemKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromPEMFailed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewRsaPasswordDecryptorFromHex creates a new RSA password decryptor from hex-encoded private key.
// The private key can be in PKCS#1 or PKCS#8 DER format.
func NewRsaPasswordDecryptorFromHex(privateKeyHex string) (PasswordDecryptor, error) {
	// Use crypto package's RSA cipher from hex
	cipher, err := crypto.NewRSAFromHex(privateKeyHex, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromHexFailed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewRsaPasswordDecryptorFromBase64 creates a new RSA password decryptor from base64-encoded private key.
// The private key can be in PKCS#1 or PKCS#8 DER format.
func NewRsaPasswordDecryptorFromBase64(privateKeyBase64 string) (PasswordDecryptor, error) {
	// Use crypto package's RSA cipher from base64
	cipher, err := crypto.NewRSAFromBase64(privateKeyBase64, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromBase64Failed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

// Decrypt decrypts the base64-encoded RSA-encrypted password using OAEP with SHA-256.
// The encrypted password is expected to be in the format: base64(RSA-OAEP(plaintext)).
func (d *RsaPasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
