package security

import (
	"crypto/rsa"
	"fmt"

	"github.com/ilxqx/vef-framework-go/crypto"
)

type RsaPasswordDecryptor struct {
	cipher crypto.Cipher
}

func NewRsaPasswordDecryptor(privateKey *rsa.PrivateKey) (PasswordDecryptor, error) {
	if privateKey == nil {
		return nil, ErrPrivateKeyNil
	}

	cipher, err := crypto.NewRSA(privateKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFailed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewRsaPasswordDecryptorFromPEM(pemKey []byte) (PasswordDecryptor, error) {
	cipher, err := crypto.NewRSAFromPEM(pemKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromPEMFailed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewRsaPasswordDecryptorFromHex(privateKeyHex string) (PasswordDecryptor, error) {
	cipher, err := crypto.NewRSAFromHex(privateKeyHex, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromHexFailed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewRsaPasswordDecryptorFromBase64(privateKeyBase64 string) (PasswordDecryptor, error) {
	cipher, err := crypto.NewRSAFromBase64(privateKeyBase64, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRSACipherFromBase64Failed, err)
	}

	return &RsaPasswordDecryptor{
		cipher: cipher,
	}, nil
}

func (d *RsaPasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
