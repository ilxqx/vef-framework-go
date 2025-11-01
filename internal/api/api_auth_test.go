package api_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/guregu/null/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/internal/app"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
	"github.com/ilxqx/vef-framework-go/internal/database"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/security"
)

// TestAuthUser is a test user model for authentication and authorization tests.
type TestAuthUser struct {
	bun.BaseModel `bun:"table:test_auth_user,alias:tau"`
	orm.Model     `bun:"extend"`

	Username string `json:"username" bun:",unique,notnull"`
	Password string `json:"-"        bun:",notnull"`
	Role     string `json:"role"     bun:",notnull"`
	Name     string `json:"name"     bun:",notnull"`
	Email    string `json:"email"    bun:",unique,notnull"`
}

// TestAuthUserSearch is the search parameters for TestAuthUser.
type TestAuthUserSearch struct {
	api.P

	Username null.String `json:"username" search:"eq"`
	Role     null.String `json:"role"     search:"eq"`
}

// TestAuthUserParams is the params for creating/updating a TestAuthUser.
type TestAuthUserParams struct {
	api.P

	Id       string `json:"id"`
	Username string `json:"username"     validate:"required"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role"         validate:"required"`
	Name     string `json:"name"         validate:"required"`
	Email    string `json:"email"        validate:"required,email"`
}

// TestAuthUserResource is a test resource for authentication and authorization tests.
type TestAuthUserResource struct {
	api.Resource
	apis.FindAllApi[TestAuthUser, TestAuthUserSearch]
	apis.CreateApi[TestAuthUser, TestAuthUserParams]
	apis.UpdateApi[TestAuthUser, TestAuthUserParams]
	apis.DeleteApi[TestAuthUser]
}

// NewTestAuthUserResource creates a new test auth user resource.
func NewTestAuthUserResource() api.Resource {
	return &TestAuthUserResource{
		Resource:   api.NewResource("test/auth_user"),
		FindAllApi: apis.NewFindAllApi[TestAuthUser, TestAuthUserSearch]().PermToken("user.query"),
		CreateApi:  apis.NewCreateApi[TestAuthUser, TestAuthUserParams]().PermToken("user.create"),
		UpdateApi:  apis.NewUpdateApi[TestAuthUser, TestAuthUserParams]().PermToken("user.update"),
		DeleteApi:  apis.NewDeleteApi[TestAuthUser]().PermToken("user.delete"),
	}
}

// TestAuthUserLoader implements security.UserLoader for testing.
type TestAuthUserLoader struct {
	db orm.Db
}

// LoadByUsername loads a user by username.
func (l *TestAuthUserLoader) LoadByUsername(ctx context.Context, username string) (*security.Principal, string, error) {
	var user TestAuthUser

	err := l.db.NewSelect().
		Model(&user).
		Where(func(cb orm.ConditionBuilder) {
			cb.Equals("username", username)
		}).
		Scan(ctx)
	if err != nil {
		return nil, "", err
	}

	principal := &security.Principal{
		Type:  security.PrincipalTypeUser,
		Id:    user.Id,
		Name:  user.Name,
		Roles: []string{user.Role},
	}

	return principal, user.Password, nil
}

// LoadById loads a user by id.
func (l *TestAuthUserLoader) LoadById(ctx context.Context, id string) (*security.Principal, error) {
	var user TestAuthUser

	err := l.db.NewSelect().
		Model(&user).
		Where(func(cb orm.ConditionBuilder) {
			cb.Equals("id", id)
		}).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	principal := &security.Principal{
		Type:  security.PrincipalTypeUser,
		Id:    user.Id,
		Name:  user.Name,
		Roles: []string{user.Role},
	}

	return principal, nil
}

// TestAuthRolePermissionsLoader implements security.RolePermissionsLoader for testing.
type TestAuthRolePermissionsLoader struct{}

// LoadPermissions loads permissions for a given role.
func (l *TestAuthRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]security.DataScope, error) {
	permissions := make(map[string]security.DataScope)
	allScope := security.NewAllDataScope()
	selfScope := security.NewSelfDataScope(constants.ColumnCreatedBy)

	switch role {
	case "admin":
		// Admin has all permissions with unrestricted data access
		permissions["user.query"] = allScope
		permissions["user.create"] = allScope
		permissions["user.update"] = allScope
		permissions["user.delete"] = allScope

	case "editor":
		// Editor can query (only their own data), create, and update, but not delete
		permissions["user.query"] = selfScope
		permissions["user.create"] = allScope
		permissions["user.update"] = selfScope
	case "viewer":
		// Viewer can only query
		permissions["user.query"] = allScope
	}

	return permissions, nil
}

