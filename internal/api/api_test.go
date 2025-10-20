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
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/app"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// TestApiBasicFlow tests the basic Api request flow.
func TestApiBasicFlow(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	// Test simple action
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "get",
		"version": "v1",
		"params": {"id": "123"}
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":{"id":"123","name":"User 123"}`)
}

// TestApiWithDatabaseAccess tests Api with database parameter injection.
func TestApiWithDatabaseAccess(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "list",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"db access works"`)
}

// TestApiWithLogger tests Api with logger parameter injection.
func TestApiWithLogger(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "log",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"logged"`)
}

// TestApiMultipleResources tests multiple resources.
func TestApiMultipleResources(t *testing.T) {
	testApp, stop := newTestApp(t,
		NewTestResource,
		NewProductResource,
	)
	defer stop()

	// Test user resource
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "get",
		"version": "v1",
		"params": {"id": "123"}
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"id":"123"`)

	// Test product resource
	resp = makeApiRequest(t, testApp, `{
		"resource": "test/product",
		"action": "list",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body = readBody(t, resp)
	require.Contains(t, body, `"data":"products"`)
}

// TestApiWithCustomParams tests Api with custom parameter struct.
func TestApiWithCustomParams(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "create",
		"version": "v1",
		"params": {
			"name": "John Doe",
			"email": "john@example.com"
		}
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"name":"John Doe"`)
	require.Contains(t, body, `"email":"john@example.com"`)
}

// TestApiNotFound tests non-existent Api.
func TestApiNotFound(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "nonexistent",
		"version": "v1"
	}`)

	require.Equal(t, 404, resp.StatusCode)
}

// TestApiInvalidRequest tests invalid request format.
func TestApiInvalidRequest(t *testing.T) {
	testApp, stop := newTestApp(t)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"invalid": "request"
	}`)

	// VEF returns 200 with error code in response body for validation errors
	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"code":1400`)
}

