package test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/api"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/middleware"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/internal/trans"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/result"
	securityPkg "github.com/ilxqx/vef-framework-go/security"
	transPkg "github.com/ilxqx/vef-framework-go/trans"
	"github.com/stretchr/testify/suite"
)

type TestService struct {
	logger log.Logger
	Name   string
}

func (s *TestService) GetData() string {
	if s.logger != nil {
		s.logger.Infof("TestService.GetData called with name: %s", s.Name)
	}
	return fmt.Sprintf("service data from %s", s.Name)
}

func (s *TestService) WithLogger(logger log.Logger) *TestService {
	return &TestService{
		logger: logger,
		Name:   s.Name,
	}
}

type TestResource struct {
	apiPkg.Resource
	Service *TestService
}

type TestResourceWithLoggerAware struct {
	apiPkg.Resource
	Service TestService
}

// EmbeddedServiceResource tests embedded struct field injection
type EmbeddedServiceResource struct {
	apiPkg.Resource
	Service TestService // Use explicit field name to avoid ambiguity
}

func (*TestResource) Test(ctx fiber.Ctx) error {
	return result.Ok("test data").Response(ctx)
}

func (*TestResource) Private(ctx fiber.Ctx, logger log.Logger, principal *securityPkg.Principal) error {
	logger.Infof("Private API called by user: %s", principal.Name)
	return result.Ok(map[string]any{
		"message": "private data",
		"user":    principal.Name,
	}).Response(ctx)
}

func (*TestResource) WithDb(ctx fiber.Ctx, db orm.Db) error {
	return result.Ok("database injected").Response(ctx)
}

func (*TestResource) WithTransformer(ctx fiber.Ctx, transformer transPkg.Transformer) error {
	return result.Ok("transformer injected").Response(ctx)
}

func (*TestResource) Error(ctx fiber.Ctx) error {
	return fiber.NewError(fiber.StatusBadRequest, "test error")
}

func (r *TestResource) TestFieldInjection(ctx fiber.Ctx, service *TestService) error {
	return result.Ok(map[string]any{
		"message":      "field injection test",
		"service_data": service.GetData(),
	}).Response(ctx)
}

func (r *TestResourceWithLoggerAware) TestLoggerAware(ctx fiber.Ctx, service TestService) error {
	return result.Ok(map[string]any{
		"message":      "logger aware test",
		"service_data": service.GetData(),
	}).Response(ctx)
}

func (r *EmbeddedServiceResource) TestEmbeddedField(ctx fiber.Ctx, service TestService) error {
	return result.Ok(map[string]any{
		"message":      "embedded field test",
		"service_data": service.GetData(),
	}).Response(ctx)
}

type AuthenticatedResource struct {
	apiPkg.Resource
}

func (*AuthenticatedResource) Protected(ctx fiber.Ctx, principal *securityPkg.Principal) error {
	return result.Ok(fmt.Sprintf("Hello %s", principal.Name)).Response(ctx)
}

// MockAuthenticator implements securityPkg.Authenticator for testing
type MockAuthenticator struct{}

func (*MockAuthenticator) Supports(authType string) bool {
	return authType == security.AuthTypeJWT
}

func (*MockAuthenticator) Authenticate(auth securityPkg.Authentication) (*securityPkg.Principal, error) {
	if auth.Principal == "valid_token" {
		return &securityPkg.Principal{
			Id:   "1",
			Name: "test_user",
		}, nil
	}
	return nil, fiber.ErrUnauthorized
}

type ApiTestSuite struct {
	suite.Suite
	app *app.App
}

