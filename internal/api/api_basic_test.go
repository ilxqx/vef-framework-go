package api_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"
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

// BasicApiTestSuite tests the basic API functionality using testify suite pattern.
// All test resources are registered once in SetupSuite for better performance.
type BasicApiTestSuite struct {
	suite.Suite

	app  *app.App
	bus  event.Bus
	stop func()
}

// SetupSuite initializes the test suite by creating a single App instance with all test resources.
func (suite *BasicApiTestSuite) SetupSuite() {
	suite.T().Log("Setting up BasicApiTestSuite - registering all test resources")

	// Collect all resource constructors
	resourceCtors := []any{
		NewTestResource,
		NewProductResource,
		NewVersionedResource,
		NewVersionedResourceV2,
		NewExplicitHandlerResource,
		NewFactoryResource,
		NewNoParamFactoryResource,
		NewNoReturnResource,
		NewFieldInjectionResource,
		NewEmbeddedProviderResource,
		NewMultipartResource,
		NewFormatsResource,
		NewFileUploadResource,
		NewAuditResource,
		NewMetaResource,
	}

	// Build fx options
	opts := make([]fx.Option, len(resourceCtors)+2)
	for i, ctor := range resourceCtors {
		opts[i] = vef.ProvideApiResource(ctor)
	}

	// Configure SQLite database
	opts[len(opts)-2] = fx.Replace(&config.DatasourceConfig{
		Type: constants.DbSQLite,
	})

	// Populate event bus for audit tests
	opts[len(opts)-1] = fx.Populate(&suite.bus)

	// Create test app
	suite.app, suite.stop = apptest.NewTestApp(suite.T(), opts...)

	suite.T().Log("BasicApiTestSuite setup complete - App instance ready")
}

// TearDownSuite cleans up the test suite by stopping the App instance.
func (suite *BasicApiTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down BasicApiTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("BasicApiTestSuite teardown complete")
}

// Helper methods for making API requests

func (suite *BasicApiTestSuite) makeApiRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// Use 10 second timeout to handle slower CI environments
	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *BasicApiTestSuite) makeApiRequestMultipart(fields map[string]string) *http.Response {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	for key, value := range fields {
		err := writer.WriteField(key, value)
		suite.Require().NoError(err, "Should write field successfully")
	}

	err := writer.Close()
	suite.Require().NoError(err, "Should close writer successfully")

	req := httptest.NewRequest(fiber.MethodPost, "/api", &buf)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())

	// Use 10 second timeout to handle slower CI environments
	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *BasicApiTestSuite) makeApiRequestWithFiles(fields map[string]string, files map[string][]FileContent) *http.Response {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	for key, value := range fields {
		err := writer.WriteField(key, value)
		suite.Require().NoError(err, "Should write field successfully")
	}

	for fieldName, fileList := range files {
		for _, file := range fileList {
			part, err := writer.CreateFormFile(fieldName, file.Filename)
			suite.Require().NoError(err, "Should create form file successfully")

			_, err = part.Write(file.Content)
			suite.Require().NoError(err, "Should write file content successfully")
		}
	}

	err := writer.Close()
	suite.Require().NoError(err, "Should close writer successfully")

	req := httptest.NewRequest(fiber.MethodPost, "/api", &buf)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())

	// Use 10 second timeout to handle slower CI environments
	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *BasicApiTestSuite) readBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err, "Should read response body successfully")

	return string(body)
}

// TestBasicFlow tests the basic API request flow.
func (suite *BasicApiTestSuite) TestBasicFlow() {
	suite.T().Log("Testing basic API request flow")

	resp := suite.makeApiRequest(`{
		"resource": "test/user",
		"action": "get",
		"version": "v1",
		"params": {"id": "123"}
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"data":{"id":"123","name":"User 123"}`, "Should return user data")
}

// TestDatabaseAccess tests API with database parameter injection.
func (suite *BasicApiTestSuite) TestDatabaseAccess() {
	suite.T().Log("Testing database parameter injection")

	resp := suite.makeApiRequest(`{
		"resource": "test/user",
		"action": "list",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"data":"db access works"`, "Should confirm database access")
}

