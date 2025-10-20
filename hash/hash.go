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

	"github.com/tjfoc/gmsm/sm3"
)

// Md5 computes the MD5 hash of the input string and returns a hex-encoded string.
// WARNING: MD5 is cryptographically broken and should not be used for security purposes.
// Use it only for non-security purposes like checksums.
func Md5(data string) string {
	return Md5Bytes([]byte(data))
}

// Md5Bytes computes the MD5 hash of the input bytes and returns a hex-encoded string.
func Md5Bytes(data []byte) string {
	hash := md5.Sum(data)

	return hex.EncodeToString(hash[:])
}

// Sha1 computes the SHA-1 hash of the input string and returns a hex-encoded string.
// WARNING: SHA-1 is cryptographically broken and should not be used for security purposes.
// Consider using Sha256 or Sha512 instead.
func Sha1(data string) string {
	return Sha1Bytes([]byte(data))
}

// Sha1Bytes computes the SHA-1 hash of the input bytes and returns a hex-encoded string.
func Sha1Bytes(data []byte) string {
	hash := sha1.Sum(data)

	return hex.EncodeToString(hash[:])
}

// Sha256 computes the SHA-256 hash of the input string and returns a hex-encoded string.
// This is the recommended hash function for most use cases.
func Sha256(data string) string {
	return Sha256Bytes([]byte(data))
}

// Sha256Bytes computes the SHA-256 hash of the input bytes and returns a hex-encoded string.
func Sha256Bytes(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

// Sha512 computes the SHA-512 hash of the input string and returns a hex-encoded string.
func Sha512(data string) string {
	return Sha512Bytes([]byte(data))
}

// Sha512Bytes computes the SHA-512 hash of the input bytes and returns a hex-encoded string.
func Sha512Bytes(data []byte) string {
	hash := sha512.Sum512(data)

	return hex.EncodeToString(hash[:])
}

// Sm3 computes the SM3 hash of the input string and returns a hex-encoded string.
// SM3 is a cryptographic hash function used in Chinese National Standard (国密算法).
func Sm3(data string) string {
	return Sm3Bytes([]byte(data))
}

// Sm3Bytes computes the SM3 hash of the input bytes and returns a hex-encoded string.
func Sm3Bytes(data []byte) string {
	hash := sm3.Sm3Sum(data)

	return hex.EncodeToString(hash)
}

// Md5Hmac computes the HMAC-MD5 of the input data with the given key.
// WARNING: MD5 is cryptographically broken. Use Sha256Hmac instead.
func Md5Hmac(key, data []byte) string {
	mac := hmac.New(md5.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-MD5: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// Sha1Hmac computes the HMAC-SHA1 of the input data with the given key.
// WARNING: SHA-1 is cryptographically weak. Use Sha256Hmac instead.
func Sha1Hmac(key, data []byte) string {
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SHA1: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// Sha256Hmac computes the HMAC-SHA256 of the input data with the given key.
// This is the recommended HMAC function for most use cases.
func Sha256Hmac(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SHA256: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// Sha512Hmac computes the HMAC-SHA512 of the input data with the given key.
func Sha512Hmac(key, data []byte) string {
	mac := hmac.New(sha512.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SHA512: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// Sm3Hmac computes the HMAC-SM3 of the input data with the given key (国密算法).
func Sm3Hmac(key, data []byte) string {
	mac := hmac.New(func() hash.Hash { return sm3.New() }, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SM3: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}
