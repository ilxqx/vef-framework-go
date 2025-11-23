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
type TestUserDeleteManyResource struct {
	api.Resource
	apis.DeleteManyApi[TestUser]
}

func NewTestUserDeleteManyResource() api.Resource {
	return &TestUserDeleteManyResource{
		Resource:      api.NewResource("test/user_delete_many"),
		DeleteManyApi: apis.NewDeleteManyApi[TestUser]().Public(),
	}
}

// Resource for composite Pk testing.
type TestCompositePkDeleteManyResource struct {
	api.Resource
	apis.DeleteManyApi[TestCompositePkItem]
}

func NewTestCompositePkDeleteManyResource() api.Resource {
	return &TestCompositePkDeleteManyResource{
		Resource:      api.NewResource("test/composite_pk_delete_many"),
		DeleteManyApi: apis.NewDeleteManyApi[TestCompositePkItem]().Public(),
	}
}

// Resource with PreDeleteMany hook.
type TestUserDeleteManyWithPreHookResource struct {
	api.Resource
	apis.DeleteManyApi[TestUser]
}

func NewTestUserDeleteManyWithPreHookResource() api.Resource {
	return &TestUserDeleteManyWithPreHookResource{
		Resource: api.NewResource("test/user_delete_many_prehook"),
		DeleteManyApi: apis.NewDeleteManyApi[TestUser]().
			Public().
			WithPreDeleteMany(func(models []TestUser, query orm.DeleteQuery, ctx fiber.Ctx, tx orm.Db) error {
				// Check if any active users in batch
				activeCount := 0

				for _, model := range models {
					if model.Status == "active" {
						activeCount++
					}
				}

				if activeCount > 0 {
					ctx.Set("X-Delete-Active-Count", strconv.Itoa(activeCount))
				}

				return nil
			}),
	}
}

// Resource with PostDeleteMany hook.
type TestUserDeleteManyWithPostHookResource struct {
	api.Resource
	apis.DeleteManyApi[TestUser]
}

func NewTestUserDeleteManyWithPostHookResource() api.Resource {
	return &TestUserDeleteManyWithPostHookResource{
		Resource: api.NewResource("test/user_delete_many_posthook"),
		DeleteManyApi: apis.NewDeleteManyApi[TestUser]().
			Public().
			WithPostDeleteMany(func(models []TestUser, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header with count
				ctx.Set("X-Deleted-Count", strconv.Itoa(len(models)))

				return nil
			}),
	}
}

// DeleteManyTestSuite tests the DeleteMany API functionality
// including basic batch delete, PreDeleteMany/PostDeleteMany hooks, negative cases, transaction rollback, and primary key formats.
type DeleteManyTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *DeleteManyTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserDeleteManyResource,
		NewTestUserDeleteManyWithPreHookResource,
		NewTestUserDeleteManyWithPostHookResource,
		NewTestCompositePkDeleteManyResource,
	)

	// Insert additional test users specifically for delete tests
	deluser001 := TestUser{Name: "Delete User 1", Email: "deluser001@example.com", Age: 25, Status: "active"}
	deluser001.Id = "deluser001"
	deluser002 := TestUser{Name: "Delete User 2", Email: "deluser002@example.com", Age: 26, Status: "active"}
	deluser002.Id = "deluser002"
	deluser003 := TestUser{Name: "Delete User 3", Email: "deluser003@example.com", Age: 27, Status: "inactive"}
	deluser003.Id = "deluser003"
	deluser004 := TestUser{Name: "Delete User 4", Email: "deluser004@example.com", Age: 28, Status: "active"}
	deluser004.Id = "deluser004"
	deluser005 := TestUser{Name: "Delete User 5", Email: "deluser005@example.com", Age: 29, Status: "active"}
	deluser005.Id = "deluser005"
	deluser006 := TestUser{Name: "Delete User 6", Email: "deluser006@example.com", Age: 30, Status: "inactive"}
	deluser006.Id = "deluser006"
	deluser007 := TestUser{Name: "Delete User 7", Email: "deluser007@example.com", Age: 31, Status: "active"}
	deluser007.Id = "deluser007"
	deluser008 := TestUser{Name: "Delete User 8", Email: "deluser008@example.com", Age: 32, Status: "active"}
	deluser008.Id = "deluser008"
	deluser009 := TestUser{Name: "Delete User 9", Email: "deluser009@example.com", Age: 33, Status: "inactive"}
	deluser009.Id = "deluser009"
	deluser010 := TestUser{Name: "Delete User 10", Email: "deluser010@example.com", Age: 34, Status: "active"}
	deluser010.Id = "deluser010"

	additionalUsers := []TestUser{
		deluser001, deluser002, deluser003, deluser004, deluser005,
		deluser006, deluser007, deluser008, deluser009, deluser010,
	}

	_, err := suite.db.NewInsert().Model(&additionalUsers).Exec(suite.ctx)
	suite.Require().NoError(err, "Failed to insert additional test users for delete tests")
}

