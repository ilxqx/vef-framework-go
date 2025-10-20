package security

import (
	"fmt"
	"time"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/id"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

const (
	tokenTypeAccess    = "access"  // Access token type
	tokenTypeRefresh   = "refresh" // Refresh token type
	accessTokenExpires = time.Hour // Access token expires
)

// JwtTokenGenerator implements the TokenGenerator interface for Jwt tokens.
// It generates both access and refresh tokens using the Jwt helper.
type JwtTokenGenerator struct {
	jwt          *security.Jwt
	tokenExpires time.Duration
}

// NewJwtTokenGenerator creates a new Jwt token generator.
func NewJwtTokenGenerator(jwt *security.Jwt, securityConfig *config.SecurityConfig) security.TokenGenerator {
	return &JwtTokenGenerator{
		jwt:          jwt,
		tokenExpires: securityConfig.TokenExpires,
	}
}

// Generate creates authentication tokens for the given principal.
// It generates both access and refresh tokens.
func (g *JwtTokenGenerator) Generate(principal *security.Principal) (*security.AuthTokens, error) {
	jwtId := id.GenerateUuid()
	// Generate access token
	accessToken, err := g.generateAccessToken(jwtId, principal)
	if err != nil {
		logger.Errorf("Failed to generate access token for principal %q: %v", principal.Id, err)

		return nil, err
	}

	// Generate refresh token using the access token's Jwt ID
	refreshToken, err := g.generateRefreshToken(jwtId, principal)
	if err != nil {
		logger.Errorf("Failed to generate refresh token for principal %q: %v", principal.Id, err)

		return nil, err
	}

	return &security.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (g *JwtTokenGenerator) generateAccessToken(jwtId string, principal *security.Principal) (string, error) {
	// Subject format: id@name for quick identity recovery in authenticator
	claimsBuilder := security.NewJwtClaimsBuilder().
		WithId(jwtId).
		WithSubject(fmt.Sprintf("%s@%s", principal.Id, principal.Name)).
		WithRoles(principal.Roles).
		WithDetails(principal.Details).
		WithType(tokenTypeAccess)

	accessToken, err := g.jwt.Generate(claimsBuilder, accessTokenExpires, time.Second*0)
	if err != nil {
		return constants.Empty, err
	}

	return accessToken, nil
}

func (g *JwtTokenGenerator) generateRefreshToken(jwtId string, principal *security.Principal) (string, error) {
	claimsBuilder := security.NewJwtClaimsBuilder().
		WithId(jwtId).
		WithSubject(fmt.Sprintf("%s@%s", principal.Id, principal.Name)).
		WithType(tokenTypeRefresh)

	// In test environment, use 0 notBefore to allow immediate token usage
	// In production, use accessTokenExpires/2 to prevent immediate reuse
	notBefore := accessTokenExpires / 2
	if testhelpers.IsTestEnv() {
		notBefore = 0
	}

	return g.jwt.Generate(claimsBuilder, g.tokenExpires, notBefore)
}
