package orm_test

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
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// DbFactoryParamResolverTestSuite tests the orm.Db factory parameter resolver.
type DbFactoryParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *DbFactoryParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up DbFactoryParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestOrmFactoryResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("DbFactoryParamResolverTestSuite setup complete")
}

func (suite *DbFactoryParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down DbFactoryParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("DbFactoryParamResolverTestSuite teardown complete")
}

func (suite *DbFactoryParamResolverTestSuite) TestDbFactoryInjection() {
	suite.T().Log("Testing orm.Db factory parameter injection")

	suite.Run("FactoryReceivedDb", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/orm_factory",
			"action": "verify_factory",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"factory_injected":true`, "Factory should have received db")
		suite.Contains(body, `"db_type":"orm.Db"`, "Should identify db type")
	})
}

func (suite *DbFactoryParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *DbFactoryParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestDbFactoryParamResolverSuite(t *testing.T) {
	suite.Run(t, new(DbFactoryParamResolverTestSuite))
}

// Test resource using factory function

type TestOrmFactoryResource struct {
	api.Resource
}

func NewTestOrmFactoryResource() api.Resource {
	return &TestOrmFactoryResource{
		Resource: api.NewResource(
			"test/orm_factory",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "verify_factory", Public: true},
			),
		),
	}
}

func (r *TestOrmFactoryResource) VerifyFactory(db orm.Db) func(ctx fiber.Ctx) error {
	factoryInjected := db != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{
			"factory_injected": factoryInjected,
			"db_type":          "orm.Db",
		}).Response(ctx)
	}
}
