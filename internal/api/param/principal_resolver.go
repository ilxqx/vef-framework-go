package param

import (
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/security"
)

type PrincipalResolver struct{}

func (*PrincipalResolver) Type() reflect.Type {
	return reflect.TypeFor[*security.Principal]()
}

func (*PrincipalResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	return reflect.ValueOf(contextx.Principal(ctx)), nil
}

func NewPrincipalResolver() api.HandlerParamResolver {
	return new(PrincipalResolver)
}