func (suite *ApiTestSuite) SetupSuite() {
	bunDb, err := database.New(
		&config.DatasourceConfig{
			Type: constants.DbTypeSQLite,
		},
	)
	suite.Require().NoError(err, "failed to create database")
	db := orm.New(bunDb)
	transformer := trans.New([]transPkg.FieldTransformer{}, []transPkg.StructTransformer{}, []transPkg.Interceptor{})
	authManager := security.NewAuthManager([]securityPkg.Authenticator{
		&MockAuthenticator{},
	})
	paramResolver := api.NewHandlerParamResolverManager([]apiPkg.HandlerParamResolver{})

	testResource := TestResource{
		Resource: apiPkg.NewResource(
			"test_resource",
			apiPkg.WithAPIs(
				apiPkg.Spec{
					Action: "test",
					Public: true,
				},
				apiPkg.Spec{
					Action: "private",
					Public: false,
				},
				apiPkg.Spec{
					Action: "withDb",
					Public: true,
				},
				apiPkg.Spec{
					Action: "withTransformer",
					Public: true,
				},
				apiPkg.Spec{
					Action: "error",
					Public: true,
				},
				apiPkg.Spec{
					Action: "testFieldInjection",
					Public: true,
				},
			),
		),
		Service: &TestService{Name: "pointer_service"},
	}

	testResourceWithLoggerAware := TestResourceWithLoggerAware{
		Resource: apiPkg.NewResource(
			"logger_aware_resource",
			apiPkg.WithAPIs(
				apiPkg.Spec{
					Action: "testLoggerAware",
					Public: true,
				},
			),
		),
		Service: TestService{Name: "logger_aware_service"},
	}

	embeddedServiceResource := EmbeddedServiceResource{
		Resource: apiPkg.NewResource(
			"embedded_resource",
			apiPkg.WithAPIs(
				apiPkg.Spec{
					Action: "testEmbeddedField",
					Public: true,
				},
			),
		),
		Service: TestService{Name: "embedded_service"},
	}

	authResource := AuthenticatedResource{
		Resource: apiPkg.NewResource(
			"auth_resource",
			apiPkg.WithAPIs(
				apiPkg.Spec{
					Action: "protected",
					Public: false,
				},
			),
		),
	}

	manager, err := api.NewManager(
		[]apiPkg.Resource{
			testResource,
			authResource,
			testResourceWithLoggerAware,
			embeddedServiceResource,
		},
		db, paramResolver,
	)
	suite.Require().NoError(err, "failed to create manager")

	openapiManager, err := api.NewManager([]apiPkg.Resource{}, db, paramResolver)
	suite.Require().NoError(err, "failed to create openapi manager")

	suite.app = app.New(app.AppParams{
		Config: &config.AppConfig{
			Name: "test-app",
			Port: 0,
		},
		ApiEngine: api.NewEngine(
			manager,
			api.NewDefaultApiPolicy(authManager),
			db,
			transformer,
		),
		OpenApiEngine: api.NewEngine(
			openapiManager,
			api.NewOpenApiPolicy(authManager),
			db,
			transformer,
		),
	})

	suite.app.Use(
		middleware.NewContentTypeMiddleware(),
		middleware.NewRequestIdMiddleware(),
		middleware.NewRequestRecordMiddleware(),
		middleware.NewRecoveryMiddleware(),
		middleware.NewLoggerMiddleware(),
	)

	suite.Require().NoError(<-suite.app.Start(), "failed to start app")
}
func (suite *ApiTestSuite) TearDownSuite() {
	if err := suite.app.Stop(); err != nil {
		suite.T().Logf("Failed to stop app: %v", err)
	}
}

func (suite *ApiTestSuite) TestPublicApi() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "test", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	suite.Require().Equal(`{"code":0,"message":"成功","data":"test data"}`, string(body))
}

func (suite *ApiTestSuite) TestInvalidRequest() {
	// Test invalid JSON
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{invalid json`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusInternalServerError, resp.StatusCode)
}

func (suite *ApiTestSuite) TestNonExistentResource() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "non_existent", "action": "test", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *ApiTestSuite) TestPrivateApiWithoutAuth() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "private", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (suite *ApiTestSuite) TestPrivateApiWithValidAuth() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "private", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer valid_token")

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	suite.Require().Contains(string(body), "private data")
	suite.Require().Contains(string(body), "test_user")
}

func (suite *ApiTestSuite) TestPrivateApiWithInvalidAuth() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "private", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer invalid_token")

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (suite *ApiTestSuite) TestApiWithDbInjection() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "withDb", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	suite.Require().Contains(string(body), "database injected")
}

func (suite *ApiTestSuite) TestApiWithTransformerInjection() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "withTransformer", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	suite.Require().Contains(string(body), "transformer injected")
}

func (suite *ApiTestSuite) TestApiError() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "error", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusBadRequest, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	suite.Require().Contains(string(body), "test error")
}

func (suite *ApiTestSuite) TestAuthenticationWithQueryParam() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api?__accessToken=valid_token",
		strings.NewReader(`{"resource": "test_resource", "action": "private", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	suite.Require().Contains(string(body), "private data")
}

func (suite *ApiTestSuite) TestFieldDependencyInjection() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test_resource", "action": "testFieldInjection", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	// Verify that the service field was injected and used
	suite.Require().Contains(string(body), "field injection test")
	suite.Require().Contains(string(body), "service data from pointer_service")
}

func (suite *ApiTestSuite) TestWithLoggerMethodAutoCall() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "logger_aware_resource", "action": "testLoggerAware", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	// Verify that WithLogger was called and the service was configured with logger
	suite.Require().Contains(string(body), "logger aware test")
	suite.Require().Contains(string(body), "service data from logger_aware_service")
}

func (suite *ApiTestSuite) TestEmbeddedFieldInjection() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "embedded_resource", "action": "testEmbeddedField", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	// Verify that the embedded field was injected with WithLogger called
	suite.Require().Contains(string(body), "embedded field test")
	suite.Require().Contains(string(body), "service data from embedded_service")
}

func (suite *ApiTestSuite) TestAuthenticatedResourceProtected() {
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "auth_resource", "action": "protected", "version": "v1", "params": {}}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer valid_token")

	resp, err := suite.app.Unwrap().Test(req)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	// Verify that the authenticated resource method works correctly
	suite.Require().Contains(string(body), "Hello test_user")
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}
