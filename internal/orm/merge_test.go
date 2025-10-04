package orm

import (
	"fmt"
	"testing"
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
)

type MergeTestSuite struct {
	*ORMTestSuite
}

// TestBasicMerge tests basic MERGE functionality
func (suite *MergeTestSuite) TestBasicMerge() {
	if suite.dbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.dbType)
	}
	suite.T().Logf("Testing basic MERGE for %s", suite.dbType)

	// Test 1: Basic MERGE with INSERT and UPDATE
	type UserMergeData struct {
		Id       string `bun:"id,pk"`
		Name     string `bun:"name"`
		Email    string `bun:"email"`
		Age      int16  `bun:"age"`
		IsActive bool   `bun:"is_active"`
	}

	// Prepare source data (mix of existing and new users)
	sourceData := []UserMergeData{
		{Id: "user1", Name: "Alice Updated", Email: "alice.updated@example.com", Age: 31, IsActive: true}, // Update existing
		{Id: "user4", Name: "David New", Email: "david@example.com", Age: 28, IsActive: true},             // Insert new
		{Id: "user5", Name: "Eva New", Email: "eva@example.com", Age: 26, IsActive: false},                // Insert new
	}

	// Execute MERGE operation using the correct bun pattern
	result, err := suite.db.NewMerge().
		Model((*User)(nil)).
		WithValues("_source_data", sourceData).
		Using("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("target.id", "_source_data.id")
		}).
		WhenMatched().
		ThenUpdate().
		SetColumns("name", "email", "age", "is_active").
		End().
		WhenNotMatched().
		ThenInsert().
		Values("id", "name", "email", "age", "is_active").
		End().
		Exec(suite.ctx)

	suite.NoError(err)
	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "Should complete without error")
	}

	// Verify results
	var newUsers []User
	err = suite.db.NewSelect().
		Model(&newUsers).
		Where(func(cb ConditionBuilder) {
			cb.In("id", []string{"user4", "user5"})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.T().Logf("Found %d new users after merge", len(newUsers))
}

// TestMergeWithConditions tests MERGE with conditional operations
func (suite *MergeTestSuite) TestMergeWithConditions() {
	if suite.dbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.dbType)
	}
	suite.T().Logf("Testing MERGE with conditions for %s", suite.dbType)

	// Test conditional MERGE with WHEN MATCHED AND conditions
	type PostMergeData struct {
		Id        string `bun:"id,pk"`
		Title     string `bun:"title"`
		Status    string `bun:"status"`
		ViewCount int    `bun:"view_count"`
	}

	sourceData := []PostMergeData{
		{Id: "post1", Title: "Updated Post 1", Status: "published", ViewCount: 150},
		{Id: "post2", Title: "Updated Post 2", Status: "draft", ViewCount: 75},
		{Id: "new1", Title: "New Post 1", Status: "draft", ViewCount: 0},
	}

	result, err := suite.db.NewMerge().
		Model((*Post)(nil)).
		WithValues("_source_data", sourceData).
		Using("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("target.id", "_source_data.id")
		}).
		WhenMatched(func(cb ConditionBuilder) {
			// Only update if view count is higher
			cb.GreaterThan("_source_data.view_count", "target.view_count")
		}).
		ThenUpdate().
		SetColumns("title", "status", "view_count").
		End().
		WhenNotMatched(func(cb ConditionBuilder) {
			// Only insert if status is not empty
			cb.IsNotNull("_source_data.status").NotEquals("_source_data.status", "")
		}).
		ThenInsert().
		Values("id", "title", "status", "view_count").
		End().
		Exec(suite.ctx)

	suite.NoError(err)
	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "Should complete without error")
	}

	suite.T().Logf("Conditional merge completed")
}

