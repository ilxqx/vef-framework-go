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

// PublisherHandlerParamResolverTestSuite tests the event.Publisher parameter resolver.
type PublisherHandlerParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *PublisherHandlerParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up PublisherHandlerParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestEventResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("PublisherHandlerParamResolverTestSuite setup complete")
}

func (suite *PublisherHandlerParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down PublisherHandlerParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("PublisherHandlerParamResolverTestSuite teardown complete")
}

func (suite *PublisherHandlerParamResolverTestSuite) TestPublisherInjection() {
	suite.T().Log("Testing event.Publisher parameter injection")

	suite.Run("PublisherInjected", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/event",
			"action": "check_publisher",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"injected":true`, "Publisher should be injected")
		suite.Contains(body, `"publisher_type":"event.Publisher"`, "Should identify publisher type")
	})
}

func (suite *PublisherHandlerParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *PublisherHandlerParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestPublisherHandlerParamResolverSuite(t *testing.T) {
	suite.Run(t, new(PublisherHandlerParamResolverTestSuite))
}

// Test resource

type TestEventResource struct {
	api.Resource
}

func NewTestEventResource() api.Resource {
	return &TestEventResource{
		Resource: api.NewResource(
			"test/event",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "check_publisher", Public: true},
			),
		),
	}
}

func (r *TestEventResource) CheckPublisher(ctx fiber.Ctx, publisher event.Publisher) error {
	if publisher == nil {
		return result.Err("event.Publisher not injected")
	}

	return result.Ok(map[string]any{
		"injected":       true,
		"publisher_type": "event.Publisher",
	}).Response(ctx)
}
