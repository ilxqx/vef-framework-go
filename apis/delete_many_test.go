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
	apis.DeleteManyAPI[TestUser]
}

func NewTestUserDeleteManyResource() api.Resource {
	return &TestUserDeleteManyResource{
		Resource:      api.NewResource("test/user_delete_many"),
		DeleteManyAPI: apis.NewDeleteManyAPI[TestUser]().Public(),
	}
}

// Resource for composite PK testing.
type TestCompositePKDeleteManyResource struct {
	api.Resource
	apis.DeleteManyAPI[TestCompositePKItem]
}

func NewTestCompositePKDeleteManyResource() api.Resource {
	return &TestCompositePKDeleteManyResource{
		Resource:      api.NewResource("test/composite_pk_delete_many"),
		DeleteManyAPI: apis.NewDeleteManyAPI[TestCompositePKItem]().Public(),
	}
}

// Resource with PreDeleteMany hook.
type TestUserDeleteManyWithPreHookResource struct {
	api.Resource
	apis.DeleteManyAPI[TestUser]
}

func NewTestUserDeleteManyWithPreHookResource() api.Resource {
	return &TestUserDeleteManyWithPreHookResource{
		Resource: api.NewResource("test/user_delete_many_prehook"),
		DeleteManyAPI: apis.NewDeleteManyAPI[TestUser]().
			Public().
			PreDeleteMany(func(models []TestUser, ctx fiber.Ctx, db orm.Db) error {
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
	apis.DeleteManyAPI[TestUser]
}

func NewTestUserDeleteManyWithPostHookResource() api.Resource {
	return &TestUserDeleteManyWithPostHookResource{
		Resource: api.NewResource("test/user_delete_many_posthook"),
		DeleteManyAPI: apis.NewDeleteManyAPI[TestUser]().
			Public().
			PostDeleteMany(func(models []TestUser, ctx fiber.Ctx, tx orm.Db) error {
				// Set custom header with count
				ctx.Set("X-Deleted-Count", strconv.Itoa(len(models)))

				return nil
			}),
	}
}

// DeleteManyTestSuite is the test suite for DeleteMany API tests.
type DeleteManyTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *DeleteManyTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserDeleteManyResource,
		NewTestUserDeleteManyWithPreHookResource,
		NewTestUserDeleteManyWithPostHookResource,
		NewTestCompositePKDeleteManyResource,
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
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_many",
			Action:   "deleteMany",
			Version:  "v1",
		},
		Params: map[string]any{
			"pks": []string{"deluser001", "deluser002", "deluser003"},
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestDeleteManyWithPreHook tests DeleteMany with PreDeleteMany hook.
func (suite *DeleteManyTestSuite) TestDeleteManyWithPreHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_many_prehook",
			Action:   "deleteMany",
			Version:  "v1",
		},
		Params: map[string]any{
			"pks": []string{"deluser004", "deluser005"}, // deluser004 and deluser005 are active
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.NotEmpty(resp.Header.Get("X-Delete-Active-Count"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestDeleteManyWithPostHook tests DeleteMany with PostDeleteMany hook.
func (suite *DeleteManyTestSuite) TestDeleteManyWithPostHook() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_delete_many_posthook",
			Action:   "deleteMany",
			Version:  "v1",
		},
		Params: map[string]any{
			"pks": []string{"deluser006"},
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.NotEmpty(resp.Header.Get("X-Deleted-Count"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
}

// TestDeleteManyNegativeCases tests negative scenarios.
func (suite *DeleteManyTestSuite) TestDeleteManyNegativeCases() {
	suite.Run("EmptyArray", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - pks must have at least 1 item
	})

	suite.Run("NonExistentUser", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"user008", "nonexistent"},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - one user not found
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound))
	})

	suite.Run("MissingIds", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				// Missing "pks"
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
		// Message should indicate PKs is required
		suite.Contains(body.Message, i18n.T("batch_delete_pks"))
	})

	suite.Run("InvalidPksType", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": "not-an-array", // Should be array
			},
		})

		// Invalid parameter type causes 500 error
		suite.Equal(500, resp.StatusCode)
	})

	suite.Run("AllNonExistent", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"nonexistent1", "nonexistent2"},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - users not found
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound))
	})

	suite.Run("DeleteTwice", func() {
		// First delete
		resp1 := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser009", "deluser010"},
			},
		})

		suite.Equal(200, resp1.StatusCode)
		body1 := suite.readBody(resp1)
		suite.True(body1.IsOk())

		// Try to delete same users again
		resp2 := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser009", "deluser010"},
			},
		})

		suite.Equal(200, resp2.StatusCode)
		body2 := suite.readBody(resp2)
		suite.False(body2.IsOk()) // Should fail - already deleted
		suite.Equal(body2.Message, i18n.T(result.ErrMessageRecordNotFound))
	})

	suite.Run("PartiallyDeleted", func() {
		// deluser001 was deleted by TestDeleteManyBasic, deluser007 still exists
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser001", "deluser007"}, // deluser001 already deleted, deluser007 still exists
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // Should fail - one user already deleted
		suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound))
	})
}