// TestLoggerInjection tests API with logger parameter injection.
func (suite *BasicApiTestSuite) TestLoggerInjection() {
	suite.T().Log("Testing logger parameter injection")

	resp := suite.makeApiRequest(`{
		"resource": "test/user",
		"action": "log",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"data":"logged"`, "Should confirm logging works")
}

// TestMultipleResources tests multiple resources registered in the same app.
func (suite *BasicApiTestSuite) TestMultipleResources() {
	suite.T().Log("Testing multiple resources")

	suite.Run("UserResource", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/user",
			"action": "get",
			"version": "v1",
			"params": {"id": "123"}
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"id":"123"`, "Should return user ID")
	})

	suite.Run("ProductResource", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/product",
			"action": "list",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"data":"products"`, "Should return products")
	})
}

// TestCustomParams tests API with custom parameter struct.
func (suite *BasicApiTestSuite) TestCustomParams() {
	suite.T().Log("Testing custom parameter struct")

	resp := suite.makeApiRequest(`{
		"resource": "test/user",
		"action": "create",
		"version": "v1",
		"params": {
			"name": "John Doe",
			"email": "john@example.com"
		}
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"name":"John Doe"`, "Should return user name")
	suite.Contains(body, `"email":"john@example.com"`, "Should return user email")
}

// TestNotFound tests non-existent API endpoint.
func (suite *BasicApiTestSuite) TestNotFound() {
	suite.T().Log("Testing non-existent API endpoint")

	resp := suite.makeApiRequest(`{
		"resource": "test/user",
		"action": "nonexistent",
		"version": "v1"
	}`)

	suite.Equal(404, resp.StatusCode, "Should return 404 Not Found")
}

// TestInvalidRequest tests invalid request format.
func (suite *BasicApiTestSuite) TestInvalidRequest() {
	suite.T().Log("Testing invalid request format")

	resp := suite.makeApiRequest(`{
		"invalid": "request"
	}`)

	suite.Equal(200, resp.StatusCode, "VEF returns 200 with error code in body")
	body := suite.readBody(resp)
	suite.Contains(body, `"code":1400`, "Should return validation error code")
}

// TestParamValidation tests parameter validation.
func (suite *BasicApiTestSuite) TestParamValidation() {
	suite.T().Log("Testing parameter validation")

	resp := suite.makeApiRequest(`{
		"resource": "test/user",
		"action": "get",
		"version": "v1",
		"params": {}
	}`)

	suite.Equal(200, resp.StatusCode, "VEF returns 200 with error code in body")
	body := suite.readBody(resp)
	suite.Contains(body, `"code":1400`, "Should return validation error for missing required parameter")
}

// TestEmailValidation tests email format validation.
func (suite *BasicApiTestSuite) TestEmailValidation() {
	suite.T().Log("Testing email format validation")

	resp := suite.makeApiRequest(`{
		"resource": "test/user",
		"action": "create",
		"version": "v1",
		"params": {
			"name": "John Doe",
			"email": "invalid-email"
		}
	}`)

	suite.Equal(200, resp.StatusCode, "VEF returns 200 with error code in body")
	body := suite.readBody(resp)
	suite.Contains(body, `"code":1400`, "Should return validation error for invalid email format")
}

// TestVersioning tests API versioning.
func (suite *BasicApiTestSuite) TestVersioning() {
	suite.T().Log("Testing API versioning")

	suite.Run("VersionV1", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/versioned",
			"action": "info",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"version":"v1"`, "Should return v1 version info")
	})

	suite.Run("VersionV2", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/versioned",
			"action": "info",
			"version": "v2"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"version":"v2"`, "Should return v2 version info")
	})
}

// TestExplicitHandler tests API with explicit handler field.
func (suite *BasicApiTestSuite) TestExplicitHandler() {
	suite.T().Log("Testing explicit handler field")

	resp := suite.makeApiRequest(`{
		"resource": "test/explicit",
		"action": "custom",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"data":"explicit handler"`, "Should use explicit handler")
}

// TestHandlerFactory tests handler factory pattern with database parameter.
func (suite *BasicApiTestSuite) TestHandlerFactory() {
	suite.T().Log("Testing handler factory pattern with db parameter")

	resp := suite.makeApiRequest(`{
		"resource": "test/factory",
		"action": "query",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"data":"factory handler with db"`, "Should use factory handler with db")
}

