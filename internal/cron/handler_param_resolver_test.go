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

// SchedulerHandlerParamResolverTestSuite tests the cron.Scheduler handler parameter resolver.
type SchedulerHandlerParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *SchedulerHandlerParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up SchedulerHandlerParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestCronResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("SchedulerHandlerParamResolverTestSuite setup complete")
}

func (suite *SchedulerHandlerParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down SchedulerHandlerParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("SchedulerHandlerParamResolverTestSuite teardown complete")
}

func (suite *SchedulerHandlerParamResolverTestSuite) TestSchedulerInjection() {
	suite.T().Log("Testing cron.Scheduler parameter injection")

	suite.Run("SchedulerInjected", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/cron",
			"action": "check_scheduler",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"injected":true`, "Scheduler should be injected")
		suite.Contains(body, `"scheduler_type":"cron.Scheduler"`, "Should identify scheduler type")
	})
}

func (suite *SchedulerHandlerParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *SchedulerHandlerParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestSchedulerHandlerParamResolverSuite(t *testing.T) {
	suite.Run(t, new(SchedulerHandlerParamResolverTestSuite))
}

// Test resource

type TestCronResource struct {
	api.Resource
}

func NewTestCronResource() api.Resource {
	return &TestCronResource{
		Resource: api.NewResource(
			"test/cron",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "check_scheduler", Public: true},
			),
		),
	}
}

func (r *TestCronResource) CheckScheduler(ctx fiber.Ctx, scheduler cron.Scheduler) error {
	if scheduler == nil {
		return result.Err("cron.Scheduler not injected")
	}

	return result.Ok(map[string]any{
		"injected":       true,
		"scheduler_type": "cron.Scheduler",
	}).Response(ctx)
}
