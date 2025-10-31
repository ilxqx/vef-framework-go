package apis_test

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type TestUserUpdateResource struct {
	api.Resource
	apis.UpdateApi[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateResource() api.Resource {
	return &TestUserUpdateResource{
		Resource:  api.NewResource("test/user_update"),
		UpdateApi: apis.NewUpdateApi[TestUser, TestUserUpdateParams]().Public(),
	}
}

type TestUserUpdateWithPreHookResource struct {
	api.Resource
	apis.UpdateApi[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateWithPreHookResource() api.Resource {
	return &TestUserUpdateWithPreHookResource{
		Resource: api.NewResource("test/user_update_prehook"),
		UpdateApi: apis.NewUpdateApi[TestUser, TestUserUpdateParams]().
			Public().
			WithPreUpdate(func(oldModel, model *TestUser, params *TestUserUpdateParams, ctx fiber.Ctx, db orm.Db) error {
				if params.Description != "" {
					model.Description = params.Description + " [Updated]"
				}

				return nil
			}),
	}
}

type TestUserUpdateWithPostHookResource struct {
	api.Resource
	apis.UpdateApi[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateWithPostHookResource() api.Resource {
	return &TestUserUpdateWithPostHookResource{
		Resource: api.NewResource("test/user_update_posthook"),
		UpdateApi: apis.NewUpdateApi[TestUser, TestUserUpdateParams]().
			Public().
			WithPostUpdate(func(oldModel, model *TestUser, params *TestUserUpdateParams, ctx fiber.Ctx, tx orm.Db) error {
				ctx.Set("X-Updated-User-Name", model.Name)

				return nil
			}),
	}
}

type TestUserUpdateParams struct {
	api.P

	Id          string `json:"id"`
	Name        string `json:"name"        validate:"required"`
	Email       string `json:"email"       validate:"required,email"`
	Description string `json:"description"`
	Age         int    `json:"age"         validate:"required,min=1,max=120"`
	Status      string `json:"status"      validate:"required,oneof=active inactive"`
}

// UpdateTestSuite tests the Update API functionality.
type UpdateTestSuite struct {
	BaseSuite
}

func (suite *UpdateTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserUpdateResource,
		NewTestUserUpdateWithPreHookResource,
		NewTestUserUpdateWithPostHookResource,
	)
}

func (suite *UpdateTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestUpdateBasic tests basic Update functionality.
func (suite *UpdateTestSuite) TestUpdateBasic() {
	suite.T().Logf("Testing Update API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
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

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Updated user001 successfully")
}

// TestUpdateWithPreHook tests Update with PreUpdate hook.
func (suite *UpdateTestSuite) TestUpdateWithPreHook() {
	suite.T().Logf("Testing Update API with PreUpdate hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
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

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Updated user002 with PreUpdate hook successfully")
}

// TestUpdateWithPostHook tests Update with PostUpdate hook.
func (suite *UpdateTestSuite) TestUpdateWithPostHook() {
	suite.T().Logf("Testing Update API with PostUpdate hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
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

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.Equal("Charlie Updated", resp.Header.Get("X-Updated-User-Name"), "Should set X-Updated-User-Name header via PostUpdate hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Updated user003 with PostUpdate hook, header: %s", resp.Header.Get("X-Updated-User-Name"))
}

// TestUpdateNegativeCases tests negative scenarios.
func (suite *UpdateTestSuite) TestUpdateNegativeCases() {
	suite.T().Logf("Testing Update API negative cases for %s", suite.dbType)

	suite.Run("NonExistentUser", func() {
		resp := suite.makeApiRequest(api.Request{
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when user does not exist")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Validation failed as expected for non-existent user")
	})

	suite.Run("MissingId", func() {
		resp := suite.makeApiRequest(api.Request{
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
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when required id is missing")
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}), "Should return primary key required message")

		suite.T().Logf("Validation failed as expected for missing id")
	})

	suite.Run("InvalidEmail", func() {
		resp := suite.makeApiRequest(api.Request{
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when email format is invalid")

		suite.T().Logf("Validation failed as expected for invalid email format")
	})

	suite.Run("InvalidAge", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update",
				Action:   "update",
				Version:  "v1",
			},
			Params: map[string]any{
				"id":     "user005",
				"name":   "Test",
				"email":  "test@example.com",
				"age":    0,
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when age is less than 1")

		suite.T().Logf("Validation failed as expected for invalid age")
	})

	suite.Run("DuplicateEmail", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update",
				Action:   "update",
				Version:  "v1",
			},
			Params: map[string]any{
				"id":          "user006",
				"name":        "Frank Miller",
				"email":       "eve@example.com",
				"description": "Sales Manager",
				"age":         35,
				"status":      "inactive",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail due to duplicate email unique constraint")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordAlreadyExists), "Should return record already exists message")

		suite.T().Logf("Validation failed as expected for duplicate email")
	})
}

// TestPartialUpdate tests partial field updates.
func (suite *UpdateTestSuite) TestPartialUpdate() {
	suite.T().Logf("Testing Update API partial update for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
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
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Partially updated user007 successfully")
}
