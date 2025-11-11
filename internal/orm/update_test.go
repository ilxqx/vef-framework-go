package orm

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/constants"
)

// UpdateTestSuite tests UPDATE operations including CTE operations, table sources,
// selection methods, filter operations, column updates, flags, ordering, RETURNING clause,
// Apply methods, and execution methods across all databases.
type UpdateTestSuite struct {
	*OrmTestSuite
}

// TestCTE tests Common Table Expression methods (With, WithValues, WithRecursive).
func (suite *UpdateTestSuite) TestCTE() {
	suite.T().Logf("Testing CTE methods for %s", suite.dbType)

	suite.Run("WithBasicCTE", func() {
		// Create CTE of active users, then update posts from those users
		result, err := suite.db.NewUpdate().
			With("active_users", func(query SelectQuery) {
				query.Model((*User)(nil)).
					Select("id").
					Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active")
					})
			}).
			Model((*Post)(nil)).
			Set("updated_by", "cte_test").
			Where(func(cb ConditionBuilder) {
				cb.InSubQuery("user_id", func(subquery SelectQuery) {
					subquery.Table("active_users").Select("id")
				})
			}).
			Exec(suite.ctx)

		suite.NoError(err, "WITH clause should work for updates")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected")
		suite.True(rowsAffected > 0, "Should update posts from active users")

		suite.T().Logf("Updated %d posts using WITH CTE", rowsAffected)
	})

	suite.Run("WithValuesCTE", func() {
		// First, get specific post IDs to update (to avoid affecting other tests)
		var postsToUpdate []Post

		err := suite.db.NewSelect().
			Model(&postsToUpdate).
			Where(func(cb ConditionBuilder) {
				cb.In("status", []string{"published", "draft"})
			}).
			OrderBy("id").
			Limit(2). // Only update 2 posts to minimize impact
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch posts to update")
		suite.True(len(postsToUpdate) >= 2, "Should have at least 2 posts to update")

		postIds := make([]string, len(postsToUpdate))
		for i, p := range postsToUpdate {
			postIds[i] = p.Id
		}

		type StatusMapping struct {
			OldStatus string `bun:"old_status"`
			NewStatus string `bun:"new_status"`
		}

		mappings := []StatusMapping{
			{OldStatus: "published", NewStatus: "archived"},
			{OldStatus: "draft", NewStatus: "deleted"},
		}

		result, err := suite.db.NewUpdate().
			WithValues("status_map", &mappings).
			Model((*Post)(nil)).
			Table("status_map", "sm").
			SetExpr("status", func(eb ExprBuilder) any {
				return eb.Column("sm.new_status")
			}).
			Where(func(cb ConditionBuilder) {
				cb.EqualsColumn("status", "sm.old_status").
					In("id", postIds) // Only update the specific posts we selected
			}).
			Exec(suite.ctx)

		suite.NoError(err, "WITH VALUES should work when supported")

		rowsAffected, _ := result.RowsAffected()
		suite.T().Logf("Updated %d posts using WITH VALUES", rowsAffected)
	})
}

