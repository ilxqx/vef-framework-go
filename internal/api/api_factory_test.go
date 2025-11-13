package api_test

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
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// FactoryHandlerTestSuite tests handler factory functionality with various parameter combinations.
type FactoryHandlerTestSuite struct {
	suite.Suite

	app  *app.App
	stop func()
}

func (suite *FactoryHandlerTestSuite) SetupSuite() {
	suite.T().Log("Setting up FactoryHandlerTestSuite")

	resourceCtors := []any{
		NewSingleParamFactoryResource,
		NewStaticFactoryResource,
		NewMultiParamFactoryResource,
		NewFactoryWithErrorResource,
		NewFieldInjectionFactoryResource,
		NewExplicitFactoryHandlerResource,
	}

	opts := make([]fx.Option, len(resourceCtors)+1)
	for i, ctor := range resourceCtors {
		opts[i] = vef.ProvideApiResource(ctor)
	}

	opts[len(opts)-1] = fx.Replace(&config.DatasourceConfig{
		Type: constants.DbSQLite,
	})

	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("FactoryHandlerTestSuite setup complete")
}

func (suite *FactoryHandlerTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down FactoryHandlerTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("FactoryHandlerTestSuite teardown complete")
}

func (suite *FactoryHandlerTestSuite) makeRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Request should not fail")

	return resp
}

func (suite *FactoryHandlerTestSuite) readBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err, "Should read response body successfully")

	return string(body)
}

// TestSingleParamFactory tests factory with single db parameter.
func (suite *FactoryHandlerTestSuite) TestSingleParamFactory() {
	suite.T().Log("Testing factory function with single db parameter")

	resp := suite.makeRequest(`{
		"resource": "factory/single",
		"action": "query",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"status":"success"`, "Should return success status")
	suite.Contains(body, `"message":"Factory with db parameter"`, "Should return correct message")
}

// TestNoParamFactory tests factory without parameters.
func (suite *FactoryHandlerTestSuite) TestNoParamFactory() {
	suite.T().Log("Testing factory function without parameters")

	resp := suite.makeRequest(`{
		"resource": "factory/noparam",
		"action": "static",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"status":"success"`, "Should return success status")
	suite.Contains(body, `"message":"Factory without parameters"`, "Should return correct message")
}

// TestMultiParamFactory tests factory with multiple parameters.
func (suite *FactoryHandlerTestSuite) TestMultiParamFactory() {
	suite.T().Log("Testing factory function with multiple parameters")

	resp := suite.makeRequest(`{
		"resource": "factory/multi",
		"action": "process",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"status":"success"`, "Should return success status")
	suite.Contains(body, `"message":"Factory with db and publisher"`, "Should return correct message")
	suite.Contains(body, `"hasDb":true`, "Should have db")
	suite.Contains(body, `"hasPublisher":true`, "Should have publisher")
}

// TestFactoryWithError tests factory that returns error.
func (suite *FactoryHandlerTestSuite) TestFactoryWithError() {
	suite.T().Log("Testing factory function with error return")

	suite.Run("SuccessCase", func() {
		resp := suite.makeRequest(`{
			"resource": "factory/error",
			"action": "validate",
			"version": "v1",
			"params": {"valid": true}
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"status":"success"`, "Should return success status")
		suite.Contains(body, `"message":"Validation passed"`, "Should return correct message")
	})

	suite.Run("ErrorCase", func() {
		resp := suite.makeRequest(`{
			"resource": "factory/error",
			"action": "validate",
			"version": "v1",
			"params": {"valid": false}
		}`)

		suite.Equal(200, resp.StatusCode, "VEF returns 200 with error code in body")
		body := suite.readBody(resp)
		suite.Contains(body, `"code":1500`, "Should return error code")
		suite.Contains(body, `"message":"Validation failed"`, "Should return error message")
	})
}

// TestFieldInjectionFactory tests factory with field injection.
func (suite *FactoryHandlerTestSuite) TestFieldInjectionFactory() {
	suite.T().Log("Testing factory function with field injection")

	resp := suite.makeRequest(`{
		"resource": "factory/field",
		"action": "check",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"status":"success"`, "Should return success status")
	suite.Contains(body, `"serviceName":"test-service"`, "Should inject service from field")
	suite.Contains(body, `"hasDb":true`, "Should have db parameter")
}

// TestExplicitFactoryHandler tests factory with explicit Handler field.
func (suite *FactoryHandlerTestSuite) TestExplicitFactoryHandler() {
	suite.T().Log("Testing explicit factory handler field")

	resp := suite.makeRequest(`{
		"resource": "factory/explicit",
		"action": "custom",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"status":"success"`, "Should return success status")
	suite.Contains(body, `"message":"Explicit factory handler"`, "Should return correct message")
	suite.Contains(body, `"hasDb":true`, "Should have db from factory")
}