// ApiAuthTestSuite is the test suite for API authentication and authorization integration tests.
type ApiAuthTestSuite struct {
	suite.Suite

	ctx         context.Context
	db          orm.Db
	app         *app.App
	stop        func()
	adminUser   TestAuthUser
	editorUser  TestAuthUser
	viewerUser  TestAuthUser
	adminToken  string
	editorToken string
	viewerToken string
}

// SetupSuite runs once before all tests in the suite.
func (suite *ApiAuthTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Create SQLite in-memory database
	dsConfig := &config.DatasourceConfig{
		Type: constants.DbSQLite,
	}

	db, err := database.New(dsConfig)
	require.NoError(suite.T(), err, "Should create database connection successfully")
	suite.db = orm.New(db)

	// Register models
	bunDb := suite.db.(orm.Unwrapper[bun.IDB]).Unwrap().(*bun.DB)
	bunDb.RegisterModel((*TestAuthUser)(nil))

	// Create tables
	_, err = bunDb.NewCreateTable().
		Model((*TestAuthUser)(nil)).
		IfNotExists().
		Exec(suite.ctx)
	require.NoError(suite.T(), err, "Should create test_auth_user table successfully")

	// Hash passwords for test users
	adminPassword, err := security.HashPassword("admin123")
	require.NoError(suite.T(), err, "Should hash admin password successfully")
	editorPassword, err := security.HashPassword("editor123")
	require.NoError(suite.T(), err, "Should hash editor password successfully")
	viewerPassword, err := security.HashPassword("viewer123")
	require.NoError(suite.T(), err, "Should hash viewer password successfully")

	// Create test users
	suite.adminUser = TestAuthUser{
		Username: "admin",
		Password: adminPassword,
		Role:     "admin",
		Name:     "Admin User",
		Email:    "admin@test.com",
	}
	suite.adminUser.Id = "admin_user_id"

	suite.editorUser = TestAuthUser{
		Username: "editor",
		Password: editorPassword,
		Role:     "editor",
		Name:     "Editor User",
		Email:    "editor@test.com",
	}
	suite.editorUser.Id = "editor_user_id"

	suite.viewerUser = TestAuthUser{
		Username: "viewer",
		Password: viewerPassword,
		Role:     "viewer",
		Name:     "Viewer User",
		Email:    "viewer@test.com",
	}
	suite.viewerUser.Id = "viewer_user_id"

	// Insert test users
	_, err = suite.db.NewInsert().
		Model(&[]TestAuthUser{suite.adminUser, suite.editorUser, suite.viewerUser}).
		Exec(suite.ctx)
	require.NoError(suite.T(), err, "Should insert test users successfully")

	// Create user loader and role permissions loader
	userLoader := &TestAuthUserLoader{db: suite.db}
	rolePermLoader := &TestAuthRolePermissionsLoader{}

	// Setup app with user resource and custom security providers
	opts := []fx.Option{
		vef.ProvideApiResource(NewTestAuthUserResource),
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
		fx.Decorate(func() bun.IDB {
			return bunDb
		}),
		fx.Supply(
			fx.Annotate(
				userLoader,
				fx.As(new(security.UserLoader)),
			),
			fx.Annotate(
				rolePermLoader,
				fx.As(new(security.RolePermissionsLoader)),
			),
		),
	}

	suite.app, suite.stop = appTest.NewTestApp(suite.T(), opts...)

	// Login and get tokens for each user
	suite.adminToken = suite.loginAndGetToken("admin", "admin123")
	suite.editorToken = suite.loginAndGetToken("editor", "editor123")
	suite.viewerToken = suite.loginAndGetToken("viewer", "viewer123")
}

// TearDownSuite runs once after all tests in the suite.
func (suite *ApiAuthTestSuite) TearDownSuite() {
	if suite.stop != nil {
		suite.stop()
	}
}

