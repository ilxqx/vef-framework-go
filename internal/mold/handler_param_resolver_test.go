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

// TransformerHandlerParamResolverTestSuite tests the mold.Transformer parameter resolver.
type TransformerHandlerParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *TransformerHandlerParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up TransformerHandlerParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestMoldResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("TransformerHandlerParamResolverTestSuite setup complete")
}

func (suite *TransformerHandlerParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down TransformerHandlerParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("TransformerHandlerParamResolverTestSuite teardown complete")
}

func (suite *TransformerHandlerParamResolverTestSuite) TestTransformerInjection() {
	suite.T().Log("Testing mold.Transformer parameter injection")

	suite.Run("TransformerInjected", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/mold",
			"action": "check_transformer",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"injected":true`, "Transformer should be injected")
		suite.Contains(body, `"transformer_type":"mold.Transformer"`, "Should identify transformer type")
	})
}

func (suite *TransformerHandlerParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *TransformerHandlerParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestTransformerHandlerParamResolverSuite(t *testing.T) {
	suite.Run(t, new(TransformerHandlerParamResolverTestSuite))
}

// Test resource

type TestMoldResource struct {
	api.Resource
}

func NewTestMoldResource() api.Resource {
	return &TestMoldResource{
		Resource: api.NewResource(
			"test/mold",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "check_transformer", Public: true},
			),
		),
	}
}

func (r *TestMoldResource) CheckTransformer(ctx fiber.Ctx, transformer mold.Transformer) error {
	if transformer == nil {
		return result.Err("mold.Transformer not injected")
	}

	return result.Ok(map[string]any{
		"injected":         true,
		"transformer_type": "mold.Transformer",
	}).Response(ctx)
}
