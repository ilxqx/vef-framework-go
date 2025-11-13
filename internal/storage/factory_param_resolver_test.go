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

// StorageServiceFactoryParamResolverTestSuite tests the storage.Service factory parameter resolver.
type StorageServiceFactoryParamResolverTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *StorageServiceFactoryParamResolverTestSuite) SetupSuite() {
	suite.T().Log("Setting up StorageServiceFactoryParamResolverTestSuite")

	opts := []fx.Option{
		vef.ProvideApiResource(NewTestStorageFactoryResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
		fx.Replace(&config.StorageConfig{
			Provider: constants.StorageMemory,
		}),
	}

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("StorageServiceFactoryParamResolverTestSuite setup complete")
}

func (suite *StorageServiceFactoryParamResolverTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down StorageServiceFactoryParamResolverTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("StorageServiceFactoryParamResolverTestSuite teardown complete")
}

func (suite *StorageServiceFactoryParamResolverTestSuite) TestStorageServiceFactoryInjection() {
	suite.T().Log("Testing storage.Service factory parameter injection")

	suite.Run("FactoryReceivedService", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/storage_factory",
			"action": "verify_factory",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"factory_injected":true`, "Factory should have received service")
		suite.Contains(body, `"service_type":"storage.Service"`, "Should identify service type")
	})
}

func (suite *StorageServiceFactoryParamResolverTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *StorageServiceFactoryParamResolverTestSuite) readBody(resp *http.Response) string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Warning: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

func TestStorageServiceFactoryParamResolverSuite(t *testing.T) {
	suite.Run(t, new(StorageServiceFactoryParamResolverTestSuite))
}

// Test resource using factory function

type TestStorageFactoryResource struct {
	api.Resource
}

func NewTestStorageFactoryResource() api.Resource {
	return &TestStorageFactoryResource{
		Resource: api.NewResource(
			"test/storage_factory",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "verify_factory", Public: true},
			),
		),
	}
}

func (r *TestStorageFactoryResource) VerifyFactory(service storage.Service) func(ctx fiber.Ctx) error {
	factoryInjected := service != nil

	return func(ctx fiber.Ctx) error {
		return result.Ok(map[string]any{
			"factory_injected": factoryInjected,
			"service_type":     "storage.Service",
		}).Response(ctx)
	}
}
