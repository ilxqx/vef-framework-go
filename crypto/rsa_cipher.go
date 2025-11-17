package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

type RsaMode string

const (
	RsaModeOaep     RsaMode = "OAEP"
	RsaModePkcs1v15 RsaMode = "PKCS1v15"
)

type RsaSignMode string

const (
	RsaSignModePss      RsaSignMode = "PSS"
	RsaSignModePkcs1v15 RsaSignMode = "PKCS1v15"
)

type rsaCipher struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	mode       RsaMode
	signMode   RsaSignMode
}

type RsaOption func(*rsaCipher)

func WithRsaMode(mode RsaMode) RsaOption {
	return func(c *rsaCipher) {
		c.mode = mode
	}
}

func WithRsaSignMode(signMode RsaSignMode) RsaOption {
	return func(c *rsaCipher) {
		c.signMode = signMode
	}
}

func NewRsa(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, opts ...RsaOption) (CipherSigner, error) {
	if privateKey == nil && publicKey == nil {
		return nil, ErrAtLeastOneKeyRequired
	}

	cipher := &rsaCipher{
		privateKey: privateKey,
		publicKey:  publicKey,
		mode:       RsaModeOaep,
		signMode:   RsaSignModePss,
	}

	for _, opt := range opts {
		opt(cipher)
	}

	if publicKey == nil && privateKey != nil {
		cipher.publicKey = &privateKey.PublicKey
	}

	return cipher, nil
}

func parseRsaKeysFromBytes(privateKeyBytes, publicKeyBytes []byte) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	var (
		privateKey *rsa.PrivateKey
		publicKey  *rsa.PublicKey
		err        error
	)

	if len(privateKeyBytes) > 0 {
		if privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBytes); err != nil {
			if key, err := x509.ParsePKCS8PrivateKey(privateKeyBytes); err != nil {
				return nil, nil, fmt.Errorf("failed to parse private key (tried PKCS1 and PKCS8): %w", err)
			} else {
				var ok bool
				if privateKey, ok = key.(*rsa.PrivateKey); !ok {
					return nil, nil, ErrNotRsaPrivateKey
				}
			}
		}
	}

	if len(publicKeyBytes) > 0 {
		if key, err := x509.ParsePKIXPublicKey(publicKeyBytes); err != nil {
			if publicKey, err = x509.ParsePKCS1PublicKey(publicKeyBytes); err != nil {
				return nil, nil, fmt.Errorf("failed to parse public key (tried PKIX and PKCS1): %w", err)
			}
		} else {
			var ok bool
			if publicKey, ok = key.(*rsa.PublicKey); !ok {
				return nil, nil, ErrNotRsaPublicKey
			}
		}
	}

	return privateKey, publicKey, nil
}

func NewRsaFromPem(privatePem, publicPem []byte, opts ...RsaOption) (CipherSigner, error) {
	var (
		privateKey *rsa.PrivateKey
		publicKey  *rsa.PublicKey
		err        error
	)

	if privatePem != nil {
		if privateKey, err = parseRsaPrivateKeyFromPem(privatePem); err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicPem != nil {
		if publicKey, err = parseRsaPublicKeyFromPem(publicPem); err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewRsa(privateKey, publicKey, opts...)
}

func NewRsaFromHex(privateKeyHex, publicKeyHex string, opts ...RsaOption) (CipherSigner, error) {
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

	privateKey, publicKey, err := parseRsaKeysFromBytes(privateBytes, publicBytes)
	if err != nil {
		return nil, err
	}

	return NewRsa(privateKey, publicKey, opts...)
}

func NewRsaFromBase64(privateKeyBase64, publicKeyBase64 string, opts ...RsaOption) (CipherSigner, error) {
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

	privateKey, publicKey, err := parseRsaKeysFromBytes(privateBytes, publicBytes)
	if err != nil {
		return nil, err
	}

	return NewRsa(privateKey, publicKey, opts...)
}

func (r *rsaCipher) Encrypt(plaintext string) (string, error) {
	if r.publicKey == nil {
		return constants.Empty, ErrPublicKeyRequiredForEncrypt
	}

	var (
		ciphertext []byte
		err        error
	)

	if r.mode == RsaModeOaep {
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

func (r *rsaCipher) Decrypt(ciphertext string) (string, error) {
	if r.privateKey == nil {
		return constants.Empty, ErrPrivateKeyRequiredForDecrypt
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	var plaintext []byte
	if r.mode == RsaModeOaep {
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

func parseRsaPrivateKeyFromPem(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePemBlock
	}

	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrNotRsaPrivateKey
		}

		return rsaKey, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedPemType, block.Type)
}

func parseRsaPublicKeyFromPem(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePemBlock
	}

	if block.Type == "PUBLIC KEY" {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, ErrNotRsaPublicKey
		}

		return rsaKey, nil
	}

	if block.Type == "RSA PUBLIC KEY" {
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedPemType, block.Type)
}

func (r *rsaCipher) Sign(data string) (string, error) {
	if r.privateKey == nil {
		return constants.Empty, ErrPrivateKeyRequiredForSign
	}

	hash := sha256.New()
	_, _ = hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	var (
		signature []byte
		err       error
	)

	if r.signMode == RsaSignModePss {
		signature, err = rsa.SignPSS(rand.Reader, r.privateKey, crypto.SHA256, hashed, nil)
	} else {
		signature, err = rsa.SignPKCS1v15(rand.Reader, r.privateKey, crypto.SHA256, hashed)
	}

	if err != nil {
		return constants.Empty, fmt.Errorf("failed to sign: %w", err)
	}

	return encoding.ToBase64(signature), nil
}

func (r *rsaCipher) Verify(data, signature string) (bool, error) {
	if r.publicKey == nil {
		return false, ErrPublicKeyRequiredForVerify
	}

	signatureBytes, err := encoding.FromBase64(signature)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	hash := sha256.New()
	_, _ = hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	if r.signMode == RsaSignModePss {
		err = rsa.VerifyPSS(r.publicKey, crypto.SHA256, hashed, signatureBytes, nil)
	} else {
		err = rsa.VerifyPKCS1v15(r.publicKey, crypto.SHA256, hashed, signatureBytes)
	}

	return err == nil, nil
}
