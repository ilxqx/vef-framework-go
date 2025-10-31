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
type TestUserCreateManyResource struct {
	api.Resource
	apis.CreateManyApi[TestUser, TestUserCreateParams]
}

func NewTestUserCreateManyResource() api.Resource {
	return &TestUserCreateManyResource{
		Resource:      api.NewResource("test/user_create_many"),
		CreateManyApi: apis.NewCreateManyApi[TestUser, TestUserCreateParams]().Public(),
	}
}

// Resource with PreCreateMany hook.
type TestUserCreateManyWithPreHookResource struct {
	api.Resource
	apis.CreateManyApi[TestUser, TestUserCreateParams]
}

func NewTestUserCreateManyWithPreHookResource() api.Resource {
	return &TestUserCreateManyWithPreHookResource{
		Resource: api.NewResource("test/user_create_many_prehook"),
		CreateManyApi: apis.NewCreateManyApi[TestUser, TestUserCreateParams]().
			Public().
			WithPreCreateMany(func(models []TestUser, paramsList []TestUserCreateParams, ctx fiber.Ctx, db orm.Db) error {
				// Add prefix to all names
				for i := range models {
					models[i].Name = "Mr. " + models[i].Name
				}

				return nil
			}),
	}
}

// Resource with PostCreateMany hook.
type TestUserCreateManyWithPostHookResource struct {
	api.Resource
	apis.CreateManyApi[TestUser, TestUserCreateParams]
}

func NewTestUserCreateManyWithPostHookResource() api.Resource {
	return &TestUserCreateManyWithPostHookResource{
		Resource: api.NewResource("test/user_create_many_posthook"),
		CreateManyApi: apis.NewCreateManyApi[TestUser, TestUserCreateParams]().
			Public().
			WithPostCreateMany(func(models []TestUser, paramsList []TestUserCreateParams, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header with count
				ctx.Set("X-Created-Count", strconv.Itoa(len(models)))

				return nil
			}),
	}
}

// CreateManyTestSuite tests the CreateMany API functionality
// including basic batch creation, PreCreateMany/PostCreateMany hooks, negative cases, and transaction rollback.
type CreateManyTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *CreateManyTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserCreateManyResource,
		NewTestUserCreateManyWithPreHookResource,
		NewTestUserCreateManyWithPostHookResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *CreateManyTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestCreateManyBasic tests basic CreateMany functionality.
