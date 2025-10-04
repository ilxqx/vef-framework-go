package test

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// Test Resources
type TestUserDeleteResource struct {
	api.Resource
	apis.DeleteAPI[TestUser]
}

func NewTestUserDeleteResource() api.Resource {
	return &TestUserDeleteResource{
		Resource:  api.NewResource("test/user_delete"),
		DeleteAPI: apis.NewDeleteAPI[TestUser]().Public(),
	}
}

// Resource with PreDelete hook
type TestUserDeleteWithPreHookResource struct {
	api.Resource
	apis.DeleteAPI[TestUser]
}

func NewTestUserDeleteWithPreHookResource() api.Resource {
	return &TestUserDeleteWithPreHookResource{
		Resource: api.NewResource("test/user_delete_prehook"),
		DeleteAPI: apis.NewDeleteAPI[TestUser]().
			Public().
			PreDelete(func(model *TestUser, ctx fiber.Ctx, db orm.Db) error {
				// Log or check conditions before delete
				if model.Status == "active" {
					ctx.Set("X-Delete-Warning", "Deleting active user")
				}
				return nil
			}),
	}
}

// Resource with PostDelete hook
type TestUserDeleteWithPostHookResource struct {
	api.Resource
	apis.DeleteAPI[TestUser]
}

func NewTestUserDeleteWithPostHookResource() api.Resource {
	return &TestUserDeleteWithPostHookResource{
		Resource: api.NewResource("test/user_delete_posthook"),
		DeleteAPI: apis.NewDeleteAPI[TestUser]().
			Public().
			PostDelete(func(model *TestUser, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header after deletion
				ctx.Set("X-Deleted-User-Id", model.Id)
				return nil
			}),
	}
}

// DeleteTestSuite is the test suite for Delete API tests
type DeleteTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite
func (suite *DeleteTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserDeleteResource,
		NewTestUserDeleteWithPreHookResource,
		NewTestUserDeleteWithPostHookResource,
	)
}

// TearDownSuite runs once after all tests in the suite
func (suite *DeleteTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestDeleteBasic tests basic Delete functionality
func (suite *DeleteTestSuite) TestDeleteBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete",
			Action:   "delete",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user001",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestDeleteWithPreHook tests Delete with PreDelete hook
func (suite *DeleteTestSuite) TestDeleteWithPreHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_prehook",
			Action:   "delete",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user002", // This is an active user
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("Deleting active user", resp.Header.Get("X-Delete-Warning"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestDeleteWithPostHook tests Delete with PostDelete hook
func (suite *DeleteTestSuite) TestDeleteWithPostHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_posthook",
			Action:   "delete",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user003",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("user003", resp.Header.Get("X-Deleted-User-Id"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestDeleteNegativeCases tests negative scenarios
func (suite *DeleteTestSuite) TestDeleteNegativeCases() {
	suite.Run("NonExistentUser", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"id": "nonexistent",
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
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				// Missing "id"
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}))
	})

	suite.Run("DeleteTwice", func() {
		// First delete
		resp1 := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"id": "user004",
			},
		})

		suite.Equal(200, resp1.StatusCode)
		body1 := suite.readBody(resp1)
		suite.True(body1.IsOk())

		// Try to delete again
		resp2 := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"id": "user004",
			},
		})

		suite.Equal(200, resp2.StatusCode)
		body2 := suite.readBody(resp2)
		suite.False(body2.IsOk()) // Should fail - already deleted
		suite.Equal(body2.Message, i18n.T(result.ErrMessageRecordNotFound))
	})
}

// TestDeleteRequiresPrimaryKey tests that delete requires primary key
func (suite *DeleteTestSuite) TestDeleteRequiresPrimaryKey() {
	suite.Run("DeleteByEmailShouldFail", func() {
		// DeleteAPI only supports deletion by primary key, not by other fields
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"email": "frank@example.com",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - primary key required
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}))
	})

	suite.Run("DeleteByStatusShouldFail", func() {
		// DeleteAPI only supports deletion by primary key, not by other fields
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": "inactive",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - primary key required
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}))
	})
}