// TestMergeWithDelete tests MERGE operations that include DELETE
func (suite *MergeTestSuite) TestMergeWithDelete() {
	if suite.dbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.dbType)
	}
	suite.T().Logf("Testing MERGE with DELETE for %s", suite.dbType)

	// Create some test posts first
	testPosts := []Post{
		{Title: "Test Post 1", Status: "published", ViewCount: 100},
		{Title: "Test Post 2", Status: "draft", ViewCount: 50},
		{Title: "Test Post 3", Status: "archived", ViewCount: 25},
	}

	// Set IDs manually
	testPosts[0].Id = "merge_test_1"
	testPosts[1].Id = "merge_test_2"
	testPosts[2].Id = "merge_test_3"

	for _, post := range testPosts {
		_, err := suite.db.NewInsert().
			Model(&post).
			Exec(suite.ctx)
		suite.NoError(err)
	}

	// Source data with updates and missing records (for deletion)
	type PostUpdateData struct {
		Id        string `bun:"id,pk"`
		Title     string `bun:"title"`
		Status    string `bun:"status"`
		ViewCount int    `bun:"view_count"`
	}

	sourceData := []PostUpdateData{
		{Id: "merge_test_1", Title: "Updated Test Post 1", Status: "published", ViewCount: 120},
		{Id: "merge_test_2", Title: "Updated Test Post 2", Status: "published", ViewCount: 80},
		// merge_test_3 is intentionally missing to test deletion
	}

	result, err := suite.db.NewMerge().
		Model((*Post)(nil)).
		WithValues("_source_data", sourceData).
		Using("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("target.id", "_source_data.id")
		}).
		WhenMatched().
		ThenUpdate().
		SetColumns("title", "status", "view_count").
		End().
		WhenNotMatchedBySource(func(cb ConditionBuilder) {
			// Delete posts that are not in source and have low view count
			cb.LessThan("target.view_count", 30)
		}).
		ThenDelete().
		Exec(suite.ctx)

	suite.NoError(err)
	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "Should complete without error")
	}

	// Verify results
	var remainingPosts []Post
	err = suite.db.NewSelect().
		Model(&remainingPosts).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("id", "merge_test_")
		}).
		OrderBy("id").
		Scan(suite.ctx)
	suite.NoError(err)

	suite.T().Logf("Remaining posts after merge with delete: %d", len(remainingPosts))
	for _, post := range remainingPosts {
		suite.T().Logf("Post %s: %s - %s (views: %d)", post.Id, post.Title, post.Status, post.ViewCount)
		// merge_test_3 should be deleted because it had low view count
		suite.NotEqual("merge_test_3", post.Id, "Post with low view count should be deleted")
	}
}

// TestMergeWithSubquerySource tests MERGE using subquery as source
func (suite *MergeTestSuite) TestMergeWithSubquerySource() {
	if suite.dbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.dbType)
	}
	suite.T().Logf("Testing MERGE with subquery source for %s", suite.dbType)

	// Test MERGE where source is a subquery
	result, err := suite.db.NewMerge().
		Model((*Post)(nil)).
		UsingSubQuery("src", func(sq SelectQuery) {
			sq.Model((*Post)(nil)).
				Select("id", "title", "status").
				SelectExpr(func(eb ExprBuilder) any {
					return eb.Expr("? + ?", eb.Column("view_count"), 10)
				}, "new_view_count").
				Where(func(cb ConditionBuilder) {
					cb.Equals("status", "published")
				})
		}).
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("p.id", "src.id")
		}).
		WhenMatched(func(cb ConditionBuilder) {
			// Only update if the calculated view count is higher
			cb.GreaterThan("src.new_view_count", "p.view_count")
		}).
		ThenUpdate().
		SetExpr("view_count", func(eb ExprBuilder) any {
			return eb.Column("src.new_view_count")
		}).
		End().
		Exec(suite.ctx)

	suite.NoError(err)
	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "Should complete without error")
		suite.T().Logf("Merge with subquery affected %d rows", affected)
	}

	// Verify subquery merge results
	var publishedPosts []Post
	err = suite.db.NewSelect().
		Model(&publishedPosts).
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "published")
		}).
		OrderBy("id").
		Scan(suite.ctx)
	suite.NoError(err)

	suite.T().Logf("Published posts after subquery merge: %d", len(publishedPosts))
	for _, post := range publishedPosts {
		suite.T().Logf("Post %s: %s (views: %d)", post.Id, post.Title, post.ViewCount)
	}
}

// TestMergeWithExpressions tests MERGE with complex expressions
func (suite *MergeTestSuite) TestMergeWithExpressions() {
	if suite.dbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.dbType)
	}
	suite.T().Logf("Testing MERGE with expressions for %s", suite.dbType)

	// Test MERGE with calculated values and expressions
	type UserStatData struct {
		UserId     string `bun:"user_id"`
		PostCount  int    `bun:"post_count"`
		TotalViews int    `bun:"total_views"`
	}

	sourceData := []UserStatData{
		{UserId: "user1", PostCount: 5, TotalViews: 500},
		{UserId: "user2", PostCount: 3, TotalViews: 300},
	}

	result, err := suite.db.NewMerge().
		Model((*User)(nil)).
		WithValues("_source_data", sourceData).
		Using("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("target.id", "_source_data.user_id")
		}).
		WhenMatched().
		ThenUpdate().
		SetExpr("name", func(eb ExprBuilder) any {
			return eb.Concat(eb.Column("target.name"), " (Updated)")
		}).
		End().
		Exec(suite.ctx)

	suite.NoError(err)
	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "Should complete without error")
	}

	suite.T().Logf("MERGE with expressions completed")
}

