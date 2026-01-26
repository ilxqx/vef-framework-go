package api_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/password"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// RESTMockUserLoader is a mock implementation of security.UserLoader for REST testing.
type RESTMockUserLoader struct {
	mock.Mock
}

func (m *RESTMockUserLoader) LoadByUsername(ctx context.Context, username string) (*security.Principal, string, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}

	return args.Get(0).(*security.Principal), args.String(1), args.Error(2)
}

func (m *RESTMockUserLoader) LoadByID(ctx context.Context, id string) (*security.Principal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*security.Principal), args.Error(1)
}

// RESTMockPermissionChecker is a mock implementation of security.PermissionChecker.
type RESTMockPermissionChecker struct {
	mock.Mock
}

func (m *RESTMockPermissionChecker) HasPermission(ctx context.Context, principal *security.Principal, permToken string) (bool, error) {
	args := m.Called(ctx, principal, permToken)

	return args.Bool(0), args.Error(1)
}

// TestRESTResource is a test resource for REST API testing.
type TestRESTResource struct {
	api.Resource
}

func NewTestRESTResource() api.Resource {
	return &TestRESTResource{
		Resource: api.NewRESTResource(
			"items",
			api.WithOperations(
				api.OperationSpec{
					Action:  "get",
					Public:  true,
					Handler: "Get",
				},
				api.OperationSpec{
					Action:  "post",
					Public:  true,
					Handler: "Post",
				},
				api.OperationSpec{
					Action:  "put",
					Handler: "Put",
				},
				api.OperationSpec{
					Action:  "delete",
					Handler: "Delete",
				},
				api.OperationSpec{
					Action:    "patch",
					Handler:   "Patch",
					PermToken: "items:patch",
				},
				api.OperationSpec{
					Action:    "post admin",
					Handler:   "Admin",
					PermToken: "items:admin",
				},
				api.OperationSpec{
					Action:  "get panic",
					Public:  true,
					Handler: "Panic",
				},
			),
		),
	}
}

type GetItemsParams struct {
	api.P

	ID      string `json:"id"`
	Keyword string `json:"keyword"`
	Page    string `json:"page"`
	Size    string `json:"size"`
}

type CreateItemParams struct {
	api.P

	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type UpdateItemParams struct {
	api.P

	ID   string `json:"id" validate:"required" label:"ID"`
	Name string `json:"name"`
}

type DeleteItemParams struct {
	api.P

	ID string `json:"id" validate:"required" label:"ID"`
}

type PatchItemParams struct {
	api.P

	ID     string `json:"id" validate:"required" label:"ID"`
	Status string `json:"status"`
}

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ItemListResponse struct {
	Items   []Item `json:"items"`
	Keyword string `json:"keyword"`
	Page    string `json:"page"`
	Size    string `json:"size"`
}

type CreatedItemResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type UpdatedItemResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UpdatedBy string `json:"updatedBy"`
}

type DeletedItemResponse struct {
	ID        string `json:"id"`
	DeletedBy string `json:"deletedBy"`
}

type PatchedItemResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Patched bool   `json:"patched"`
}

type AdminResponse struct {
	Action string `json:"action"`
	Admin  bool   `json:"admin"`
}

func (*TestRESTResource) Get(ctx fiber.Ctx, params *GetItemsParams) error {
	// If id is provided, return single item
	if params.ID != "" {
		if params.ID == "notfound" {
			return result.Result{Code: result.ErrCodeRecordNotFound, Message: i18n.T(result.ErrMessageRecordNotFound)}.Response(ctx)
		}

		return result.Ok(&Item{
			ID:   params.ID,
			Name: "Item " + params.ID,
		}).Response(ctx)
	}

	// Default values
	page := params.Page
	if page == "" {
		page = "1"
	}
	size := params.Size
	if size == "" {
		size = "10"
	}

	return result.Ok(&ItemListResponse{
		Items: []Item{
			{ID: "1", Name: "Item 1"},
			{ID: "2", Name: "Item 2"},
		},
		Keyword: params.Keyword,
		Page:    page,
		Size:    size,
	}).Response(ctx)
}

func (*TestRESTResource) Post(ctx fiber.Ctx, params *CreateItemParams) error {
	return result.Ok(&CreatedItemResponse{
		ID:    "new-id",
		Name:  params.Name,
		Price: params.Price,
	}).Response(ctx)
}

func (*TestRESTResource) Put(ctx fiber.Ctx, params *UpdateItemParams, principal *security.Principal) error {
	return result.Ok(&UpdatedItemResponse{
		ID:        params.ID,
		Name:      params.Name,
		UpdatedBy: principal.ID,
	}).Response(ctx)
}