// TestHandlerFactoryNoParam tests handler factory pattern without parameters.
func (suite *BasicApiTestSuite) TestHandlerFactoryNoParam() {
	suite.T().Log("Testing handler factory pattern without parameters")

	resp := suite.makeApiRequest(`{
		"resource": "test/noparam",
		"action": "static",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"data":"factory handler without params"`, "Should use factory handler without params")
}

// TestNoReturnValue tests handler without return value.
func (suite *BasicApiTestSuite) TestNoReturnValue() {
	suite.T().Log("Testing handler without return value")

	resp := suite.makeApiRequest(`{
		"resource": "test/noreturn",
		"action": "ping",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
}

// TestFieldInjection tests parameter injection from resource fields.
func (suite *BasicApiTestSuite) TestFieldInjection() {
	suite.T().Log("Testing parameter injection from resource fields")

	resp := suite.makeApiRequest(`{
		"resource": "test/field",
		"action": "check",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"service":"injected"`, "Should inject service from field")
}

// TestEmbeddedProvider tests API from embedded Provider.
func (suite *BasicApiTestSuite) TestEmbeddedProvider() {
	suite.T().Log("Testing API from embedded Provider")

	resp := suite.makeApiRequest(`{
		"resource": "test/embedded",
		"action": "provided",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"data":"from provider"`, "Should use API from embedded provider")
}

// TestMultipartFormData tests API request with multipart/form-data format.
func (suite *BasicApiTestSuite) TestMultipartFormData() {
	suite.T().Log("Testing multipart/form-data request format")

	resp := suite.makeApiRequestMultipart(map[string]string{
		"resource": "test/multipart",
		"action":   "import",
		"version":  "v1",
		"params":   `{"name":"John Doe","email":"john@example.com"}`,
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"name":"John Doe"`, "Should return user name")
	suite.Contains(body, `"email":"john@example.com"`, "Should return user email")
}

// TestRequestFormats tests both JSON and multipart/form-data request formats.
func (suite *BasicApiTestSuite) TestRequestFormats() {
	suite.T().Log("Testing different request formats")

	suite.Run("JSONFormat", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/formats",
			"action": "echo",
			"version": "v1"
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"message":"request received"`, "Should confirm request received")
		suite.Contains(body, `"data":"application/json`, "Should detect JSON content type")
	})

	suite.Run("MultipartFormat", func() {
		resp := suite.makeApiRequestMultipart(map[string]string{
			"resource": "test/formats",
			"action":   "echo",
			"version":  "v1",
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"message":"request received"`, "Should confirm request received")
		suite.Contains(body, `"data":"multipart/form-data`, "Should detect multipart content type")
	})
}

// TestMultipartWithMultipleFiles tests uploading multiple files with different keys.
func (suite *BasicApiTestSuite) TestMultipartWithMultipleFiles() {
	suite.T().Log("Testing multiple files with different keys")

	files := map[string][]FileContent{
		"avatar": {
			{Filename: "avatar.png", Content: []byte("fake avatar image data")},
		},
		"document": {
			{Filename: "resume.pdf", Content: []byte("fake pdf content")},
		},
	}

	resp := suite.makeApiRequestWithFiles(map[string]string{
		"resource": "test/upload",
		"action":   "multiple_keys",
		"version":  "v1",
		"params":   `{"userId":"123"}`,
	}, files)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"userId":"123"`, "Should return user ID")
	suite.Contains(body, `"avatar":"avatar.png"`, "Should return avatar filename")
	suite.Contains(body, `"document":"resume.pdf"`, "Should return document filename")
}

// TestMultipartWithSameKeyFiles tests uploading multiple files with the same key.
func (suite *BasicApiTestSuite) TestMultipartWithSameKeyFiles() {
	suite.T().Log("Testing multiple files with same key")

	files := map[string][]FileContent{
		"attachments": {
			{Filename: "file1.txt", Content: []byte("content of file 1")},
			{Filename: "file2.txt", Content: []byte("content of file 2")},
			{Filename: "file3.txt", Content: []byte("content of file 3")},
		},
	}

	resp := suite.makeApiRequestWithFiles(map[string]string{
		"resource": "test/upload",
		"action":   "same_key",
		"version":  "v1",
		"params":   `{"category":"documents"}`,
	}, files)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"category":"documents"`, "Should return category")
	suite.Contains(body, `"fileCount":3`, "Should return correct file count")
	suite.Contains(body, `"attachments":["file1.txt","file2.txt","file3.txt"]`, "Should return all filenames")
}

// TestMultipartFilesWithParams tests uploading files along with other parameters.
func (suite *BasicApiTestSuite) TestMultipartFilesWithParams() {
	suite.T().Log("Testing file upload with additional parameters")

	files := map[string][]FileContent{
		"image": {
			{Filename: "photo.jpg", Content: []byte("fake image data")},
		},
	}

	resp := suite.makeApiRequestWithFiles(map[string]string{
		"resource": "test/upload",
		"action":   "with_params",
		"version":  "v1",
		"params":   `{"title":"My Photo","description":"A beautiful sunset","tags":["nature","sunset"]}`,
	}, files)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"title":"My Photo"`, "Should return title")
	suite.Contains(body, `"description":"A beautiful sunset"`, "Should return description")
	suite.Contains(body, `"tags":["nature","sunset"]`, "Should return tags")
	suite.Contains(body, `"image":"photo.jpg"`, "Should return image filename")
}

// FileContent represents a file to be uploaded.
type FileContent struct {
	Filename string
	Content  []byte
}

// Test Resources

type TestUserResource struct {
	api.Resource
}

func NewTestResource() api.Resource {
	return &TestUserResource{
		Resource: api.NewResource(
			"test/user",
			api.WithApis(
				api.Spec{Action: "get", Public: true},
				api.Spec{Action: "list", Public: true},
				api.Spec{Action: "create", Public: true},
				api.Spec{Action: "log", Public: true},
			),
		),
	}
}

type GetUserParams struct {
	api.P

	ID string `json:"id" validate:"required"`
}

func (r *TestUserResource) Get(ctx fiber.Ctx, params GetUserParams) error {
	return result.Ok(map[string]string{
		"id":   params.ID,
		"name": "User " + params.ID,
	}).Response(ctx)
}

func (r *TestUserResource) List(ctx fiber.Ctx, db orm.Db) error {
	// Just verify db is injected
	if db != nil {
		return result.Ok("db access works").Response(ctx)
	}

	return result.Err("db not injected")
}

type CreateUserParams struct {
	api.P

	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func (r *TestUserResource) Create(ctx fiber.Ctx, params CreateUserParams) error {
	return result.Ok(map[string]string{
		"name":  params.Name,
		"email": params.Email,
	}).Response(ctx)
}

func (r *TestUserResource) Log(ctx fiber.Ctx, logger log.Logger) error {
	logger.Info("Test log message")

	return result.Ok("logged").Response(ctx)
}

// Product Resource

type ProductResource struct {
	api.Resource
}

func NewProductResource() api.Resource {
	return &ProductResource{
		Resource: api.NewResource(
			"test/product",
			api.WithApis(
				api.Spec{Action: "list", Public: true},
			),
		),
	}
}

func (r *ProductResource) List(ctx fiber.Ctx) error {
	return result.Ok("products").Response(ctx)
}

// Versioned Resource

type VersionedResource struct {
	api.Resource
}

func NewVersionedResource() api.Resource {
	return &VersionedResource{
		Resource: api.NewResource(
			"test/versioned",
			api.WithVersion(api.VersionV1),
			api.WithApis(
				api.Spec{Action: "info", Public: true},
			),
		),
	}
}

func (r *VersionedResource) Info(ctx fiber.Ctx) error {
	return result.Ok(map[string]string{
		"version": api.VersionV1,
	}).Response(ctx)
}

// V2 Resource

type VersionedResourceV2 struct {
	api.Resource
}

func NewVersionedResourceV2() api.Resource {
	return &VersionedResourceV2{
		Resource: api.NewResource(
			"test/versioned",
			api.WithVersion(api.VersionV2),
			api.WithApis(
				api.Spec{Action: "info", Public: true},
			),
		),
	}
}

func (r *VersionedResourceV2) Info(ctx fiber.Ctx) error {
	return result.Ok(map[string]string{
		"version": api.VersionV2,
	}).Response(ctx)
}

// Explicit Handler Resource - tests Spec.Handler field

type ExplicitHandlerResource struct {
	api.Resource
}

func NewExplicitHandlerResource() api.Resource {
	return &ExplicitHandlerResource{
		Resource: api.NewResource(
			"test/explicit",
			api.WithApis(
				api.Spec{
					Action: "custom",
					Public: true,
					Handler: func(ctx fiber.Ctx) error {
						return result.Ok("explicit handler").Response(ctx)
					},
				},
			),
		),
	}
}

// Factory Resource - tests handler factory pattern

type FactoryResource struct {
	api.Resource
}

func NewFactoryResource() api.Resource {
	return &FactoryResource{
		Resource: api.NewResource(
			"test/factory",
			api.WithApis(
				api.Spec{Action: "query", Public: true},
			),
		),
	}
}

func (r *FactoryResource) Query(db orm.Db) func(ctx fiber.Ctx) error {
	// This is a handler factory - it receives db and returns a handler
	return func(ctx fiber.Ctx) error {
		if db != nil {
			return result.Ok("factory handler with db").Response(ctx)
		}

		return result.Err("db not available")
	}
}

// NoParamFactory Resource - tests handler factory without parameters

type NoParamFactoryResource struct {
	api.Resource
}

func NewNoParamFactoryResource() api.Resource {
	return &NoParamFactoryResource{
		Resource: api.NewResource(
			"test/noparam",
			api.WithApis(
				api.Spec{Action: "static", Public: true},
			),
		),
	}
}

func (r *NoParamFactoryResource) Static() func(ctx fiber.Ctx) error {
	// This is a handler factory without parameters - it returns a handler
	return func(ctx fiber.Ctx) error {
		return result.Ok("factory handler without params").Response(ctx)
	}
}

// NoReturn Resource - tests handler without return value

type NoReturnResource struct {
	api.Resource
}

func NewNoReturnResource() api.Resource {
	return &NoReturnResource{
		Resource: api.NewResource(
			"test/noreturn",
			api.WithApis(
				api.Spec{Action: "ping", Public: true},
			),
		),
	}
}

func (r *NoReturnResource) Ping(ctx fiber.Ctx, logger log.Logger) {
	// No return value
	if err := result.Ok("pong").Response(ctx); err != nil {
		logger.Errorf("Failed to send response: %v", err)
	}
}

// Field Injection Resource - tests parameter injection from struct fields

type TestService struct {
	Name string
}

type FieldInjectionResource struct {
	api.Resource

	Service *TestService
}

func NewFieldInjectionResource() api.Resource {
	return &FieldInjectionResource{
		Resource: api.NewResource(
			"test/field",
			api.WithApis(
				api.Spec{Action: "check", Public: true},
			),
		),
		Service: &TestService{Name: "injected"},
	}
}

func (r *FieldInjectionResource) Check(ctx fiber.Ctx, service *TestService) error {
	if service != nil {
		return result.Ok(map[string]string{
			"service": service.Name,
		}).Response(ctx)
	}

	return result.Err("service not injected")
}

// Embedded Provider Resource - tests Api from embedded Provider

type ProvidedApi struct{}

func (p *ProvidedApi) Provide() api.Spec {
	return api.Spec{
		Action: "provided",
		Public: true,
		Handler: func(ctx fiber.Ctx) error {
			return result.Ok("from provider").Response(ctx)
		},
	}
}

type EmbeddedProviderResource struct {
	api.Resource
	*ProvidedApi
}

func NewEmbeddedProviderResource() api.Resource {
	return &EmbeddedProviderResource{
		Resource:    api.NewResource("test/embedded"),
		ProvidedApi: &ProvidedApi{},
	}
}

// Multipart Resource - tests multipart/form-data request handling

type MultipartResource struct {
	api.Resource
}

func NewMultipartResource() api.Resource {
	return &MultipartResource{
		Resource: api.NewResource(
			"test/multipart",
			api.WithApis(
				api.Spec{Action: "import", Public: true},
			),
		),
	}
}

type ImportParams struct {
	api.P

	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func (r *MultipartResource) Import(ctx fiber.Ctx, params ImportParams) error {
	return result.Ok(params).Response(ctx)
}

// Formats Resource - tests both JSON and multipart/form-data request formats

type FormatsResource struct {
	api.Resource
}

func NewFormatsResource() api.Resource {
	return &FormatsResource{
		Resource: api.NewResource(
			"test/formats",
			api.WithApis(
				api.Spec{Action: "echo", Public: true},
			),
		),
	}
}

func (r *FormatsResource) Echo(ctx fiber.Ctx) error {
	// Just verify the request was processed successfully
	return result.Ok(ctx.Get(fiber.HeaderContentType), result.WithMessage("request received")).
		Response(ctx)
}

// FileUpload Resource - tests file upload handling

type FileUploadResource struct {
	api.Resource
}

func NewFileUploadResource() api.Resource {
	return &FileUploadResource{
		Resource: api.NewResource(
			"test/upload",
			api.WithApis(
				api.Spec{Action: "multiple_keys", Public: true},
				api.Spec{Action: "same_key", Public: true},
				api.Spec{Action: "with_params", Public: true},
			),
		),
	}
}

type MultipleKeysParams struct {
	api.P

	UserId   string `json:"userId" validate:"required"`
	Avatar   *multipart.FileHeader
	Document *multipart.FileHeader
}

func (r *FileUploadResource) MultipleKeys(ctx fiber.Ctx, params MultipleKeysParams) error {
	response := fiber.Map{
		"userId":   params.UserId,
		"avatar":   params.Avatar.Filename,
		"document": params.Document.Filename,
	}

	return result.Ok(response).Response(ctx)
}

type SameKeyParams struct {
	api.P

	Category    string `json:"category" validate:"required"`
	Attachments []*multipart.FileHeader
}

func (r *FileUploadResource) SameKey(ctx fiber.Ctx, params SameKeyParams) error {
	response := fiber.Map{
		"category":  params.Category,
		"fileCount": len(params.Attachments),
		"attachments": lo.Map(params.Attachments, func(attachment *multipart.FileHeader, _ int) string {
			return attachment.Filename
		}),
	}

	return result.Ok(response).Response(ctx)
}

type WithParamsParams struct {
	api.P

	Title       string   `json:"title"       validate:"required"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Image       *multipart.FileHeader
}

func (r *FileUploadResource) WithParams(ctx fiber.Ctx, params WithParamsParams) error {
	response := fiber.Map{
		"title":       params.Title,
		"description": params.Description,
		"tags":        params.Tags,
		"image":       params.Image.Filename,
	}

	return result.Ok(response).Response(ctx)
}

// Audit Resource - tests audit log functionality

type AuditResource struct {
	api.Resource
}

func NewAuditResource() api.Resource {
	return &AuditResource{
		Resource: api.NewResource(
			"test/audit",
			api.WithApis(
				api.Spec{Action: "success", EnableAudit: true, Public: true},
				api.Spec{Action: "failure", EnableAudit: true, Public: true},
				api.Spec{Action: "no_audit", EnableAudit: false, Public: true},
			),
		),
	}
}

type AuditSuccessParams struct {
	api.P

	Name string `json:"name" validate:"required"`
}

func (r *AuditResource) Success(ctx fiber.Ctx, params AuditSuccessParams) error {
	return result.Ok(fiber.Map{
		"name":    params.Name,
		"message": "success",
	}).Response(ctx)
}

func (r *AuditResource) Failure(ctx fiber.Ctx) error {
	return result.Err("Record not found", result.WithCode(result.ErrCodeRecordNotFound))
}

func (r *AuditResource) NoAudit(ctx fiber.Ctx) error {
	return result.Ok("no audit").Response(ctx)
}

// TestAuditSuccess tests audit event for successful requests.
func (suite *BasicApiTestSuite) TestAuditSuccess() {
	suite.T().Log("Testing audit event for successful request")

	var (
		auditEvents []*api.AuditEvent
		mu          sync.Mutex
	)

	unsubscribe := api.SubscribeAuditEvent(suite.bus, func(ctx context.Context, evt *api.AuditEvent) {
		mu.Lock()
		defer mu.Unlock()

		auditEvents = append(auditEvents, evt)
	})
	defer unsubscribe()

	resp := suite.makeApiRequest(`{
		"resource": "test/audit",
		"action": "success",
		"version": "v1",
		"params": {"name": "test-user"}
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	suite.Len(auditEvents, 1, "Should receive exactly one audit event")

	evt := auditEvents[0]
	suite.Equal("test/audit", evt.Resource, "Resource should match")
	suite.Equal("success", evt.Action, "Action should match")
	suite.Equal("v1", evt.Version, "Version should match")
	suite.Equal(result.OkCode, evt.ResultCode, "Result code should be success")
	suite.NotEmpty(evt.RequestId, "Request ID should be set")
	suite.NotEmpty(evt.RequestIP, "Request IP should be set")
	suite.NotNil(evt.RequestParams, "Request params should not be nil")
	suite.Equal("test-user", evt.RequestParams["name"], "Request params should contain name")
	suite.GreaterOrEqual(evt.ElapsedTime, 0, "Elapsed time should be non-negative")
}

// TestAuditFailure tests audit event for failed requests.
func (suite *BasicApiTestSuite) TestAuditFailure() {
	suite.T().Log("Testing audit event for failed request")

	var (
		auditEvents []*api.AuditEvent
		mu          sync.Mutex
	)

	unsubscribe := api.SubscribeAuditEvent(suite.bus, func(ctx context.Context, evt *api.AuditEvent) {
		mu.Lock()
		defer mu.Unlock()

		auditEvents = append(auditEvents, evt)
	})
	defer unsubscribe()

	resp := suite.makeApiRequest(`{
		"resource": "test/audit",
		"action": "failure",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	suite.Len(auditEvents, 1, "Should receive exactly one audit event")

	evt := auditEvents[0]
	suite.Equal("test/audit", evt.Resource, "Resource should match")
	suite.Equal("failure", evt.Action, "Action should match")
	suite.Equal("v1", evt.Version, "Version should match")
	suite.Equal(result.ErrCodeRecordNotFound, evt.ResultCode, "Result code should be record not found")
	suite.Equal("Record not found", evt.ResultMessage, "Result message should match")
	suite.NotEmpty(evt.RequestId, "Request ID should be set")
	suite.GreaterOrEqual(evt.ElapsedTime, 0, "Elapsed time should be non-negative")
}

// TestAuditDisabled tests that audit events are not published when disabled.
func (suite *BasicApiTestSuite) TestAuditDisabled() {
	suite.T().Log("Testing audit disabled")

	var (
		auditEvents []*api.AuditEvent
		mu          sync.Mutex
	)

	unsubscribe := api.SubscribeAuditEvent(suite.bus, func(ctx context.Context, evt *api.AuditEvent) {
		mu.Lock()
		defer mu.Unlock()

		auditEvents = append(auditEvents, evt)
	})
	defer unsubscribe()

	resp := suite.makeApiRequest(`{
		"resource": "test/audit",
		"action": "no_audit",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	suite.Len(auditEvents, 0, "Should not receive any audit events when audit is disabled")
}

// Meta Resource - tests meta parameter injection

type MetaResource struct {
	api.Resource
}

func NewMetaResource() api.Resource {
	return &MetaResource{
		Resource: api.NewResource(
			"test/meta",
			api.WithApis(
				api.Spec{Action: "with_meta", Public: true},
				api.Spec{Action: "with_both", Public: true},
				api.Spec{Action: "meta_validation", Public: true},
			),
		),
	}
}

type RequestMeta struct {
	api.M

	RequestId string `json:"requestId"`
	ClientIP  string `json:"clientIp"`
	UserAgent string `json:"userAgent"`
}

func (r *MetaResource) WithMeta(ctx fiber.Ctx, meta RequestMeta) error {
	return result.Ok(fiber.Map{
		"requestId": meta.RequestId,
		"clientIp":  meta.ClientIP,
		"userAgent": meta.UserAgent,
	}).Response(ctx)
}

type CreateItemParams struct {
	api.P

	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type CreateItemMeta struct {
	api.M

	Source    string `json:"source" validate:"required"`
	Timestamp int64  `json:"timestamp"`
}

func (r *MetaResource) WithBoth(ctx fiber.Ctx, params CreateItemParams, meta CreateItemMeta) error {
	return result.Ok(fiber.Map{
		"params": fiber.Map{
			"name":        params.Name,
			"description": params.Description,
		},
		"meta": fiber.Map{
			"source":    meta.Source,
			"timestamp": meta.Timestamp,
		},
	}).Response(ctx)
}

type ValidatedMeta struct {
	api.M

	ApiKey  string `json:"apiKey" validate:"required,min=10"`
	Version string `json:"version" validate:"required,oneof=v1 v2 v3"`
}

func (r *MetaResource) MetaValidation(ctx fiber.Ctx, meta ValidatedMeta) error {
	return result.Ok(fiber.Map{
		"apiKey":  meta.ApiKey,
		"version": meta.Version,
	}).Response(ctx)
}

// TestWithMeta tests API with meta parameter injection.
func (suite *BasicApiTestSuite) TestWithMeta() {
	suite.T().Log("Testing meta parameter injection")

	resp := suite.makeApiRequest(`{
		"resource": "test/meta",
		"action": "with_meta",
		"version": "v1",
		"meta": {
			"requestId": "req-12345",
			"clientIp": "192.168.1.1",
			"userAgent": "TestClient/1.0"
		}
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"requestId":"req-12345"`, "Should return request ID")
	suite.Contains(body, `"clientIp":"192.168.1.1"`, "Should return client IP")
	suite.Contains(body, `"userAgent":"TestClient/1.0"`, "Should return user agent")
}

// TestWithParamsAndMeta tests API with both params and meta injection.
func (suite *BasicApiTestSuite) TestWithParamsAndMeta() {
	suite.T().Log("Testing both params and meta injection")

	resp := suite.makeApiRequest(`{
		"resource": "test/meta",
		"action": "with_both",
		"version": "v1",
		"params": {
			"name": "Test Item",
			"description": "This is a test item"
		},
		"meta": {
			"source": "web-app",
			"timestamp": 1234567890
		}
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"name":"Test Item"`, "Should return item name")
	suite.Contains(body, `"description":"This is a test item"`, "Should return item description")
	suite.Contains(body, `"source":"web-app"`, "Should return meta source")
	suite.Contains(body, `"timestamp":1234567890`, "Should return meta timestamp")
}

// TestMetaValidation tests meta parameter validation.
func (suite *BasicApiTestSuite) TestMetaValidation() {
	suite.T().Log("Testing meta parameter validation")

	suite.Run("ValidMeta", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/meta",
			"action": "meta_validation",
			"version": "v1",
			"meta": {
				"apiKey": "valid-key-12345",
				"version": "v2"
			}
		}`)

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")
		body := suite.readBody(resp)
		suite.Contains(body, `"apiKey":"valid-key-12345"`, "Should return API key")
		suite.Contains(body, `"version":"v2"`, "Should return version")
	})

	suite.Run("InvalidApiKey", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/meta",
			"action": "meta_validation",
			"version": "v1",
			"meta": {
				"apiKey": "short",
				"version": "v2"
			}
		}`)

		suite.Equal(200, resp.StatusCode, "VEF returns 200 with error code in body")
		body := suite.readBody(resp)
		suite.Contains(body, `"code":1400`, "Should return validation error for short API key")
	})

	suite.Run("InvalidVersion", func() {
		resp := suite.makeApiRequest(`{
			"resource": "test/meta",
			"action": "meta_validation",
			"version": "v1",
			"meta": {
				"apiKey": "valid-key-12345",
				"version": "v99"
			}
		}`)

		suite.Equal(200, resp.StatusCode, "VEF returns 200 with error code in body")
		body := suite.readBody(resp)
		suite.Contains(body, `"code":1400`, "Should return validation error for invalid version")
	})
}

// TestMissingMeta tests API when meta is not provided.
func (suite *BasicApiTestSuite) TestMissingMeta() {
	suite.T().Log("Testing missing meta parameter")

	resp := suite.makeApiRequest(`{
		"resource": "test/meta",
		"action": "with_meta",
		"version": "v1"
	}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")
	body := suite.readBody(resp)
	suite.Contains(body, `"requestId":""`, "Should return empty request ID")
	suite.Contains(body, `"clientIp":""`, "Should return empty client IP")
	suite.Contains(body, `"userAgent":""`, "Should return empty user agent")
}

// TestApiBasicSuite runs the basic api test suite.
func TestApiBasicSuite(t *testing.T) {
	suite.Run(t, new(BasicApiTestSuite))
}
