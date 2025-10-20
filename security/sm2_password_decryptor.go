package security

import (
	"fmt"

	"github.com/tjfoc/gmsm/sm2"

	"github.com/ilxqx/vef-framework-go/crypto"
)

// Sm2PasswordDecryptor implements PasswordDecryptor using SM2 encryption (国密算法).
// The encrypted password should be base64-encoded.
type Sm2PasswordDecryptor struct {
	cipher crypto.Cipher
}

// NewSm2PasswordDecryptor creates a new SM2 password decryptor with the given private key.
func NewSm2PasswordDecryptor(privateKey *sm2.PrivateKey) (PasswordDecryptor, error) {
	if privateKey == nil {
		return nil, ErrPrivateKeyNil
	}

	// Use crypto package's SM2 cipher
	cipher, err := crypto.NewSM2(privateKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM2CipherFailed, err)
	}

	return &Sm2PasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewSm2PasswordDecryptorFromPEM creates a new SM2 password decryptor from PEM-encoded private key.
func NewSm2PasswordDecryptorFromPEM(pemKey []byte) (PasswordDecryptor, error) {
	// Use crypto package's SM2 cipher from PEM
	cipher, err := crypto.NewSM2FromPEM(pemKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM2CipherFromPEMFailed, err)
	}

	return &Sm2PasswordDecryptor{
		cipher: cipher,
	}, nil
}

// NewSm2PasswordDecryptorFromHex creates a new SM2 password decryptor from hex-encoded private key.
func NewSm2PasswordDecryptorFromHex(privateKeyHex string) (PasswordDecryptor, error) {
	// Use crypto package's SM2 cipher from hex
	cipher, err := crypto.NewSM2FromHex(privateKeyHex, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM2CipherFromHexFailed, err)
	}

	return &Sm2PasswordDecryptor{
		cipher: cipher,
	}, nil
}

// Decrypt decrypts the base64-encoded SM2-encrypted password.
// The encrypted password is expected to be in the format: base64(SM2(plaintext)).
func (d *Sm2PasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
