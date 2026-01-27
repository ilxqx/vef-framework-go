package param_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/cron"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
	"github.com/ilxqx/vef-framework-go/storage"
)

type ParamResolversTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *ParamResolversTestSuite) SetupSuite() {
	suite.T().Log("Setting up ParamResolversTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestParamResolversResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.SQLite,
		}),
		fx.Replace(&config.StorageConfig{
			Provider: constants.StorageMemory,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("ParamResolversTestSuite setup complete")
}

func (suite *ParamResolversTestSuite) TearDownSuite() {
	if suite.stop != nil {
		suite.stop()
	}
}

func (suite *ParamResolversTestSuite) TestCtxResolver() {
	suite.Run("InjectFiberCtx", func() {
		resp := suite.makeApiRequest("verify_ctx", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestDBResolver() {
	suite.Run("InjectOrmDB", func() {
		resp := suite.makeApiRequest("verify_db", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestLoggerResolver() {
	suite.Run("InjectLogger", func() {
		resp := suite.makeApiRequest("verify_logger", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestPrincipalResolver() {
	suite.Run("InjectPrincipal", func() {
		resp := suite.makeApiRequest("verify_principal", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestDBFactoryResolver() {
	suite.Run("InjectDBToFactory", func() {
		resp := suite.makeApiRequest("verify_db_factory", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"factory_injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestStorageResolver() {
	suite.Run("InjectStorageService", func() {
		resp := suite.makeApiRequest("verify_storage", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestStorageFactoryResolver() {
	suite.Run("InjectStorageToFactory", func() {
		resp := suite.makeApiRequest("verify_storage_factory", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"factory_injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestMoldResolver() {
	suite.Run("InjectMoldTransformer", func() {
		resp := suite.makeApiRequest("verify_mold", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestMoldFactoryResolver() {
	suite.Run("InjectMoldToFactory", func() {
		resp := suite.makeApiRequest("verify_mold_factory", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"factory_injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestEventResolver() {
	suite.Run("InjectEventPublisher", func() {
		resp := suite.makeApiRequest("verify_event", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestEventFactoryResolver() {
	suite.Run("InjectEventToFactory", func() {
		resp := suite.makeApiRequest("verify_event_factory", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"factory_injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestCronResolver() {
	suite.Run("InjectCronScheduler", func() {
		resp := suite.makeApiRequest("verify_cron", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"injected":true`)
	})
}

func (suite *ParamResolversTestSuite) TestCronFactoryResolver() {
	suite.Run("InjectCronToFactory", func() {
		resp := suite.makeApiRequest("verify_cron_factory", "{}")
		suite.Equal(200, resp.StatusCode)
		suite.Contains(suite.readBody(resp), `"factory_injected":true`)
	})
}

func (suite *ParamResolversTestSuite) makeApiRequest(action, body string) *http.Response {
	fullBody := `{"resource": "test/param_resolvers", "action": "` + action + `", "version": "v1", "params": ` + body + `}`

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(fullBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	return resp
}

func (suite *ParamResolversTestSuite) readBody(resp *http.Response) string {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	return string(body)
}

func TestParamResolversTestSuite(t *testing.T) {
	suite.Run(t, new(ParamResolversTestSuite))
}

// Resource Definition

type TestParamResolversResource struct {
	api.Resource
}

func NewTestParamResolversResource() api.Resource {
	return &TestParamResolversResource{
		Resource: api.NewRPCResource(
			"test/param_resolvers",
			api.WithVersion(api.VersionV1),
			api.WithOperations(
				api.OperationSpec{Action: "verify_ctx", Public: true},
				api.OperationSpec{Action: "verify_db", Public: true},
				api.OperationSpec{Action: "verify_logger", Public: true},
				api.OperationSpec{Action: "verify_principal", Public: true},
				api.OperationSpec{Action: "verify_db_factory", Public: true},
				api.OperationSpec{Action: "verify_storage", Public: true},
				api.OperationSpec{Action: "verify_storage_factory", Public: true},
				api.OperationSpec{Action: "verify_mold", Public: true},
				api.OperationSpec{Action: "verify_mold_factory", Public: true},
				api.OperationSpec{Action: "verify_event", Public: true},
				api.OperationSpec{Action: "verify_event_factory", Public: true},
				api.OperationSpec{Action: "verify_cron", Public: true},
				api.OperationSpec{Action: "verify_cron_factory", Public: true},
			),
		),
	}
}

func (*TestParamResolversResource) VerifyCtx(ctx fiber.Ctx) error {
	injected := ctx != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyDB(ctx fiber.Ctx, db orm.DB) error {
	injected := db != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyLogger(ctx fiber.Ctx, logger log.Logger) error {
	injected := logger != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyPrincipal(ctx fiber.Ctx, principal *security.Principal) error {
	injected := principal != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyDbFactory(db orm.DB) func(ctx fiber.Ctx) error {
	injected := db != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{"factory_injected": injected}).Response(ctx)
	}
}

func (*TestParamResolversResource) VerifyStorage(ctx fiber.Ctx, service storage.Service) error {
	injected := service != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyStorageFactory(service storage.Service) func(ctx fiber.Ctx) error {
	injected := service != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{"factory_injected": injected}).Response(ctx)
	}
}

func (*TestParamResolversResource) VerifyMold(ctx fiber.Ctx, transformer mold.Transformer) error {
	injected := transformer != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyMoldFactory(transformer mold.Transformer) func(ctx fiber.Ctx) error {
	injected := transformer != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{"factory_injected": injected}).Response(ctx)
	}
}

func (*TestParamResolversResource) VerifyEvent(ctx fiber.Ctx, publisher event.Publisher) error {
	injected := publisher != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyEventFactory(publisher event.Publisher) func(ctx fiber.Ctx) error {
	injected := publisher != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{"factory_injected": injected}).Response(ctx)
	}
}

func (*TestParamResolversResource) VerifyCron(ctx fiber.Ctx, scheduler cron.Scheduler) error {
	injected := scheduler != nil

	return result.Ok(map[string]any{"injected": injected}).Response(ctx)
}

func (*TestParamResolversResource) VerifyCronFactory(scheduler cron.Scheduler) func(ctx fiber.Ctx) error {
	injected := scheduler != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{"factory_injected": injected}).Response(ctx)
	}
}