// TestDeleteManyTransactionRollback tests that the entire batch rolls back on error.
func (suite *DeleteManyTestSuite) TestDeleteManyTransactionRollback() {
	suite.Run("AllOrNothingSemantics", func() {
		// Try to delete a batch where the second item doesn't exist
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser007", "nonexistent_rollback"},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())

		// Verify that the first user was not deleted (transaction rolled back)
		count, err := suite.db.NewSelect().Model((*TestUser)(nil)).Where(func(cb orm.ConditionBuilder) {
			cb.Equals("id", "deluser007")
		}).Count(suite.ctx)
		suite.NoError(err)
		suite.Equal(int64(1), count, "First user should still exist - transaction should have rolled back")
	})
}

// TestDeleteManyPrimaryKeyFormats tests different primary key format support.
func (suite *DeleteManyTestSuite) TestDeleteManyPrimaryKeyFormats() {
	suite.Run("SinglePK_DirectValues", func() {
		// Single PK with direct value array: ["id1", "id2"]
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []string{"deluser008"},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
	})

	suite.Run("SinglePK_MapFormat", func() {
		// Single PK with map format: [{"id": "value1"}, {"id": "value2"}]
		// Test the map format with already deleted users (from DeleteTwice)
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					map[string]any{"id": "deluser009"},
					map[string]any{"id": "deluser010"},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // These users were already deleted by DeleteTwice
	})

	suite.Run("SinglePK_MixedFormat", func() {
		// Mixed format - both direct values and maps
		// Use already deleted users to demonstrate the mixed format
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					"deluser001",                       // direct value - already deleted
					map[string]any{"id": "deluser002"}, // map format - already deleted
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk()) // These users were already deleted
	})

	// Composite PK tests with TestCompositePKItem model
	suite.Run("CompositePK_MapFormatRequired", func() {
		// Test with map format (correct for composite PKs)
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/composite_pk_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					map[string]any{"tenantId": "tenant001", "itemCode": "item001"},
					map[string]any{"tenantId": "tenant001", "itemCode": "item002"},
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Composite PK deletion with map format should succeed")

		// Verify items were deleted
		count, err := suite.db.NewSelect().
			Model((*TestCompositePKItem)(nil)).
			Where(func(cb orm.ConditionBuilder) {
				cb.Equals("tenant_id", "tenant001")
			}).
			Count(suite.ctx)
		suite.NoError(err)
		suite.Equal(int64(0), count, "Both items for tenant001 should be deleted")
	})

	suite.Run("CompositePK_PartialKeys", func() {
		// Test with missing one of the composite keys
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/composite_pk_delete_many",
				Action:   "deleteMany",
				Version:  "v1",
			},
			Params: map[string]any{
				"pks": []any{
					map[string]any{"tenantId": "tenant002"}, // Missing itemCode
				},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when missing composite PK fields")
		suite.Contains(body.Message, "itemCode", "Error should mention missing itemCode field")
	})
}