// Test Resources

type SingleParamFactoryResource struct {
	api.Resource
}

func NewSingleParamFactoryResource() api.Resource {
	return &SingleParamFactoryResource{
		Resource: api.NewResource(
			"factory/single",
			api.WithApis(
				api.Spec{Action: "query", Public: true},
			),
		),
	}
}

func (r *SingleParamFactoryResource) Query(db orm.Db) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		if db == nil {
			return result.Err("db not injected")
		}

		return result.Ok(fiber.Map{
			"status":  "success",
			"message": "Factory with db parameter",
		}).Response(ctx)
	}
}

type StaticFactoryResource struct {
	api.Resource
}

func NewStaticFactoryResource() api.Resource {
	return &StaticFactoryResource{
		Resource: api.NewResource(
			"factory/noparam",
			api.WithApis(
				api.Spec{Action: "static", Public: true},
			),
		),
	}
}

func (r *StaticFactoryResource) Static() func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		return result.Ok(fiber.Map{
			"status":  "success",
			"message": "Factory without parameters",
		}).Response(ctx)
	}
}

type MultiParamFactoryResource struct {
	api.Resource
}

func NewMultiParamFactoryResource() api.Resource {
	return &MultiParamFactoryResource{
		Resource: api.NewResource(
			"factory/multi",
			api.WithApis(
				api.Spec{Action: "process", Public: true},
			),
		),
	}
}

func (r *MultiParamFactoryResource) Process(db orm.Db, publisher event.Publisher) func(ctx fiber.Ctx, logger log.Logger) error {
	// Factory receives lifecycle dependencies (db, publisher)
	// Handler receives runtime dependencies (logger)
	return func(ctx fiber.Ctx, logger log.Logger) error {
		logger.Info("Factory handler with logger")

		return result.Ok(fiber.Map{
			"status":       "success",
			"message":      "Factory with db and publisher",
			"hasDb":        db != nil,
			"hasPublisher": publisher != nil,
		}).Response(ctx)
	}
}

type FactoryWithErrorResource struct {
	api.Resource
}

func NewFactoryWithErrorResource() api.Resource {
	return &FactoryWithErrorResource{
		Resource: api.NewResource(
			"factory/error",
			api.WithApis(
				api.Spec{Action: "validate", Public: true},
			),
		),
	}
}

type ValidateParams struct {
	api.P

	Valid bool `json:"valid"`
}

func (r *FactoryWithErrorResource) Validate(db orm.Db) (func(ctx fiber.Ctx, params ValidateParams) error, error) {
	if db == nil {
		return nil, result.Err("db not available")
	}

	return func(ctx fiber.Ctx, params ValidateParams) error {
		if !params.Valid {
			return result.Err("Validation failed", result.WithCode(1500))
		}

		return result.Ok(fiber.Map{
			"status":  "success",
			"message": "Validation passed",
		}).Response(ctx)
	}, nil
}

type FactoryService struct {
	Name string
}

type FieldInjectionFactoryResource struct {
	api.Resource

	Service *FactoryService
}

func NewFieldInjectionFactoryResource() api.Resource {
	return &FieldInjectionFactoryResource{
		Resource: api.NewResource(
			"factory/field",
			api.WithApis(
				api.Spec{Action: "check", Public: true},
			),
		),
		Service: &FactoryService{Name: "test-service"},
	}
}

func (r *FieldInjectionFactoryResource) Check(db orm.Db, service *FactoryService) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		return result.Ok(fiber.Map{
			"status":      "success",
			"serviceName": service.Name,
			"hasDb":       db != nil,
		}).Response(ctx)
	}
}

type ExplicitFactoryHandlerResource struct {
	api.Resource
}

func NewExplicitFactoryHandlerResource() api.Resource {
	return &ExplicitFactoryHandlerResource{
		Resource: api.NewResource(
			"factory/explicit",
			api.WithApis(
				api.Spec{
					Action: "custom",
					Public: true,
					Handler: func(db orm.Db) func(ctx fiber.Ctx) error {
						return func(ctx fiber.Ctx) error {
							return result.Ok(fiber.Map{
								"status":  "success",
								"message": "Explicit factory handler",
								"hasDb":   db != nil,
							}).Response(ctx)
						}
					},
				},
			),
		),
	}
}

func TestFactoryHandlerSuite(t *testing.T) {
	suite.Run(t, new(FactoryHandlerTestSuite))
}