// TestApiVersioning tests Api versioning.
func TestApiVersioning(t *testing.T) {
	testApp, stop := newTestApp(t,
		NewVersionedResource,
		NewVersionedResourceV2,
	)
	defer stop()

	// Test v1
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/versioned",
		"action": "info",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"version":"v1"`)

	// Test v2
	resp = makeApiRequest(t, testApp, `{
		"resource": "test/versioned",
		"action": "info",
		"version": "v2"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body = readBody(t, resp)
	require.Contains(t, body, `"version":"v2"`)
}

// TestApiParamValidation tests parameter validation.
func TestApiParamValidation(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	// Missing required parameter
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "get",
		"version": "v1",
		"params": {}
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	// Should have validation error
	require.Contains(t, body, `"code":1400`)
}

// TestApiEmailValidation tests email format validation.
func TestApiEmailValidation(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	// Invalid email format
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/user",
		"action": "create",
		"version": "v1",
		"params": {
			"name": "John Doe",
			"email": "invalid-email"
		}
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	// Should have validation error for email
	require.Contains(t, body, `"code":1400`)
}

// TestApiExplicitHandler tests Api.Handler field.
func TestApiExplicitHandler(t *testing.T) {
	testApp, stop := newTestApp(t, NewExplicitHandlerResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/explicit",
		"action": "custom",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"explicit handler"`)
}

// TestApiHandlerFactory tests handler factory pattern with db parameter.
func TestApiHandlerFactory(t *testing.T) {
	testApp, stop := newTestApp(t, NewFactoryResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/factory",
		"action": "query",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"factory handler with db"`)
}

// TestApiHandlerFactoryNoParam tests handler factory pattern without parameters.
func TestApiHandlerFactoryNoParam(t *testing.T) {
	testApp, stop := newTestApp(t, NewNoParamFactoryResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/noparam",
		"action": "static",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"factory handler without params"`)
}

// TestApiNoReturnValue tests handler without return value.
func TestApiNoReturnValue(t *testing.T) {
	testApp, stop := newTestApp(t, NewNoReturnResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/noreturn",
		"action": "ping",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
}

// TestApiFieldInjection tests parameter injection from resource fields.
func TestApiFieldInjection(t *testing.T) {
	testApp, stop := newTestApp(t, NewFieldInjectionResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/field",
		"action": "check",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"service":"injected"`)
}

// TestApiEmbeddedProvider tests Api from embedded Provider.
func TestApiEmbeddedProvider(t *testing.T) {
	testApp, stop := newTestApp(t, NewEmbeddedProviderResource)
	defer stop()

	resp := makeApiRequest(t, testApp, `{
		"resource": "test/embedded",
		"action": "provided",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"from provider"`)
}

// TestApiMultipartFormData tests Api request with multipart/form-data format.
func TestApiMultipartFormData(t *testing.T) {
	testApp, stop := newTestApp(t, NewMultipartResource)
	defer stop()

	resp := makeApiRequestMultipart(t, testApp, map[string]string{
		"resource": "test/multipart",
		"action":   "import",
		"version":  "v1",
		"params":   `{"name":"John Doe","email":"john@example.com"}`,
	})

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"name":"John Doe"`)
	require.Contains(t, body, `"email":"john@example.com"`)
}

// TestApiRequestFormats tests both JSON and multipart/form-data request formats.
func TestApiRequestFormats(t *testing.T) {
	testApp, stop := newTestApp(t, NewFormatsResource)
	defer stop()

	// Test JSON format
	jsonResp := makeApiRequest(t, testApp, `{
		"resource": "test/formats",
		"action": "echo",
		"version": "v1"
	}`)

	require.Equal(t, 200, jsonResp.StatusCode)
	jsonBody := readBody(t, jsonResp)
	require.Contains(t, jsonBody, `"message":"request received"`)
	require.Contains(t, jsonBody, `"data":"application/json`)

	// Test multipart/form-data format
	formResp := makeApiRequestMultipart(t, testApp, map[string]string{
		"resource": "test/formats",
		"action":   "echo",
		"version":  "v1",
	})

	require.Equal(t, 200, formResp.StatusCode)
	formBody := readBody(t, formResp)
	require.Contains(t, formBody, `"message":"request received"`)
	require.Contains(t, formBody, `"data":"multipart/form-data`)
}

// TestApiMultipartWithMultipleFiles tests uploading multiple files with different keys.
func TestApiMultipartWithMultipleFiles(t *testing.T) {
	testApp, stop := newTestApp(t, NewFileUploadResource)
	defer stop()

	files := map[string][]FileContent{
		"avatar": {
			{Filename: "avatar.png", Content: []byte("fake avatar image data")},
		},
		"document": {
			{Filename: "resume.pdf", Content: []byte("fake pdf content")},
		},
	}

	resp := makeApiRequestWithFiles(t, testApp, map[string]string{
		"resource": "test/upload",
		"action":   "multiple_keys",
		"version":  "v1",
		"params":   `{"userId":"123"}`,
	}, files)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"userId":"123"`)
	require.Contains(t, body, `"avatar":"avatar.png"`)
	require.Contains(t, body, `"document":"resume.pdf"`)
}

// TestApiMultipartWithSameKeyFiles tests uploading multiple files with the same key.
func TestApiMultipartWithSameKeyFiles(t *testing.T) {
	testApp, stop := newTestApp(t, NewFileUploadResource)
	defer stop()

	files := map[string][]FileContent{
		"attachments": {
			{Filename: "file1.txt", Content: []byte("content of file 1")},
			{Filename: "file2.txt", Content: []byte("content of file 2")},
			{Filename: "file3.txt", Content: []byte("content of file 3")},
		},
	}

	resp := makeApiRequestWithFiles(t, testApp, map[string]string{
		"resource": "test/upload",
		"action":   "same_key",
		"version":  "v1",
		"params":   `{"category":"documents"}`,
	}, files)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"category":"documents"`)
	require.Contains(t, body, `"fileCount":3`)
	require.Contains(t, body, `"attachments":["file1.txt","file2.txt","file3.txt"]`)
}

