package security

import (
	"fmt"
	"time"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/id"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	tokenTypeAccess    = "access"
	tokenTypeRefresh   = "refresh"
	accessTokenExpires = time.Minute * 30
)

type JWTTokenGenerator struct {
	jwt          *security.JWT
	tokenExpires time.Duration
}

func NewJWTTokenGenerator(jwt *security.JWT, securityConfig *config.SecurityConfig) security.TokenGenerator {
	return &JWTTokenGenerator{
		jwt:          jwt,
		tokenExpires: securityConfig.TokenExpires,
	}
}

func (g *JWTTokenGenerator) Generate(principal *security.Principal) (*security.AuthTokens, error) {
	jwtID := id.GenerateUUID()

	accessToken, err := g.generateAccessToken(jwtID, principal)
	if err != nil {
		logger.Errorf("Failed to generate access token for principal %q: %v", principal.ID, err)

		return nil, err
	}

	refreshToken, err := g.generateRefreshToken(jwtID, principal)
	if err != nil {
		logger.Errorf("Failed to generate refresh token for principal %q: %v", principal.ID, err)

		return nil, err
	}

	return &security.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateAccessToken encodes id@name in subject to avoid DB lookups during authentication.
func (g *JWTTokenGenerator) generateAccessToken(jwtID string, principal *security.Principal) (string, error) {
	claimsBuilder := security.NewJWTClaimsBuilder().
		WithID(jwtID).
		WithSubject(fmt.Sprintf("%s@%s", principal.ID, principal.Name)).
		WithRoles(principal.Roles).
		WithDetails(principal.Details).
		WithType(tokenTypeAccess)

	return g.jwt.Generate(claimsBuilder, accessTokenExpires, 0)
}

func (g *JWTTokenGenerator) generateRefreshToken(jwtID string, principal *security.Principal) (string, error) {
	claimsBuilder := security.NewJWTClaimsBuilder().
		WithID(jwtID).
		WithSubject(fmt.Sprintf("%s@%s", principal.ID, principal.Name)).
		WithType(tokenTypeRefresh)

	return g.jwt.Generate(claimsBuilder, g.tokenExpires, refreshTokenNotBefore)
}