func (*TestRESTResource) Delete(ctx fiber.Ctx, params *DeleteItemParams, principal *security.Principal) error {
	return result.Ok(&DeletedItemResponse{
		ID:        params.ID,
		DeletedBy: principal.ID,
	}).Response(ctx)
}

func (*TestRESTResource) Patch(ctx fiber.Ctx, params *PatchItemParams) error {
	return result.Ok(&PatchedItemResponse{
		ID:      params.ID,
		Status:  params.Status,
		Patched: true,
	}).Response(ctx)
}

func (*TestRESTResource) Admin(ctx fiber.Ctx) error {
	return result.Ok(&AdminResponse{
		Action: "admin",
		Admin:  true,
	}).Response(ctx)
}

func (*TestRESTResource) Panic(_ fiber.Ctx) error {
	panic("intentional panic for testing")
}

// RESTEngineTestSuite tests REST API engine functionality.
type RESTEngineTestSuite struct {
	suite.Suite

	ctx               context.Context
	app               *app.App
	stop              func()
	userLoader        *RESTMockUserLoader
	permissionChecker *RESTMockPermissionChecker
	jwtSecret         string
	testUser          *security.Principal
	adminUser         *security.Principal
	hashedPassword    string
}

func (suite *RESTEngineTestSuite) SetupSuite() {
	suite.T().Log("Setting up RESTEngineTestSuite")

	suite.ctx = context.Background()
	suite.jwtSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	suite.testUser = security.NewUser("user001", "Test User", "admin", "user")
	suite.testUser.Details = map[string]any{
		"email": "test@example.com",
	}

	// Admin user with items:admin permission
	suite.adminUser = security.NewUser("admin001", "Admin User", "superadmin")
	suite.adminUser.Details = map[string]any{
		"email": "admin@example.com",
	}

	var err error
	suite.hashedPassword, err = password.NewBcryptEncoder().Encode("password123")
	suite.Require().NoError(err)

	suite.userLoader = new(RESTMockUserLoader)
	suite.permissionChecker = new(RESTMockPermissionChecker)

	suite.setupTestApp()

	suite.T().Log("RESTEngineTestSuite setup complete")
}

func (suite *RESTEngineTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down RESTEngineTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("RESTEngineTestSuite teardown complete")
}

func (suite *RESTEngineTestSuite) SetupTest() {
	suite.userLoader.Calls = nil
	suite.permissionChecker.Calls = nil
}

func (suite *RESTEngineTestSuite) setupTestApp() {
	suite.userLoader.On("LoadByUsername", mock.Anything, "testuser").
		Return(suite.testUser, suite.hashedPassword, nil).
		Maybe()

	suite.userLoader.On("LoadByID", mock.Anything, "user001").
		Return(suite.testUser, nil).
		Maybe()

	// Admin user for admin-allow test
	suite.userLoader.On("LoadByUsername", mock.Anything, "adminuser").
		Return(suite.adminUser, suite.hashedPassword, nil).
		Maybe()

	suite.userLoader.On("LoadByID", mock.Anything, "admin001").
		Return(suite.adminUser, nil).
		Maybe()

	suite.permissionChecker.On("HasPermission", mock.Anything, mock.Anything, "items:patch").
		Return(true, nil).
		Maybe()

	// Permission denied for items:admin for regular user
	suite.permissionChecker.On("HasPermission", mock.Anything, suite.testUser, "items:admin").
		Return(false, nil).
		Maybe()

	// Permission allowed for items:admin for admin user
	suite.permissionChecker.On("HasPermission", mock.Anything, suite.adminUser, "items:admin").
		Return(true, nil).
		Maybe()

	suite.app, suite.stop = apptest.NewTestApp(
		suite.T(),
		fx.Supply(
			fx.Annotate(
				suite.userLoader,
				fx.As(new(security.UserLoader)),
			),
		),
		fx.Decorate(func() security.PermissionChecker {
			return suite.permissionChecker
		}),
		fx.Replace(
			&config.DatasourceConfig{
				Type: constants.SQLite,
			},
			&security.JWTConfig{
				Secret:   suite.jwtSecret,
				Audience: "test-app",
			},
		),
		fx.Provide(
			fx.Annotate(
				NewTestRESTResource,
				fx.As(new(api.Resource)),
				fx.ResultTags(`group:"vef:api:resources"`),
			),
		),
	)
}

func (suite *RESTEngineTestSuite) makeRESTRequest(method, path string, body string) *http.Response {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	return resp
}