// TestTableSource tests table source methods (Model, ModelTable, Table, TableFrom, TableExpr, TableSubQuery).
func (suite *UpdateTestSuite) TestTableSource() {
	suite.T().Logf("Testing table source methods for %s", suite.dbType)

	suite.Run("ModelBasic", func() {
		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("updated_by", "model_test").
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "Model should set table correctly")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("Updated %d user using Model", rowsAffected)
	})

	suite.Run("ModelTable", func() {
		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			ModelTable("test_user", "u").
			Set("updated_by", "model_table_test").
			Where(func(cb ConditionBuilder) {
				cb.Equals("u.email", "alice@example.com")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "ModelTable should override table name and alias")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("Updated %d user using ModelTable", rowsAffected)
	})

	suite.Run("TableDirect", func() {
		testPosts := []Post{
			{Title: "Table Direct Post", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
		}

		_, err := suite.db.NewInsert().Model(&testPosts).Exec(suite.ctx)
		suite.NoError(err, "Should insert test post")

		result, err := suite.db.NewUpdate().
			Table("test_post", "p").
			Set("status", "published").
			Where(func(cb ConditionBuilder) {
				cb.Equals("p.title", "Table Direct Post")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "Table method should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected")
		suite.Equal(int64(1), rowsAffected, "Should update 1 post using Table")

		suite.T().Logf("Updated %d post using Table method", rowsAffected)
	})

	suite.Run("TableFrom", func() {
		testPosts := []Post{
			{Title: "TableFrom Post", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
		}

		_, err := suite.db.NewInsert().Model(&testPosts).Exec(suite.ctx)
		suite.NoError(err, "Should insert test post")

		result, err := suite.db.NewUpdate().
			TableFrom((*Post)(nil), "p").
			Set("status", "archived").
			Where(func(cb ConditionBuilder) {
				cb.Equals("p.title", "TableFrom Post")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "TableFrom method should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected")
		suite.Equal(int64(1), rowsAffected, "Should update 1 post using TableFrom")

		suite.T().Logf("Updated %d post using TableFrom method", rowsAffected)
	})

	suite.Run("TableExpr", func() {
		// Setup: Create an inactive user and their post
		testUser := &User{
			Name:     "Inactive Author",
			Email:    "inactive.author@example.com",
			Age:      30,
			IsActive: false,
		}
		_, err := suite.db.NewInsert().Model(testUser).Exec(suite.ctx)
		suite.NoError(err, "Should insert test user")

		testPost := &Post{
			Title:      "Post by Inactive User",
			Content:    "This post should be archived",
			UserId:     testUser.Id,
			CategoryId: "cat1",
			Status:     "published",
		}
		_, err = suite.db.NewInsert().Model(testPost).Exec(suite.ctx)
		suite.NoError(err, "Should insert test post")

		var result sql.Result

		result, err = suite.db.NewUpdate().
			Model((*Post)(nil)).
			TableExpr(func(eb ExprBuilder) any {
				return eb.SubQuery(func(sq SelectQuery) {
					sq.Model((*User)(nil))
				})
			}, "u").
			Set("status", "archived").
			Set("updated_by", "multi_table_update").
			Where(func(cb ConditionBuilder) {
				cb.EqualsColumn("user_id", "u.id").
					IsFalse("u.is_active")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "TableExpr for multi-table update should work")

		rowsAffected, _ := result.RowsAffected()
		suite.True(rowsAffected > 0, "Should update posts from inactive users")
		suite.T().Logf("Multi-table update affected %d posts on %s", rowsAffected, suite.dbType)
	})
}

// TestSelectionMethods tests selection methods for controlling which columns are included in UPDATE SET clause.
// These methods (Select, Exclude, SelectAll, ExcludeAll) only work when updating with a model instance,
// as the framework reflects all columns from the model and these methods control which ones to include in SET.
func (suite *UpdateTestSuite) TestSelectionMethods() {
	suite.T().Logf("Testing selection methods for %s", suite.dbType)

	suite.Run("SelectSpecificColumns", func() {
		// First, get the user ID
		var existingUser User

		err := suite.db.NewSelect().
			Model(&existingUser).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch user")

		originalEmail := existingUser.Email
		originalIsActive := existingUser.IsActive
		userId := existingUser.Id

		// Create a NEW user instance (not from database query)
		user := User{
			Model: Model{
				Id: userId,
			},
			Name: "Alice Updated",
			Age:  99,
		}

		result, err := suite.db.NewUpdate().
			Model(&user).
			Select("name", "age").
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "Select should control which columns are updated")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		// Verify only selected columns were updated
		var updatedUser User

		err = suite.db.NewSelect().
			Model(&updatedUser).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(userId)
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch updated user")

		suite.Equal("Alice Updated", updatedUser.Name, "Name should be updated")
		suite.Equal(int16(99), updatedUser.Age, "Age should be updated")
		suite.Equal(originalEmail, updatedUser.Email, "Email should NOT be updated")
		suite.Equal(originalIsActive, updatedUser.IsActive, "IsActive should NOT be updated")

		suite.T().Logf("Select updated only name and age: name=%s, age=%d, email=%s (unchanged), is_active=%v (unchanged)",
			updatedUser.Name, updatedUser.Age, updatedUser.Email, updatedUser.IsActive)
	})

	suite.Run("SelectAllColumns", func() {
		var user User

		err := suite.db.NewSelect().
			Model(&user).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "bob@example.com")
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch user")

		// Modify all fields
		user.Name = "Bob Updated"
		user.Age = 88
		user.Email = "bob.new@example.com"
		user.IsActive = false

		result, err := suite.db.NewUpdate().
			Model(&user).
			SelectAll().
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "SelectAll should update all columns")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		// Verify all columns were updated
		var updatedUser User

		err = suite.db.NewSelect().
			Model(&updatedUser).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch updated user")

		suite.Equal("Bob Updated", updatedUser.Name, "Name should be updated")
		suite.Equal(int16(88), updatedUser.Age, "Age should be updated")
		suite.Equal("bob.new@example.com", updatedUser.Email, "Email should be updated")
		suite.Equal(false, updatedUser.IsActive, "IsActive should be updated")

		suite.T().Logf("SelectAll updated all columns: name=%s, age=%d, email=%s, is_active=%v",
			updatedUser.Name, updatedUser.Age, updatedUser.Email, updatedUser.IsActive)
	})

	suite.Run("ExcludeSpecificColumns", func() {
		var user User

		err := suite.db.NewSelect().
			Model(&user).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "charlie@example.com")
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch user")

		originalEmail := user.Email
		originalAge := user.Age

		// Modify multiple fields but exclude some from update
		user.Name = "Charlie Updated"
		user.Age = 77
		user.Email = "charlie.new@example.com"
		user.IsActive = true

		result, err := suite.db.NewUpdate().
			Model(&user).
			Exclude("email", "age").
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "Exclude should prevent specific columns from being updated")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		// Verify excluded columns were NOT updated
		var updatedUser User

		err = suite.db.NewSelect().
			Model(&updatedUser).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch updated user")

		suite.Equal("Charlie Updated", updatedUser.Name, "Name should be updated")
		suite.Equal(true, updatedUser.IsActive, "IsActive should be updated")
		suite.Equal(originalEmail, updatedUser.Email, "Email should NOT be updated (excluded)")
		suite.Equal(originalAge, updatedUser.Age, "Age should NOT be updated (excluded)")

		suite.T().Logf("Exclude prevented email and age from updating: name=%s (updated), email=%s (unchanged), age=%d (unchanged)",
			updatedUser.Name, updatedUser.Email, updatedUser.Age)
	})

	suite.Run("ExcludeAllColumns", func() {
		var user User

		err := suite.db.NewSelect().
			Model(&user).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch user")

		originalName := user.Name
		originalEmail := user.Email

		// Modify multiple fields
		user.Name = "Alice Updated Again"
		user.Age = 77 // Use a different value than previous tests
		user.Email = "alice.new@example.com"

		// Select name and email first, then ExcludeAll clears them, then Select age
		// Result: only age should be updated
		result, err := suite.db.NewUpdate().
			Model(&user).
			Select("name", "email").
			ExcludeAll().
			Select("age").
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "ExcludeAll should clear previous Select and allow new Select")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		// Verify only age was updated
		var updatedUser User

		err = suite.db.NewSelect().
			Model(&updatedUser).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch updated user")

		suite.Equal(originalName, updatedUser.Name, "Name should NOT be updated (cleared by ExcludeAll)")
		suite.Equal(originalEmail, updatedUser.Email, "Email should NOT be updated (cleared by ExcludeAll)")
		suite.Equal(int16(77), updatedUser.Age, "Age should be updated (selected after ExcludeAll)")

		suite.T().Logf("ExcludeAll cleared previous Select, only age updated: name=%s (unchanged), email=%s (unchanged), age=%d (updated)",
			updatedUser.Name, updatedUser.Email, updatedUser.Age)
	})
}

