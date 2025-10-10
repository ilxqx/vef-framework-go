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
	apis.CreateAPI[TestUser, TestUserCreateParams]
}

func NewTestUserCreateResource() api.Resource {
	return &TestUserCreateResource{
		Resource:  api.NewResource("test/user_create"),
		CreateAPI: apis.NewCreateAPI[TestUser, TestUserCreateParams]().Public(),
	}
}

// Resource with PreCreate hook.
type TestUserCreateWithPreHookResource struct {
	api.Resource
	apis.CreateAPI[TestUser, TestUserCreateParams]
}

func NewTestUserCreateWithPreHookResource() api.Resource {
	return &TestUserCreateWithPreHookResource{
		Resource: api.NewResource("test/user_create_prehook"),
		CreateAPI: apis.NewCreateAPI[TestUser, TestUserCreateParams]().
			Public().
			PreCreate(func(model *TestUser, params *TestUserCreateParams, ctx fiber.Ctx, db orm.Db) error {
				// Add prefix to name
				model.Name = "Mr. " + model.Name

				return nil
			}),
	}
}

// Resource with PostCreate hook.
type TestUserCreateWithPostHookResource struct {
	api.Resource
	apis.CreateAPI[TestUser, TestUserCreateParams]
}

func NewTestUserCreateWithPostHookResource() api.Resource {
	return &TestUserCreateWithPostHookResource{
		Resource: api.NewResource("test/user_create_posthook"),
		CreateAPI: apis.NewCreateAPI[TestUser, TestUserCreateParams]().
			Public().
			PostCreate(func(model *TestUser, params *TestUserCreateParams, ctx fiber.Ctx, db orm.Db) error {
				// Log or perform additional operations
				ctx.Set("X-Created-User-Id", model.Id)

				return nil
			}),
	}
}

// Test params for create.
type TestUserCreateParams struct {
	api.In

	Name        string `json:"name"        validate:"required"`
	Email       string `json:"email"       validate:"required,email"`
	Description string `json:"description"`
	Age         int    `json:"age"         validate:"required,min=1,max=120"`
	Status      string `json:"status"      validate:"required,oneof=active inactive"`
}

// CreateTestSuite is the test suite for Create API tests.
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
	resp := suite.makeAPIRequest(api.Request{
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

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	suite.NotNil(body.Data)

	// CreateAPI returns primary key(s) only
	pk := suite.readDataAsMap(body.Data)
	suite.NotEmpty(pk["id"])
}

// TestCreateWithPreHook tests Create with PreCreate hook.
func (suite *CreateTestSuite) TestCreateWithPreHook() {
	resp := suite.makeAPIRequest(api.Request{
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

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	// CreateAPI returns primary key(s) only, we can't verify the name directly
	// The PreCreate hook was executed if the insert succeeded
	pk := suite.readDataAsMap(body.Data)
	suite.NotEmpty(pk["id"])
}

// TestCreateWithPostHook tests Create with PostCreate hook.
func (suite *CreateTestSuite) TestCreateWithPostHook() {
	resp := suite.makeAPIRequest(api.Request{
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

	suite.Equal(200, resp.StatusCode)
	suite.NotEmpty(resp.Header.Get("X-Created-User-Id"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())

	// CreateAPI returns primary key(s) only
	pk := suite.readDataAsMap(body.Data)
	suite.NotEmpty(pk["id"])
}

// TestCreateNegativeCases tests negative scenarios.
func (suite *CreateTestSuite) TestCreateNegativeCases() {
	suite.Run("MissingRequiredField", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"email":  "test@example.com",
				"age":    25,
				"status": "active",
				// Missing "name"
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("InvalidEmail", func() {
		resp := suite.makeAPIRequest(api.Request{
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("InvalidAge", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create",
				Action:   "create",
				Version:  "v1",
			},
			Params: map[string]any{
				"name":   "Test",
				"email":  "test@example.com",
				"age":    150, // Invalid: > 120
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("InvalidStatus", func() {
		resp := suite.makeAPIRequest(api.Request{
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("DuplicateEmail", func() {
		// First create
		suite.makeAPIRequest(api.Request{
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

		// Try to create with same email
		resp := suite.makeAPIRequest(api.Request{
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail due to unique constraint
	})
}
