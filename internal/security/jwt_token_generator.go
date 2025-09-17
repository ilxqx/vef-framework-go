package security

import (
	"fmt"
	"time"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/utils"
)

const (
	tokenTypeAccess    = "access"  // Access token type
	tokenTypeRefresh   = "refresh" // Refresh token type
	accessTokenExpires = time.Hour // Access token expires
)

// jwtTokenGenerator implements the TokenGenerator interface for JWT tokens.
// It generates both access and refresh tokens using the JWT helper.
type jwtTokenGenerator struct {
	jwt          *security.JWT
	tokenExpires time.Duration
}

// newJWTTokenGenerator creates a new JWT token generator.
func newJWTTokenGenerator(jwt *security.JWT, securityConfig *config.SecurityConfig) security.TokenGenerator {
	return &jwtTokenGenerator{
		jwt:          jwt,
		tokenExpires: securityConfig.TokenExpires,
	}
}

// Generate creates authentication tokens for the given principal.
// It generates both access and refresh tokens.
func (g *jwtTokenGenerator) Generate(principal *security.Principal) (*security.TokenCredentials, error) {
	jwtId := utils.GenerateId()
	// Generate access token
	accessToken, err := g.generateAccessToken(jwtId, principal)
	if err != nil {
		logger.Errorf("failed to generate access token for principal '%s': %v", principal.Id, err)
		return nil, err
	}

	// Generate refresh token using the access token's JWT ID
	refreshToken, err := g.generateRefreshToken(jwtId, principal)
	if err != nil {
		logger.Errorf("failed to generate refresh token for principal '%s': %v", principal.Id, err)
		return nil, err
	}

	logger.Infof("successfully generated tokens for principal '%s'", principal.Id)
	return &security.TokenCredentials{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (g *jwtTokenGenerator) generateAccessToken(jwtId string, principal *security.Principal) (string, error) {
	// Subject format: id@name for quick identity recovery in authenticator
	claimsBuilder := security.NewJWTClaimsBuilder().
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

func (g *jwtTokenGenerator) generateRefreshToken(jwtId string, principal *security.Principal) (string, error) {
	claimsBuilder := security.NewJWTClaimsBuilder().
		WithId(jwtId).
		WithSubject(fmt.Sprintf("%s@%s", principal.Id, principal.Name)).
		WithType(tokenTypeRefresh)

	return g.jwt.Generate(claimsBuilder, g.tokenExpires, accessTokenExpires/2)
}