// TestFilterOperations tests filter methods (Where, WherePk, WhereDeleted, IncludeDeleted).
func (suite *UpdateTestSuite) TestFilterOperations() {
	suite.T().Logf("Testing filter operations for %s", suite.dbType)

	suite.Run("WhereBasic", func() {
		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("updated_by", "where_test").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "Where should work correctly")

		rowsAffected, _ := result.RowsAffected()
		suite.True(rowsAffected > 0, "Should update active users")

		suite.T().Logf("Updated %d users with Where", rowsAffected)
	})

	suite.Run("WherePk", func() {
		var user User

		err := suite.db.NewSelect().
			Model(&user).
			OrderBy("id").
			Limit(1).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch a user")

		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("updated_by", "where_pk_test").
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.ctx)

		suite.NoError(err, "WherePk should work correctly")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("Updated user by PK: %s", user.Id)
	})

	suite.Run("WhereDeletedAndIncludeDeleted", func() {
		type SoftDeleteArticle struct {
			bun.BaseModel `bun:"table:test_update_soft_delete,alias:tusd"`
			Model

			Title     string    `json:"title" bun:"title,notnull"`
			Status    string    `json:"status" bun:"status,notnull"`
			DeletedAt time.Time `json:"deletedAt" bun:",soft_delete,nullzero"`
		}

		bunDB := suite.getBunDb()
		_, err := bunDB.NewDropTable().Model((*SoftDeleteArticle)(nil)).IfExists().Exec(suite.ctx)
		suite.Require().NoError(err, "Should drop existing soft delete table")

		_, err = bunDB.NewCreateTable().Model((*SoftDeleteArticle)(nil)).IfNotExists().Exec(suite.ctx)

		suite.Require().NoError(err, "Should create soft delete table")
		defer func() {
			_, dropErr := bunDB.NewDropTable().Model((*SoftDeleteArticle)(nil)).IfExists().Exec(suite.ctx)
			suite.Require().NoError(dropErr, "Should cleanup soft delete table")
		}()

		records := []*SoftDeleteArticle{
			{Title: "Soft delete target", Status: "draft"},
			{Title: "Active record", Status: "draft"},
		}

		_, err = suite.db.NewInsert().Model(&records).Exec(suite.ctx)
		suite.Require().NoError(err, "Should insert soft delete records")

		_, err = suite.db.NewDelete().
			Model((*SoftDeleteArticle)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("id", records[0].Id)
			}).
			Exec(suite.ctx)
		suite.Require().NoError(err, "Should soft delete first record")

		result, err := suite.db.NewUpdate().
			Model((*SoftDeleteArticle)(nil)).
			Set("status", "archived").
			Where(func(cb ConditionBuilder) {
				cb.Equals("id", records[0].Id)
			}).
			WhereDeleted().
			Exec(suite.ctx)
		suite.Require().NoError(err, "WhereDeleted should target soft-deleted rows")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "WhereDeleted should update exactly one record")

		result, err = suite.db.NewUpdate().
			Model((*SoftDeleteArticle)(nil)).
			Set("status", "reviewed").
			IncludeDeleted().
			Where(func(cb ConditionBuilder) {
				cb.In("id", []string{records[0].Id, records[1].Id})
			}).
			Exec(suite.ctx)
		suite.Require().NoError(err, "IncludeDeleted should allow updating soft-deleted rows")

		rowsAffected, _ = result.RowsAffected()
		suite.Equal(int64(2), rowsAffected, "IncludeDeleted should update both records")

		var fetched []SoftDeleteArticle

		err = suite.db.NewSelect().
			Model(&fetched).
			IncludeDeleted().
			Where(func(cb ConditionBuilder) {
				cb.In("id", []string{records[0].Id, records[1].Id})
			}).
			OrderBy("title").
			Scan(suite.ctx)
		suite.Require().NoError(err, "Should fetch all records including deleted")
		suite.Len(fetched, 2, "Should return two records")

		for _, record := range fetched {
			suite.Equal("reviewed", record.Status, "Status should reflect IncludeDeleted update")

			if record.Id == records[0].Id {
				suite.False(record.DeletedAt.IsZero(), "First record should remain soft deleted")
			} else {
				suite.True(record.DeletedAt.IsZero(), "Second record should remain active")
			}
		}
	})
}

