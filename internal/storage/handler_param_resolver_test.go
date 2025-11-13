package storage_test

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
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

// StorageServiceHandlerParamResolverTestSuite tests the storage.Service parameter resolver.
type StorageServiceHandlerParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *StorageServiceHandlerParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up StorageServiceHandlerParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestStorageResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
		fx.Replace(&config.StorageConfig{
			Provider: constants.StorageMemory,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("StorageServiceHandlerParamResolverTestSuite setup complete")
}

func (suite *StorageServiceHandlerParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down StorageServiceHandlerParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("StorageServiceHandlerParamResolverTestSuite teardown complete")
}

func (suite *StorageServiceHandlerParamResolverTestSuite) TestStorageServiceInjection() {
	suite.T().Log("Testing storage.Service parameter injection")

	suite.Run("ServiceInjected", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/storage",
			"action": "check_service",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"injected":true`, "Service should be injected")
		suite.Contains(body, `"service_type":"storage.Service"`, "Should identify service type")
	})
}

func (suite *StorageServiceHandlerParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *StorageServiceHandlerParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestStorageServiceHandlerParamResolverSuite(t *testing.T) {
	suite.Run(t, new(StorageServiceHandlerParamResolverTestSuite))
}

// Test resource

type TestStorageResource struct {
	api.Resource
}

func NewTestStorageResource() api.Resource {
	return &TestStorageResource{
		Resource: api.NewResource(
			"test/storage",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "check_service", Public: true},
			),
		),
	}
}

func (r *TestStorageResource) CheckService(ctx fiber.Ctx, service storage.Service) error {
	if service == nil {
		return result.Err("storage.Service not injected")
	}

	return result.Ok(map[string]any{
		"injected":     true,
		"service_type": "storage.Service",
	}).Response(ctx)
}
