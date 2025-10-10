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
type TestUserUpdateResource struct {
	api.Resource
	apis.UpdateAPI[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateResource() api.Resource {
	return &TestUserUpdateResource{
		Resource:  api.NewResource("test/user_update"),
		UpdateAPI: apis.NewUpdateAPI[TestUser, TestUserUpdateParams]().Public(),
	}
}

// Resource with PreUpdate hook.
type TestUserUpdateWithPreHookResource struct {
	api.Resource
	apis.UpdateAPI[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateWithPreHookResource() api.Resource {
	return &TestUserUpdateWithPreHookResource{
		Resource: api.NewResource("test/user_update_prehook"),
		UpdateAPI: apis.NewUpdateAPI[TestUser, TestUserUpdateParams]().
			Public().
			PreUpdate(func(oldModel, model *TestUser, params *TestUserUpdateParams, ctx fiber.Ctx, db orm.Db) error {
				// Add suffix to description
				if params.Description != "" {
					model.Description = params.Description + " [Updated]"
				}

				return nil
			}),
	}
}

// Resource with PostUpdate hook.
type TestUserUpdateWithPostHookResource struct {
	api.Resource
	apis.UpdateAPI[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateWithPostHookResource() api.Resource {
	return &TestUserUpdateWithPostHookResource{
		Resource: api.NewResource("test/user_update_posthook"),
		UpdateAPI: apis.NewUpdateAPI[TestUser, TestUserUpdateParams]().
			Public().
			PostUpdate(func(oldModel, model *TestUser, params *TestUserUpdateParams, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header
				ctx.Set("X-Updated-User-Name", model.Name)

				return nil
			}),
	}
}

// Test params for update (includes id).
type TestUserUpdateParams struct {
	api.In
	orm.ModelPK `json:",inline"`

	Name        string `json:"name"        validate:"required"`
	Email       string `json:"email"       validate:"required,email"`
	Description string `json:"description"`
	Age         int    `json:"age"         validate:"required,min=1,max=120"`
	Status      string `json:"status"      validate:"required,oneof=active inactive"`
}

// UpdateTestSuite is the test suite for Update API tests.
type UpdateTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *UpdateTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserUpdateResource,
		NewTestUserUpdateWithPreHookResource,
		NewTestUserUpdateWithPostHookResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *UpdateTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestUpdateBasic tests basic Update functionality.
func (suite *UpdateTestSuite) TestUpdateBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update",
			Action:   "update",
			Version:  "v1",
		},
		Params: map[string]any{
			"id":          "user001",
			"name":        "Updated Alice",
			"email":       "alice.updated@example.com",
			"description": "Updated description",
			"age":         26,
			"status":      "inactive",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	// UpdateAPI returns no data, just success status
}

// TestUpdateWithPreHook tests Update with PreUpdate hook.
func (suite *UpdateTestSuite) TestUpdateWithPreHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_prehook",
			Action:   "update",
			Version:  "v1",
		},
		Params: map[string]any{
			"id":          "user002",
			"name":        "Bob Updated",
			"email":       "bob@example.com",
			"description": "New description",
			"age":         31,
			"status":      "active",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	// UpdateAPI returns no data, PreUpdate hook was executed if update succeeded
}

// TestUpdateWithPostHook tests Update with PostUpdate hook.
func (suite *UpdateTestSuite) TestUpdateWithPostHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_posthook",
			Action:   "update",
			Version:  "v1",
		},
		Params: map[string]any{
			"id":     "user003",
			"name":   "Charlie Updated",
			"email":  "charlie@example.com",
			"age":    29,
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("Charlie Updated", resp.Header.Get("X-Updated-User-Name"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestUpdateNegativeCases tests negative scenarios.
func (suite *UpdateTestSuite) TestUpdateNegativeCases() {
	suite.Run("NonExistentUser", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update",
				Action:   "update",
				Version:  "v1",
			},
			Params: map[string]any{
				"id":     "nonexistent",
				"name":   "Test",
				"email":  "test@example.com",
				"age":    25,
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - user not found
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound))
	})

	suite.Run("MissingId", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update",
				Action:   "update",
				Version:  "v1",
			},
			Params: map[string]any{
				"name":   "Test",
				"email":  "test@example.com",
				"age":    25,
				"status": "active",
				// Missing "id"
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}))
	})

	suite.Run("InvalidEmail", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update",
				Action:   "update",
				Version:  "v1",
			},
			Params: map[string]any{
				"id":     "user004",
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
				Resource: "test/user_update",
				Action:   "update",
				Version:  "v1",
			},
			Params: map[string]any{
				"id":     "user005",
				"name":   "Test",
				"email":  "test@example.com",
				"age":    0, // Invalid: min=1
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("DuplicateEmail", func() {
		// Try to update user006's email to user005's email (which hasn't been modified in tests)
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update",
				Action:   "update",
				Version:  "v1",
			},
			Params: map[string]any{
				"id":          "user006",
				"name":        "Frank Miller",
				"email":       "eve@example.com", // user005's email - should cause unique constraint violation
				"description": "Sales Manager",
				"age":         35,
				"status":      "inactive",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail due to unique constraint violation
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordAlreadyExists))
	})
}

// TestPartialUpdate tests updating only some fields.
func (suite *UpdateTestSuite) TestPartialUpdate() {
	// First, get original user
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update",
			Action:   "update",
			Version:  "v1",
		},
		Params: map[string]any{
			"id":     "user007",
			"name":   "Grace Updated",
			"email":  "grace@example.com",
			"age":    30,
			"status": "active",
			// Not updating description
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	// UpdateAPI returns no data, update succeeded
}
