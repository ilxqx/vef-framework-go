package crypto

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

type sm2Cipher struct {
	privateKey *sm2.PrivateKey
	publicKey  *sm2.PublicKey
}

func NewSm2(privateKey *sm2.PrivateKey, publicKey *sm2.PublicKey) (CipherSigner, error) {
	if privateKey == nil && publicKey == nil {
		return nil, ErrAtLeastOneKeyRequired
	}

	if publicKey == nil && privateKey != nil {
		publicKey = &privateKey.PublicKey
	}

	return &sm2Cipher{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func NewSm2FromPem(privatePem, publicPem []byte) (CipherSigner, error) {
	var (
		privateKey *sm2.PrivateKey
		publicKey  *sm2.PublicKey
		err        error
	)

	if privatePem != nil {
		privateKey, err = parseSm2PrivateKeyFromPem(privatePem)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicPem != nil {
		publicKey, err = parseSm2PublicKeyFromPem(publicPem)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewSm2(privateKey, publicKey)
}

func NewSm2FromHex(privateKeyHex, publicKeyHex string) (CipherSigner, error) {
	var (
		privateKey *sm2.PrivateKey
		publicKey  *sm2.PublicKey
	)

	if privateKeyHex != constants.Empty {
		if keyBytes, err := encoding.FromHex(privateKeyHex); err != nil {
			return nil, fmt.Errorf("failed to decode private key from hex: %w", err)
		} else if privateKey, err = x509.ParseSm2PrivateKey(keyBytes); err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicKeyHex != constants.Empty {
		if keyBytes, err := encoding.FromHex(publicKeyHex); err != nil {
			return nil, fmt.Errorf("failed to decode public key from hex: %w", err)
		} else if publicKey, err = x509.ParseSm2PublicKey(keyBytes); err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewSm2(privateKey, publicKey)
}

func NewSm2FromBase64(privateKeyBase64, publicKeyBase64 string) (CipherSigner, error) {
	var (
		privateKey *sm2.PrivateKey
		publicKey  *sm2.PublicKey
	)

	if privateKeyBase64 != constants.Empty {
		if keyBytes, err := encoding.FromBase64(privateKeyBase64); err != nil {
			return nil, fmt.Errorf("failed to decode private key from base64: %w", err)
		} else if privateKey, err = x509.ParseSm2PrivateKey(keyBytes); err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicKeyBase64 != constants.Empty {
		if keyBytes, err := encoding.FromBase64(publicKeyBase64); err != nil {
			return nil, fmt.Errorf("failed to decode public key from base64: %w", err)
		} else if publicKey, err = x509.ParseSm2PublicKey(keyBytes); err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewSm2(privateKey, publicKey)
}

func (s *sm2Cipher) Encrypt(plaintext string) (string, error) {
	if s.publicKey == nil {
		return constants.Empty, ErrPublicKeyRequiredForEncrypt
	}

	ciphertext, err := sm2.Encrypt(s.publicKey, []byte(plaintext), rand.Reader, sm2.C1C3C2)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encoding.ToBase64(ciphertext), nil
}

func (s *sm2Cipher) Decrypt(ciphertext string) (string, error) {
	if s.privateKey == nil {
		return constants.Empty, ErrPrivateKeyRequiredForDecrypt
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := sm2.Decrypt(s.privateKey, encryptedData, sm2.C1C3C2)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func parseSm2PrivateKeyFromPem(pemData []byte) (*sm2.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePemBlock
	}

	return x509.ParseSm2PrivateKey(block.Bytes)
}

func parseSm2PublicKeyFromPem(pemData []byte) (*sm2.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePemBlock
	}

	return x509.ParseSm2PublicKey(block.Bytes)
}

func (s *sm2Cipher) Sign(data string) (string, error) {
	if s.privateKey == nil {
		return constants.Empty, ErrPrivateKeyRequiredForSign
	}

	signature, err := s.privateKey.Sign(rand.Reader, []byte(data), nil)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to sign: %w", err)
	}

	return encoding.ToBase64(signature), nil
}

func (s *sm2Cipher) Verify(data, signature string) (bool, error) {
	if s.publicKey == nil {
		return false, ErrPublicKeyRequiredForVerify
	}

	signatureBytes, err := encoding.FromBase64(signature)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	valid := s.publicKey.Verify([]byte(data), signatureBytes)

	return valid, nil
}