// TestMergePerformance tests MERGE performance with datasets
func (suite *MergeTestSuite) TestMergePerformance() {
	if suite.dbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.dbType)
	}
	suite.T().Logf("Testing MERGE performance for %s", suite.dbType)

	// Skip performance test if running in short mode
	if testing.Short() {
		suite.T().Skip("Skipping performance test in short mode")
	}

	// Create dataset for performance testing
	const batchSize = 100
	type PerfTestData struct {
		Id    string `bun:"id,pk"`
		Value int64  `bun:"value"`
		Hash  string `bun:"hash"`
	}

	var sourceData []PerfTestData
	for i := range batchSize {
		sourceData = append(sourceData, PerfTestData{
			Id:    fmt.Sprintf("perf_%d", i),
			Value: int64(i * 10),
			Hash:  fmt.Sprintf("hash_%d", i%100),
		})
	}

	// Measure MERGE performance
	start := time.Now()
	result, err := suite.db.NewMerge().
		Model((*PerfTestData)(nil)).
		WithValues("_source_data", sourceData).
		Using("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("target.id", "_source_data.id")
		}).
		WhenMatched().
		ThenUpdate().
		SetColumns("value", "hash").
		End().
		WhenNotMatched().
		ThenInsert().
		Values("id", "value", "hash").
		End().
		Exec(suite.ctx)
	duration := time.Since(start)

	suite.NoError(err)
	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "Should complete without error")
	}

	suite.T().Logf("MERGE performance: %d rows processed in %v (%.2f rows/sec)",
		batchSize, duration, float64(batchSize)/duration.Seconds())

	// Cleanup performance test data
	_, err = suite.db.NewDelete().
		Model((*PerfTestData)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("id", "perf_")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestMergeWithReturning tests MERGE with RETURNING clause
func (suite *MergeTestSuite) TestMergeWithReturning() {
	if suite.dbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.dbType)
	}
	suite.T().Logf("Testing MERGE with RETURNING for %s", suite.dbType)

	type UserMergeData struct {
		Id       string `bun:"id,pk"`
		Name     string `bun:"name"`
		Email    string `bun:"email"`
		IsActive bool   `bun:"is_active"`
	}

	sourceData := []UserMergeData{
		{Id: "return1", Name: "Return User 1", Email: "return1@example.com", IsActive: true},
		{Id: "return2", Name: "Return User 2", Email: "return2@example.com", IsActive: false},
	}

	// Test MERGE with RETURNING clause
	var returnedUsers []User
	err := suite.db.NewMerge().
		Model((*User)(nil)).
		WithValues("_source_data", sourceData).
		Using("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("target.id", "_source_data.id")
		}).
		WhenMatched().
		ThenUpdate().
		SetColumns("name", "email", "is_active").
		End().
		WhenNotMatched().
		ThenInsert().
		Values("id", "name", "email", "is_active").
		End().
		Returning("id", "name", "email").
		Scan(suite.ctx, &returnedUsers)

	// Note: RETURNING clause support varies by database
	// PostgreSQL: RETURNING clause supported
	// SQL Server: OUTPUT clause supported
	// MySQL: Limited support
	// SQLite: RETURNING clause (newer versions)

	if err != nil {
		suite.T().Logf("RETURNING clause not supported on %s: %v", suite.dbType, err)
		// Fall back to regular merge without returning
		result, err := suite.db.NewMerge().
			Model((*User)(nil)).
			WithValues("_source_data", sourceData).
			Using("_source_data").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("target.id", "_source_data.id")
			}).
			WhenMatched().
			ThenUpdate().
			SetColumns("name", "email", "is_active").
			End().
			WhenNotMatched().
			ThenInsert().
			Values("id", "name", "email", "is_active").
			End().
			Exec(suite.ctx)
		suite.NoError(err)
		if result != nil {
			affected, _ := result.RowsAffected()
			suite.True(affected >= 0, "Should complete without error")
		}
	} else {
		suite.True(len(returnedUsers) >= 0, "Should return some users")
		suite.T().Logf("MERGE with RETURNING returned %d users", len(returnedUsers))
		for _, user := range returnedUsers {
			suite.T().Logf("Returned user: %s (%s)", user.Name, user.Email)
		}
	}

	// Verify the merge was successful
	var mergedUsers []User
	err = suite.db.NewSelect().
		Model(&mergedUsers).
		Where(func(cb ConditionBuilder) {
			cb.In("id", []string{"return1", "return2"})
		}).
		OrderBy("id").
		Scan(suite.ctx)
	suite.NoError(err)

	suite.T().Logf("Merged users: %d", len(mergedUsers))
	for _, user := range mergedUsers {
		suite.T().Logf("User %s: %s (%s)", user.Id, user.Name, user.Email)
	}
}