// TestColumnUpdates tests column update methods (Column, ColumnExpr, Set, SetExpr).
func (suite *UpdateTestSuite) TestColumnUpdates() {
	suite.T().Logf("Testing column update methods for %s", suite.dbType)

	var post Post

	err := suite.db.NewSelect().
		Model(&post).
		Where(func(cb ConditionBuilder) {
			cb.Contains("title", "Introduction")
		}).
		Scan(suite.ctx)
	suite.NoError(err, "Should fetch a post")

	suite.Run("ColumnMethod", func() {
		// Column method is used to override model values during update
		// When using Model(&post), Column can override specific field values
		var post Post

		err := suite.db.NewSelect().
			Model(&post).
			Where(func(cb ConditionBuilder) {
				cb.Equals("title", "Database Design Basics")
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch post")

		originalViewCount := post.ViewCount

		// Modify the model
		post.Title = "Updated Title from Model"
		post.ViewCount = 999

		// Use Column to override model's view_count value
		result, err := suite.db.NewUpdate().
			Model(&post).
			Column("view_count", 555). // Override model's 999 with 555
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "Column should override model value")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		// Verify the update
		var updatedPost Post

		err = suite.db.NewSelect().
			Model(&updatedPost).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(post.Id)
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should retrieve updated post")
		suite.Equal("Updated Title from Model", updatedPost.Title, "Title should be from model")
		suite.Equal(int32(555), int32(updatedPost.ViewCount), "View count should be from Column, not model")

		suite.T().Logf("Column overrode model: view_count=%d (model was 999, original was %d)",
			updatedPost.ViewCount, originalViewCount)
	})

	suite.Run("ColumnExprMethod", func() {
		// ColumnExpr method is used to override model values with expressions
		// When using Model(&post), ColumnExpr can override with calculated values
		var post Post

		err := suite.db.NewSelect().
			Model(&post).
			Where(func(cb ConditionBuilder) {
				cb.Equals("title", "Machine Learning Basics")
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch post")

		originalViewCount := post.ViewCount

		// Modify the model
		post.Title = "Updated Title from Model"
		post.ViewCount = 888

		// Use ColumnExpr to override model's view_count with an expression
		result, err := suite.db.NewUpdate().
			Model(&post).
			ColumnExpr("view_count", func(eb ExprBuilder) any {
				// Increment view_count by 100
				return eb.Add(eb.Column("view_count"), 100)
			}).
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "ColumnExpr should override model value with expression")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		// Verify the update
		var updatedPost Post

		err = suite.db.NewSelect().
			Model(&updatedPost).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(post.Id)
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should retrieve updated post")
		suite.Equal("Updated Title from Model", updatedPost.Title, "Title should be from model")
		suite.Equal(int32(originalViewCount+100), int32(updatedPost.ViewCount), "View count should be from ColumnExpr")

		suite.T().Logf("ColumnExpr overrode model: view_count=%d (was %d, model was 888)",
			updatedPost.ViewCount, originalViewCount)
	})

	suite.Run("SetMethod", func() {
		result, err := suite.db.NewUpdate().
			Model((*Post)(nil)).
			Set("updated_by", "set_test").
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(post.Id)
			}).
			Exec(suite.ctx)

		suite.NoError(err, "Set should work as alias for Column")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("Updated using Set method")
	})

	suite.Run("SetExprMethod", func() {
		result, err := suite.db.NewUpdate().
			Model((*Post)(nil)).
			SetExpr("title", func(eb ExprBuilder) any {
				return eb.Concat(eb.Column("title"), " [Updated]")
			}).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(post.Id)
			}).
			Exec(suite.ctx)

		suite.NoError(err, "SetExpr should work as alias for ColumnExpr")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("Updated title using SetExpr")
	})
}