// loginAndGetToken logs in a user and returns the access token.
func (suite *ApiAuthTestSuite) loginAndGetToken(username, password string) string {
	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "security/auth",
			Action:   "login",
		},
		Params: map[string]any{
			"type":        "password",
			"principal":   username,
			"credentials": password,
		},
	}

	resp := suite.makeApiRequest(body, "")
	res := suite.readBody(resp)

	require.Equal(suite.T(), result.OkCode, res.Code, "Login should succeed for user %s", username)
	data := suite.readDataAsMap(res.Data)
	accessToken, ok := data["accessToken"].(string)
	require.True(suite.T(), ok, "Response should contain accessToken")
	require.NotEmpty(suite.T(), accessToken, "Access token should not be empty")

	suite.T().Logf("Successfully logged in user '%s' and obtained access token", username)

	return accessToken
}

// makeApiRequest makes an API request with optional authorization token.
func (suite *ApiAuthTestSuite) makeApiRequest(body api.Request, token string) *http.Response {
	jsonBody, err := encoding.ToJson(body)
	suite.Require().NoError(err, "Should encode request body to JSON successfully")

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	if token != "" {
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	}

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err, "Should make API request successfully")

	return resp
}

// readBody reads and parses the response body.
func (suite *ApiAuthTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			suite.T().Errorf("failed to close response body: %v", closeErr)
		}
	}()

	suite.Require().NoError(err, "Should read response body successfully")
	res, err := encoding.FromJson[result.Result](string(body))
	suite.Require().NoError(err, "Should parse response body as JSON successfully")

	return *res
}

// readDataAsMap reads data as a map.
func (suite *ApiAuthTestSuite) readDataAsMap(data any) map[string]any {
	m, ok := data.(map[string]any)
	suite.Require().True(ok, "Expected data to be a map")

	return m
}

// readDataAsSlice reads data as a slice.
func (suite *ApiAuthTestSuite) readDataAsSlice(data any) []any {
	slice, ok := data.([]any)
	suite.Require().True(ok, "Expected data to be a slice")

	return slice
}

// TestAdminCanQueryUsers verifies that admin can query all users (unrestricted data access).
func (suite *ApiAuthTestSuite) TestAdminCanQueryUsers() {
	suite.T().Log("Testing admin can query all users with unrestricted data access")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "find_all",
		},
	}

	resp := suite.makeApiRequest(body, suite.adminToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Admin should be able to query all users")
	users := suite.readDataAsSlice(res.Data)
	// At least the initial 3 users (admin, editor, viewer) should be visible
	assert.GreaterOrEqual(suite.T(), len(users), 3, "Admin should see at least 3 users")
}

// TestEditorCanQueryUsers verifies that editor can only query users they created (SelfDataScope).
func (suite *ApiAuthTestSuite) TestEditorCanQueryUsers() {
	suite.T().Log("Testing editor can only query users they created (SelfDataScope)")

	// Get current count of editor's users
	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "find_all",
		},
	}

	resp := suite.makeApiRequest(body, suite.editorToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Editor should be able to query users")
	initialUsers := suite.readDataAsSlice(res.Data)
	initialCount := len(initialUsers)

	// Verify all existing users were created by editor
	for _, user := range initialUsers {
		userMap := user.(map[string]any)
		createdBy, ok := userMap["createdBy"].(string)
		assert.True(suite.T(), ok, "createdBy should be a string")
		assert.Equal(suite.T(), suite.editorUser.Id, createdBy, "Editor should only see users they created")
		suite.T().Logf("Editor query result - Username: %v, CreatedBy: %s", userMap["username"], createdBy)
	}

	// Now editor creates a new user
	createBody := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "create",
		},
		Params: map[string]any{
			"username": "editor_scope_test",
			"password": "hash123",
			"role":     "viewer",
			"name":     "Editor Scope Test",
			"email":    "editor.scope.test@test.com",
		},
	}

	createResp := suite.makeApiRequest(createBody, suite.editorToken)
	createRes := suite.readBody(createResp)
	assert.Equal(suite.T(), result.OkCode, createRes.Code, "Editor should be able to create a user")

	// Query again should return initialCount + 1 records
	resp = suite.makeApiRequest(body, suite.editorToken)
	res = suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Editor should be able to query users after creating one")
	users := suite.readDataAsSlice(res.Data)
	assert.Equal(suite.T(), initialCount+1, len(users), "Editor should see one more user after creating")

	// Verify all returned users have correct createdBy
	for _, user := range users {
		userMap := user.(map[string]any)
		userCreatedBy, ok := userMap["createdBy"].(string)
		assert.True(suite.T(), ok, "User's createdBy should be a string")
		assert.Equal(suite.T(), suite.editorUser.Id, userCreatedBy, "All users should be created by editor")
		suite.T().Logf("Editor query after create - Username: %v, CreatedBy: %s", userMap["username"], userCreatedBy)
	}
}

