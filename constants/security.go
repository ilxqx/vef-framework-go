package constants

const (
	AuthSchemeBearer = "Bearer"

	QueryKeyAccessToken = "__accessToken"

	// System internal principals that are not allowed to login.
	PrincipalSystem    = OperatorSystem    // PrincipalSystem is the system principal
	PrincipalCronJob   = OperatorCronJob   // PrincipalCronJob is the cron job principal
	PrincipalAnonymous = OperatorAnonymous // PrincipalAnonymous is the anonymous principal
)
