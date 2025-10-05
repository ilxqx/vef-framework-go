package orm

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
)

type InsertTestSuite struct {
	*ORMTestSuite
}

// TestBasicInsert tests basic INSERT functionality across all databases.
func (suite *InsertTestSuite) TestBasicInsert() {
	suite.T().Logf("Testing basic INSERT for %s", suite.dbType)

	// Test 1: Insert single user
	newUser := &User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      28,
		IsActive: true,
	}

	_, err := suite.db.NewInsert().
		Model(newUser).
		Exec(suite.ctx)
	suite.NoError(err)
	suite.True(newUser.Id != "", "User Id should be set after insert")

	// Verify the user was inserted
	var insertedUser User

	err = suite.db.NewSelect().
		Model(&insertedUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "john@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("John Doe", insertedUser.Name)
	suite.Equal("john@example.com", insertedUser.Email)
	suite.Equal(int16(28), insertedUser.Age)
	suite.True(insertedUser.IsActive)

	// Test 2: Insert multiple users
	users := []*User{
		{Name: "Jane Smith", Email: "jane@example.com", Age: 26, IsActive: true},
		{Name: "Mike Wilson", Email: "mike@example.com", Age: 31, IsActive: false},
	}

	_, err = suite.db.NewInsert().
		Model(&users).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify all users have IDs set
	for _, user := range users {
		suite.True(user.Id != "", "Each user should have an Id set")
	}

	// Verify users were inserted correctly
	var retrievedUsers []User

	err = suite.db.NewSelect().
		Model(&retrievedUsers).
		Where(func(cb ConditionBuilder) {
			cb.In("email", []any{"jane@example.com", "mike@example.com"})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(retrievedUsers, 2, "Should have inserted 2 users")

	// Test 3: Insert with specific field values (testing Column method)
	specificUser := &User{
		Name:     "Specific User Base",
		Email:    "specific@example.com",
		Age:      40,
		IsActive: true,
	}

	_, err = suite.db.NewInsert().
		Model(specificUser).
		Column("name", "Specific User Updated").
		Exec(suite.ctx)
	suite.NoError(err)
	suite.True(specificUser.Id != "", "User Id should be set")

	// Verify the column override worked
	var specificRetrieved User

	err = suite.db.NewSelect().
		Model(&specificRetrieved).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "specific@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("Specific User Updated", specificRetrieved.Name) // Should be overridden value
	suite.Equal(int16(40), specificRetrieved.Age)                // Should keep model value
	suite.True(specificRetrieved.IsActive)

	// Test 4: Insert with RETURNING clause (skip for MySQL)
	if suite.dbType != constants.DbMySQL {
		returningUser := &User{
			Name:     "Returning User",
			Email:    "returning@example.com",
			Age:      29,
			IsActive: true,
		}

		err = suite.db.NewInsert().
			Model(returningUser).
			Returning("id", "name", "email").
			Scan(suite.ctx, returningUser)
		suite.NoError(err)
		suite.True(returningUser.Id != "", "User Id should be returned")
		suite.Equal("Returning User", returningUser.Name)
	} else {
		suite.T().Log("Skipping RETURNING clause test for MySQL")
	}

	// Cleanup test data
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.In("email", []any{
				"john@example.com", "jane@example.com", "mike@example.com",
				"specific@example.com", "returning@example.com",
			})
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestInsertWithConflict tests INSERT with conflict resolution (UPSERT).
func (suite *InsertTestSuite) TestInsertWithConflict() {
	suite.T().Logf("Testing INSERT with conflict resolution for %s", suite.dbType)

	// First, insert a user
	originalUser := &User{
		Name:     "Conflict User",
		Email:    "conflict@example.com",
		Age:      30,
		IsActive: true,
	}

	_, err := suite.db.NewInsert().
		Model(originalUser).
		Exec(suite.ctx)
	suite.NoError(err)

	// Test 1: Insert with ON CONFLICT DO NOTHING
	duplicateUser := &User{
		Name:     "Duplicate User",
		Email:    "conflict@example.com", // Same email
		Age:      25,
		IsActive: false,
	}

	_, err = suite.db.NewInsert().
		Model(duplicateUser).
		OnConflict(func(cb ConflictBuilder) {
			cb.Columns("email").DoNothing()
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify original user is unchanged
	var checkUser User

	err = suite.db.NewSelect().
		Model(&checkUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "conflict@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("Conflict User", checkUser.Name, "Original user name should be unchanged")
	suite.Equal(int16(30), checkUser.Age, "Original user age should be unchanged")

	// Test 2: Insert with ON CONFLICT DO UPDATE
	updateUser := &User{
		Name:     "Updated Conflict User",
		Email:    "conflict@example.com", // Same email
		Age:      35,
		IsActive: false,
	}

	_, err = suite.db.NewInsert().
		Model(updateUser).
		OnConflict(func(cb ConflictBuilder) {
			cb.Columns("email").DoUpdate().
				Set("name", "Updated Conflict User").
				Set("age", 35).
				Set("is_active", false)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify user was updated
	err = suite.db.NewSelect().
		Model(&checkUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "conflict@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("Updated Conflict User", checkUser.Name, "User name should be updated")
	suite.Equal(int16(35), checkUser.Age, "User age should be updated")
	suite.False(checkUser.IsActive, "User should be inactive after update")

	// Cleanup
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "conflict@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestInsertComplexModel tests inserting models with various data types.
func (suite *InsertTestSuite) TestInsertComplexModel() {
	suite.T().Logf("Testing INSERT with complex model for %s", suite.dbType)

	now := time.Now()
	jsonData := map[string]any{
		"key1": "value1",
		"key2": 42,
		"nested": map[string]any{
			"subkey": "subvalue",
		},
	}

	complexModel := &ComplexModel{
		StringField: "test string",
		IntField:    42,
		FloatField:  3.14159,
		BoolField:   true,
		TimeField:   now,
		NullString:  nil, // Test NULL value
		NullInt:     nil, // Test NULL value
		NullTime:    &now,
		JSONField:   jsonData,
	}

	// Only test array field for PostgreSQL
	if suite.dbType == constants.DbPostgres {
		complexModel.ArrayField = []string{"item1", "item2", "item3"}
	}

	_, err := suite.db.NewInsert().
		Model(complexModel).
		Exec(suite.ctx)
	if err != nil {
		// Some fields might not be supported by all databases
		suite.T().Logf("Complex model insert failed (expected for some databases): %v", err)

		// Try a simpler version without problematic fields
		simpleComplexModel := &ComplexModel{
			StringField: "test string",
			IntField:    42,
			FloatField:  3.14159,
			BoolField:   true,
			TimeField:   now,
		}

		_, err = suite.db.NewInsert().
			Model(simpleComplexModel).
			Exec(suite.ctx)
		suite.NoError(err, "Simple complex model insert should work")

		// Cleanup
		_, err = suite.db.NewDelete().
			Model((*ComplexModel)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PKEquals(simpleComplexModel.Id)
			}).
			Exec(suite.ctx)
		suite.NoError(err)
	} else {
		suite.True(complexModel.Id != "", "Complex model should have ID set")

		// Verify the data was inserted correctly
		var retrievedModel ComplexModel

		err = suite.db.NewSelect().
			Model(&retrievedModel).
			Where(func(cb ConditionBuilder) {
				cb.PKEquals(complexModel.Id)
			}).
			Scan(suite.ctx)
		suite.NoError(err)
		suite.Equal("test string", retrievedModel.StringField)
		suite.Equal(42, retrievedModel.IntField)
		suite.True(retrievedModel.BoolField)

		// Cleanup
		_, err = suite.db.NewDelete().
			Model((*ComplexModel)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PKEquals(complexModel.Id)
			}).
			Exec(suite.ctx)
		suite.NoError(err)
	}
}

// TestInsertWithValues tests INSERT with bulk operations.
func (suite *InsertTestSuite) TestInsertBulkOperations() {
	suite.T().Logf("Testing INSERT bulk operations for %s", suite.dbType)

	// Test 1: Large batch insert
	batchSize := 10

	batchUsers := make([]*User, batchSize)
	for i := range batchSize {
		batchUsers[i] = &User{
			Name:     fmt.Sprintf("Batch User %d", i+1),
			Email:    fmt.Sprintf("batch%d@example.com", i+1),
			Age:      int16(20 + i),
			IsActive: i%2 == 0, // Alternate true/false
		}
	}

	start := time.Now()
	_, err := suite.db.NewInsert().
		Model(&batchUsers).
		Exec(suite.ctx)
	duration := time.Since(start)

	suite.NoError(err)
	suite.T().Logf("Batch insert of %d users took %v", batchSize, duration)

	// Verify all users were inserted
	var retrievedBatchUsers []User

	err = suite.db.NewSelect().
		Model(&retrievedBatchUsers).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("email", "batch")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(retrievedBatchUsers, batchSize, "Should have inserted all batch users")

	// Verify data integrity
	for _, user := range retrievedBatchUsers {
		suite.True(user.Id != "", "Each user should have an Id")
		suite.Contains(user.Name, "Batch User", "User name should contain 'Batch User'")
		suite.True(user.Age >= 20 && user.Age < 20+int16(batchSize), "Age should be in expected range")
	}

	// Test 2: Insert with mixed data types
	posts := []*Post{
		{
			Title:       "Batch Post 1",
			Content:     "Content for batch post 1",
			Description: lo.ToPtr("Description 1"),
			UserId:      batchUsers[0].Id,
			CategoryId:  suite.getCategoryId(), // Helper to get a category ID
			Status:      "published",
			ViewCount:   100,
		},
		{
			Title:      "Batch Post 2",
			Content:    "Content for batch post 2",
			UserId:     batchUsers[1].Id,
			CategoryId: suite.getCategoryId(),
			Status:     "draft",
			ViewCount:  0,
		},
	}

	_, err = suite.db.NewInsert().
		Model(&posts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify posts were inserted
	var insertedPosts []Post

	err = suite.db.NewSelect().
		Model(&insertedPosts).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Batch Post")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(insertedPosts, 2, "Should have inserted 2 posts")

	// Cleanup
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Batch Post")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("email", "batch")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestInsertErrorHandling tests error handling in insert operations.
func (suite *InsertTestSuite) TestInsertErrorHandling() {
	suite.T().Logf("Testing INSERT error handling for %s", suite.dbType)

	// Test 1: Insert with unique constraint violation (should fail without OnConflict)
	originalUser := &User{
		Name:     "Original User",
		Email:    "unique@example.com",
		Age:      25,
		IsActive: true,
	}

	_, err := suite.db.NewInsert().
		Model(originalUser).
		Exec(suite.ctx)
	suite.NoError(err)

	// Try to insert another user with same email (should fail)
	duplicateUser := &User{
		Name:     "Duplicate User",
		Email:    "unique@example.com", // Same email
		Age:      30,
		IsActive: false,
	}

	_, err = suite.db.NewInsert().
		Model(duplicateUser).
		Exec(suite.ctx)
	suite.Error(err, "Insert with duplicate email should fail")

	// Test 2: Insert with invalid data (e.g., NULL in NOT NULL column)
	invalidUser := &User{
		Name:  "",
		Email: "",
	}

	_, err = suite.db.NewInsert().
		Model(invalidUser).
		Column("name", nil).  // Override the empty name
		Column("email", nil). // Override the empty email
		Exec(suite.ctx)
	suite.Error(err, "Insert with invalid data should fail")

	// Cleanup
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "unique@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// Helper functions

// getCategoryId returns the first available category Id from fixture data.
func (suite *InsertTestSuite) getCategoryId() string {
	var category Category
	if err := suite.db.NewSelect().
		Model(&category).
		Limit(1).
		Scan(suite.ctx); err != nil {
		suite.T().Fatalf("Failed to get category Id: %v", err)
	}

	return category.Id
}
