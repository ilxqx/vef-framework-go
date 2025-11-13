package security

type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Authentication struct {
	Type        string `json:"type"`
	Principal   string `json:"principal"`
	Credentials any    `json:"credentials"`
}

type ExternalAppConfig struct {
	Enabled     bool   `json:"enabled"`
	IpWhitelist string `json:"ipWhitelist"`
}