// TestUpdateFlags tests special flags (OmitZero, Bulk).
func (suite *UpdateTestSuite) TestUpdateFlags() {
	suite.T().Logf("Testing update flags for %s", suite.dbType)

	suite.Run("OmitZeroFlag", func() {
		var user User

		err := suite.db.NewSelect().
			Model(&user).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch user")

		// Update with a full User model but only set Name, leave Age as zero
		partialUpdate := &User{
			Model: Model{Id: user.Id},
			Name:  "Alice Updated with OmitZero",
			// Age will be zero, OmitZero should skip it
		}

		result, err := suite.db.NewUpdate().
			Model(partialUpdate).
			OmitZero().
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "OmitZero should skip zero-value fields")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("Updated with OmitZero flag")
	})

	suite.Run("BulkUpdate", func() {
		// Fetch multiple posts to update (use all posts to ensure we have enough)
		var posts []Post

		err := suite.db.NewSelect().
			Model(&posts).
			OrderBy("id").
			Limit(3).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch posts for bulk update")
		suite.True(len(posts) >= 2, "Should have at least 2 posts")

		originalCount := len(posts)
		suite.T().Logf("Fetched %d posts for bulk update", originalCount)

		// Store original view counts for verification
		originalViewCounts := make([]int, len(posts))
		for i := range posts {
			originalViewCounts[i] = posts[i].ViewCount
		}

		// Modify each post with different values
		for i := range posts {
			posts[i].ViewCount = 1000 + i*100 // 1000, 1100, 1200, ...
			posts[i].UpdatedBy = "bulk_test"
		}

		// Perform bulk update using Bulk() method
		result, err := suite.db.NewUpdate().
			Model(&posts).
			Bulk().
			WherePk().
			Exec(suite.ctx)

		suite.NoError(err, "Bulk update should work")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(originalCount), rowsAffected, "Should update all fetched posts")

		// Verify each post was updated with its specific value
		for i, post := range posts {
			var updatedPost Post

			err = suite.db.NewSelect().
				Model(&updatedPost).
				Where(func(cb ConditionBuilder) {
					cb.PkEquals(post.Id)
				}).
				Scan(suite.ctx)
			suite.NoError(err, "Should retrieve updated post")
			suite.Equal(1000+i*100, updatedPost.ViewCount, "View count should match bulk update value")
			suite.Equal("bulk_test", updatedPost.UpdatedBy, "UpdatedBy should be bulk_test")

			suite.T().Logf("Post %d: view_count=%d (was %d, expected %d)",
				i, updatedPost.ViewCount, originalViewCounts[i], 1000+i*100)
		}

		suite.T().Logf("Bulk updated %d posts with different values", originalCount)
	})
}

