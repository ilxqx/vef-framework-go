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
			PreCreateMany(func(models []TestUser, paramsList []TestUserCreateParams, ctx fiber.Ctx, db orm.Db) error {
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
			PostCreateMany(func(models []TestUser, paramsList []TestUserCreateParams, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header with count
				ctx.Set("X-Created-Count", strconv.Itoa(len(models)))

				return nil
			}),
	}
}

// CreateManyTestSuite is the test suite for CreateMany Api tests.
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

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	suite.NotNil(body.Data)

	// CreateManyApi returns array of primary keys
	pks := suite.readDataAsSlice(body.Data)
	suite.Len(pks, 3)

	for _, pk := range pks {
		pkMap := suite.readDataAsMap(pk)
		suite.NotEmpty(pkMap["id"])
	}
}

// TestCreateManyWithPreHook tests CreateMany with PreCreateMany hook.
func (suite *CreateManyTestSuite) TestCreateManyWithPreHook() {
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

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	// CreateManyApi returns array of primary keys
	pks := suite.readDataAsSlice(body.Data)
	suite.Len(pks, 2)
}

// TestCreateManyWithPostHook tests CreateMany with PostCreateMany hook.
func (suite *CreateManyTestSuite) TestCreateManyWithPostHook() {
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

	suite.Equal(200, resp.StatusCode)
	suite.NotEmpty(resp.Header.Get("X-Created-Count"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())

	pks := suite.readDataAsSlice(body.Data)
	suite.Len(pks, 2)
}

// TestCreateManyNegativeCases tests negative scenarios.
func (suite *CreateManyTestSuite) TestCreateManyNegativeCases() {
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - list must have at least 1 item
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail due to unique constraint
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail due to unique constraint
	})
}

// TestCreateManyTransactionRollback tests that the entire batch rolls back on error.
func (suite *CreateManyTestSuite) TestCreateManyTransactionRollback() {
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())

		// Verify that the first user was not created (transaction rolled back)
		count, err := suite.db.NewSelect().
			Model((*TestUser)(nil)).
			Where(func(cb orm.ConditionBuilder) {
				cb.Equals("email", "rollback1@example.com")
			}).
			Count(suite.ctx)
		suite.NoError(err)
		suite.Equal(int64(0), count, "First user should not exist - transaction should have rolled back")
	})
}
