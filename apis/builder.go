package apis

import (
	"time"

	"github.com/ilxqx/vef-framework-go/api"
)

type baseAPIBuilder[T any] struct {
	action      string
	version     string
	enableAudit bool
	timeout     time.Duration
	public      bool
	permToken   string
	rateLimit   api.RateLimit

	self T
}

func (b *baseAPIBuilder[T]) Action(action string) T {
	b.action = action

	return b.self
}

func (b *baseAPIBuilder[T]) EnableAudit() T {
	b.enableAudit = true

	return b.self
}

func (b *baseAPIBuilder[T]) Timeout(timeout time.Duration) T {
	b.timeout = timeout

	return b.self
}

func (b *baseAPIBuilder[T]) Public() T {
	b.public = true

	return b.self
}

func (b *baseAPIBuilder[T]) PermToken(token string) T {
	b.permToken = token

	return b.self
}

func (b *baseAPIBuilder[T]) RateLimit(max int, expiration time.Duration) T {
	b.rateLimit = api.RateLimit{
		Max:        max,
		Expiration: expiration,
	}

	return b.self
}

func (b *baseAPIBuilder[T]) Build(handler any) api.Spec {
	return api.Spec{
		Action:      b.action,
		Version:     b.version,
		EnableAudit: b.enableAudit,
		Timeout:     b.timeout,
		Public:      b.public,
		PermToken:   b.permToken,
		Limit:       b.rateLimit,
		Handler:     handler,
	}
}
