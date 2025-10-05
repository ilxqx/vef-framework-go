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

// MD5 computes the MD5 hash of the input string and returns a hex-encoded string.
// WARNING: MD5 is cryptographically broken and should not be used for security purposes.
// Use it only for non-security purposes like checksums.
func MD5(data string) string {
	return MD5Bytes([]byte(data))
}

// MD5Bytes computes the MD5 hash of the input bytes and returns a hex-encoded string.
func MD5Bytes(data []byte) string {
	hash := md5.Sum(data)

	return hex.EncodeToString(hash[:])
}

// SHA1 computes the SHA-1 hash of the input string and returns a hex-encoded string.
// WARNING: SHA-1 is cryptographically broken and should not be used for security purposes.
// Consider using SHA256 or SHA512 instead.
func SHA1(data string) string {
	return SHA1Bytes([]byte(data))
}

// SHA1Bytes computes the SHA-1 hash of the input bytes and returns a hex-encoded string.
func SHA1Bytes(data []byte) string {
	hash := sha1.Sum(data)

	return hex.EncodeToString(hash[:])
}

// SHA256 computes the SHA-256 hash of the input string and returns a hex-encoded string.
// This is the recommended hash function for most use cases.
func SHA256(data string) string {
	return SHA256Bytes([]byte(data))
}

// SHA256Bytes computes the SHA-256 hash of the input bytes and returns a hex-encoded string.
func SHA256Bytes(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

// SHA512 computes the SHA-512 hash of the input string and returns a hex-encoded string.
func SHA512(data string) string {
	return SHA512Bytes([]byte(data))
}

// SHA512Bytes computes the SHA-512 hash of the input bytes and returns a hex-encoded string.
func SHA512Bytes(data []byte) string {
	hash := sha512.Sum512(data)

	return hex.EncodeToString(hash[:])
}

// SM3 computes the SM3 hash of the input string and returns a hex-encoded string.
// SM3 is a cryptographic hash function used in Chinese National Standard (国密算法).
func SM3(data string) string {
	return SM3Bytes([]byte(data))
}

// SM3Bytes computes the SM3 hash of the input bytes and returns a hex-encoded string.
func SM3Bytes(data []byte) string {
	hash := sm3.Sm3Sum(data)

	return hex.EncodeToString(hash)
}

// MD5Hmac computes the HMAC-MD5 of the input data with the given key.
// WARNING: MD5 is cryptographically broken. Use SHA256Hmac instead.
func MD5Hmac(key, data []byte) string {
	mac := hmac.New(md5.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-MD5: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// SHA1Hmac computes the HMAC-SHA1 of the input data with the given key.
// WARNING: SHA-1 is cryptographically weak. Use SHA256Hmac instead.
func SHA1Hmac(key, data []byte) string {
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SHA1: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// SHA256Hmac computes the HMAC-SHA256 of the input data with the given key.
// This is the recommended HMAC function for most use cases.
func SHA256Hmac(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SHA256: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// SHA512Hmac computes the HMAC-SHA512 of the input data with the given key.
func SHA512Hmac(key, data []byte) string {
	mac := hmac.New(sha512.New, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SHA512: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// SM3Hmac computes the HMAC-SM3 of the input data with the given key (国密算法).
func SM3Hmac(key, data []byte) string {
	mac := hmac.New(func() hash.Hash { return sm3.New() }, key)
	if _, err := mac.Write(data); err != nil {
		panic(
			fmt.Errorf("hash: failed to write data to HMAC-SM3: %w", err),
		)
	}

	return hex.EncodeToString(mac.Sum(nil))
}
