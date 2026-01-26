package auth

import (
	"context"

	isecurity "github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/security"
)

type AccessTokenAuthenticator struct {
	manager security.AuthManager
}

func (a *AccessTokenAuthenticator) Authenticate(ctx context.Context, token string) (*security.Principal, error) {
	return a.manager.Authenticate(ctx, security.Authentication{
		Kind:      isecurity.AuthKindToken,
		Principal: token,
	})
}

func NewAccessTokenAuthenticator(manager security.AuthManager) TokenAuthenticator {
	return &AccessTokenAuthenticator{
		manager: manager,
	}
}
