package mold_test

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
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/result"
)

// TransformerFactoryParamResolverTestSuite tests the mold.Transformer factory parameter resolver.
type TransformerFactoryParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *TransformerFactoryParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up TransformerFactoryParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestMoldFactoryResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("TransformerFactoryParamResolverTestSuite setup complete")
}

func (suite *TransformerFactoryParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down TransformerFactoryParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("TransformerFactoryParamResolverTestSuite teardown complete")
}

func (suite *TransformerFactoryParamResolverTestSuite) TestTransformerFactoryInjection() {
	suite.T().Log("Testing mold.Transformer factory parameter injection")

	suite.Run("FactoryReceivedTransformer", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/mold_factory",
			"action": "verify_factory",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"factory_injected":true`, "Factory should have received transformer")
		suite.Contains(body, `"transformer_type":"mold.Transformer"`, "Should identify transformer type")
	})
}

func (suite *TransformerFactoryParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *TransformerFactoryParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestTransformerFactoryParamResolverSuite(t *testing.T) {
	suite.Run(t, new(TransformerFactoryParamResolverTestSuite))
}

// Test resource using factory function

type TestMoldFactoryResource struct {
	api.Resource
}

func NewTestMoldFactoryResource() api.Resource {
	return &TestMoldFactoryResource{
		Resource: api.NewResource(
			"test/mold_factory",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "verify_factory", Public: true},
			),
		),
	}
}

func (r *TestMoldFactoryResource) VerifyFactory(transformer mold.Transformer) func(ctx fiber.Ctx) error {
	factoryInjected := transformer != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{
			"factory_injected":  factoryInjected,
			"transformer_type": "mold.Transformer",
		}).Response(ctx)
	}
}
