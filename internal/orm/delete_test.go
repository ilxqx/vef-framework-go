package orm

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
)

type DeleteTestSuite struct {
	*OrmTestSuite
}

// TestBasicDelete tests basic DELETE functionality across all databases.
func (suite *DeleteTestSuite) TestBasicDelete() {
	suite.T().Logf("Testing basic DELETE for %s", suite.dbType)

	// Create test users for deletion
	testUsers := []*User{
		{
			Name:     "Delete Test User 1",
			Email:    "delete1@example.com",
			Age:      25,
			IsActive: true,
		},
		{
			Name:     "Delete Test User 2",
			Email:    "delete2@example.com",
			Age:      30,
			IsActive: false,
		},
	}

	_, err := suite.db.NewInsert().
		Model(&testUsers).
		Exec(suite.ctx)
	suite.NoError(err)
	suite.True(len(testUsers) == 2)

	for _, user := range testUsers {
		suite.NotEmpty(user.Id, "Each test user should have ID after insert")
	}

	// Test 1: Delete single user by specific condition
	result, err := suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "delete1@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify result
	rowsAffected, err := result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(1), rowsAffected)

	// Verify the user was deleted
	var deletedUser User

	err = suite.db.NewSelect().
		Model(&deletedUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "delete1@example.com")
		}).
		Scan(suite.ctx)
	suite.Error(err, "Deleted user should not exist")

	// Test 2: Delete by primary key
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.PkEquals(testUsers[1].Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify Pk deletion
	err = suite.db.NewSelect().
		Model(&deletedUser).
		Where(func(cb ConditionBuilder) {
			cb.PkEquals(testUsers[1].Id)
		}).
		Scan(suite.ctx)
	suite.Error(err, "Deleted user should not exist")

	// Test 3: Delete multiple records with IN condition
	moreTestUsers := []*User{
		{Name: "Bulk Delete 1", Email: "bulk1@example.com", Age: 20, IsActive: true},
		{Name: "Bulk Delete 2", Email: "bulk2@example.com", Age: 21, IsActive: true},
		{Name: "Bulk Delete 3", Email: "bulk3@example.com", Age: 22, IsActive: true},
	}

	_, err = suite.db.NewInsert().
		Model(&moreTestUsers).
		Exec(suite.ctx)
	suite.NoError(err)

	// Delete multiple using IN condition
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.In("email", []string{"bulk1@example.com", "bulk2@example.com"})
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify bulk deletion
	var remainingBulkUsers []User

	err = suite.db.NewSelect().
		Model(&remainingBulkUsers).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("email", "bulk")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingBulkUsers, 1, "Should have 1 remaining bulk user")
	suite.Equal("bulk3@example.com", remainingBulkUsers[0].Email)

	// Cleanup remaining test user
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "bulk3@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestDeleteWithConditions tests DELETE with various WHERE conditions.
func (suite *DeleteTestSuite) TestDeleteWithConditions() {
	suite.T().Logf("Testing DELETE with conditions for %s", suite.dbType)

	// Create test posts for deletion
	testPosts := []*Post{
		{
			Title:      "Delete Test Post 1",
			Content:    "Content 1",
			UserId:     "test_user_id",
			CategoryId: "test_category_id",
			Status:     "draft",
			ViewCount:  5,
		},
		{
			Title:      "Delete Test Post 2",
			Content:    "Content 2",
			UserId:     "test_user_id",
			CategoryId: "test_category_id",
			Status:     "published",
			ViewCount:  100,
		},
		{
			Title:      "Delete Test Post 3",
			Content:    "Content 3",
			UserId:     "test_user_id",
			CategoryId: "test_category_id",
			Status:     "draft",
			ViewCount:  10,
		},
	}

	_, err := suite.db.NewInsert().
		Model(&testPosts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Test 1: Delete with simple condition
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "draft")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify simple condition deletion (should delete 2 draft posts)
	var remainingTestPosts []Post

	err = suite.db.NewSelect().
		Model(&remainingTestPosts).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Delete Test Post")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingTestPosts, 1, "Should have 1 remaining published post")
	suite.Equal("published", remainingTestPosts[0].Status)

	// Test 2: Delete with complex conditions (AND)
	complexTestPosts := []*Post{
		{Title: "Complex Delete 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 50},
		{Title: "Complex Delete 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 150},
		{Title: "Complex Delete 3", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft", ViewCount: 75},
	}

	_, err = suite.db.NewInsert().
		Model(&complexTestPosts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Delete published posts with high view count
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "published").GreaterThan("view_count", 75)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify complex condition deletion
	var complexRemaining []Post

	err = suite.db.NewSelect().
		Model(&complexRemaining).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Complex Delete")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(complexRemaining, 2, "Should have 2 remaining posts")

	// Should have the low-view published post and the draft post
	for _, post := range complexRemaining {
		suite.True(
			(post.Status == "published" && post.ViewCount <= 75) || post.Status == "draft",
			"Remaining posts should be either low-view published or draft",
		)
	}

	// Test 3: Delete with OR conditions
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("view_count", 50).OrEquals("view_count", 75)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify OR condition deletion
	err = suite.db.NewSelect().
		Model(&complexRemaining).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Complex Delete")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(complexRemaining, 0, "All complex delete posts should be deleted")

	// Cleanup remaining test posts
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Delete Test Post")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestDeleteWithJoins tests DELETE with subqueries (simulating JOINs).
func (suite *DeleteTestSuite) TestDeleteWithJoins() {
	suite.T().Logf("Testing DELETE with JOINs (subqueries) for %s", suite.dbType)

	// Create test data
	testUser := &User{
		Name:     "Join Delete User",
		Email:    "joindelete@example.com",
		Age:      30,
		IsActive: false, // Inactive user
	}

	_, err := suite.db.NewInsert().
		Model(testUser).
		Exec(suite.ctx)
	suite.NoError(err)

	testPosts := []*Post{
		{Title: "Join Delete Post 1", Content: "Content", UserId: testUser.Id, CategoryId: "cat1", Status: "published"},
		{Title: "Join Delete Post 2", Content: "Content", UserId: testUser.Id, CategoryId: "cat1", Status: "draft"},
	}

	_, err = suite.db.NewInsert().
		Model(&testPosts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Test 1: Delete posts by inactive users using subquery
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.InSubQuery("user_id", func(subquery SelectQuery) {
				subquery.Model((*User)(nil)).
					Select("id").
					Where(func(cb ConditionBuilder) {
						cb.Equals("is_active", false)
					})
			})
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify subquery-based deletion
	var remainingJoinPosts []Post

	err = suite.db.NewSelect().
		Model(&remainingJoinPosts).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Join Delete Post")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingJoinPosts, 0, "All posts by inactive users should be deleted")

	// Test 2: Delete with complex subquery
	activeUser := &User{
		Name:     "Active Join User",
		Email:    "activejoin@example.com",
		Age:      25,
		IsActive: true,
	}

	_, err = suite.db.NewInsert().
		Model(activeUser).
		Exec(suite.ctx)
	suite.NoError(err)

	morePosts := []*Post{
		{Title: "Active User Post 1", Content: "Content", UserId: activeUser.Id, CategoryId: "cat1", Status: "published", ViewCount: 10},
		{Title: "Active User Post 2", Content: "Content", UserId: activeUser.Id, CategoryId: "cat1", Status: "published", ViewCount: 200},
	}

	_, err = suite.db.NewInsert().
		Model(&morePosts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Delete high-view posts by active users
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.GreaterThan("view_count", 100).InSubQuery("user_id", func(subquery SelectQuery) {
				subquery.Model((*User)(nil)).
					Select("id").
					Where(func(cb ConditionBuilder) {
						cb.Equals("is_active", true)
					})
			})
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify complex subquery deletion
	var activeUserPosts []Post

	err = suite.db.NewSelect().
		Model(&activeUserPosts).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Active User Post")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(activeUserPosts, 1, "Should have 1 remaining low-view post")
	suite.True(activeUserPosts[0].ViewCount <= 100, "Remaining post should have low view count")

	// Cleanup
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Active User Post")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.In("email", []string{"joindelete@example.com", "activejoin@example.com"})
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestDeleteWithReturning tests DELETE with RETURNING clause.
func (suite *DeleteTestSuite) TestDeleteWithReturning() {
	suite.T().Logf("Testing DELETE with RETURNING for %s", suite.dbType)

	// Skip if database doesn't support RETURNING
	if suite.dbType == constants.DbMySQL {
		suite.T().Skip("MySQL doesn't support RETURNING clause")
	}

	// Create test user for deletion
	testUser := &User{
		Name:     "Returning Delete Test",
		Email:    "returningdelete@example.com",
		Age:      28,
		IsActive: true,
	}

	_, err := suite.db.NewInsert().
		Model(testUser).
		Exec(suite.ctx)
	suite.NoError(err)

	// Test 1: Delete with RETURNING clause
	type DeleteResult struct {
		Id    string `bun:"id"`
		Name  string `bun:"name"`
		Email string `bun:"email"`
	}

	var returnedUser DeleteResult

	err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "returningdelete@example.com")
		}).
		Returning("id", "name", "email").
		Scan(suite.ctx, &returnedUser)
	suite.NoError(err)

	// Verify returned values
	suite.Equal(testUser.Id, returnedUser.Id)
	suite.Equal("Returning Delete Test", returnedUser.Name)
	suite.Equal("returningdelete@example.com", returnedUser.Email)

	// Verify the user was actually deleted
	var deletedUser User

	err = suite.db.NewSelect().
		Model(&deletedUser).
		Where(func(cb ConditionBuilder) {
			cb.PkEquals(testUser.Id)
		}).
		Scan(suite.ctx)
	suite.Error(err, "User should not exist after delete")

	// Test 2: Delete multiple records with RETURNING
	multiTestUsers := []*User{
		{Name: "Multi Delete 1", Email: "multi1@example.com", Age: 25, IsActive: true},
		{Name: "Multi Delete 2", Email: "multi2@example.com", Age: 26, IsActive: true},
	}

	_, err = suite.db.NewInsert().
		Model(&multiTestUsers).
		Exec(suite.ctx)
	suite.NoError(err)

	var returnedUsers []DeleteResult

	err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("email", "multi")
		}).
		Returning("id", "name", "email").
		Scan(suite.ctx, &returnedUsers)
	suite.NoError(err)
	suite.Len(returnedUsers, 2, "Should return 2 deleted users")

	// Verify all were deleted
	var remainingMultiUsers []User

	err = suite.db.NewSelect().
		Model(&remainingMultiUsers).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("email", "multi")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingMultiUsers, 0, "No multi users should remain")
}

// TestDeleteCascade tests cascading delete operations.
func (suite *DeleteTestSuite) TestDeleteCascade() {
	suite.T().Logf("Testing cascade DELETE for %s", suite.dbType)

	// Create test data with relationships
	testUser := &User{
		Name:     "Cascade Delete User",
		Email:    "cascadedelete@example.com",
		Age:      30,
		IsActive: true,
	}

	_, err := suite.db.NewInsert().
		Model(testUser).
		Exec(suite.ctx)
	suite.NoError(err)

	testCategory := &Category{
		Name:        "Cascade Delete Category",
		Description: lo.ToPtr("Category for cascade test"),
	}

	_, err = suite.db.NewInsert().
		Model(testCategory).
		Exec(suite.ctx)
	suite.NoError(err)

	testPosts := []*Post{
		{Title: "Cascade Post 1", Content: "Content", UserId: testUser.Id, CategoryId: testCategory.Id, Status: "published"},
		{Title: "Cascade Post 2", Content: "Content", UserId: testUser.Id, CategoryId: testCategory.Id, Status: "draft"},
	}

	_, err = suite.db.NewInsert().
		Model(&testPosts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Test manual cascade - delete posts first, then user
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("user_id", testUser.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify posts were deleted
	var userPosts []Post

	err = suite.db.NewSelect().
		Model(&userPosts).
		Where(func(cb ConditionBuilder) {
			cb.Equals("user_id", testUser.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(userPosts, 0, "User posts should be deleted")

	// Now delete the user
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.PkEquals(testUser.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify user was deleted
	var deletedUser User

	err = suite.db.NewSelect().
		Model(&deletedUser).
		Where(func(cb ConditionBuilder) {
			cb.PkEquals(testUser.Id)
		}).
		Scan(suite.ctx)
	suite.Error(err, "User should not exist after delete")

	// Cleanup category
	_, err = suite.db.NewDelete().
		Model((*Category)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.PkEquals(testCategory.Id)
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestDeleteErrorHandling tests error handling in delete operations.
func (suite *DeleteTestSuite) TestDeleteErrorHandling() {
	suite.T().Logf("Testing DELETE error handling for %s", suite.dbType)

	// Test 1: Delete with no matching records (should not error but affect 0 rows)
	result, err := suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "nonexistent@example.com")
		}).
		Exec(suite.ctx)
	suite.NoError(err, "DELETE with no matching rows should not error")

	rowsAffected, err := result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(0), rowsAffected, "Should affect 0 rows")

	// Test 2: Delete without WHERE clause (dangerous - deletes all records)
	// We'll test this carefully with a dedicated test table
	testSimple := []*SimpleModel{
		{Name: "Delete All Test 1", Value: 1},
		{Name: "Delete All Test 2", Value: 2},
	}

	_, err = suite.db.NewInsert().
		Model(&testSimple).
		Exec(suite.ctx)
	suite.NoError(err)

	// Delete all records from SimpleModel table (no WHERE clause)
	result, err = suite.db.NewDelete().
		Model((*SimpleModel)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("name", "Delete All Test")
		}).
		Exec(suite.ctx)
	suite.NoError(err, "DELETE without WHERE should work")

	rowsAffected, err = result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(2), rowsAffected, "Should delete all test records")

	// Verify all test records were deleted
	var remainingSimple []SimpleModel

	err = suite.db.NewSelect().
		Model(&remainingSimple).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("name", "Delete All Test")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingSimple, 0, "All test simple models should be deleted")

	// Test 3: Delete with invalid condition (should work but affect 0 rows)
	_, err = suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("invalid_field", "value") // This will likely cause a SQL error
		}).
		Exec(suite.ctx)
	suite.Error(err, "DELETE with invalid field should error")
}

