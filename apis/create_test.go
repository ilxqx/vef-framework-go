package apis_test

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// Test Resources.
type TestUserCreateResource struct {
	api.Resource
	apis.CreateApi[TestUser, TestUserCreateParams]
}

func NewTestUserCreateResource() api.Resource {
	return &TestUserCreateResource{
		Resource:  api.NewResource("test/user_create"),
		CreateApi: apis.NewCreateApi[TestUser, TestUserCreateParams]().Public(),
	}
}

// Resource with PreCreate hook.
type TestUserCreateWithPreHookResource struct {
	api.Resource
	apis.CreateApi[TestUser, TestUserCreateParams]
}

func NewTestUserCreateWithPreHookResource() api.Resource {
	return &TestUserCreateWithPreHookResource{
		Resource: api.NewResource("test/user_create_prehook"),
		CreateApi: apis.NewCreateApi[TestUser, TestUserCreateParams]().
			Public().
			WithPreCreate(func(model *TestUser, params *TestUserCreateParams, query orm.InsertQuery, ctx fiber.Ctx, tx orm.Db) error {
				// Add prefix to name
				model.Name = "Mr. " + model.Name

				return nil
			}),
	}
}

// Resource with PostCreate hook.
type TestUserCreateWithPostHookResource struct {
	api.Resource
	apis.CreateApi[TestUser, TestUserCreateParams]
}

func NewTestUserCreateWithPostHookResource() api.Resource {
	return &TestUserCreateWithPostHookResource{
		Resource: api.NewResource("test/user_create_posthook"),
		CreateApi: apis.NewCreateApi[TestUser, TestUserCreateParams]().
			Public().
			WithPostCreate(func(model *TestUser, params *TestUserCreateParams, ctx fiber.Ctx, db orm.Db) error {
				// Log or perform additional operations
				ctx.Set("X-Created-User-Id", model.Id)

				return nil
			}),
	}
}

// Test params for create.
type TestUserCreateParams struct {
	api.P

	Name        string `json:"name"        validate:"required"`
	Email       string `json:"email"       validate:"required,email"`
	Description string `json:"description"`
	Age         int    `json:"age"         validate:"required,min=1,max=120"`
	Status      string `json:"status"      validate:"required,oneof=active inactive"`
}

// CreateTestSuite tests the Create API functionality
// including basic create, PreCreate/PostCreate hooks, and negative cases.
type CreateTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *CreateTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserCreateResource,
		NewTestUserCreateWithPreHookResource,
		NewTestUserCreateWithPostHookResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *CreateTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestCreateBasic tests basic Create functionality.
func (suite *CreateTestSuite) TestCreateBasic() {
	suite.T().Logf("Testing Create API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_create",
			Action:   "create",
			Version:  "v1",
		},
		Params: map[string]any{
			"name":        "New User",
			"email":       "newuser@example.com",
			"description": "Test user",
			"age":         25,
			"status":      "active",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	suite.NotNil(body.Data, "Should return data")

	pk := suite.readDataAsMap(body.Data)
	suite.NotEmpty(pk["id"], "Should return created user id")

	suite.T().Logf("Created user with id: %v", pk["id"])
}

// TestCreateWithPreHook tests Create with PreCreate hook.
func (suite *CreateTestSuite) TestCreateWithPreHook() {
	suite.T().Logf("Testing Create API with PreCreate hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_create_prehook",
			Action:   "create",
			Version:  "v1",
		},
		Params: map[string]any{
			"name":   "John",
			"email":  "john@example.com",
			"age":    30,
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	pk := suite.readDataAsMap(body.Data)
	suite.NotEmpty(pk["id"], "Should return created user id")

	suite.T().Logf("Created user with PreCreate hook, id: %v", pk["id"])
}

// TestCreateWithPostHook tests Create with PostCreate hook.
func (suite *CreateTestSuite) TestCreateWithPostHook() {
	suite.T().Logf("Testing Create API with PostCreate hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_create_posthook",
			Action:   "create",
			Version:  "v1",
		},
		Params: map[string]any{
			"name":   "Jane",
			"email":  "jane@example.com",
			"age":    28,
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.NotEmpty(resp.Header.Get("X-Created-User-Id"), "Should set X-Created-User-Id header via PostCreate hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	pk := suite.readDataAsMap(body.Data)
	suite.NotEmpty(pk["id"], "Should return created user id")

	suite.T().Logf("Created user with PostCreate hook, id: %v, header: %s", pk["id"], resp.Header.Get("X-Created-User-Id"))
}

// TestCreateNegativeCases tests negative scenarios.
func (suite *CreateTestSuite) TestCreateNegativeCases() {
	suite.T().Logf("Testing Create API negative cases for %s", suite.dbType)

	suite.Run("MissingRequiredField", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"email":  "test@example.com",
				"age":    25,
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when required field 'name' is missing")

		suite.T().Logf("Validation failed as expected for missing required field")
	})

	suite.Run("InvalidEmail", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"name":   "Test",
				"email":  "invalid-email",
				"age":    25,
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when email format is invalid")

		suite.T().Logf("Validation failed as expected for invalid email format")
	})

	suite.Run("InvalidAge", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"name":   "Test",
				"email":  "test@example.com",
				"age":    150,
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when age is greater than 120")

		suite.T().Logf("Validation failed as expected for invalid age")
	})

	suite.Run("InvalidStatus", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"name":   "Test",
				"email":  "test@example.com",
				"age":    25,
				"status": "invalid_status",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when status is not 'active' or 'inactive'")

		suite.T().Logf("Validation failed as expected for invalid status")
	})

	suite.Run("DuplicateEmail", func() {
		suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"name":   "First User",
				"email":  "duplicate@example.com",
				"age":    25,
				"status": "active",
			},
		})

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"name":   "Second User",
				"email":  "duplicate@example.com",
				"age":    30,
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail due to duplicate email unique constraint")

		suite.T().Logf("Validation failed as expected for duplicate email")
	})
}