// TestApiMultipartFilesWithParams tests uploading files along with other parameters.
func TestApiMultipartFilesWithParams(t *testing.T) {
	testApp, stop := newTestApp(t, NewFileUploadResource)
	defer stop()

	files := map[string][]FileContent{
		"image": {
			{Filename: "photo.jpg", Content: []byte("fake image data")},
		},
	}

	resp := makeApiRequestWithFiles(t, testApp, map[string]string{
		"resource": "test/upload",
		"action":   "with_params",
		"version":  "v1",
		"params":   `{"title":"My Photo","description":"A beautiful sunset","tags":["nature","sunset"]}`,
	}, files)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"title":"My Photo"`)
	require.Contains(t, body, `"description":"A beautiful sunset"`)
	require.Contains(t, body, `"tags":["nature","sunset"]`)
	require.Contains(t, body, `"image":"photo.jpg"`)
}

// Helper functions

func newTestApp(t *testing.T, resourceCtors ...any) (*app.App, func()) {
	opts := make([]fx.Option, len(resourceCtors)+1)
	for i, ctor := range resourceCtors {
		opts[i] = vef.ProvideApiResource(ctor)
	}

	opts[len(opts)-1] = fx.Replace(&config.DatasourceConfig{
		Type: constants.DbSQLite,
	})

	return appTest.NewTestApp(t, opts...)
}

func newTestAppWithBus(t *testing.T, resourceCtors ...any) (*app.App, event.Bus, func()) {
	var bus event.Bus

	opts := make([]fx.Option, len(resourceCtors)+2)
	for i, ctor := range resourceCtors {
		opts[i] = vef.ProvideApiResource(ctor)
	}

	opts[len(opts)-2] = fx.Replace(&config.DatasourceConfig{
		Type: constants.DbSQLite,
	})
	opts[len(opts)-1] = fx.Populate(&bus)

	testApp, stop := appTest.NewTestApp(t, opts...)

	return testApp, bus, stop
}

func makeApiRequest(t *testing.T, app interface {
	Test(req *http.Request, timeout ...time.Duration) (*http.Response, error)
}, body string,
) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

func makeApiRequestMultipart(t *testing.T, app interface {
	Test(req *http.Request, timeout ...time.Duration) (*http.Response, error)
}, fields map[string]string,
) *http.Response {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	// Add form fields
	for key, value := range fields {
		err := writer.WriteField(key, value)
		require.NoError(t, err)
	}

	err := writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", &buf)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

// FileContent represents a file to be uploaded.
type FileContent struct {
	Filename string
	Content  []byte
}

func makeApiRequestWithFiles(t *testing.T, app interface {
	Test(req *http.Request, timeout ...time.Duration) (*http.Response, error)
}, fields map[string]string, files map[string][]FileContent,
) *http.Response {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	// Add form fields
	for key, value := range fields {
		err := writer.WriteField(key, value)
		require.NoError(t, err)
	}

	// Add files
	for fieldName, fileList := range files {
		for _, file := range fileList {
			part, err := writer.CreateFormFile(fieldName, file.Filename)
			require.NoError(t, err)

			_, err = part.Write(file.Content)
			require.NoError(t, err)
		}
	}

	err := writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", &buf)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

func readBody(t *testing.T, resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	require.NoError(t, err)

	return string(body)
}

// Test Resources

type TestUserResource struct {
	apiPkg.Resource
}

func NewTestResource() apiPkg.Resource {
	return &TestUserResource{
		Resource: apiPkg.NewResource(
			"test/user",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "get", Public: true},
				apiPkg.Spec{Action: "list", Public: true},
				apiPkg.Spec{Action: "create", Public: true},
				apiPkg.Spec{Action: "log", Public: true},
			),
		),
	}
}

type GetUserParams struct {
	apiPkg.In

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
	apiPkg.In

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
	apiPkg.Resource
}

func NewProductResource() apiPkg.Resource {
	return &ProductResource{
		Resource: apiPkg.NewResource(
			"test/product",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "list", Public: true},
			),
		),
	}
}

func (r *ProductResource) List(ctx fiber.Ctx) error {
	return result.Ok("products").Response(ctx)
}

// Versioned Resource

type VersionedResource struct {
	apiPkg.Resource
}

func NewVersionedResource() apiPkg.Resource {
	return &VersionedResource{
		Resource: apiPkg.NewResource(
			"test/versioned",
			apiPkg.WithVersion(apiPkg.VersionV1),
			apiPkg.WithApis(
				apiPkg.Spec{Action: "info", Public: true},
			),
		),
	}
}

func (r *VersionedResource) Info(ctx fiber.Ctx) error {
	return result.Ok(map[string]string{
		"version": apiPkg.VersionV1,
	}).Response(ctx)
}

// V2 Resource

type VersionedResourceV2 struct {
	apiPkg.Resource
}

func NewVersionedResourceV2() apiPkg.Resource {
	return &VersionedResourceV2{
		Resource: apiPkg.NewResource(
			"test/versioned",
			apiPkg.WithVersion(apiPkg.VersionV2),
			apiPkg.WithApis(
				apiPkg.Spec{Action: "info", Public: true},
			),
		),
	}
}

func (r *VersionedResourceV2) Info(ctx fiber.Ctx) error {
	return result.Ok(map[string]string{
		"version": apiPkg.VersionV2,
	}).Response(ctx)
}

// Explicit Handler Resource - tests Spec.Handler field

type ExplicitHandlerResource struct {
	apiPkg.Resource
}

func NewExplicitHandlerResource() apiPkg.Resource {
	return &ExplicitHandlerResource{
		Resource: apiPkg.NewResource(
			"test/explicit",
			apiPkg.WithApis(
				apiPkg.Spec{
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
	apiPkg.Resource
}

func NewFactoryResource() apiPkg.Resource {
	return &FactoryResource{
		Resource: apiPkg.NewResource(
			"test/factory",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "query", Public: true},
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
	apiPkg.Resource
}

func NewNoParamFactoryResource() apiPkg.Resource {
	return &NoParamFactoryResource{
		Resource: apiPkg.NewResource(
			"test/noparam",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "static", Public: true},
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
	apiPkg.Resource
}

func NewNoReturnResource() apiPkg.Resource {
	return &NoReturnResource{
		Resource: apiPkg.NewResource(
			"test/noreturn",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "ping", Public: true},
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
	apiPkg.Resource

	Service *TestService
}

func NewFieldInjectionResource() apiPkg.Resource {
	return &FieldInjectionResource{
		Resource: apiPkg.NewResource(
			"test/field",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "check", Public: true},
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

func (p *ProvidedApi) Provide() apiPkg.Spec {
	return apiPkg.Spec{
		Action: "provided",
		Public: true,
		Handler: func(ctx fiber.Ctx) error {
			return result.Ok("from provider").Response(ctx)
		},
	}
}

type EmbeddedProviderResource struct {
	apiPkg.Resource
	*ProvidedApi
}

func NewEmbeddedProviderResource() apiPkg.Resource {
	return &EmbeddedProviderResource{
		Resource:    apiPkg.NewResource("test/embedded"),
		ProvidedApi: &ProvidedApi{},
	}
}

// Multipart Resource - tests multipart/form-data request handling

type MultipartResource struct {
	apiPkg.Resource
}

func NewMultipartResource() apiPkg.Resource {
	return &MultipartResource{
		Resource: apiPkg.NewResource(
			"test/multipart",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "import", Public: true},
			),
		),
	}
}

type ImportParams struct {
	apiPkg.In

	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func (r *MultipartResource) Import(ctx fiber.Ctx, params ImportParams) error {
	return result.Ok(params).Response(ctx)
}

// Formats Resource - tests both JSON and multipart/form-data request formats

type FormatsResource struct {
	apiPkg.Resource
}

func NewFormatsResource() apiPkg.Resource {
	return &FormatsResource{
		Resource: apiPkg.NewResource(
			"test/formats",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "echo", Public: true},
			),
		),
	}
}

func (r *FormatsResource) Echo(ctx fiber.Ctx) error {
	// Just verify the request was processed successfully
	return result.OkWithMessage("request received", ctx.Get(fiber.HeaderContentType)).
		Response(ctx)
}

// FileUpload Resource - tests file upload handling

type FileUploadResource struct {
	apiPkg.Resource
}

func NewFileUploadResource() apiPkg.Resource {
	return &FileUploadResource{
		Resource: apiPkg.NewResource(
			"test/upload",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "multiple_keys", Public: true},
				apiPkg.Spec{Action: "same_key", Public: true},
				apiPkg.Spec{Action: "with_params", Public: true},
			),
		),
	}
}

type MultipleKeysParams struct {
	apiPkg.In

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
	apiPkg.In

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
	apiPkg.In

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
	apiPkg.Resource
}

func NewAuditResource() apiPkg.Resource {
	return &AuditResource{
		Resource: apiPkg.NewResource(
			"test/audit",
			apiPkg.WithApis(
				apiPkg.Spec{Action: "success", EnableAudit: true, Public: true},
				apiPkg.Spec{Action: "failure", EnableAudit: true, Public: true},
				apiPkg.Spec{Action: "no_audit", EnableAudit: false, Public: true},
			),
		),
	}
}

type AuditSuccessParams struct {
	apiPkg.In

	Name string `json:"name" validate:"required"`
}

func (r *AuditResource) Success(ctx fiber.Ctx, params AuditSuccessParams) error {
	return result.Ok(fiber.Map{
		"name":    params.Name,
		"message": "success",
	}).Response(ctx)
}

func (r *AuditResource) Failure(ctx fiber.Ctx) error {
	return result.ErrWithCode(result.ErrCodeRecordNotFound, "Record not found")
}

func (r *AuditResource) NoAudit(ctx fiber.Ctx) error {
	return result.Ok("no audit").Response(ctx)
}

// TestApiAuditSuccess tests audit event for successful requests.
func TestApiAuditSuccess(t *testing.T) {
	testApp, bus, stop := newTestAppWithBus(t, NewAuditResource)
	defer stop()

	// Capture audit events
	var (
		auditEvents []*apiPkg.AuditEvent
		mu          sync.Mutex
	)

	unsubscribe := apiPkg.SubscribeAuditEvent(bus, func(ctx context.Context, evt *apiPkg.AuditEvent) {
		mu.Lock()
		defer mu.Unlock()

		auditEvents = append(auditEvents, evt)
	})
	defer unsubscribe()

	// Make successful request
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/audit",
		"action": "success",
		"version": "v1",
		"params": {"name": "test-user"}
	}`)

	require.Equal(t, 200, resp.StatusCode)

	// Wait for async event processing
	time.Sleep(100 * time.Millisecond)

	// Verify audit event was published
	mu.Lock()
	defer mu.Unlock()

	require.Len(t, auditEvents, 1, "should receive exactly one audit event")

	evt := auditEvents[0]
	require.Equal(t, "test/audit", evt.Resource, "resource should match")
	require.Equal(t, "success", evt.Action, "action should match")
	require.Equal(t, "v1", evt.Version, "version should match")
	require.Equal(t, result.OkCode, evt.ResultCode, "result code should be success")
	require.NotEmpty(t, evt.RequestId, "request ID should be set")
	require.NotEmpty(t, evt.RequestIP, "request IP should be set")
	require.NotNil(t, evt.RequestParams, "request params should not be nil")
	require.Equal(t, "test-user", evt.RequestParams["name"], "request params should contain name")
	require.GreaterOrEqual(t, evt.ElapsedTime, 0, "elapsed time should be non-negative")
}

