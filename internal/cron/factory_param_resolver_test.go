package cron_test

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
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/result"
)

// SchedulerFactoryParamResolverTestSuite tests the cron.Scheduler factory parameter resolver.
type SchedulerFactoryParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *SchedulerFactoryParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up SchedulerFactoryParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestCronFactoryResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("SchedulerFactoryParamResolverTestSuite setup complete")
}

func (suite *SchedulerFactoryParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down SchedulerFactoryParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("SchedulerFactoryParamResolverTestSuite teardown complete")
}

func (suite *SchedulerFactoryParamResolverTestSuite) TestSchedulerFactoryInjection() {
	suite.T().Log("Testing cron.Scheduler factory parameter injection")

	suite.Run("FactoryReceivedScheduler", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/cron_factory",
			"action": "verify_factory",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"factory_injected":true`, "Factory should have received scheduler")
		suite.Contains(body, `"scheduler_type":"cron.Scheduler"`, "Should identify scheduler type")
	})
}

func (suite *SchedulerFactoryParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *SchedulerFactoryParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestSchedulerFactoryParamResolverSuite(t *testing.T) {
	suite.Run(t, new(SchedulerFactoryParamResolverTestSuite))
}

// Test resource using factory function

type TestCronFactoryResource struct {
	api.Resource
}

func NewTestCronFactoryResource() api.Resource {
	return &TestCronFactoryResource{
		Resource: api.NewResource(
			"test/cron_factory",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "verify_factory", Public: true},
			),
		),
	}
}

func (r *TestCronFactoryResource) VerifyFactory(scheduler cron.Scheduler) func(ctx fiber.Ctx) error {
	factoryInjected := scheduler != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{
			"factory_injected": factoryInjected,
			"scheduler_type":   "cron.Scheduler",
		}).Response(ctx)
	}
}