func (suite *RESTEngineTestSuite) makeRESTRequestWithToken(method, path string, body string, token string) *http.Response {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set(fiber.HeaderAuthorization, constants.AuthSchemeBearer+" "+token)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	return resp
}

func (suite *RESTEngineTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			suite.T().Errorf("failed to close response body: %v", closeErr)
		}
	}()

	suite.Require().NoError(err)
	res, err := encoding.FromJSON[result.Result](string(body))
	suite.Require().NoError(err)

	return *res
}

func (suite *RESTEngineTestSuite) readDataAsMap(data any) map[string]any {
	m, ok := data.(map[string]any)
	suite.Require().True(ok, "Data should be a map")

	return m
}

func (suite *RESTEngineTestSuite) login() string {
	jsonBody, err := encoding.ToJSON(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"kind":        "password",
			"principal":   "testuser",
			"credentials": "password123",
		},
	})
	suite.Require().NoError(err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(body.Data)

	return tokens["accessToken"].(string)
}

func (suite *RESTEngineTestSuite) loginAsAdmin() string {
	jsonBody, err := encoding.ToJSON(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"kind":        "password",
			"principal":   "adminuser",
			"credentials": "password123",
		},
	})
	suite.Require().NoError(err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Admin login should succeed")

	tokens := suite.readDataAsMap(body.Data)

	return tokens["accessToken"].(string)
}

func (suite *RESTEngineTestSuite) TestGetList() {
	suite.T().Log("Testing GET list endpoint")

	resp := suite.makeRESTRequest(fiber.MethodGet, "/api/items", "")

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "GET list should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.NotNil(data["items"], "Should have items")
	suite.Equal("1", data["page"], "Default page should be 1")
	suite.Equal("10", data["size"], "Default size should be 10")
}

func (suite *RESTEngineTestSuite) TestGetListWithQueryParams() {
	suite.T().Log("Testing GET list with query parameters")

	resp := suite.makeRESTRequest(fiber.MethodGet, "/api/items?keyword=test&page=2&size=20", "")

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "GET list should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("test", data["keyword"], "Keyword should match")
	suite.Equal("2", data["page"], "Page should match")
	suite.Equal("20", data["size"], "Size should match")
}

func (suite *RESTEngineTestSuite) TestGetByID() {
	suite.T().Log("Testing GET by ID endpoint")

	resp := suite.makeRESTRequest(fiber.MethodGet, "/api/items?id=123", "")

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "GET by ID should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("123", data["id"], "ID should match")
	suite.Equal("Item 123", data["name"], "Name should match")
}

func (suite *RESTEngineTestSuite) TestGetByIDNotFound() {
	suite.T().Log("Testing GET by ID not found")

	resp := suite.makeRESTRequest(fiber.MethodGet, "/api/items?id=notfound", "")

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "GET by ID should fail for not found")
	suite.Equal(result.ErrCodeRecordNotFound, body.Code, "Should return record not found error")
}

func (suite *RESTEngineTestSuite) TestPostCreate() {
	suite.T().Log("Testing POST create endpoint")

	resp := suite.makeRESTRequest(fiber.MethodPost, "/api/items", `{"name":"New Item","price":100}`)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "POST create should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("new-id", data["id"], "Should have generated ID")
	suite.Equal("New Item", data["name"], "Name should match")
	suite.Equal(float64(100), data["price"], "Price should match")
}

func (suite *RESTEngineTestSuite) TestPostCreateInvalidBody() {
	suite.T().Log("Testing POST create with invalid body")

	resp := suite.makeRESTRequest(fiber.MethodPost, "/api/items", "{invalid json}")

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "POST create should fail with invalid body")
	suite.Equal(result.ErrCodeBadRequest, body.Code, "Should return bad request error")
}

func (suite *RESTEngineTestSuite) TestPutUpdateWithoutToken() {
	suite.T().Log("Testing PUT update without token")

	resp := suite.makeRESTRequest(fiber.MethodPut, "/api/items", `{"id":"123","name":"Updated Item"}`)

	suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")
}

