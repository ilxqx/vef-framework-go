package test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/app"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// TestAPIBasicFlow tests the basic API request flow
func TestAPIBasicFlow(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	// Test simple action
	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/user",
		"action": "get",
		"version": "v1",
		"params": {"id": "123"}
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":{"id":"123","name":"User 123"}`)
}

// TestAPIWithDatabaseAccess tests API with database parameter injection
func TestAPIWithDatabaseAccess(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/user",
		"action": "list",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"db access works"`)
}

// TestAPIWithLogger tests API with logger parameter injection
func TestAPIWithLogger(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/user",
		"action": "log",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"logged"`)
}

// TestAPIMultipleResources tests multiple resources
func TestAPIMultipleResources(t *testing.T) {
	testApp, stop := newTestApp(t,
		NewTestResource,
		NewProductResource,
	)
	defer stop()

	// Test user resource
	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/user",
		"action": "get",
		"version": "v1",
		"params": {"id": "123"}
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"id":"123"`)

	// Test product resource
	resp = makeAPIRequest(t, testApp, `{
		"resource": "test/product",
		"action": "list",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body = readBody(t, resp)
	require.Contains(t, body, `"data":"products"`)
}

// TestAPIWithCustomParams tests API with custom parameter struct
func TestAPIWithCustomParams(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
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

// TestAPINotFound tests non-existent API
func TestAPINotFound(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/user",
		"action": "nonexistent",
		"version": "v1"
	}`)

	require.Equal(t, 404, resp.StatusCode)
}

// TestAPIInvalidRequest tests invalid request format
func TestAPIInvalidRequest(t *testing.T) {
	testApp, stop := newTestApp(t)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"invalid": "request"
	}`)

	// VEF returns 200 with error code in response body for validation errors
	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"code":1400`)
}

// TestAPIVersioning tests API versioning
func TestAPIVersioning(t *testing.T) {
	testApp, stop := newTestApp(t,
		NewVersionedResource,
		NewVersionedResourceV2,
	)
	defer stop()

	// Test v1
	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/versioned",
		"action": "info",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"version":"v1"`)

	// Test v2
	resp = makeAPIRequest(t, testApp, `{
		"resource": "test/versioned",
		"action": "info",
		"version": "v2"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body = readBody(t, resp)
	require.Contains(t, body, `"version":"v2"`)
}

// TestAPIParamValidation tests parameter validation
func TestAPIParamValidation(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	// Missing required parameter
	resp := makeAPIRequest(t, testApp, `{
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

// TestAPIEmailValidation tests email format validation
func TestAPIEmailValidation(t *testing.T) {
	testApp, stop := newTestApp(t, NewTestResource)
	defer stop()

	// Invalid email format
	resp := makeAPIRequest(t, testApp, `{
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

// TestAPIExplicitHandler tests Spec.Handler field
func TestAPIExplicitHandler(t *testing.T) {
	testApp, stop := newTestApp(t, NewExplicitHandlerResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/explicit",
		"action": "custom",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"explicit handler"`)
}

// TestAPIHandlerFactory tests handler factory pattern with db parameter
func TestAPIHandlerFactory(t *testing.T) {
	testApp, stop := newTestApp(t, NewFactoryResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/factory",
		"action": "query",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"factory handler with db"`)
}

// TestAPIHandlerFactoryNoParam tests handler factory pattern without parameters
func TestAPIHandlerFactoryNoParam(t *testing.T) {
	testApp, stop := newTestApp(t, NewNoParamFactoryResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/noparam",
		"action": "static",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"factory handler without params"`)
}

// TestAPINoReturnValue tests handler without return value
func TestAPINoReturnValue(t *testing.T) {
	testApp, stop := newTestApp(t, NewNoReturnResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/noreturn",
		"action": "ping",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
}

// TestAPIFieldInjection tests parameter injection from resource fields
func TestAPIFieldInjection(t *testing.T) {
	testApp, stop := newTestApp(t, NewFieldInjectionResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/field",
		"action": "check",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"service":"injected"`)
}

// TestAPIEmbeddedProvider tests API from embedded Provider
func TestAPIEmbeddedProvider(t *testing.T) {
	testApp, stop := newTestApp(t, NewEmbeddedProviderResource)
	defer stop()

	resp := makeAPIRequest(t, testApp, `{
		"resource": "test/embedded",
		"action": "provided",
		"version": "v1"
	}`)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"data":"from provider"`)
}

// TestAPIMultipartFormData tests API request with multipart/form-data format
func TestAPIMultipartFormData(t *testing.T) {
	testApp, stop := newTestApp(t, NewMultipartResource)
	defer stop()

	resp := makeAPIRequestMultipart(t, testApp, map[string]string{
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

// TestAPIRequestFormats tests both JSON and multipart/form-data request formats
func TestAPIRequestFormats(t *testing.T) {
	testApp, stop := newTestApp(t, NewFormatsResource)
	defer stop()

	// Test JSON format
	jsonResp := makeAPIRequest(t, testApp, `{
		"resource": "test/formats",
		"action": "echo",
		"version": "v1"
	}`)

	require.Equal(t, 200, jsonResp.StatusCode)
	jsonBody := readBody(t, jsonResp)
	require.Contains(t, jsonBody, `"message":"request received"`)
	require.Contains(t, jsonBody, `"data":"application/json`)

	// Test multipart/form-data format
	formResp := makeAPIRequestMultipart(t, testApp, map[string]string{
		"resource": "test/formats",
		"action":   "echo",
		"version":  "v1",
	})

	require.Equal(t, 200, formResp.StatusCode)
	formBody := readBody(t, formResp)
	require.Contains(t, formBody, `"message":"request received"`)
	require.Contains(t, formBody, `"data":"multipart/form-data`)
}

// TestAPIMultipartWithMultipleFiles tests uploading multiple files with different keys
func TestAPIMultipartWithMultipleFiles(t *testing.T) {
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

	resp := makeAPIRequestWithFiles(t, testApp, map[string]string{
		"resource": "test/upload",
		"action":   "multipleKeys",
		"version":  "v1",
		"params":   `{"userId":"123"}`,
	}, files)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"userId":"123"`)
	require.Contains(t, body, `"avatar":"avatar.png"`)
	require.Contains(t, body, `"document":"resume.pdf"`)
}

// TestAPIMultipartWithSameKeyFiles tests uploading multiple files with the same key
func TestAPIMultipartWithSameKeyFiles(t *testing.T) {
	testApp, stop := newTestApp(t, NewFileUploadResource)
	defer stop()

	files := map[string][]FileContent{
		"attachments": {
			{Filename: "file1.txt", Content: []byte("content of file 1")},
			{Filename: "file2.txt", Content: []byte("content of file 2")},
			{Filename: "file3.txt", Content: []byte("content of file 3")},
		},
	}

	resp := makeAPIRequestWithFiles(t, testApp, map[string]string{
		"resource": "test/upload",
		"action":   "sameKey",
		"version":  "v1",
		"params":   `{"category":"documents"}`,
	}, files)

	require.Equal(t, 200, resp.StatusCode)
	body := readBody(t, resp)
	require.Contains(t, body, `"category":"documents"`)
	require.Contains(t, body, `"fileCount":3`)
	require.Contains(t, body, `"attachments":["file1.txt","file2.txt","file3.txt"]`)
}

// TestAPIMultipartFilesWithParams tests uploading files along with other parameters
func TestAPIMultipartFilesWithParams(t *testing.T) {
	testApp, stop := newTestApp(t, NewFileUploadResource)
	defer stop()

	files := map[string][]FileContent{
		"image": {
			{Filename: "photo.jpg", Content: []byte("fake image data")},
		},
	}

	resp := makeAPIRequestWithFiles(t, testApp, map[string]string{
		"resource": "test/upload",
		"action":   "withParams",
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
		opts[i] = vef.ProvideAPIResource(ctor)
	}

	opts[len(opts)-1] = fx.Replace(&config.DatasourceConfig{
		Type: constants.DbSQLite,
	})

	return appTest.NewTestApp(t, opts...)
}

func makeAPIRequest(t *testing.T, app interface {
	Test(req *http.Request, timeout ...time.Duration) (*http.Response, error)
}, body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func makeAPIRequestMultipart(t *testing.T, app interface {
	Test(req *http.Request, timeout ...time.Duration) (*http.Response, error)
}, fields map[string]string) *http.Response {
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

// FileContent represents a file to be uploaded
type FileContent struct {
	Filename string
	Content  []byte
}

func makeAPIRequestWithFiles(t *testing.T, app interface {
	Test(req *http.Request, timeout ...time.Duration) (*http.Response, error)
}, fields map[string]string, files map[string][]FileContent) *http.Response {
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
			apiPkg.WithAPIs(
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
	Name  string `json:"name" validate:"required"`
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
			apiPkg.WithAPIs(
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
			apiPkg.WithAPIs(
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
			apiPkg.WithAPIs(
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
			apiPkg.WithAPIs(
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
			apiPkg.WithAPIs(
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
			apiPkg.WithAPIs(
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
			apiPkg.WithAPIs(
				apiPkg.Spec{Action: "ping", Public: true},
			),
		),
	}
}

func (r *NoReturnResource) Ping(ctx fiber.Ctx) {
	// No return value
	result.Ok("pong").Response(ctx)
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
			apiPkg.WithAPIs(
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

// Embedded Provider Resource - tests API from embedded Provider

type ProvidedAPI struct{}

func (p *ProvidedAPI) Provide() apiPkg.Spec {
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
	*ProvidedAPI
}

func NewEmbeddedProviderResource() apiPkg.Resource {
	return &EmbeddedProviderResource{
		Resource:    apiPkg.NewResource("test/embedded"),
		ProvidedAPI: &ProvidedAPI{},
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
			apiPkg.WithAPIs(
				apiPkg.Spec{Action: "import", Public: true},
			),
		),
	}
}

type ImportParams struct {
	apiPkg.In

	Name  string `json:"name" validate:"required"`
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
			apiPkg.WithAPIs(
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
			apiPkg.WithAPIs(
				apiPkg.Spec{Action: "multipleKeys", Public: true},
				apiPkg.Spec{Action: "sameKey", Public: true},
				apiPkg.Spec{Action: "withParams", Public: true},
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

	Title       string   `json:"title" validate:"required"`
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
