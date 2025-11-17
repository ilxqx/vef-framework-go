package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
)

type EcdsaCurve string

const (
	EcdsaCurveP224 EcdsaCurve = "P224"
	EcdsaCurveP256 EcdsaCurve = "P256"
	EcdsaCurveP384 EcdsaCurve = "P384"
	EcdsaCurveP521 EcdsaCurve = "P521"
)

type ecdsaCipher struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

type EcdsaOption func(*ecdsaCipher)

func NewEcdsa(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, opts ...EcdsaOption) (Signer, error) {
	if privateKey == nil && publicKey == nil {
		return nil, ErrAtLeastOneKeyRequired
	}

	cipher := &ecdsaCipher{
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	for _, opt := range opts {
		opt(cipher)
	}

	if publicKey == nil && privateKey != nil {
		cipher.publicKey = &privateKey.PublicKey
	}

	return cipher, nil
}

func NewEcdsaFromPem(privatePem, publicPem []byte, opts ...EcdsaOption) (Signer, error) {
	var (
		privateKey *ecdsa.PrivateKey
		publicKey  *ecdsa.PublicKey
		err        error
	)

	if privatePem != nil {
		if privateKey, err = parseEcdsaPrivateKeyFromPem(privatePem); err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicPem != nil {
		if publicKey, err = parseEcdsaPublicKeyFromPem(publicPem); err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	return NewEcdsa(privateKey, publicKey, opts...)
}

func parseEcdsaKeysFromBytes(privateKeyBytes, publicKeyBytes []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	var (
		privateKey *ecdsa.PrivateKey
		publicKey  *ecdsa.PublicKey
		err        error
	)

	if len(privateKeyBytes) > 0 {
		if privateKey, err = x509.ParseECPrivateKey(privateKeyBytes); err != nil {
			key, err2 := x509.ParsePKCS8PrivateKey(privateKeyBytes)
			if err2 != nil {
				return nil, nil, fmt.Errorf("failed to parse private key (tried EC and PKCS8): %w", err)
			}

			var ok bool
			if privateKey, ok = key.(*ecdsa.PrivateKey); !ok {
				return nil, nil, ErrNotEcdsaPrivateKey
			}
		}
	}

	if len(publicKeyBytes) > 0 {
		if key, err := x509.ParsePKIXPublicKey(publicKeyBytes); err != nil {
			return nil, nil, fmt.Errorf("failed to parse public key: %w", err)
		} else {
			var ok bool
			if publicKey, ok = key.(*ecdsa.PublicKey); !ok {
				return nil, nil, ErrNotEcdsaPublicKey
			}
		}
	}

	return privateKey, publicKey, nil
}

func NewEcdsaFromHex(privateKeyHex, publicKeyHex string, opts ...EcdsaOption) (Signer, error) {
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

	privateKey, publicKey, err := parseEcdsaKeysFromBytes(privateBytes, publicBytes)
	if err != nil {
		return nil, err
	}

	return NewEcdsa(privateKey, publicKey, opts...)
}

func NewEcdsaFromBase64(privateKeyBase64, publicKeyBase64 string, opts ...EcdsaOption) (Signer, error) {
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

	privateKey, publicKey, err := parseEcdsaKeysFromBytes(privateBytes, publicBytes)
	if err != nil {
		return nil, err
	}

	return NewEcdsa(privateKey, publicKey, opts...)
}

func GenerateEcdsaKey(curve EcdsaCurve) (*ecdsa.PrivateKey, error) {
	var ellipticCurve elliptic.Curve

	switch curve {
	case EcdsaCurveP224:
		ellipticCurve = elliptic.P224()
	case EcdsaCurveP256:
		ellipticCurve = elliptic.P256()
	case EcdsaCurveP384:
		ellipticCurve = elliptic.P384()
	case EcdsaCurveP521:
		ellipticCurve = elliptic.P521()
	default:
		ellipticCurve = elliptic.P256()
	}

	return ecdsa.GenerateKey(ellipticCurve, rand.Reader)
}

type ecdsaSignature struct {
	R, S *big.Int
}

func (e *ecdsaCipher) Sign(data string) (string, error) {
	if e.privateKey == nil {
		return constants.Empty, ErrPrivateKeyRequiredForSign
	}

	hash := sha256.New()
	_, _ = hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, e.privateKey, hashed)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to sign: %w", err)
	}

	signature, err := asn1.Marshal(ecdsaSignature{R: r, S: s})
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to marshal signature: %w", err)
	}

	return encoding.ToBase64(signature), nil
}

func (e *ecdsaCipher) Verify(data, signature string) (bool, error) {
	if e.publicKey == nil {
		return false, ErrPublicKeyRequiredForVerify
	}

	signatureBytes, err := encoding.FromBase64(signature)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	var sig ecdsaSignature
	if _, err := asn1.Unmarshal(signatureBytes, &sig); err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	hash := sha256.New()
	_, _ = hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	valid := ecdsa.Verify(e.publicKey, hashed, sig.R, sig.S)

	return valid, nil
}

func parseEcdsaPrivateKeyFromPem(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePemBlock
	}

	if block.Type == "EC PRIVATE KEY" {
		return x509.ParseECPrivateKey(block.Bytes)
	}

	if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		ecdsaKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, ErrNotEcdsaPrivateKey
		}

		return ecdsaKey, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedPemType, block.Type)
}

func parseEcdsaPublicKeyFromPem(pemData []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePemBlock
	}

	if block.Type == "PUBLIC KEY" {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		ecdsaKey, ok := key.(*ecdsa.PublicKey)
		if !ok {
			return nil, ErrNotEcdsaPublicKey
		}

		return ecdsaKey, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedPemType, block.Type)
}

var _ Signer = (*ecdsaCipher)(nil)