// TearDownSuite runs once after all tests in the suite.
func (suite *DeleteManyTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestDeleteManyBasic tests basic DeleteMany functionality.
func (suite *DeleteManyTestSuite) TestDeleteManyBasic() {
	suite.T().Logf("Testing DeleteMany API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_many",
			Action:   "delete_many",
			Version:  "v1",
		},
		Params: map[string]any{
			"pks": []string{"deluser001", "deluser002", "deluser003"},
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	suite.T().Logf("Successfully deleted 3 users in batch")
}

// TestDeleteManyWithPreHook tests DeleteMany with PreDeleteMany hook.
func (suite *DeleteManyTestSuite) TestDeleteManyWithPreHook() {
	suite.T().Logf("Testing DeleteMany API with PreDeleteMany hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_many_prehook",
			Action:   "delete_many",
			Version:  "v1",
		},
		Params: map[string]any{
			"pks": []string{"deluser004", "deluser005"}, // deluser004 and deluser005 are active
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.NotEmpty(resp.Header.Get("X-Delete-Active-Count"), "Should set X-Delete-Active-Count header via PreDeleteMany hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	activeCount := resp.Header.Get("X-Delete-Active-Count")
	suite.T().Logf("Deleted 2 users with PreDeleteMany hook, active count: %s", activeCount)
}

// TestDeleteManyWithPostHook tests DeleteMany with PostDeleteMany hook.
func (suite *DeleteManyTestSuite) TestDeleteManyWithPostHook() {
	suite.T().Logf("Testing DeleteMany API with PostDeleteMany hook for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_many_posthook",
			Action:   "delete_many",
			Version:  "v1",
		},
		Params: map[string]any{
			"pks": []string{"deluser006"},
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	suite.NotEmpty(resp.Header.Get("X-Deleted-Count"), "Should set X-Deleted-Count header via PostDeleteMany hook")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")

	deletedCount := resp.Header.Get("X-Deleted-Count")
	suite.T().Logf("Deleted %s users with PostDeleteMany hook", deletedCount)
}

// TestDeleteManyNegativeCases tests negative scenarios.
func (suite *DeleteManyTestSuite) TestDeleteManyNegativeCases() {
	suite.T().Logf("Testing DeleteMany API negative cases for %s", suite.dbType)

	suite.Run("EmptyArray", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when pks list is empty")

		suite.T().Logf("Validation failed as expected for empty pks list")
	})

	suite.Run("NonExistentUser", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"user008", "nonexistent"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when one user does not exist in batch")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Validation failed as expected for non-existent user in batch")
	})

	suite.Run("MissingIds", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				// Missing "pks"
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when pks parameter is missing")
		suite.Contains(body.Message, i18n.T("batch_delete_pks"), "Message should indicate pks is required")

		suite.T().Logf("Validation failed as expected for missing pks parameter")
	})

	suite.Run("InvalidPksType", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": "not-an-array", // Should be array
			},
		})

		suite.Equal(500, resp.StatusCode, "Should return 500 status code for invalid parameter type")

		suite.T().Logf("Validation failed as expected for invalid pks parameter type")
	})

	suite.Run("AllNonExistent", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"nonexistent1", "nonexistent2"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when all users do not exist")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Validation failed as expected for all non-existent users")
	})

	suite.Run("DeleteTwice", func() {
		// First delete
		resp1 := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser009", "deluser010"},
			},
		})

		suite.Equal(200, resp1.StatusCode, "Should return 200 status code")
		body1 := suite.readBody(resp1)
		suite.True(body1.IsOk(), "Should return successful response on first delete")

		suite.T().Logf("First delete of deluser009 and deluser010 succeeded")

		// Try to delete same users again
		resp2 := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser009", "deluser010"},
			},
		})

		suite.Equal(200, resp2.StatusCode, "Should return 200 status code")
		body2 := suite.readBody(resp2)
		suite.False(body2.IsOk(), "Should fail when trying to delete already deleted users")
		suite.Equal(body2.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Second delete of deluser009 and deluser010 failed as expected - users already deleted")
	})

	suite.Run("PartiallyDeleted", func() {
		// deluser001 was deleted by TestDeleteManyBasic, deluser007 still exists
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser001", "deluser007"}, // deluser001 already deleted, deluser007 still exists
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when one user is already deleted")
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")

		suite.T().Logf("Validation failed as expected for partially deleted batch")
	})
}

