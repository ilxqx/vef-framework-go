package event_test

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
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/result"
)

// PublisherFactoryParamResolverTestSuite tests the event.Publisher factory parameter resolver.
type PublisherFactoryParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *PublisherFactoryParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up PublisherFactoryParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestEventFactoryResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("PublisherFactoryParamResolverTestSuite setup complete")
}

func (suite *PublisherFactoryParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down PublisherFactoryParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("PublisherFactoryParamResolverTestSuite teardown complete")
}

func (suite *PublisherFactoryParamResolverTestSuite) TestPublisherFactoryInjection() {
	suite.T().Log("Testing event.Publisher factory parameter injection")

	suite.Run("FactoryReceivedPublisher", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/event_factory",
			"action": "verify_factory",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"factory_injected":true`, "Factory should have received publisher")
		suite.Contains(body, `"publisher_type":"event.Publisher"`, "Should identify publisher type")
	})
}

func (suite *PublisherFactoryParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *PublisherFactoryParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestPublisherFactoryParamResolverSuite(t *testing.T) {
	suite.Run(t, new(PublisherFactoryParamResolverTestSuite))
}

// Test resource using factory function

type TestEventFactoryResource struct {
	api.Resource
}

func NewTestEventFactoryResource() api.Resource {
	return &TestEventFactoryResource{
		Resource: api.NewResource(
			"test/event_factory",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "verify_factory", Public: true},
			),
		),
	}
}

func (r *TestEventFactoryResource) VerifyFactory(publisher event.Publisher) func(ctx fiber.Ctx) error {
	factoryInjected := publisher != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{
			"factory_injected": factoryInjected,
			"publisher_type":   "event.Publisher",
		}).Response(ctx)
	}
}
