package security

import (
	"fmt"

	"github.com/tjfoc/gmsm/sm2"

	"github.com/ilxqx/vef-framework-go/crypto"
)

type Sm2PasswordDecryptor struct {
	cipher crypto.Cipher
}

func NewSm2PasswordDecryptor(privateKey *sm2.PrivateKey) (PasswordDecryptor, error) {
	if privateKey == nil {
		return nil, ErrPrivateKeyNil
	}

	cipher, err := crypto.NewSM2(privateKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM2CipherFailed, err)
	}

	return &Sm2PasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewSm2PasswordDecryptorFromPEM(pemKey []byte) (PasswordDecryptor, error) {
	cipher, err := crypto.NewSM2FromPEM(pemKey, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM2CipherFromPEMFailed, err)
	}

	return &Sm2PasswordDecryptor{
		cipher: cipher,
	}, nil
}

func NewSm2PasswordDecryptorFromHex(privateKeyHex string) (PasswordDecryptor, error) {
	cipher, err := crypto.NewSM2FromHex(privateKeyHex, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSM2CipherFromHexFailed, err)
	}

	return &Sm2PasswordDecryptor{
		cipher: cipher,
	}, nil
}

func (d *Sm2PasswordDecryptor) Decrypt(encryptedPassword string) (string, error) {
	return d.cipher.Decrypt(encryptedPassword)
}
