package hash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/tjfoc/gmsm/sm3"
)

func Md5(data string) string {
	return Md5Bytes([]byte(data))
}

func Md5Bytes(data []byte) string {
	hash := md5.Sum(data)

	return hex.EncodeToString(hash[:])
}

func Sha1(data string) string {
	return Sha1Bytes([]byte(data))
}

func Sha1Bytes(data []byte) string {
	hash := sha1.Sum(data)

	return hex.EncodeToString(hash[:])
}

func Sha256(data string) string {
	return Sha256Bytes([]byte(data))
}

func Sha256Bytes(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

func Sha512(data string) string {
	return Sha512Bytes([]byte(data))
}

func Sha512Bytes(data []byte) string {
	hash := sha512.Sum512(data)

	return hex.EncodeToString(hash[:])
}

// Sm3 uses Chinese National Standard cryptographic hash (国密算法).
func Sm3(data string) string {
	return Sm3Bytes([]byte(data))
}

func Sm3Bytes(data []byte) string {
	hash := sm3.Sm3Sum(data)

	return hex.EncodeToString(hash)
}

func Md5Hmac(key, data []byte) (string, error) {
	mac := hmac.New(md5.New, key)
	if _, err := mac.Write(data); err != nil {
		return constants.Empty, fmt.Errorf("hash: failed to write data to HMAC-MD5: %w", err)
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

func Sha1Hmac(key, data []byte) (string, error) {
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(data); err != nil {
		return constants.Empty, fmt.Errorf("hash: failed to write data to HMAC-SHA1: %w", err)
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

func Sha256Hmac(key, data []byte) (string, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(data); err != nil {
		return constants.Empty, fmt.Errorf("hash: failed to write data to HMAC-SHA256: %w", err)
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

func Sha512Hmac(key, data []byte) (string, error) {
	mac := hmac.New(sha512.New, key)
	if _, err := mac.Write(data); err != nil {
		return constants.Empty, fmt.Errorf("hash: failed to write data to HMAC-SHA512: %w", err)
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

// Sm3Hmac uses Chinese National Standard cryptographic hash (国密算法).
func Sm3Hmac(key, data []byte) (string, error) {
	mac := hmac.New(func() hash.Hash { return sm3.New() }, key)
	if _, err := mac.Write(data); err != nil {
		return constants.Empty, fmt.Errorf("hash: failed to write data to HMAC-SM3: %w", err)
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}