func (suite *CreateManyTestSuite) TestCreateManyBasic() {
	suite.T().Logf("Testing CreateMany API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_create_many",
			Action:   "create_many",
			Version:  "v1",
		},
		Params: map[string]any{
			"list": []any{
				map[string]any{
					"name":        "User One",
					"email":       "user1@example.com",
					"description": "First user",
					"age":         25,
					"status":      "active",
				},
				map[string]any{
					"name":        "User Two",
					"email":       "user2@example.com",
					"description": "Second user",
					"age":         30,
					"status":      "inactive",
				},
				map[string]any{
					"name":   "User Three",
					"email":  "user3@example.com",
					"age":    35,
					"status": "active",
				},
			},
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	suite.NotNil(body.Data, "Should return data")

	// CreateManyApi returns array of primary keys
	pks := suite.readDataAsSlice(body.Data)
	suite.Len(pks, 3, "Should create 3 users")

	for i, pk := range pks {
		pkMap := suite.readDataAsMap(pk)
		suite.NotEmpty(pkMap["id"], "Should return created user id for user %d", i+1)
		suite.T().Logf("Created user %d with id: %v", i+1, pkMap["id"])
	}
}

// TestCreateManyWithPreHook tests CreateMany with PreCreateMany hook.
func (suite *CreateManyTestSuite) TestCreateManyWithPreHook() {
	suite.T().Logf("Testing CreateMany API with PreCreateMany hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_create_many_prehook",
			Action:   "create_many",
			Version:  "v1",
		},
		Params: map[string]any{
			"list": []any{
				map[string]any{
					"name":   "John",
					"email":  "john.batch@example.com",
					"age":    28,
					"status": "active",
				},
				map[string]any{
					"name":   "Jane",
					"email":  "jane.batch@example.com",
					"age":    26,
					"status": "active",
				},
			},
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	// CreateManyApi returns array of primary keys
	pks := suite.readDataAsSlice(body.Data)
	suite.Len(pks, 2, "Should create 2 users with PreCreateMany hook")

	for i, pk := range pks {
		pkMap := suite.readDataAsMap(pk)
		suite.T().Logf("Created user %d with PreCreateMany hook, id: %v", i+1, pkMap["id"])
	}
}

// TestCreateManyWithPostHook tests CreateMany with PostCreateMany hook.
func (suite *CreateManyTestSuite) TestCreateManyWithPostHook() {
	suite.T().Logf("Testing CreateMany API with PostCreateMany hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_create_many_posthook",
			Action:   "create_many",
			Version:  "v1",
		},
		Params: map[string]any{
			"list": []any{
				map[string]any{
					"name":   "Alice",
					"email":  "alice.batch@example.com",
					"age":    29,
					"status": "active",
				},
				map[string]any{
					"name":   "Bob",
					"email":  "bob.batch@example.com",
					"age":    31,
					"status": "inactive",
				},
			},
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.NotEmpty(resp.Header.Get("X-Created-Count"), "Should set X-Created-Count header via PostCreateMany hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	pks := suite.readDataAsSlice(body.Data)
	suite.Len(pks, 2, "Should create 2 users with PostCreateMany hook")

	createdCount := resp.Header.Get("X-Created-Count")
	suite.T().Logf("Created %s users with PostCreateMany hook", createdCount)

	for i, pk := range pks {
		pkMap := suite.readDataAsMap(pk)
		suite.T().Logf("Created user %d with id: %v", i+1, pkMap["id"])
	}
}

// TestCreateManyNegativeCases tests negative scenarios.
func (suite *CreateManyTestSuite) TestCreateManyNegativeCases() {
	suite.T().Logf("Testing CreateMany API negative cases for %s", suite.dbType)

	suite.Run("EmptyArray", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when list is empty")

		suite.T().Logf("Validation failed as expected for empty list")
	})

	suite.Run("MissingRequiredField", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"name":   "Valid User",
						"email":  "valid@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"email":  "invalid@example.com",
						"age":    30,
						"status": "active",
						// Missing "name"
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when required field 'name' is missing in batch")

		suite.T().Logf("Validation failed as expected for missing required field")
	})

	suite.Run("InvalidEmailInBatch", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"name":   "Valid User",
						"email":  "valid@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"name":   "Invalid User",
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

	suite.Run("InvalidAgeInBatch", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"name":   "Valid User",
						"email":  "valid2@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"name":   "Invalid User",
						"email":  "invalid2@example.com",
						"age":    150, // Invalid: > 120
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when age is greater than 120 in batch")

		suite.T().Logf("Validation failed as expected for invalid age")
	})

	suite.Run("DuplicateEmailInSameBatch", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"name":   "User A",
						"email":  "duplicate.batch@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"name":   "User B",
						"email":  "duplicate.batch@example.com",
						"age":    30,
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail due to duplicate email in same batch")

		suite.T().Logf("Validation failed as expected for duplicate email in batch")
	})

	suite.Run("DuplicateWithExistingRecord", func() {
		// First create a user
		suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"name":   "Existing User",
						"email":  "existing.batch@example.com",
						"age":    25,
						"status": "active",
					},
				},
			},
		})

		// Try to create batch with duplicate email
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"name":   "New User",
						"email":  "new.batch@example.com",
						"age":    30,
						"status": "active",
					},
					map[string]any{
						"name":   "Duplicate User",
						"email":  "existing.batch@example.com",
						"age":    35,
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail due to duplicate email with existing record")

		suite.T().Logf("Validation failed as expected for duplicate with existing record")
	})
}

// TestCreateManyTransactionRollback tests that the entire batch rolls back on error.
func (suite *CreateManyTestSuite) TestCreateManyTransactionRollback() {
	suite.T().Logf("Testing CreateMany API transaction rollback for %s", suite.dbType)

	suite.Run("AllOrNothingSemantics", func() {
		// Try to create a batch where the second item will fail
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_create_many",
				Action:   "create_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"list": []any{
					map[string]any{
						"name":   "Should Not Be Created",
						"email":  "rollback1@example.com",
						"age":    25,
						"status": "active",
					},
					map[string]any{
						"name":   "Invalid User",
						"email":  "rollback2@example.com",
						"age":    0, // Invalid: min=1
						"status": "active",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when one item in batch is invalid")

		// Verify that the first user was not created (transaction rolled back)
		count, err := suite.db.NewSelect().
			Model((*TestUser)(nil)).
			Where(func(cb orm.ConditionBuilder) {
				cb.Equals("email", "rollback1@example.com")
			}).
			Count(suite.ctx)
		suite.NoError(err, "Should successfully query database")
		suite.Equal(int64(0), count, "First user should not exist - transaction should have rolled back")

		suite.T().Logf("Transaction rollback verified: first user was not created")
	})
}
