package security

type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Authentication struct {
	Kind        string `json:"kind"`
	Principal   string `json:"principal"`
	Credentials any    `json:"credentials"`
}

type ExternalAppConfig struct {
	Enabled     bool   `json:"enabled"`
	IPWhitelist string `json:"ipWhitelist"`
}