func (suite *RESTEngineTestSuite) TestPutUpdateWithToken() {
	suite.T().Log("Testing PUT update with token")

	token := suite.login()

	resp := suite.makeRESTRequestWithToken(fiber.MethodPut, "/api/items", `{"id":"123","name":"Updated Item"}`, token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "PUT update should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("123", data["id"], "ID should match")
	suite.Equal("Updated Item", data["name"], "Name should match")
	suite.Equal("user001", data["updatedBy"], "UpdatedBy should match principal ID")
}

func (suite *RESTEngineTestSuite) TestDeleteWithoutToken() {
	suite.T().Log("Testing DELETE without token")

	resp := suite.makeRESTRequest(fiber.MethodDelete, "/api/items?id=123", "")

	suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")
}

func (suite *RESTEngineTestSuite) TestDeleteWithToken() {
	suite.T().Log("Testing DELETE with token")

	token := suite.login()

	resp := suite.makeRESTRequestWithToken(fiber.MethodDelete, "/api/items?id=123", "", token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "DELETE should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("123", data["id"], "ID should match")
	suite.Equal("user001", data["deletedBy"], "DeletedBy should match principal ID")
}

func (suite *RESTEngineTestSuite) TestPatchWithPermission() {
	suite.T().Log("Testing PATCH with permission")

	token := suite.login()

	resp := suite.makeRESTRequestWithToken(fiber.MethodPatch, "/api/items", `{"id":"123","status":"active"}`, token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "PATCH should succeed with permission")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("123", data["id"], "ID should match")
	suite.Equal("active", data["status"], "Status should match")
	suite.Equal(true, data["patched"], "Patched flag should be true")
}

func (suite *RESTEngineTestSuite) TestRouteNotFound() {
	suite.T().Log("Testing route not found")

	resp := suite.makeRESTRequest(fiber.MethodGet, "/api/nonexistent", "")

	// REST router returns 404 for unmatched routes since each route has its own middleware
	suite.Equal(404, resp.StatusCode, "Should return 404 Not Found for unmatched route")
}

func (suite *RESTEngineTestSuite) TestMethodNotAllowed() {
	suite.T().Log("Testing method not allowed")

	resp := suite.makeRESTRequest(fiber.MethodOptions, "/api/items", "")

	// REST router returns 405 Method Not Allowed for unregistered methods on existing paths
	suite.Equal(405, resp.StatusCode, "Should return 405 Method Not Allowed for unregistered method")
}

func (suite *RESTEngineTestSuite) TestInvalidToken() {
	suite.T().Log("Testing invalid token")

	resp := suite.makeRESTRequestWithToken(fiber.MethodPut, "/api/items", `{"id":"123","name":"Test"}`, "invalid.token.here")

	suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with invalid token")
	suite.Equal(result.ErrCodeTokenInvalid, body.Code, "Should return token invalid error")
}

func (suite *RESTEngineTestSuite) TestI18nErrorMessages() {
	suite.T().Log("Testing i18n error messages")

	resp := suite.makeRESTRequest(fiber.MethodPut, "/api/items", `{"id":"123","name":"Test"}`)

	suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail without token")
	suite.Equal(i18n.T(result.ErrMessageUnauthenticated), body.Message, "Should return i18n translated message")
}

func (suite *RESTEngineTestSuite) TestEmptyRequestBody() {
	suite.T().Log("Testing empty request body for POST")

	req := httptest.NewRequest(fiber.MethodPost, "/api/items", strings.NewReader(""))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	// Empty body is allowed for REST POST - handler receives empty params
	suite.True(body.IsOk(), "Should succeed with empty body")
}

func (suite *RESTEngineTestSuite) TestTokenInQueryParam() {
	suite.T().Log("Testing token in query parameter")

	token := suite.login()

	req := httptest.NewRequest(fiber.MethodPut, "/api/items?"+constants.QueryKeyAccessToken+"="+token, strings.NewReader(`{"id":"123","name":"Test"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "PUT should succeed with token in query param")
}

func (suite *RESTEngineTestSuite) TestComplexQueryParams() {
	suite.T().Log("Testing complex query parameters")

	resp := suite.makeRESTRequest(fiber.MethodGet, "/api/items?keyword=test%20item&page=1&size=10", "")

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "GET should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("test item", data["keyword"], "URL-encoded keyword should be decoded")
}

func (suite *RESTEngineTestSuite) TestPermissionDenied() {
	suite.T().Log("Testing permission denied (403)")

	token := suite.login()

	resp := suite.makeRESTRequestWithToken(fiber.MethodPost, "/api/items/admin", `{}`, token)

	suite.Equal(403, resp.StatusCode, "Should return 403 Forbidden")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with permission denied")
	suite.Equal(result.ErrCodeAccessDenied, body.Code, "Should return access denied error code")
}

func (suite *RESTEngineTestSuite) TestAdminWithPermission() {
	suite.T().Log("Testing admin action with permission")

	token := suite.loginAsAdmin()

	resp := suite.makeRESTRequestWithToken(fiber.MethodPost, "/api/items/admin", `{}`, token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Admin action should succeed with permission")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("admin", data["action"], "Action should be admin")
	suite.Equal(true, data["admin"], "Admin flag should be true")
}

func (suite *RESTEngineTestSuite) TestPutMissingRequiredID() {
	suite.T().Log("Testing PUT with missing required ID")

	token := suite.login()

	// PUT without id in body
	resp := suite.makeRESTRequestWithToken(fiber.MethodPut, "/api/items", `{"name":"Test"}`, token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with validation error")
	suite.Equal(result.ErrCodeBadRequest, body.Code, "Should return bad request error code")
}

func (suite *RESTEngineTestSuite) TestDeleteMissingRequiredID() {
	suite.T().Log("Testing DELETE with missing required ID")

	token := suite.login()

	// DELETE without id - send empty JSON body
	resp := suite.makeRESTRequestWithToken(fiber.MethodDelete, "/api/items", `{}`, token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with validation error")
	suite.Equal(result.ErrCodeBadRequest, body.Code, "Should return bad request error code")
}

func (suite *RESTEngineTestSuite) TestPatchMissingRequiredID() {
	suite.T().Log("Testing PATCH with missing required ID")

	token := suite.login()

	// PATCH without id in body
	resp := suite.makeRESTRequestWithToken(fiber.MethodPatch, "/api/items", `{"status":"active"}`, token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with validation error")
	suite.Equal(result.ErrCodeBadRequest, body.Code, "Should return bad request error code")
}

func (suite *RESTEngineTestSuite) TestHandlerPanic() {
	suite.T().Log("Testing handler panic returns 500")

	resp := suite.makeRESTRequest(fiber.MethodGet, "/api/items/panic", "")

	suite.Equal(500, resp.StatusCode, "Should return 500 Internal Server Error")
}

func (suite *RESTEngineTestSuite) TestContentTypeValidation() {
	suite.T().Log("Testing content type validation for REST")

	req := httptest.NewRequest(fiber.MethodPost, "/api/items", strings.NewReader(`{"name":"Test"}`))
	req.Header.Set(fiber.HeaderContentType, "text/plain")

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)

	suite.Equal(415, resp.StatusCode, "Should return 415 Unsupported Media Type")
}

func (suite *RESTEngineTestSuite) TestPutInvalidJSON() {
	suite.T().Log("Testing PUT with invalid JSON")

	token := suite.login()

	resp := suite.makeRESTRequestWithToken(fiber.MethodPut, "/api/items", "{invalid json}", token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with invalid JSON")
	suite.Equal(result.ErrCodeBadRequest, body.Code, "Should return bad request error code")
}

func (suite *RESTEngineTestSuite) TestPatchInvalidJSON() {
	suite.T().Log("Testing PATCH with invalid JSON")

	token := suite.login()

	resp := suite.makeRESTRequestWithToken(fiber.MethodPatch, "/api/items", "{invalid json}", token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with invalid JSON")
	suite.Equal(result.ErrCodeBadRequest, body.Code, "Should return bad request error code")
}

func (suite *RESTEngineTestSuite) TestDeleteInvalidJSON() {
	suite.T().Log("Testing DELETE with invalid JSON body")

	token := suite.login()

	resp := suite.makeRESTRequestWithToken(fiber.MethodDelete, "/api/items", "{invalid json}", token)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should fail with invalid JSON")
	suite.Equal(result.ErrCodeBadRequest, body.Code, "Should return bad request error code")
}

func (suite *RESTEngineTestSuite) TestUserLoaderCalledOnLogin() {
	suite.T().Log("Testing UserLoader is called during login")

	_ = suite.login()

	suite.userLoader.AssertCalled(suite.T(), "LoadByUsername", mock.Anything, "testuser")
}

func (suite *RESTEngineTestSuite) TestPermissionCheckerCalledOnPatch() {
	suite.T().Log("Testing PermissionChecker is called for PATCH")

	token := suite.login()

	suite.permissionChecker.Calls = nil

	_ = suite.makeRESTRequestWithToken(fiber.MethodPatch, "/api/items", `{"id":"123","status":"active"}`, token)

	suite.permissionChecker.AssertCalled(suite.T(), "HasPermission", mock.Anything, mock.Anything, "items:patch")
}

func (suite *RESTEngineTestSuite) TestPermissionCheckerCalledOnAdmin() {
	suite.T().Log("Testing PermissionChecker is called for admin action")

	token := suite.login()

	suite.permissionChecker.Calls = nil

	_ = suite.makeRESTRequestWithToken(fiber.MethodPost, "/api/items/admin", `{}`, token)

	suite.permissionChecker.AssertCalled(suite.T(), "HasPermission", mock.Anything, mock.Anything, "items:admin")
}

func TestRESTEngineSuite(t *testing.T) {
	suite.Run(t, new(RESTEngineTestSuite))
}