// TestOrderingAndLimits tests ordering and limit methods (OrderBy, OrderByDesc, OrderByExpr, Limit).
func (suite *UpdateTestSuite) TestOrderingAndLimits() {
	suite.T().Logf("Testing ordering and limits for %s", suite.dbType)

	// Only MySQL supports ORDER BY and LIMIT in UPDATE statements
	if suite.dbType != constants.DbMySQL {
		suite.T().Skipf("Database %s doesn't support ORDER BY and LIMIT in UPDATE statements", suite.dbType)
	}

	suite.Run("OrderByBasic", func() {
		result, err := suite.db.NewUpdate().
			Model((*Post)(nil)).
			Set("updated_by", "order_test").
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(2).
			Exec(suite.ctx)

		suite.NoError(err, "OrderBy should work when supported")

		rowsAffected, _ := result.RowsAffected()
		suite.True(rowsAffected > 0, "Should update posts with ordering")
		suite.T().Logf("Updated %d posts with OrderBy", rowsAffected)
	})

	suite.Run("OrderByDesc", func() {
		_, err := suite.db.NewUpdate().
			Model((*Post)(nil)).
			Set("updated_by", "order_desc_test").
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "draft")
			}).
			OrderByDesc("view_count").
			Limit(1).
			Exec(suite.ctx)

		suite.NoError(err, "OrderByDesc should work when supported")
		suite.T().Logf("Updated with OrderByDesc")
	})

	suite.Run("OrderByExpr", func() {
		_, err := suite.db.NewUpdate().
			Model((*Post)(nil)).
			Set("updated_by", "order_expr_test").
			Where(func(cb ConditionBuilder) {
				cb.NotEquals("status", "deleted")
			}).
			OrderByExpr(func(eb ExprBuilder) any {
				return eb.Column("view_count")
			}).
			Limit(1).
			Exec(suite.ctx)

		suite.NoError(err, "OrderByExpr should work when supported")
		suite.T().Logf("Updated with OrderByExpr")
	})

	suite.Run("LimitOnly", func() {
		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("updated_by", "limit_test").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			Limit(1).
			Exec(suite.ctx)

		suite.NoError(err, "Limit should work when supported")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should limit updates to 1 row")
		suite.T().Logf("Limited update to %d row", rowsAffected)
	})
}

