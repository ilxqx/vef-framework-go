package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

type EciesCurve string

const (
	EciesCurveP256   EciesCurve = "P256"
	EciesCurveP384   EciesCurve = "P384"
	EciesCurveP521   EciesCurve = "P521"
	EciesCurveX25519 EciesCurve = "X25519"
)

type eciesCipher struct {
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
}

type EciesOption func(*eciesCipher)

func NewEcies(privateKey *ecdh.PrivateKey, publicKey *ecdh.PublicKey, opts ...EciesOption) (Cipher, error) {
	if privateKey == nil && publicKey == nil {
		return nil, ErrAtLeastOneKeyRequired
	}

	cipher := &eciesCipher{
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	for _, opt := range opts {
		opt(cipher)
	}

	if publicKey == nil && privateKey != nil {
		cipher.publicKey = privateKey.PublicKey()
	}

	return cipher, nil
}

func NewEciesFromBytes(privateKeyBytes, publicKeyBytes []byte, curve EciesCurve, opts ...EciesOption) (Cipher, error) {
	var (
		privateKey *ecdh.PrivateKey
		publicKey  *ecdh.PublicKey
		err        error
	)

	ecdhCurve := getCurve(curve)

	if len(privateKeyBytes) > 0 {
		if privateKey, err = ecdhCurve.NewPrivateKey(privateKeyBytes); err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if len(publicKeyBytes) > 0 {
		if publicKey, err = ecdhCurve.NewPublicKey(publicKeyBytes); err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewEcies(privateKey, publicKey, opts...)
}

func NewEciesFromHex(privateKeyHex, publicKeyHex string, curve EciesCurve, opts ...EciesOption) (Cipher, error) {
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

	return NewEciesFromBytes(privateBytes, publicBytes, curve, opts...)
}

func NewEciesFromBase64(privateKeyBase64, publicKeyBase64 string, curve EciesCurve, opts ...EciesOption) (Cipher, error) {
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

	return NewEciesFromBytes(privateBytes, publicBytes, curve, opts...)
}

func GenerateEciesKey(curve EciesCurve) (*ecdh.PrivateKey, error) {
	ecdhCurve := getCurve(curve)

	return ecdhCurve.GenerateKey(rand.Reader)
}

func getCurve(curve EciesCurve) ecdh.Curve {
	switch curve {
	case EciesCurveP256:
		return ecdh.P256()
	case EciesCurveP384:
		return ecdh.P384()
	case EciesCurveP521:
		return ecdh.P521()
	case EciesCurveX25519:
		return ecdh.X25519()
	default:
		return ecdh.P256()
	}
}

func (e *eciesCipher) Encrypt(plaintext string) (string, error) {
	if e.publicKey == nil {
		return constants.Empty, ErrPublicKeyRequiredForEncrypt
	}

	ephemeralKey, err := e.publicKey.Curve().GenerateKey(rand.Reader)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	sharedSecret, err := ephemeralKey.ECDH(e.publicKey)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to derive shared secret: %w", err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)

	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(kdf, aesKey); err != nil {
		return constants.Empty, fmt.Errorf("failed to derive AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return constants.Empty, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	ephemeralPublicKey := ephemeralKey.PublicKey().Bytes()
	result := make([]byte, 0, len(ephemeralPublicKey)+len(nonce)+len(ciphertext))
	result = append(result, ephemeralPublicKey...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return encoding.ToBase64(result), nil
}

func (e *eciesCipher) Decrypt(ciphertext string) (string, error) {
	if e.privateKey == nil {
		return constants.Empty, ErrPrivateKeyRequiredForDecrypt
	}

	encryptedData, err := encoding.FromBase64(ciphertext)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decode base64: %w", err)
	}

	publicKeySize := e.privateKey.PublicKey().Bytes()
	publicKeyLen := len(publicKeySize)

	if len(encryptedData) < publicKeyLen+12 {
		return constants.Empty, ErrCiphertextTooShort
	}

	ephemeralPublicKeyBytes := encryptedData[:publicKeyLen]

	ephemeralPublicKey, err := e.privateKey.Curve().NewPublicKey(ephemeralPublicKeyBytes)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to parse ephemeral public key: %w", err)
	}

	sharedSecret, err := e.privateKey.ECDH(ephemeralPublicKey)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to derive shared secret: %w", err)
	}

	kdf := hkdf.New(sha256.New, sharedSecret, nil, nil)

	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(kdf, aesKey); err != nil {
		return constants.Empty, fmt.Errorf("failed to derive AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < publicKeyLen+nonceSize {
		return constants.Empty, ErrCiphertextTooShort
	}

	nonce := encryptedData[publicKeyLen : publicKeyLen+nonceSize]
	ciphertextData := encryptedData[publicKeyLen+nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

var _ Cipher = (*eciesCipher)(nil)