// TestViewerCanQueryUsers verifies that viewer can query all users (unrestricted data access).
func (suite *ApiAuthTestSuite) TestViewerCanQueryUsers() {
	suite.T().Log("Testing viewer can query all users with unrestricted data access")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "find_all",
		},
	}

	resp := suite.makeApiRequest(body, suite.viewerToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Viewer should be able to query all users")
	users := suite.readDataAsSlice(res.Data)
	// At least the initial 3 users (admin, editor, viewer) should be visible
	assert.GreaterOrEqual(suite.T(), len(users), 3, "Viewer should see at least 3 users")
}

// TestAdminCanCreateUser verifies that admin can create users.
func (suite *ApiAuthTestSuite) TestAdminCanCreateUser() {
	suite.T().Log("Testing admin can create users")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "create",
		},
		Params: map[string]any{
			"username": "newuser",
			"password": "hash123",
			"role":     "viewer",
			"name":     "New User",
			"email":    "newuser@test.com",
		},
	}

	resp := suite.makeApiRequest(body, suite.adminToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Admin should be able to create a user")
}

// TestEditorCanCreateUser verifies that editor can create users.
func (suite *ApiAuthTestSuite) TestEditorCanCreateUser() {
	suite.T().Log("Testing editor can create users")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "create",
		},
		Params: map[string]any{
			"username": "newuser2",
			"password": "hash123",
			"role":     "viewer",
			"name":     "New User 2",
			"email":    "newuser2@test.com",
		},
	}

	resp := suite.makeApiRequest(body, suite.editorToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Editor should be able to create a user")
}

// TestViewerCannotCreateUser verifies that viewer cannot create users.
func (suite *ApiAuthTestSuite) TestViewerCannotCreateUser() {
	suite.T().Log("Testing viewer cannot create users (permission denied)")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "create",
		},
		Params: map[string]any{
			"username": "shouldnotcreate",
			"password": "hash123",
			"role":     "viewer",
			"name":     "Should Not Create",
			"email":    "shouldnotcreate@test.com",
		},
	}

	resp := suite.makeApiRequest(body, suite.viewerToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.ErrCodeAccessDenied, res.Code, "Viewer should not be able to create users")
}

// TestAdminCanUpdateUser verifies that admin can update users.
func (suite *ApiAuthTestSuite) TestAdminCanUpdateUser() {
	suite.T().Log("Testing admin can update users")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "update",
		},
		Params: map[string]any{
			"id":       suite.viewerUser.Id,
			"username": suite.viewerUser.Username,
			"password": suite.viewerUser.Password,
			"role":     suite.viewerUser.Role,
			"name":     "Updated Viewer Name",
			"email":    suite.viewerUser.Email,
		},
	}

	resp := suite.makeApiRequest(body, suite.adminToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Admin should be able to update users")
}

// TestEditorCanUpdateUser verifies that editor can update users they created (SelfDataScope).
func (suite *ApiAuthTestSuite) TestEditorCanUpdateUser() {
	suite.T().Log("Testing editor can update users they created (SelfDataScope)")

	// First, editor creates a user
	createBody := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "create",
		},
		Params: map[string]any{
			"username": "editor_owned_user",
			"password": "hash123",
			"role":     "viewer",
			"name":     "Editor Owned User",
			"email":    "editor.owned@test.com",
		},
	}

	createResp := suite.makeApiRequest(createBody, suite.editorToken)
	createRes := suite.readBody(createResp)
	assert.Equal(suite.T(), result.OkCode, createRes.Code, "Editor should be able to create a user to update")

	createdUser := suite.readDataAsMap(createRes.Data)
	userId := createdUser["id"].(string)
	suite.T().Logf("Editor created user with ID: %s", userId)

	// Now editor updates the user they just created
	updateBody := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "update",
		},
		Params: map[string]any{
			"id":       userId,
			"username": "editor_owned_user",
			"password": "hash123",
			"role":     "viewer",
			"name":     "Editor Owned User Updated",
			"email":    "editor.owned.updated@test.com",
		},
	}

	updateResp := suite.makeApiRequest(updateBody, suite.editorToken)
	updateRes := suite.readBody(updateResp)

	// Update should succeed
	assert.Equal(suite.T(), result.OkCode, updateRes.Code, "Editor should be able to update users they created")
}

