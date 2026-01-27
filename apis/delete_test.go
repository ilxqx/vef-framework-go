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
type TestUserDeleteResource struct {
	api.Resource
	apis.Delete[TestUser]
}

func NewTestUserDeleteResource() api.Resource {
	return &TestUserDeleteResource{
		Resource: api.NewRPCResource("test/user_delete"),
		Delete:   apis.NewDelete[TestUser]().Public(),
	}
}

// Resource with PreDelete hook.
type TestUserDeleteWithPreHookResource struct {
	api.Resource
	apis.Delete[TestUser]
}

func NewTestUserDeleteWithPreHookResource() api.Resource {
	return &TestUserDeleteWithPreHookResource{
		Resource: api.NewRPCResource("test/user_delete_prehook"),
		Delete: apis.NewDelete[TestUser]().
			Public().
			WithPreDelete(func(model *TestUser, _ orm.DeleteQuery, ctx fiber.Ctx, _ orm.DB) error {
				// Log or check conditions before delete
				if model.Status == "active" {
					ctx.Set("X-Delete-Warning", "Deleting active user")
				}

				return nil
			}),
	}
}

// Resource with PostDelete hook.
type TestUserDeleteWithPostHookResource struct {
	api.Resource
	apis.Delete[TestUser]
}

func NewTestUserDeleteWithPostHookResource() api.Resource {
	return &TestUserDeleteWithPostHookResource{
		Resource: api.NewRPCResource("test/user_delete_posthook"),
		Delete: apis.NewDelete[TestUser]().
			Public().
			WithPostDelete(func(model *TestUser, ctx fiber.Ctx, _ orm.DB) error {
				// Set custom header after deletion
				ctx.Set("X-Deleted-User-ID", model.ID)

				return nil
			}),
	}
}

// DeleteTestSuite tests the Delete API functionality
// including basic delete, PreDelete/PostDelete hooks, negative cases, and primary key requirements.
type DeleteTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *DeleteTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserDeleteResource,
		NewTestUserDeleteWithPreHookResource,
		NewTestUserDeleteWithPostHookResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *DeleteTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestDeleteBasic tests basic Delete functionality.
func (suite *DeleteTestSuite) TestDeleteBasic() {
	suite.T().Logf("Testing Delete API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete",
			Action:   "delete",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user001",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Deleted user001 successfully")
}

// TestDeleteWithPreHook tests Delete with PreDelete hook.
func (suite *DeleteTestSuite) TestDeleteWithPreHook() {
	suite.T().Logf("Testing Delete API with PreDelete hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_prehook",
			Action:   "delete",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user002", // This is an active user
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.Equal("Deleting active user", resp.Header.Get("X-Delete-Warning"), "Should set X-Delete-Warning header via PreDelete hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Deleted user002 with PreDelete hook, warning: %s", resp.Header.Get("X-Delete-Warning"))
}

// TestDeleteWithPostHook tests Delete with PostDelete hook.
func (suite *DeleteTestSuite) TestDeleteWithPostHook() {
	suite.T().Logf("Testing Delete API with PostDelete hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_posthook",
			Action:   "delete",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user003",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.Equal("user003", resp.Header.Get("X-Deleted-User-ID"), "Should set X-Deleted-User-ID header via PostDelete hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Deleted user003 with PostDelete hook, user id: %s", resp.Header.Get("X-Deleted-User-ID"))
}

// TestDeleteNegativeCases tests negative scenarios.
func (suite *DeleteTestSuite) TestDeleteNegativeCases() {
	suite.T().Logf("Testing Delete API negative cases for %s", suite.dbType)

	suite.Run("NonExistentUser", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"id": "nonexistent",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when user does not exist")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Validation failed as expected for non-existent user")
	})

	suite.Run("MissingID", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				// Missing "id"
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when required id is missing")
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}), "Should return primary key required message")

		suite.T().Logf("Validation failed as expected for missing id")
	})

	suite.Run("DeleteTwice", func() {
		// First delete
		resp1 := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"id": "user004",
			},
		})

		suite.Equal(200, resp1.StatusCode, "Should return 200 status code")
		body1 := suite.readBody(resp1)
		suite.True(body1.IsOk(), "Should return successful response on first delete")

		suite.T().Logf("First delete of user004 succeeded")

		// Try to delete again
		resp2 := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"id": "user004",
			},
		})

		suite.Equal(200, resp2.StatusCode, "Should return 200 status code")
		body2 := suite.readBody(resp2)
		suite.False(body2.IsOk(), "Should fail when trying to delete already deleted user")
		suite.Equal(body2.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Second delete of user004 failed as expected - user already deleted")
	})
}

// TestDeleteRequiresPrimaryKey tests that delete requires primary key.
func (suite *DeleteTestSuite) TestDeleteRequiresPrimaryKey() {
	suite.T().Logf("Testing Delete API primary key requirement for %s", suite.dbType)

	suite.Run("DeleteByEmailShouldFail", func() {
		// DeleteApi only supports deletion by primary key, not by other fields
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"email": "frank@example.com",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when trying to delete by email instead of primary key")
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}), "Should return primary key required message")

		suite.T().Logf("Validation failed as expected - cannot delete by email, primary key required")
	})

	suite.Run("DeleteByStatusShouldFail", func() {
		// DeleteApi only supports deletion by primary key, not by other fields
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete",
				Action:   "delete",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": "inactive",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when trying to delete by status instead of primary key")
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}), "Should return primary key required message")

		suite.T().Logf("Validation failed as expected - cannot delete by status, primary key required")
	})
}
