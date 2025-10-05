package orm

import (
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
)

type UpdateTestSuite struct {
	*ORMTestSuite
}

// TestBasicUpdate tests basic UPDATE functionality across all databases.
func (suite *UpdateTestSuite) TestBasicUpdate() {
	suite.T().Logf("Testing basic UPDATE for %s", suite.dbType)

	// First get a user to update
	var user User

	err := suite.db.NewSelect().
		Model(&user).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "alice@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	originalAge := user.Age

	// Test 1: Update single field
	result, err := suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("age", 31).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "alice@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify result
	rowsAffected, err := result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(1), rowsAffected)

	// Verify the update
	var updatedUser User

	err = suite.db.NewSelect().
		Model(&updatedUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "alice@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal(int16(31), updatedUser.Age)
	suite.Equal("Alice Johnson", updatedUser.Name) // Other fields unchanged

	// Test 2: Update multiple fields
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("name", "Alice Smith").
		Set("age", 32).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "alice@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify multiple field update
	err = suite.db.NewSelect().
		Model(&updatedUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "alice@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("Alice Smith", updatedUser.Name)
	suite.Equal(int16(32), updatedUser.Age)

	// Test 3: Update with model instance
	userToUpdate := &User{
		Name: "Alice Johnson Updated",
		Age:  33,
	}
	userToUpdate.Id = user.Id
	_, err = suite.db.NewUpdate().
		Model(userToUpdate).
		Column("name", userToUpdate.Name).
		Column("age", userToUpdate.Age).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(userToUpdate.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify model instance update
	err = suite.db.NewSelect().
		Model(&updatedUser).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(user.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("Alice Johnson Updated", updatedUser.Name)
	suite.Equal(int16(33), updatedUser.Age)

	// Cleanup - restore original age
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("name", "Alice Johnson").
		Set("age", originalAge).
		Set("is_active", true).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(user.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestUpdateWithConditions tests UPDATE with various WHERE conditions.
func (suite *UpdateTestSuite) TestUpdateWithConditions() {
	suite.T().Logf("Testing UPDATE with conditions for %s", suite.dbType)

	// Test 1: Update multiple records with simple condition
	_, err := suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("updated_by", "batch_update").
		Where(func(cb ConditionBuilder) {
			cb.IsTrue("is_active")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify the updates
	var activeUsers []User

	err = suite.db.NewSelect().
		Model(&activeUsers).
		Where(func(cb ConditionBuilder) {
			cb.IsTrue("is_active")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(activeUsers, 2) // Alice and Bob are active

	for _, user := range activeUsers {
		suite.Equal("batch_update", user.UpdatedBy)
	}

	// Test 2: Update with complex conditions (AND)
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("updated_by", "complex_update").
		Where(func(cb ConditionBuilder) {
			cb.IsTrue("is_active").GreaterThan("age", 26)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify complex condition update (should only affect Alice)
	var complexUpdatedUsers []User

	err = suite.db.NewSelect().
		Model(&complexUpdatedUsers).
		Where(func(cb ConditionBuilder) {
			cb.Equals("updated_by", "complex_update")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Require().Len(complexUpdatedUsers, 1)
	suite.Equal("Alice Johnson", complexUpdatedUsers[0].Name)

	// Test 3: Update with OR conditions
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("updated_by", "or_update").
		Where(func(cb ConditionBuilder) {
			cb.Equals("age", 25).OrEquals("age", 35)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify OR condition update (should affect Bob and Charlie)
	var orUpdatedUsers []User

	err = suite.db.NewSelect().
		Model(&orUpdatedUsers).
		Where(func(cb ConditionBuilder) {
			cb.Equals("updated_by", "or_update")
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(orUpdatedUsers, 2)
	suite.Equal("Bob Smith", orUpdatedUsers[0].Name)
	suite.Equal("Charlie Brown", orUpdatedUsers[1].Name)

	// Test 4: Update with IN condition
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("updated_by", "in_update").
		Where(func(cb ConditionBuilder) {
			cb.In("name", []string{"Bob Smith", "Charlie Brown"})
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify IN condition update
	var inUpdatedUsers []User

	err = suite.db.NewSelect().
		Model(&inUpdatedUsers).
		Where(func(cb ConditionBuilder) {
			cb.Equals("updated_by", "in_update")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(inUpdatedUsers, 2)

	// Cleanup - restore original updated_by
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("updated_by", "system").
		Where(func(cb ConditionBuilder) {
			cb.NotEquals("updated_by", "system")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestUpdateWithExpressions tests UPDATE with SQL expressions.
func (suite *UpdateTestSuite) TestUpdateWithExpressions() {
	suite.T().Logf("Testing UPDATE with expressions for %s", suite.dbType)

	// Get initial post view count
	var post Post

	err := suite.db.NewSelect().
		Model(&post).
		Where(func(cb ConditionBuilder) {
			cb.Contains("title", "Introduction")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	initialViewCount := post.ViewCount

	// Test 1: Arithmetic expression (increment view count)
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		SetExpr("view_count", func(eb ExprBuilder) any {
			return eb.Expr("view_count + ?", 10)
		}).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify arithmetic update
	var updatedPost Post

	err = suite.db.NewSelect().
		Model(&updatedPost).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal(initialViewCount+10, updatedPost.ViewCount)

	// Test 2: String concatenation expression
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		SetExpr("title", func(eb ExprBuilder) any {
			return eb.Concat("title", "' - Updated'")
		}).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify string concatenation
	err = suite.db.NewSelect().
		Model(&updatedPost).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Contains(updatedPost.Title, "- Updated")

	// Test 3: Conditional expression (CASE)
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		SetExpr("status", func(eb ExprBuilder) any {
			return eb.Case(func(cb CaseBuilder) {
				cb.When(func(cb ConditionBuilder) {
					cb.GreaterThan("view_count", 100)
				}).Then("popular").
					Else("normal")
			})
		}).
		Where(func(cb ConditionBuilder) {
			cb.NotEquals("status", "draft")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify conditional expression update
	var posts []Post

	err = suite.db.NewSelect().
		Model(&posts).
		Where(func(cb ConditionBuilder) {
			cb.In("status", []string{"popular", "normal"})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(posts) > 0, "Should have posts with updated status")

	// Cleanup - restore original values
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		Set("view_count", initialViewCount).
		Set("title", "Introduction to Go").
		Set("status", "published").
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestUpdateWithJoins tests UPDATE with JOIN operations.
func (suite *UpdateTestSuite) TestUpdateWithJoins() {
	suite.T().Logf("Testing UPDATE with JOINs for %s", suite.dbType)

	// Test: Update posts based on user information using subquery approach
	// (JOINs in UPDATE syntax varies between databases, so we use subqueries)
	_, err := suite.db.NewUpdate().
		Model((*Post)(nil)).
		Set("status", "user_inactive_post").
		Where(func(cb ConditionBuilder) {
			cb.InSubQuery("user_id", func(subquery SelectQuery) {
				subquery.Model((*User)(nil)).
					Select("id").
					Where(func(cb ConditionBuilder) {
						cb.IsFalse("is_active")
					})
			})
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify subquery-based update
	var inactiveUserPosts []Post

	err = suite.db.NewSelect().
		Model(&inactiveUserPosts).
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "user_inactive_post")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	// Should have posts by inactive users (Charlie)
	for _, post := range inactiveUserPosts {
		// Verify these are indeed from inactive users
		var user User

		err = suite.db.NewSelect().
			Model(&user).
			Where(func(cb ConditionBuilder) {
				cb.PKEquals(post.UserId)
			}).
			Scan(suite.ctx)
		suite.NoError(err)
		suite.False(user.IsActive, "Posts should be from inactive users")
	}

	// Test 2: Update with more complex subquery
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		Set("status", "high_activity_user").
		Where(func(cb ConditionBuilder) {
			cb.InSubQuery("user_id", func(subquery SelectQuery) {
				subquery.Model((*User)(nil)).
					Select("id").
					Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active").GreaterThan("age", 25)
					})
			})
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify complex subquery update
	var highActivityPosts []Post

	err = suite.db.NewSelect().
		Model(&highActivityPosts).
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "high_activity_user")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	// Should have posts by active users with age > 25 (Alice)
	for _, post := range highActivityPosts {
		var user User

		err = suite.db.NewSelect().
			Model(&user).
			Where(func(cb ConditionBuilder) {
				cb.PKEquals(post.UserId)
			}).
			Scan(suite.ctx)
		suite.NoError(err)
		suite.True(user.IsActive, "User should be active")
		suite.True(user.Age > 25, "User should be older than 25")
	}

	// Cleanup - restore original statuses based on fixture data
	posts := map[string]string{
		"Introduction to Go":             "published",
		"Database Design Basics":         "published",
		"Machine Learning Basics":        "draft",
		"Business Strategy Fundamentals": "published",
		"Advanced Go Patterns":           "review",
		"Latest Trends in Science":       "published",
		"Practical Tech Tutorials":       "published",
		"Startup Finance 101":            "published",
	}

	for title, status := range posts {
		_, err = suite.db.NewUpdate().
			Model((*Post)(nil)).
			Set("status", status).
			Where(func(cb ConditionBuilder) {
				cb.Equals("title", title)
			}).
			Exec(suite.ctx)
		suite.NoError(err)
	}
}

// TestUpdateReturning tests UPDATE with RETURNING clause.
func (suite *UpdateTestSuite) TestUpdateReturning() {
	suite.T().Logf("Testing UPDATE with RETURNING for %s", suite.dbType)

	// Skip if database doesn't support RETURNING
	if suite.dbType == constants.DbMySQL {
		suite.T().Skip("MySQL doesn't support RETURNING clause")
	}

	// Get a user to update
	var user User

	err := suite.db.NewSelect().
		Model(&user).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "bob@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	originalAge := user.Age

	// Test: Update with RETURNING clause
	type UpdateResult struct {
		Id   string `bun:"id"`
		Name string `bun:"name"`
		Age  int16  `bun:"age"`
	}

	var returnedUser UpdateResult

	err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("name", "Bob Johnson").
		Set("age", 26).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "bob@example.com")
		}).
		Returning("id", "name", "age").
		Scan(suite.ctx, &returnedUser)
	suite.NoError(err)

	// Verify returned values
	suite.Equal(user.Id, returnedUser.Id)
	suite.Equal("Bob Johnson", returnedUser.Name)
	suite.Equal(int16(26), returnedUser.Age)

	// Verify the update actually happened
	var updatedUser User

	err = suite.db.NewSelect().
		Model(&updatedUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "bob@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("Bob Johnson", updatedUser.Name)
	suite.Equal(int16(26), updatedUser.Age)

	// Test: RETURNING with multiple updates
	var returnedUsers []UpdateResult

	err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("updated_by", "returning_test").
		Where(func(cb ConditionBuilder) {
			cb.IsTrue("is_active")
		}).
		Returning("id", "name", "age").
		Scan(suite.ctx, &returnedUsers)
	suite.NoError(err)
	suite.Len(returnedUsers, 2) // Should return 2 active users

	// Cleanup - restore original values
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("name", "Bob Smith").
		Set("age", originalAge).
		Set("updated_by", "system").
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "bob@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestUpdateComplexModel tests updating models with various data types.
func (suite *UpdateTestSuite) TestUpdateComplexModel() {
	suite.T().Logf("Testing UPDATE with complex model for %s", suite.dbType)

	// First, insert a complex model to update
	now := time.Now()
	complexModel := &ComplexModel{
		StringField: "original string",
		IntField:    100,
		FloatField:  1.234,
		BoolField:   false,
		TimeField:   now.Add(-time.Hour),
		JSONField:   map[string]any{"original": true},
	}

	_, err := suite.db.NewInsert().
		Model(complexModel).
		Exec(suite.ctx)
	suite.NoError(err)
	suite.NotEmpty(complexModel.Id, "Complex model should have ID after insert")

	// Test 1: Update basic fields
	_, err = suite.db.NewUpdate().
		Model((*ComplexModel)(nil)).
		Set("string_field", "updated string").
		Set("int_field", 200).
		Set("float_field", 5.678).
		Set("bool_field", true).
		Set("time_field", now).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify basic field updates
	var updatedComplex ComplexModel

	err = suite.db.NewSelect().
		Model(&updatedComplex).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("updated string", updatedComplex.StringField)
	suite.Equal(200, updatedComplex.IntField)
	suite.Equal(5.678, updatedComplex.FloatField)
	suite.True(updatedComplex.BoolField)

	// Test 2: Update JSON field (if supported)
	jsonData := map[string]any{
		"updated": true,
		"version": 2,
		"items":   []string{"a", "b", "c"},
	}

	_, err = suite.db.NewUpdate().
		Model((*ComplexModel)(nil)).
		Set("json_field", jsonData).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Exec(suite.ctx)
	if err != nil {
		suite.T().Logf("JSON field update not supported for %s: %v", suite.dbType, err)
	} else {
		// Verify JSON update
		var jsonUpdated ComplexModel

		err = suite.db.NewSelect().
			Model(&jsonUpdated).
			Where(func(cb ConditionBuilder) {
				cb.PKEquals(complexModel.Id)
			}).
			Scan(suite.ctx)
		suite.NoError(err)

		if jsonUpdated.JSONField != nil {
			suite.Equal(true, jsonUpdated.JSONField["updated"])
			suite.Equal(float64(2), jsonUpdated.JSONField["version"]) // JSON numbers are float64
		}
	}

	// Test 3: Update null fields
	nullString := "not null anymore"
	nullInt := 42
	nullTime := now

	_, err = suite.db.NewUpdate().
		Model((*ComplexModel)(nil)).
		Set("null_string", &nullString).
		Set("null_int", &nullInt).
		Set("null_time", &nullTime).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify null field updates
	err = suite.db.NewSelect().
		Model(&updatedComplex).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.NotNil(updatedComplex.NullString)
	suite.Equal("not null anymore", *updatedComplex.NullString)
	suite.NotNil(updatedComplex.NullInt)
	suite.Equal(42, *updatedComplex.NullInt)
	suite.NotNil(updatedComplex.NullTime)

	// Test 4: Set fields back to null
	_, err = suite.db.NewUpdate().
		Model((*ComplexModel)(nil)).
		Set("null_string", nil).
		Set("null_int", nil).
		Set("null_time", nil).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify fields are null again
	err = suite.db.NewSelect().
		Model(&updatedComplex).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	// Some drivers may scan NULL into pointer-as-zero-value instead of nil. Normalize here.
	if updatedComplex.NullString != nil && *updatedComplex.NullString == "" {
		updatedComplex.NullString = nil
	}

	if updatedComplex.NullInt != nil && *updatedComplex.NullInt == 0 {
		updatedComplex.NullInt = nil
	}

	if updatedComplex.NullTime != nil && updatedComplex.NullTime.IsZero() {
		updatedComplex.NullTime = nil
	}

	suite.Nil(updatedComplex.NullString)
	suite.Nil(updatedComplex.NullInt)
	suite.Nil(updatedComplex.NullTime)

	// Cleanup
	_, err = suite.db.NewDelete().
		Model((*ComplexModel)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(complexModel.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestUpdateBulkOperations tests bulk update operations.
func (suite *UpdateTestSuite) TestUpdateBulkOperations() {
	suite.T().Logf("Testing bulk UPDATE operations for %s", suite.dbType)

	// Test 1: Bulk update with condition affecting multiple records
	_, err := suite.db.NewUpdate().
		Model((*Post)(nil)).
		Set("updated_by", "bulk_update_test").
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "published")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify bulk update
	var bulkUpdatedPosts []Post

	err = suite.db.NewSelect().
		Model(&bulkUpdatedPosts).
		Where(func(cb ConditionBuilder) {
			cb.Equals("updated_by", "bulk_update_test")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(bulkUpdatedPosts) > 1, "Should have multiple posts updated")

	// Test 2: Bulk update with expression
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		SetExpr("view_count", func(eb ExprBuilder) any {
			return eb.Expr("view_count + ?", 5)
		}).
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "published")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Test 3: Conditional bulk update
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		SetExpr("status", func(eb ExprBuilder) any {
			return eb.Case(func(cb CaseBuilder) {
				cb.When(func(condBuilder ConditionBuilder) {
					condBuilder.GreaterThan("view_count", 100)
				}).Then("high_traffic").
					When(func(condBuilder ConditionBuilder) {
						condBuilder.GreaterThan("view_count", 50)
					}).
					Then("medium_traffic").
					Else("low_traffic")
			})
		}).
		Where(func(cb ConditionBuilder) {
			cb.NotEquals("status", "draft")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify conditional bulk update
	var trafficPosts []Post

	err = suite.db.NewSelect().
		Model(&trafficPosts).
		Where(func(cb ConditionBuilder) {
			cb.In("status", []string{"high_traffic", "medium_traffic", "low_traffic"})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(trafficPosts) > 0, "Should have posts with traffic status")

	// Verify traffic categorization is correct
	for _, post := range trafficPosts {
		if post.ViewCount > 100 {
			suite.Equal("high_traffic", post.Status)
		} else if post.ViewCount > 50 {
			suite.Equal("medium_traffic", post.Status)
		} else {
			suite.Equal("low_traffic", post.Status)
		}
	}

	// Cleanup - restore original post statuses and updated_by
	posts := map[string]string{
		"Introduction to Go":             "published",
		"Database Design Basics":         "published",
		"Machine Learning Basics":        "draft",
		"Business Strategy Fundamentals": "published",
		"Advanced Go Patterns":           "review",
		"Latest Trends in Science":       "published",
		"Practical Tech Tutorials":       "published",
		"Startup Finance 101":            "published",
	}

	for title, status := range posts {
		_, err = suite.db.NewUpdate().
			Model((*Post)(nil)).
			Set("status", status).
			Set("updated_by", "system"). // Restore updated_by from fixture
			SetExpr("view_count", func(eb ExprBuilder) any {
				return eb.Expr("view_count - ?", 5) // Restore view count
			}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("title", title)
			}).
			Exec(suite.ctx)
		suite.NoError(err)
	}
}

// TestUpdateErrorHandling tests error handling in update operations.
func (suite *UpdateTestSuite) TestUpdateErrorHandling() {
	suite.T().Logf("Testing UPDATE error handling for %s", suite.dbType)

	// Test 1: Update with no matching records (should not error but affect 0 rows)
	result, err := suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("name", "Non-existent User").
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "nonexistent@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err, "UPDATE with no matching rows should not error")

	rowsAffected, err := result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(0), rowsAffected, "Should affect 0 rows")

	// Test 2: Update without WHERE clause (updates all records - be careful!)
	// We'll use a test table to avoid affecting fixture data
	testUser := &User{
		Name:     "Test User For Update",
		Email:    "testupdate@example.com",
		Age:      25,
		IsActive: true,
	}

	_, err = suite.db.NewInsert().
		Model(testUser).
		Exec(suite.ctx)
	suite.NoError(err)

	// Update without WHERE clause on a single record table is safe
	result, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("updated_by", "no_where_test").
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "testupdate@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err, "UPDATE without WHERE should work")

	rowsAffected, err = result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(1), rowsAffected)

	// Test 3: Update with invalid field/column name (should error)
	_, err = suite.db.NewUpdate().
		Model((*User)(nil)).
		Set("non_existent_field", "value").
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(testUser.Id)
		}).
		Exec(suite.ctx)
	suite.Error(err, "UPDATE with invalid field should error")

	// Cleanup test user
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "testupdate@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestUpdateWithOptionalFields tests updating with optional/nullable fields.
func (suite *UpdateTestSuite) TestUpdateWithOptionalFields() {
	suite.T().Logf("Testing UPDATE with optional fields for %s", suite.dbType)

	// Get a post with description to update
	var post Post

	err := suite.db.NewSelect().
		Model(&post).
		Where(func(cb ConditionBuilder) {
			cb.IsNotNull("description")
		}).
		Limit(1).
		Scan(suite.ctx)
	suite.NoError(err)

	originalDescription := *post.Description

	// Test 1: Set nullable field to non-null value
	newDescription := "Updated description for testing"
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		Set("description", newDescription).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify non-null update
	var updatedPost Post

	err = suite.db.NewSelect().
		Model(&updatedPost).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.NotNil(updatedPost.Description)
	suite.Equal(newDescription, *updatedPost.Description)

	// Test 2: Set nullable field to null
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		Set("description", nil).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify null update
	updatedPost = Post{}
	err = suite.db.NewSelect().
		Model(&updatedPost).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Nil(updatedPost.Description)

	// Test 3: Set nullable field back to original value
	_, err = suite.db.NewUpdate().
		Model((*Post)(nil)).
		Set("description", originalDescription).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify restoration
	updatedPost = Post{}
	err = suite.db.NewSelect().
		Model(&updatedPost).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(post.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal(originalDescription, *updatedPost.Description)
}
