package password

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

type sha256Encoder struct {
	salt         string
	saltPosition string
}

// Sha256Option configures sha256Encoder.
type Sha256Option func(*sha256Encoder)

// WithSha256Salt sets a static salt value.
// WARNING: Static salts provide minimal security. Use modern algorithms like Argon2 for new systems.
func WithSha256Salt(salt string) Sha256Option {
	return func(e *sha256Encoder) {
		e.salt = salt
	}
}

// WithSha256SaltPosition sets where the salt is placed ("prefix" or "suffix").
func WithSha256SaltPosition(position string) Sha256Option {
	return func(e *sha256Encoder) {
		e.saltPosition = position
	}
}

// NewSha256Encoder creates a new SHA-256-based password encoder.
// WARNING: Use only for legacy system compatibility.
func NewSha256Encoder(opts ...Sha256Option) Encoder {
	encoder := &sha256Encoder{
		saltPosition: "suffix",
	}

	for _, opt := range opts {
		opt(encoder)
	}

	return encoder
}

func (e *sha256Encoder) Encode(password string) (string, error) {
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

	hash := sha256.Sum256([]byte(input))
	hexHash := hex.EncodeToString(hash[:])

	if e.salt != constants.Empty {
		return fmt.Sprintf("{sha256}$%s$%s", e.salt, hexHash), nil
	}

	return hexHash, nil
}

func (e *sha256Encoder) Matches(password, encodedPassword string) bool {
	if strings.HasPrefix(encodedPassword, "{sha256}$") {
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

		hash := sha256.Sum256([]byte(input))
		actualHash := hex.EncodeToString(hash[:])

		return actualHash == expectedHash
	}

	hash := sha256.Sum256([]byte(password))
	actualHash := hex.EncodeToString(hash[:])

	return actualHash == encodedPassword
}

func (e *sha256Encoder) UpgradeEncoding(encodedPassword string) bool {
	return true
}
