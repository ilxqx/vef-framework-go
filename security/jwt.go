package security

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/result"
)

const (
	jwtIssuer          = "vef"                                                              // Issuer
	defaultJwtAudience = "vef-app"                                                          // Audience
	defaultJwtSecret   = "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2" // Secret
)

var jwtParseOptions = []jwt.ParserOption{
	jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	jwt.WithIssuer(jwtIssuer),
	jwt.WithLeeway(10 * time.Second),
	jwt.WithIssuedAt(),
	jwt.WithExpirationRequired(),
}

// Jwt provides low-level Jwt token operations.
// It handles token generation, parsing, and validation without business logic.
type Jwt struct {
	config *JwtConfig
	secret []byte
}

// NewJwt creates a new Jwt instance with the given configuration.
// Secret expects a hex-encoded string; invalid hex will cause a panic during initialization.
// Audience will be defaulted when empty.
func NewJwt(config *JwtConfig) (*Jwt, error) {
	var (
		secret []byte
		err    error
	)
	if secret, err = hex.DecodeString(lo.CoalesceOrEmpty(config.Secret, defaultJwtSecret)); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeJwtSecretFailed, err)
	}

	config.Audience = lo.CoalesceOrEmpty(config.Audience, defaultJwtAudience)

	return &Jwt{
		config: config,
		secret: secret,
	}, nil
}

// Generate creates a Jwt token with the given claims and expires.
// The expiration is computed as now + expires; iat and nbf are set to now.
func (j *Jwt) Generate(claimsBuilder *JwtClaimsBuilder, expires, notBefore time.Duration) (string, error) {
	claims := claimsBuilder.build()
	// Set standard claims
	now := time.Now()
	claims[claimIssuer] = jwtIssuer
	claims[claimAudience] = j.config.Audience
	claims[claimIssuedAt] = now.Unix()
	claims[claimNotBefore] = now.Add(notBefore).Unix()
	claims[claimExpiresAt] = now.Add(expires).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.secret)
}

// Parse parses and validates a Jwt token.
// It returns a read-only claims accessor which performs safe conversions and never panics.
func (j *Jwt) Parse(tokenString string) (*JwtClaimsAccessor, error) {
	options := make([]jwt.ParserOption, 0, len(jwtParseOptions)+1)
	options = append(options, jwtParseOptions...)
	options = append(options, jwt.WithAudience(j.config.Audience))

	token, err := jwt.NewParser(options...).
		Parse(
			tokenString,
			func(token *jwt.Token) (any, error) {
				return j.secret, nil
			},
		)
	if err != nil {
		return nil, mapJwtError(err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, result.ErrTokenInvalid
	}

	return NewJwtClaimsAccessor(claims), nil
}

// mapJwtError maps Jwt library errors to framework errors.
func mapJwtError(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return result.ErrTokenExpired
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return result.ErrTokenNotValidYet
	case errors.Is(err, jwt.ErrTokenInvalidIssuer):
		return result.ErrTokenInvalidIssuer
	case errors.Is(err, jwt.ErrTokenInvalidAudience):
		return result.ErrTokenInvalidAudience
	default:
		return result.ErrTokenInvalid
	}
}
