package security

import (
	"fmt"
	"time"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/id"
	"github.com/ilxqx/vef-framework-go/security"
)

const (
	tokenTypeAccess    = "access"
	tokenTypeRefresh   = "refresh"
	accessTokenExpires = time.Minute * 30
)

type JwtTokenGenerator struct {
	jwt          *security.Jwt
	tokenExpires time.Duration
}

func NewJwtTokenGenerator(jwt *security.Jwt, securityConfig *config.SecurityConfig) security.TokenGenerator {
	return &JwtTokenGenerator{
		jwt:          jwt,
		tokenExpires: securityConfig.TokenExpires,
	}
}

func (g *JwtTokenGenerator) Generate(principal *security.Principal) (*security.AuthTokens, error) {
	jwtId := id.GenerateUuid()

	accessToken, err := g.generateAccessToken(jwtId, principal)
	if err != nil {
		logger.Errorf("Failed to generate access token for principal %q: %v", principal.Id, err)

		return nil, err
	}

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

// generateAccessToken encodes id@name in subject to avoid DB lookups during authentication.
func (g *JwtTokenGenerator) generateAccessToken(jwtId string, principal *security.Principal) (string, error) {
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

	return g.jwt.Generate(claimsBuilder, g.tokenExpires, refreshTokenNotBefore)
}