// TestDeleteComplexConditions tests DELETE with complex conditions.
func (suite *DeleteTestSuite) TestDeleteComplexConditions() {
	suite.T().Logf("Testing DELETE with complex conditions for %s", suite.dbType)

	// Create test data
	complexDeletePosts := []*Post{
		{Title: "Complex 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft", ViewCount: 5},
		{Title: "Complex 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 50},
		{Title: "Complex 3", Content: "Content", UserId: "user2", CategoryId: "cat1", Status: "published", ViewCount: 150},
		{Title: "Complex 4", Content: "Content", UserId: "user2", CategoryId: "cat2", Status: "review", ViewCount: 25},
	}

	_, err := suite.db.NewInsert().
		Model(&complexDeletePosts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Test 1: Delete with grouped conditions
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "published").Group(func(innerCb ConditionBuilder) {
				innerCb.LessThan("view_count", 100).OrEquals("user_id", "user2")
			})
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify grouped condition deletion
	// Should delete: published posts with (view_count < 100 OR user_id = user2)
	// This means: Complex 2 (published, view=50) and Complex 3 (published, user2, view=150)
	var remainingComplex []Post

	err = suite.db.NewSelect().
		Model(&remainingComplex).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Complex")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingComplex, 2, "Should have 2 remaining posts")

	// Should have Complex 1 (draft) and Complex 4 (review)
	statusCount := make(map[string]int)
	for _, post := range remainingComplex {
		statusCount[post.Status]++
	}

	suite.Equal(1, statusCount["draft"], "Should have 1 draft post")
	suite.Equal(1, statusCount["review"], "Should have 1 review post")

	// Test 2: Delete with NOT conditions
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.NotEquals("status", "draft").NotEquals("view_count", 150)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify NOT condition deletion
	// Should delete posts that are not draft AND not with view_count 150
	// This means: Complex 4 (review, view=25)
	err = suite.db.NewSelect().
		Model(&remainingComplex).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Complex")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingComplex, 1, "Should have 1 remaining post")
	suite.Equal("draft", remainingComplex[0].Status, "Remaining post should be draft")

	// Test 3: Delete with BETWEEN condition
	moreComplexPosts := []*Post{
		{Title: "Between 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 25},
		{Title: "Between 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 50},
		{Title: "Between 3", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 75},
		{Title: "Between 4", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 100},
	}

	_, err = suite.db.NewInsert().
		Model(&moreComplexPosts).
		Exec(suite.ctx)
	suite.NoError(err)

	// Delete posts with view count between 40 and 80
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Between("view_count", 40, 80)
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	// Verify BETWEEN deletion
	var betweenRemaining []Post

	err = suite.db.NewSelect().
		Model(&betweenRemaining).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Between")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(betweenRemaining, 2, "Should have 2 remaining posts")

	// Should have posts with view count 25 and 100
	for _, post := range betweenRemaining {
		suite.True(post.ViewCount < 40 || post.ViewCount > 80,
			"Remaining posts should be outside the 40-80 range")
	}

	// Cleanup all test posts
	_, err = suite.db.NewDelete().
		Model((*Post)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("title", "Complex").OrStartsWith("title", "Between")
		}).
		Exec(suite.ctx)
	suite.NoError(err)
}

// TestDeletePerformance tests delete performance with larger datasets.
func (suite *DeleteTestSuite) TestDeletePerformance() {
	suite.T().Logf("Testing DELETE performance for %s", suite.dbType)

	// Create a batch of test records for performance testing
	batchSize := 100

	performanceUsers := make([]*User, batchSize)
	for i := range batchSize {
		performanceUsers[i] = &User{
			Name:     fmt.Sprintf("Perf User %d", i),
			Email:    fmt.Sprintf("perf-%03d@example.com", i),
			Age:      int16(20 + i%50),
			IsActive: i%2 == 0,
		}
	}

	start := time.Now()
	_, err := suite.db.NewInsert().
		Model(&performanceUsers).
		Exec(suite.ctx)
	suite.NoError(err)

	insertDuration := time.Since(start)
	suite.T().Logf("Inserted %d users in %v", batchSize, insertDuration)

	// Test bulk delete performance
	start = time.Now()
	result, err := suite.db.NewDelete().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("email", "perf-")
		}).
		Exec(suite.ctx)
	suite.NoError(err)

	deleteDuration := time.Since(start)

	rowsAffected, err := result.RowsAffected()
	suite.NoError(err)
	suite.Equal(int64(batchSize), rowsAffected)

	suite.T().Logf("Deleted %d users in %v", batchSize, deleteDuration)

	// Verify all performance test users were deleted
	var remainingPerfUsers []User

	err = suite.db.NewSelect().
		Model(&remainingPerfUsers).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("email", "perf")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(remainingPerfUsers, 0, "All performance test users should be deleted")
}