// TestReturningClause tests RETURNING clause methods (Returning, ReturningAll, ReturningNone).
func (suite *UpdateTestSuite) TestReturningClause() {
	suite.T().Logf("Testing RETURNING clause methods for %s", suite.dbType)

	if suite.dbType == constants.DbMySQL {
		suite.T().Skip("MySQL doesn't support RETURNING clause")
	}

	suite.Run("ReturningSpecificColumns", func() {
		type UpdateResult struct {
			Id   string `bun:"id"`
			Name string `bun:"name"`
			Age  int16  `bun:"age"`
		}

		var result UpdateResult

		err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("age", 28).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "bob@example.com")
			}).
			Returning("id", "name", "age").
			Scan(suite.ctx, &result)

		suite.NoError(err, "Returning should work with specific columns")
		suite.NotEmpty(result.Id, "ID should be returned")
		suite.Equal(int16(28), result.Age, "Age should be updated value")

		suite.T().Logf("Returned: id=%s, name=%s, age=%d", result.Id, result.Name, result.Age)
	})

	suite.Run("ReturningAllColumns", func() {
		type UpdateResult struct {
			Id    string `bun:"id"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
			Age   int16  `bun:"age"`
		}

		var result UpdateResult

		err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("age", 29).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "charlie@example.com")
			}).
			ReturningAll().
			Scan(suite.ctx, &result)

		suite.NoError(err, "ReturningAll should return all columns")
		suite.NotEmpty(result.Id, "ID should be returned")
		suite.NotEmpty(result.Email, "Email should be returned")

		suite.T().Logf("ReturningAll: id=%s, email=%s", result.Id, result.Email)
	})

	suite.Run("ReturningNone", func() {
		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("age", 30).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			ReturningNone().
			Exec(suite.ctx)

		suite.NoError(err, "ReturningNone should return no columns")

		rowsAffected, _ := result.RowsAffected()
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("ReturningNone: updated %d row", rowsAffected)
	})
}

// TestApplyMethods tests Apply and ApplyIf methods.
func (suite *UpdateTestSuite) TestApplyMethods() {
	suite.T().Logf("Testing Apply methods for %s", suite.dbType)

	suite.Run("ApplyBasic", func() {
		applyActive := func(query UpdateQuery) {
			query.Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			})
		}

		applyUpdatedBy := func(query UpdateQuery) {
			query.Set("updated_by", "apply_test")
		}

		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("age", 35).
			Apply(applyActive, applyUpdatedBy).
			Exec(suite.ctx)

		suite.NoError(err, "Apply should work correctly")

		rowsAffected, _ := result.RowsAffected()
		suite.True(rowsAffected > 0, "Should update active users")

		suite.T().Logf("Applied functions updated %d users", rowsAffected)
	})

	suite.Run("ApplyIfTrue", func() {
		addFilter := func(query UpdateQuery) {
			query.Where(func(cb ConditionBuilder) {
				cb.GreaterThan("age", 25)
			})
		}

		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("updated_by", "apply_if_true").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			ApplyIf(true, addFilter).
			Exec(suite.ctx)

		suite.NoError(err, "ApplyIf(true) should apply function")

		rowsAffected, _ := result.RowsAffected()
		suite.True(rowsAffected >= 0, "Should execute with condition")

		suite.T().Logf("ApplyIf(true) updated %d users", rowsAffected)
	})

	suite.Run("ApplyIfFalse", func() {
		addFilter := func(query UpdateQuery) {
			query.Where(func(cb ConditionBuilder) {
				cb.LessThan("age", 18)
			})
		}

		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("updated_by", "apply_if_false").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			ApplyIf(false, addFilter).
			Exec(suite.ctx)

		suite.NoError(err, "ApplyIf(false) should skip function")

		rowsAffected, _ := result.RowsAffected()
		suite.True(rowsAffected > 0, "Should update without applied filter")

		suite.T().Logf("ApplyIf(false) updated %d users", rowsAffected)
	})
}

// TestExecution tests execution methods (Exec, Scan).
func (suite *UpdateTestSuite) TestExecution() {
	suite.T().Logf("Testing execution methods for %s", suite.dbType)

	suite.Run("ExecBasic", func() {
		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("age", 40).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "Exec should execute update successfully")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected")
		suite.Equal(int64(1), rowsAffected, "Should affect 1 row")

		suite.T().Logf("Exec updated %d row", rowsAffected)
	})

	suite.Run("ScanWithReturning", func() {
		if suite.dbType == constants.DbMySQL {
			suite.T().Skip("MySQL doesn't support RETURNING clause")
		}

		type UpdateResult struct {
			Id   string `bun:"id"`
			Name string `bun:"name"`
		}

		var result UpdateResult

		err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("age", 41).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "bob@example.com")
			}).
			Returning("id", "name").
			Scan(suite.ctx, &result)

		suite.NoError(err, "Scan should work with RETURNING")
		suite.NotEmpty(result.Id, "ID should be scanned")
		suite.NotEmpty(result.Name, "Name should be scanned")

		suite.T().Logf("Scanned result: id=%s, name=%s", result.Id, result.Name)
	})

	suite.Run("ExecNoMatchingRows", func() {
		result, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("age", 999).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "nonexistent@example.com")
			}).
			Exec(suite.ctx)

		suite.NoError(err, "Exec with no matching rows should not error")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected")
		suite.Equal(int64(0), rowsAffected, "Should affect 0 rows")

		suite.T().Logf("Exec with no matches: %d rows affected", rowsAffected)
	})

	suite.Run("ExecInvalidField", func() {
		_, err := suite.db.NewUpdate().
			Model((*User)(nil)).
			Set("nonexistent_field", "value").
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Exec(suite.ctx)

		suite.Error(err, "Exec with invalid field should error")

		suite.T().Logf("Exec with invalid field correctly returned error")
	})
}