// TestEditorCannotUpdateOthersUser verifies that editor cannot update users created by others (SelfDataScope).
func (suite *ApiAuthTestSuite) TestEditorCannotUpdateOthersUser() {
	suite.T().Log("Testing editor cannot update users created by others (SelfDataScope)")
	suite.T().Logf("Attempting to update viewerUser (ID: %s) which was not created by editor", suite.viewerUser.Id)

	// Try to update viewerUser (which was created during setup, not by editor)
	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "update",
		},
		Params: map[string]any{
			"id":       suite.viewerUser.Id,
			"username": suite.viewerUser.Username,
			"password": suite.viewerUser.Password,
			"role":     suite.viewerUser.Role,
			"name":     "Should Not Update",
			"email":    suite.viewerUser.Email,
		},
	}

	resp := suite.makeApiRequest(body, suite.editorToken)
	res := suite.readBody(resp)

	// Should fail because editor can only update users they created
	// The record won't be found due to SelfDataScope filtering
	assert.Equal(suite.T(), result.ErrCodeRecordNotFound, res.Code, "Editor should not be able to update users created by others")
}

// TestViewerCannotUpdateUser verifies that viewer cannot update users.
func (suite *ApiAuthTestSuite) TestViewerCannotUpdateUser() {
	suite.T().Log("Testing viewer cannot update users (permission denied)")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "update",
		},
		Params: map[string]any{
			"id":   suite.viewerUser.Id,
			"name": "Should Not Update",
		},
	}

	resp := suite.makeApiRequest(body, suite.viewerToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.ErrCodeAccessDenied, res.Code, "Viewer should not be able to update users")
}

// TestAdminCanDeleteUser verifies that admin can delete users.
func (suite *ApiAuthTestSuite) TestAdminCanDeleteUser() {
	suite.T().Log("Testing admin can delete users")

	// First create a user to delete
	newUser := TestAuthUser{
		Username: "todelete",
		Password: "hash123",
		Role:     "viewer",
		Name:     "To Delete",
		Email:    "todelete@test.com",
	}
	newUser.Id = "todelete_id"

	bunDb := suite.db.(orm.Unwrapper[bun.IDB]).Unwrap()
	_, err := bunDb.NewInsert().
		Model(&newUser).
		Exec(suite.ctx)
	require.NoError(suite.T(), err, "Should insert user to delete successfully")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "delete",
		},
		Params: map[string]any{
			"id": newUser.Id,
		},
	}

	resp := suite.makeApiRequest(body, suite.adminToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.OkCode, res.Code, "Admin should be able to delete users")
}

// TestEditorCannotDeleteUser verifies that editor cannot delete users.
func (suite *ApiAuthTestSuite) TestEditorCannotDeleteUser() {
	suite.T().Log("Testing editor cannot delete users (permission denied)")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "delete",
		},
		Params: map[string]any{
			"id": suite.viewerUser.Id,
		},
	}

	resp := suite.makeApiRequest(body, suite.editorToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.ErrCodeAccessDenied, res.Code, "Editor should not be able to delete users")
}

// TestViewerCannotDeleteUser verifies that viewer cannot delete users.
func (suite *ApiAuthTestSuite) TestViewerCannotDeleteUser() {
	suite.T().Log("Testing viewer cannot delete users (permission denied)")

	body := api.Request{
		Identifier: api.Identifier{
			Version:  api.VersionV1,
			Resource: "test/auth_user",
			Action:   "delete",
		},
		Params: map[string]any{
			"id": suite.viewerUser.Id,
		},
	}

	resp := suite.makeApiRequest(body, suite.viewerToken)
	res := suite.readBody(resp)

	assert.Equal(suite.T(), result.ErrCodeAccessDenied, res.Code, "Viewer should not be able to delete users")
}

// TestApiAuthSuite runs the authentication and authorization test suite.
func TestApiAuthSuite(t *testing.T) {
	suite.Run(t, new(ApiAuthTestSuite))
}
