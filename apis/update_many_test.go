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
	apis.UpdateManyApi[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateManyResource() api.Resource {
	return &TestUserUpdateManyResource{
		Resource:      api.NewResource("test/user_update_many"),
		UpdateManyApi: apis.NewUpdateManyApi[TestUser, TestUserUpdateParams]().Public(),
	}
}

// Resource with PreUpdateMany hook.
type TestUserUpdateManyWithPreHookResource struct {
	api.Resource
	apis.UpdateManyApi[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateManyWithPreHookResource() api.Resource {
	return &TestUserUpdateManyWithPreHookResource{
		Resource: api.NewResource("test/user_update_many_prehook"),
		UpdateManyApi: apis.NewUpdateManyApi[TestUser, TestUserUpdateParams]().
			Public().
			WithPreUpdateMany(func(oldModels, models []TestUser, paramsList []TestUserUpdateParams, ctx fiber.Ctx, db orm.Db) error {
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
	apis.UpdateManyApi[TestUser, TestUserUpdateParams]
}

func NewTestUserUpdateManyWithPostHookResource() api.Resource {
	return &TestUserUpdateManyWithPostHookResource{
		Resource: api.NewResource("test/user_update_many_posthook"),
		UpdateManyApi: apis.NewUpdateManyApi[TestUser, TestUserUpdateParams]().
			Public().
			WithPostUpdateMany(func(oldModels, models []TestUser, paramsList []TestUserUpdateParams, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header with count
				ctx.Set("X-Updated-Count", strconv.Itoa(len(models)))

				return nil
			}),
	}
}

// UpdateManyTestSuite tests the UpdateMany API functionality
// including basic batch update, PreUpdateMany/PostUpdateMany hooks, negative cases, transaction rollback, and partial updates.
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
	suite.T().Logf("Testing UpdateMany API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many",
			Action:   "update_many",
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

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	// UpdateManyApi returns no data, just success status

	suite.T().Logf("Successfully updated 2 users in batch")
}

// TestUpdateManyWithPreHook tests UpdateMany with PreUpdateMany hook.
func (suite *UpdateManyTestSuite) TestUpdateManyWithPreHook() {
	suite.T().Logf("Testing UpdateMany API with PreUpdateMany hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many_prehook",
			Action:   "update_many",
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

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Successfully updated 2 users with PreUpdateMany hook")
}

// TestUpdateManyWithPostHook tests UpdateMany with PostUpdateMany hook.
func (suite *UpdateManyTestSuite) TestUpdateManyWithPostHook() {
	suite.T().Logf("Testing UpdateMany API with PostUpdateMany hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many_posthook",
			Action:   "update_many",
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

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.NotEmpty(resp.Header.Get("X-Updated-Count"), "Should set X-Updated-Count header via PostUpdateMany hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	updatedCount := resp.Header.Get("X-Updated-Count")
	suite.T().Logf("Updated %s users with PostUpdateMany hook", updatedCount)
}

// TestUpdateManyNegativeCases tests negative scenarios.
func (suite *UpdateManyTestSuite) TestUpdateManyNegativeCases() {
	suite.T().Logf("Testing UpdateMany API negative cases for %s", suite.dbType)

	suite.Run("EmptyArray", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
				Version:  "v1",
			},
			Params: map[string]any{"list": []any{}},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when list is empty")

		suite.T().Logf("Validation failed as expected for empty list")
	})

	suite.Run("NonExistentUser", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when one user does not exist")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Validation failed as expected for non-existent user in batch")
	})

	suite.Run("MissingId", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when required id is missing")
		suite.Equal(body.Message, i18n.T("primary_key_required", map[string]any{"field": "id"}), "Should return primary key required message")

		suite.T().Logf("Validation failed as expected for missing id")
	})

	suite.Run("InvalidEmail", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when email format is invalid in batch")

		suite.T().Logf("Validation failed as expected for invalid email format")
	})

	suite.Run("InvalidAge", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when age is less than 1 in batch")

		suite.T().Logf("Validation failed as expected for invalid age")
	})

	suite.Run("DuplicateEmailInBatch", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail due to duplicate email in batch")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordAlreadyExists), "Should return record already exists message")

		suite.T().Logf("Validation failed as expected for duplicate email in batch")
	})

	suite.Run("DuplicateEmailWithExisting", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail due to duplicate email with existing record")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordAlreadyExists), "Should return record already exists message")

		suite.T().Logf("Validation failed as expected for duplicate email with existing record")
	})
}

// TestUpdateManyTransactionRollback tests that the entire batch rolls back on error.
func (suite *UpdateManyTestSuite) TestUpdateManyTransactionRollback() {
	suite.T().Logf("Testing UpdateMany API transaction rollback for %s", suite.dbType)

	suite.Run("AllOrNothingSemantics", func() {
		// Try to update a batch where the second item will fail
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_update_many",
				Action:   "update_many",
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when one item in batch is invalid")

		// Verify that the first user was not updated (transaction rolled back)
		var user TestUser

		err := suite.db.NewSelect().Model(&user).Where(func(cb orm.ConditionBuilder) {
			cb.Equals("id", "user001")
		}).Scan(suite.ctx, &user)
		suite.NoError(err, "Should successfully query database")
		suite.NotEqual("rollback1@example.com", user.Email, "Email should not have been updated - transaction should have rolled back")

		suite.T().Logf("Transaction rollback verified: first user was not updated")
	})
}

// TestUpdateManyPartialUpdate tests updating only some fields.
func (suite *UpdateManyTestSuite) TestUpdateManyPartialUpdate() {
	suite.T().Logf("Testing UpdateMany API partial update for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_update_many",
			Action:   "update_many",
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

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Successfully partially updated 2 users")
}