// TestApiAuditFailure tests audit event for failed requests.
func TestApiAuditFailure(t *testing.T) {
	testApp, bus, stop := newTestAppWithBus(t, NewAuditResource)
	defer stop()

	// Capture audit events
	var (
		auditEvents []*apiPkg.AuditEvent
		mu          sync.Mutex
	)

	unsubscribe := apiPkg.SubscribeAuditEvent(bus, func(ctx context.Context, evt *apiPkg.AuditEvent) {
		mu.Lock()
		defer mu.Unlock()

		auditEvents = append(auditEvents, evt)
	})
	defer unsubscribe()

	// Make failed request
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/audit",
		"action": "failure",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)

	// Wait for async event processing
	time.Sleep(100 * time.Millisecond)

	// Verify audit event was published
	mu.Lock()
	defer mu.Unlock()

	require.Len(t, auditEvents, 1, "should receive exactly one audit event")

	evt := auditEvents[0]
	require.Equal(t, "test/audit", evt.Resource, "resource should match")
	require.Equal(t, "failure", evt.Action, "action should match")
	require.Equal(t, "v1", evt.Version, "version should match")
	require.Equal(t, result.ErrCodeRecordNotFound, evt.ResultCode, "result code should be record not found")
	require.Equal(t, "Record not found", evt.ResultMessage, "result message should match")
	require.NotEmpty(t, evt.RequestId, "request ID should be set")
	require.GreaterOrEqual(t, evt.ElapsedTime, 0, "elapsed time should be non-negative")
}

// TestApiAuditDisabled tests that audit events are not published when disabled.
func TestApiAuditDisabled(t *testing.T) {
	testApp, bus, stop := newTestAppWithBus(t, NewAuditResource)
	defer stop()

	// Capture audit events
	var (
		auditEvents []*apiPkg.AuditEvent
		mu          sync.Mutex
	)

	unsubscribe := apiPkg.SubscribeAuditEvent(bus, func(ctx context.Context, evt *apiPkg.AuditEvent) {
		mu.Lock()
		defer mu.Unlock()

		auditEvents = append(auditEvents, evt)
	})
	defer unsubscribe()

	// Make request to endpoint with audit disabled
	resp := makeApiRequest(t, testApp, `{
		"resource": "test/audit",
		"action": "no_audit",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)

	// Wait for potential async event processing
	time.Sleep(100 * time.Millisecond)

	// Verify no audit event was published
	mu.Lock()
	defer mu.Unlock()

	require.Len(t, auditEvents, 0, "should not receive any audit events when audit is disabled")
}
