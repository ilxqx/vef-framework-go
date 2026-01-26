package constants

// Authentication constants.
const (
	AuthSchemeBearer    = "Bearer"
	QueryKeyAccessToken = "__accessToken"
)

// System internal principals (not allowed to login).
const (
	PrincipalSystem    = OperatorSystem
	PrincipalCronJob   = OperatorCronJob
	PrincipalAnonymous = OperatorAnonymous
)
