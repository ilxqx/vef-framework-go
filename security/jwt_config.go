package security

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
)

// Custom and standard Jwt claim keys.
// Short keys are used for custom claims to keep token size small.
const (
	claimJwtId     = "jti" // Jwt ID
	claimSubject   = "sub" // Subject
	claimIssuer    = "iss" // Issuer
	claimAudience  = "aud" // Audience
	claimIssuedAt  = "iat" // Issued At
	claimNotBefore = "nbf" // Not Before
	claimExpiresAt = "exp" // Expires At
	claimType      = "typ" // Token Type
	claimRoles     = "rls" // User Roles
	claimDetails   = "det" // User Details
)

// JwtConfig is the configuration for the Jwt token.
type JwtConfig struct {
	Secret   string `config:"secret"`   // Secret key for Jwt signing
	Audience string `config:"audience"` // Jwt audience
}

// JwtClaimsBuilder helps build Jwt claims for different token types.
type JwtClaimsBuilder struct {
	claims jwt.MapClaims
}

// NewJwtClaimsBuilder creates a new Jwt claims builder.
func NewJwtClaimsBuilder() *JwtClaimsBuilder {
	return &JwtClaimsBuilder{
		claims: make(jwt.MapClaims),
	}
}

// WithId sets the Jwt ID claim.
func (b *JwtClaimsBuilder) WithId(id string) *JwtClaimsBuilder {
	b.claims[claimJwtId] = id

	return b
}

// Id returns the Jwt ID claim.
func (b *JwtClaimsBuilder) Id() (string, bool) {
	id, ok := b.claims[claimJwtId]

	return cast.ToString(id), ok
}

// WithSubject sets the subject claim.
func (b *JwtClaimsBuilder) WithSubject(subject string) *JwtClaimsBuilder {
	b.claims[claimSubject] = subject

	return b
}

// Subject returns the subject claim.
func (b *JwtClaimsBuilder) Subject() (string, bool) {
	subject, ok := b.claims[claimSubject]

	return cast.ToString(subject), ok
}

// WithRoles sets the roles claim.
func (b *JwtClaimsBuilder) WithRoles(roles []string) *JwtClaimsBuilder {
	b.claims[claimRoles] = roles

	return b
}

// Roles returns the roles claim.
func (b *JwtClaimsBuilder) Roles() ([]string, bool) {
	roles, ok := b.claims[claimRoles]

	return cast.ToStringSlice(roles), ok
}

// WithDetails sets the details claim.
func (b *JwtClaimsBuilder) WithDetails(details any) *JwtClaimsBuilder {
	b.claims[claimDetails] = details

	return b
}

// Details returns the details claim.
func (b *JwtClaimsBuilder) Details() (any, bool) {
	details, ok := b.claims[claimDetails]

	return details, ok
}

// WithType sets the token type claim.
func (b *JwtClaimsBuilder) WithType(typ string) *JwtClaimsBuilder {
	b.claims[claimType] = typ

	return b
}

// Type returns the token type claim.
func (b *JwtClaimsBuilder) Type() (string, bool) {
	typ, ok := b.claims[claimType]

	return cast.ToString(typ), ok
}

// WithClaim sets a custom claim.
func (b *JwtClaimsBuilder) WithClaim(key string, value any) *JwtClaimsBuilder {
	b.claims[key] = value

	return b
}

// Claim returns a custom claim.
func (b *JwtClaimsBuilder) Claim(key string) (any, bool) {
	claim, ok := b.claims[key]

	return claim, ok
}

// build returns the built claims.
func (b *JwtClaimsBuilder) build() jwt.MapClaims {
	return b.claims
}

type JwtClaimsAccessor struct {
	claims jwt.MapClaims
}

// NewJwtClaimsAccessor creates a new Jwt claims accessor.
func NewJwtClaimsAccessor(claims jwt.MapClaims) *JwtClaimsAccessor {
	return &JwtClaimsAccessor{
		claims: claims,
	}
}

// Id returns the Jwt ID claim.
// Returns empty string if the claim is missing or not a string.
func (a *JwtClaimsAccessor) Id() string {
	return cast.ToString(a.claims[claimJwtId])
}

// Subject returns the subject claim.
// Returns empty string if the claim is missing or not a string.
func (a *JwtClaimsAccessor) Subject() string {
	return cast.ToString(a.claims[claimSubject])
}

// Roles returns the roles claim.
// Supports both []string and []any payloads; returns empty slice if absent.
func (a *JwtClaimsAccessor) Roles() []string {
	return cast.ToStringSlice(a.claims[claimRoles])
}

// Details returns the details claim.
func (a *JwtClaimsAccessor) Details() any {
	return a.claims[claimDetails]
}

// Type returns the token type claim.
// Returns empty string if the claim is missing or not a string.
func (a *JwtClaimsAccessor) Type() string {
	return cast.ToString(a.claims[claimType])
}

// Claim returns the claim.
func (a *JwtClaimsAccessor) Claim(key string) any {
	return a.claims[key]
}
