package apis_test

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// Test Resources.
type TestUserUpdateManyResource struct {
	api.Resource
	apis.UpdateManyAPI[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateManyResource() api.Resource {
	return &TestUserUpdateManyResource{
		Resource:      api.NewResource("test/user_update_many"),
		UpdateManyAPI: apis.NewUpdateManyAPI[TestUser, TestUserUpdateParams]().Public(),
	}
}

// Resource with PreUpdateMany hook.
type TestUserUpdateManyWithPreHookResource struct {
	api.Resource
	apis.UpdateManyAPI[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateManyWithPreHookResource() api.Resource {
	return &TestUserUpdateManyWithPreHookResource{
		Resource: api.NewResource("test/user_update_many_prehook"),
		UpdateManyAPI: apis.NewUpdateManyAPI[TestUser, TestUserUpdateParams]().
			Public().
			PreUpdateMany(func(oldModels, models []TestUser, paramsList []TestUserUpdateParams, ctx fiber.Ctx, db orm.Db) error {
				// Add suffix to all descriptions
				for i := range models {
					if paramsList[i].Description != "" {
						models[i].Description = paramsList[i].Description + " [Batch Updated]"
					}
				}

				return nil
			}),
	}
}

// Resource with PostUpdateMany hook.
type TestUserUpdateManyWithPostHookResource struct {
	api.Resource
	apis.UpdateManyAPI[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateManyWithPostHookResource() api.Resource {
	return &TestUserUpdateManyWithPostHookResource{
		Resource: api.NewResource("test/user_update_many_posthook"),
		UpdateManyAPI: apis.NewUpdateManyAPI[TestUser, TestUserUpdateParams]().
			Public().
			PostUpdateMany(func(oldModels, models []TestUser, paramsList []TestUserUpdateParams, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header with count
				ctx.Set("X-Updated-Count", strconv.Itoa(len(models)))

				return nil
			}),
	}
}

// UpdateManyTestSuite is the test suite for UpdateMany API tests.
type UpdateManyTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *UpdateManyTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserUpdateManyResource,
		NewTestUserUpdateManyWithPreHookResource,
		NewTestUserUpdateManyWithPostHookResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *UpdateManyTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestUpdateManyBasic tests basic UpdateMany functionality.
func (suite *UpdateManyTestSuite) TestUpdateManyBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many",
			Action:   "updateMany",
			Version:  "v1",
		},
		Params: map[string]any{
			"list": []any{
				map[string]any{
					"id":          "user001",
					"name":        "Updated Alice",
					"email":       "alice.updated@example.com",
					"description": "Updated description",
					"age":         26,
					"status":      "inactive",
				},
				map[string]any{
					"id":     "user002",
					"name":   "Updated Bob",
					"email":  "bob.updated@example.com",
					"age":    31,
					"status": "active",
				},
			},
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	// UpdateManyAPI returns no data, just success status
}

// TestUpdateManyWithPreHook tests UpdateMany with PreUpdateMany hook.
func (suite *UpdateManyTestSuite) TestUpdateManyWithPreHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many_prehook",
			Action:   "updateMany",
			Version:  "v1",
		},
		Params: map[string]any{
			"list": []any{
				map[string]any{
					"id":          "user003",
					"name":        "Charlie Updated",
					"email":       "charlie.updated@example.com",
					"description": "New description",
					"age":         29,
					"status":      "active",
				},
				map[string]any{
					"id":          "user004",
					"name":        "David Updated",
					"email":       "david.updated@example.com",
					"description": "Another description",
					"age":         33,
					"status":      "inactive",
				},
			},
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestUpdateManyWithPostHook tests UpdateMany with PostUpdateMany hook.
func (suite *UpdateManyTestSuite) TestUpdateManyWithPostHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many_posthook",
			Action:   "updateMany",
			Version:  "v1",
		},
		Params: map[string]any{
			"list": []any{
				map[string]any{
					"id":     "user005",
					"name":   "Eve Updated",
					"email":  "eve.updated@example.com",
					"age":    28,
					"status": "active",
				},
				map[string]any{
					"id":     "user006",
					"name":   "Frank Updated",
					"email":  "frank.updated@example.com",
					"age":    36,
					"status": "inactive",
				},
			},
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.NotEmpty(resp.Header.Get("X-Updated-Count"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestUpdateManyNegativeCases tests negative scenarios.
func (suite *UpdateManyTestSuite) TestUpdateManyNegativeCases() {
	suite.Run("EmptyArray", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{"list": []any{}},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - list must have at least 1 item
	})

	suite.Run("NonExistentUser", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"id":     "user007",
						"name":   "Valid Update",
						"email":  "valid@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"id":     "nonexistent",
						"name":   "Invalid Update",
						"email":  "invalid@example.com",
						"age":    30,
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - one user not found
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound))
	})

	suite.Run("MissingId", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"id":     "user008",
						"name":   "Valid Update",
						"email":  "valid@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						// Missing "id"
						"name":   "Invalid Update",
						"email":  "invalid@example.com",
						"age":    30,
						"status": "active",
					},
				},
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
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"id":     "user009",
						"name":   "Valid Update",
						"email":  "valid@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"id":     "user010",
						"name":   "Invalid Update",
						"email":  "not-an-email",
						"age":    30,
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("InvalidAge", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"id":     "user011",
						"name":   "Valid Update",
						"email":  "valid@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"id":     "user012",
						"name":   "Invalid Update",
						"email":  "invalid@example.com",
						"age":    0, // Invalid: min=1
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("DuplicateEmailInBatch", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"id":     "user001",
						"name":   "User One",
						"email":  "duplicate.batch.update@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"id":     "user002",
						"name":   "User Two",
						"email":  "duplicate.batch.update@example.com", // Duplicate with previous
						"age":    30,
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail due to unique constraint
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordAlreadyExists))
	})

	suite.Run("DuplicateEmailWithExisting", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"id":     "user001",
						"name":   "User One",
						"email":  "grace@example.com", // Existing email from user007
						"age":    25,
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail due to unique constraint
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordAlreadyExists))
	})
}

// TestUpdateManyTransactionRollback tests that the entire batch rolls back on error.
func (suite *UpdateManyTestSuite) TestUpdateManyTransactionRollback() {
	suite.Run("AllOrNothingSemantics", func() {
		// Try to update a batch where the second item will fail
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "updateMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"id":     "user001",
						"name":   "Should Not Be Updated",
						"email":  "rollback1@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"id":     "nonexistent_rollback",
						"name":   "Invalid User",
						"email":  "rollback2@example.com",
						"age":    30,
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())

		// Verify that the first user was not updated (transaction rolled back)
		var user TestUser

		err := suite.db.NewSelect().Model(&user).Where(func(cb orm.ConditionBuilder) {
			cb.Equals("id", "user001")
		}).Scan(suite.ctx, &user)
		suite.NoError(err)
		suite.NotEqual("rollback1@example.com", user.Email, "Email should not have been updated - transaction should have rolled back")
	})
}

// TestUpdateManyPartialUpdate tests updating only some fields.
func (suite *UpdateManyTestSuite) TestUpdateManyPartialUpdate() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many",
			Action:   "updateMany",
			Version:  "v1",
		},
		Params: map[string]any{
			"list": []any{
				map[string]any{
					"id":     "user009",
					"name":   "Partial Update 1",
					"email":  "partial1@example.com",
					"age":    25,
					"status": "active",
					// Not updating description
				},
				map[string]any{
					"id":     "user010",
					"name":   "Partial Update 2",
					"email":  "partial2@example.com",
					"age":    30,
					"status": "inactive",
					// Not updating description
				},
			},
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}
