package password

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

type md5Encoder struct {
	salt         string
	saltPosition string
}

// Md5Option configures md5Encoder.
type Md5Option func(*md5Encoder)

// WithMd5Salt sets a static salt value.
// WARNING: Static salts provide minimal security. Use modern algorithms like Argon2 for new systems.
func WithMd5Salt(salt string) Md5Option {
	return func(e *md5Encoder) {
		e.salt = salt
	}
}

// WithMd5SaltPosition sets where the salt is placed ("prefix" or "suffix").
func WithMd5SaltPosition(position string) Md5Option {
	return func(e *md5Encoder) {
		e.saltPosition = position
	}
}

// NewMd5Encoder creates a new MD5-based password encoder.
// WARNING: Use only for legacy system compatibility.
func NewMd5Encoder(opts ...Md5Option) Encoder {
	encoder := &md5Encoder{
		saltPosition: "suffix",
	}

	for _, opt := range opts {
		opt(encoder)
	}

	return encoder
}

func (e *md5Encoder) Encode(password string) (string, error) {
	var input string
	if e.salt != constants.Empty {
		if e.saltPosition == "prefix" {
			input = e.salt + password
		} else {
			input = password + e.salt
		}
	} else {
		input = password
	}

	hash := md5.Sum([]byte(input))
	hexHash := hex.EncodeToString(hash[:])

	if e.salt != constants.Empty {
		return fmt.Sprintf("{md5}$%s$%s", e.salt, hexHash), nil
	}

	return hexHash, nil
}

func (e *md5Encoder) Matches(password, encodedPassword string) bool {
	if strings.HasPrefix(encodedPassword, "{md5}$") {
		parts := strings.Split(encodedPassword, "$")
		if len(parts) != 3 {
			return false
		}

		salt := parts[1]
		expectedHash := parts[2]

		var input string
		if e.saltPosition == "prefix" {
			input = salt + password
		} else {
			input = password + salt
		}

		hash := md5.Sum([]byte(input))
		actualHash := hex.EncodeToString(hash[:])

		return actualHash == expectedHash
	}

	hash := md5.Sum([]byte(password))
	actualHash := hex.EncodeToString(hash[:])

	return actualHash == encodedPassword
}

func (e *md5Encoder) UpgradeEncoding(encodedPassword string) bool {
	return true
}