// TestDeleteManyTransactionRollback tests that the entire batch rolls back on error.
func (suite *DeleteManyTestSuite) TestDeleteManyTransactionRollback() {
	suite.T().Logf("Testing DeleteMany API transaction rollback for %s", suite.dbType)

	suite.Run("AllOrNothingSemantics", func() {
		// Try to delete a batch where the second item doesn't exist
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser007", "nonexistent_rollback"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when one item in batch does not exist")

		// Verify that the first user was not deleted (transaction rolled back)
		count, err := suite.db.NewSelect().Model((*TestUser)(nil)).Where(func(cb orm.ConditionBuilder) {
			cb.Equals("id", "deluser007")
		}).Count(suite.ctx)
		suite.NoError(err, "Should successfully query database")
		suite.Equal(int64(1), count, "First user should still exist - transaction should have rolled back")

		suite.T().Logf("Transaction rollback verified: first user was not deleted")
	})
}

// TestDeleteManyPrimaryKeyFormats tests different primary key format support.
func (suite *DeleteManyTestSuite) TestDeleteManyPrimaryKeyFormats() {
	suite.T().Logf("Testing DeleteMany API primary key formats for %s", suite.dbType)

	suite.Run("SinglePk_DirectValues", func() {
		// Single Pk with direct value array: ["id1", "id2"]
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser008"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response for single primary key with direct values")

		suite.T().Logf("Successfully deleted user using single primary key direct value format")
	})

	suite.Run("SinglePk_MapFormat", func() {
		// Single Pk with map format: [{"id": "value1"}, {"id": "value2"}]
		// Test the map format with already deleted users (from DeleteTwice)
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					map[string]any{"id": "deluser009"},
					map[string]any{"id": "deluser010"},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when users were already deleted")

		suite.T().Logf("Map format correctly handled - users were already deleted by DeleteTwice")
	})

	suite.Run("SinglePk_MixedFormat", func() {
		// Mixed format - both direct values and maps
		// Use already deleted users to demonstrate the mixed format
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					"deluser001",                       // direct value - already deleted
					map[string]any{"id": "deluser002"}, // map format - already deleted
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when users were already deleted")

		suite.T().Logf("Mixed format correctly handled - users were already deleted")
	})

	// Composite Pk tests with TestCompositePkItem model
	suite.Run("CompositePk_MapFormatRequired", func() {
		// Test with map format (correct for composite Pks)
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/composite_pk_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					map[string]any{"tenantId": "tenant001", "itemCode": "item001"},
					map[string]any{"tenantId": "tenant001", "itemCode": "item002"},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Composite primary key deletion with map format should succeed")

		// Verify items were deleted
		count, err := suite.db.NewSelect().
			Model((*TestCompositePkItem)(nil)).
			Where(func(cb orm.ConditionBuilder) {
				cb.Equals("tenant_id", "tenant001")
			}).
			Count(suite.ctx)
		suite.NoError(err, "Should successfully query database")
		suite.Equal(int64(0), count, "Both items for tenant001 should be deleted")

		suite.T().Logf("Successfully deleted 2 items using composite primary key map format")
	})

	suite.Run("CompositePk_PartialKeys", func() {
		// Test with missing one of the composite keys
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/composite_pk_delete_many",
				Action:   "delete_many",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					map[string]any{"tenantId": "tenant002"}, // Missing itemCode
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when missing composite primary key fields")
		suite.Contains(body.Message, "itemCode", "Error message should mention missing itemCode field")

		suite.T().Logf("Validation failed as expected for missing composite primary key field")
	})
}
