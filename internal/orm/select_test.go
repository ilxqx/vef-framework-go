package orm

import "github.com/ilxqx/vef-framework-go/constants"

type SelectTestSuite struct {
	*ORMTestSuite
}

// TestBasicSelect tests basic SELECT functionality across all databases
func (suite *SelectTestSuite) TestBasicSelect() {
	suite.T().Logf("Testing basic SELECT for %s", suite.dbType)

	// Test 1: Select all users
	var users []User
	err := suite.db.NewSelect().
		Model(&users).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(users, 3, "Should have 3 users from fixture")

	// Verify user data matches fixture
	userEmails := make(map[string]User)
	for _, user := range users {
		userEmails[user.Email] = user
	}

	alice := userEmails["alice@example.com"]
	suite.Equal("Alice Johnson", alice.Name)
	suite.Equal(int16(30), alice.Age)
	suite.True(alice.IsActive)

	bob := userEmails["bob@example.com"]
	suite.Equal("Bob Smith", bob.Name)
	suite.Equal(int16(25), bob.Age)
	suite.True(bob.IsActive)

	charlie := userEmails["charlie@example.com"]
	suite.Equal("Charlie Brown", charlie.Name)
	suite.Equal(int16(35), charlie.Age)
	suite.False(charlie.IsActive)

	// Test 2: Select single user
	var singleUser User
	err = suite.db.NewSelect().
		Model(&singleUser).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "alice@example.com")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Equal("Alice Johnson", singleUser.Name)
	suite.Equal("alice@example.com", singleUser.Email)

	// Test 3: Count users
	count, err := suite.db.NewSelect().
		Model((*User)(nil)).
		Count(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(3), count)

	// Test 4: Check existence
	exists, err := suite.db.NewSelect().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "alice@example.com")
		}).
		Exists(suite.ctx)
	suite.NoError(err)
	suite.True(exists)

	// Test 5: Check non-existence
	notExists, err := suite.db.NewSelect().
		Model((*User)(nil)).
		Where(func(cb ConditionBuilder) {
			cb.Equals("email", "nonexistent@example.com")
		}).
		Exists(suite.ctx)
	suite.NoError(err)
	suite.False(notExists)
}

// TestSelectWithConditions tests SELECT with various WHERE conditions
func (suite *SelectTestSuite) TestSelectWithConditions() {
	suite.T().Logf("Testing SELECT with conditions for %s", suite.dbType)

	// Test 1: Equals condition
	var activeUsers []User
	err := suite.db.NewSelect().
		Model(&activeUsers).
		Where(func(cb ConditionBuilder) {
			cb.Equals("is_active", true)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(activeUsers, 2, "Should have 2 active users")
	for _, user := range activeUsers {
		suite.True(user.IsActive)
	}

	// Test 2: Greater than condition
	var olderUsers []User
	err = suite.db.NewSelect().
		Model(&olderUsers).
		Where(func(cb ConditionBuilder) {
			cb.GreaterThan("age", 28)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(olderUsers, 2, "Should have 2 users older than 28")
	for _, user := range olderUsers {
		suite.True(user.Age > 28)
	}

	// Test 3: Between condition
	var middleAgedUsers []User
	err = suite.db.NewSelect().
		Model(&middleAgedUsers).
		Where(func(cb ConditionBuilder) {
			cb.Between("age", 25, 30)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(middleAgedUsers, 2, "Should have 2 users between age 25-30")
	for _, user := range middleAgedUsers {
		suite.True(user.Age >= 25 && user.Age <= 30)
	}

	// Test 4: IN condition
	var specificUsers []User
	err = suite.db.NewSelect().
		Model(&specificUsers).
		Where(func(cb ConditionBuilder) {
			cb.In("email", []string{"alice@example.com", "bob@example.com"})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(specificUsers, 2, "Should have 2 specific users")

	// Test 5: LIKE condition (Contains)
	var johnsonUsers []User
	err = suite.db.NewSelect().
		Model(&johnsonUsers).
		Where(func(cb ConditionBuilder) {
			cb.Contains("name", "Johnson")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(johnsonUsers, 1, "Should have 1 user with Johnson in name")
	suite.Equal("Alice Johnson", johnsonUsers[0].Name)

	// Test 6: StartsWith condition
	var aliceUsers []User
	err = suite.db.NewSelect().
		Model(&aliceUsers).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("name", "Alice")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(aliceUsers, 1, "Should have 1 user starting with Alice")
	suite.Equal("Alice Johnson", aliceUsers[0].Name)

	// Test 7: Complex AND conditions
	var complexUsers []User
	err = suite.db.NewSelect().
		Model(&complexUsers).
		Where(func(cb ConditionBuilder) {
			cb.IsTrue("is_active").GreaterThan("age", 26)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(complexUsers, 1, "Should have 1 active user older than 26 (Alice)")
	userNames := []string{complexUsers[0].Name}
	suite.Contains(userNames, "Alice Johnson")

	// Test 8: OR conditions
	var orUsers []User
	err = suite.db.NewSelect().
		Model(&orUsers).
		Where(func(cb ConditionBuilder) {
			cb.Equals("age", 25).OrEquals("age", 35)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(orUsers, 2, "Should have 2 users with age 25 or 35")
}

// TestSelectWithJoins tests SELECT with JOIN operations
func (suite *SelectTestSuite) TestSelectWithJoins() {
	suite.T().Logf("Testing SELECT with JOINs for %s", suite.dbType)

	// Test 1: Select posts with user information
	var posts []Post
	err := suite.db.NewSelect().
		Model(&posts).
		Join((*User)(nil), func(cb ConditionBuilder) {
			cb.EqualsColumn("u.id", "p.user_id")
		}).
		Where(func(cb ConditionBuilder) {
			cb.Equals("u.name", "Alice Johnson")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(posts) >= 1, "Should have posts by Alice")

	// Test 2: Select posts with category using relation
	var postsWithRelation []Post
	err = suite.db.NewSelect().
		Model(&postsWithRelation).
		Relation("User").
		Relation("Category").
		Where(func(cb ConditionBuilder) {
			cb.Equals("status", "published")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(postsWithRelation) > 0, "Should have published posts")

	// Verify relations are loaded
	for _, post := range postsWithRelation {
		if post.User != nil {
			suite.NotEmpty(post.User.Name, "User relation should be loaded")
		}
		if post.Category != nil {
			suite.NotEmpty(post.Category.Name, "Category relation should be loaded")
		}
	}

	// Test 3: Left join with posts and categories
	var categoriesWithPosts []Category
	err = suite.db.NewSelect().
		Model(&categoriesWithPosts).
		LeftJoin((*Post)(nil), func(cb ConditionBuilder) {
			cb.EqualsColumn("c.id", "p.category_id")
		}, "p").
		GroupBy("c.id", "c.name", "c.description").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(categoriesWithPosts) > 0, "Should have categories")
}

// TestSelectWithAggregation tests SELECT with aggregation functions
func (suite *SelectTestSuite) TestSelectWithAggregation() {
	suite.T().Logf("Testing SELECT with aggregation for %s", suite.dbType)

	// Test 1: Count posts by status
	type StatusCount struct {
		Status string `bun:"status"`
		Count  int64  `bun:"post_count"`
	}

	var statusCounts []StatusCount
	err := suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("status").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.CountAll()
		}, "post_count").
		GroupBy("status").
		OrderBy("status").
		Scan(suite.ctx, &statusCounts)
	suite.NoError(err)
	suite.True(len(statusCounts) > 0, "Should have status counts")

	// Verify counts make sense
	totalCount := int64(0)
	for _, sc := range statusCounts {
		suite.True(sc.Count > 0, "Each status should have at least 1 post")
		totalCount += sc.Count
	}

	// Verify total matches expected post count from fixture
	expectedPosts := 8 // From fixture.yaml
	suite.Equal(int64(expectedPosts), totalCount, "Total count should match fixture posts")

	// Test 2: Average age of users
	type AgeStats struct {
		AvgAge   float64 `bun:"avg_age"`
		MinAge   int16   `bun:"min_age"`
		MaxAge   int16   `bun:"max_age"`
		CountAge int64   `bun:"count_age"`
	}

	var ageStats AgeStats
	err = suite.db.NewSelect().
		Model((*User)(nil)).
		SelectExpr(func(eb ExprBuilder) any {
			return eb.AvgColumn("age")
		}, "avg_age").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.MinColumn("age")
		}, "min_age").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.MaxColumn("age")
		}, "max_age").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.CountColumn("age")
		}, "count_age").
		Scan(suite.ctx, &ageStats)
	suite.NoError(err)
	suite.Equal(int64(3), ageStats.CountAge, "Should count 3 users")
	suite.Equal(int16(25), ageStats.MinAge, "Min age should be 25")
	suite.Equal(int16(35), ageStats.MaxAge, "Max age should be 35")
	suite.InDelta(30.0, ageStats.AvgAge, 1.0, "Average age should be around 30")

	// Test 3: Sum of view counts
	type ViewStats struct {
		TotalViews int64   `bun:"total_views"`
		AvgViews   float64 `bun:"avg_views"`
	}

	var viewStats ViewStats
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		SelectExpr(func(eb ExprBuilder) any {
			return eb.SumColumn("view_count")
		}, "total_views").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.AvgColumn("view_count")
		}, "avg_views").
		Scan(suite.ctx, &viewStats)
	suite.NoError(err)
	suite.True(viewStats.TotalViews > 0, "Should have total views")
	suite.True(viewStats.AvgViews > 0, "Should have average views")
}

// TestSelectWithSubQueries tests SELECT with subquery conditions
func (suite *SelectTestSuite) TestSelectWithSubQueries() {
	suite.T().Logf("Testing SELECT with subqueries for %s", suite.dbType)

	// Test 1: Users who have published posts
	var usersWithPublishedPosts []User
	err := suite.db.NewSelect().
		Model(&usersWithPublishedPosts).
		Where(func(cb ConditionBuilder) {
			cb.InSubQuery("id", func(subquery SelectQuery) {
				subquery.Model((*Post)(nil)).
					Select("user_id").
					Where(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
			})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(usersWithPublishedPosts) > 0, "Should have users with published posts")

	// Test 2: Posts with above average view count
	var popularPosts []Post
	err = suite.db.NewSelect().
		Model(&popularPosts).
		Where(func(cb ConditionBuilder) {
			cb.GreaterThanSubQuery("view_count", func(subquery SelectQuery) {
				subquery.Model((*Post)(nil)).
					SelectExpr(func(eb ExprBuilder) any {
						return eb.AvgColumn("view_count")
					})
			})
		}).
		OrderByDesc("view_count").
		Scan(suite.ctx)
	suite.NoError(err)

	if len(popularPosts) > 0 {
		// Verify posts are actually above average
		var avgViewCount float64
		err = suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}).
			Scan(suite.ctx, &avgViewCount)
		suite.NoError(err)

		for _, post := range popularPosts {
			suite.True(float64(post.ViewCount) > avgViewCount,
				"Post %s should have above average view count", post.Title)
		}
	}

	// Test 3: Categories with posts
	var categoriesWithPosts []Category
	err = suite.db.NewSelect().
		Model(&categoriesWithPosts).
		Where(func(cb ConditionBuilder) {
			cb.InSubQuery("id", func(subquery SelectQuery) {
				subquery.Model((*Post)(nil)).
					Select("category_id").
					Distinct()
			})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(categoriesWithPosts) > 0, "Should have categories with posts")
}

// TestSelectWithOrderingAndLimits tests SELECT with ORDER BY and LIMIT clauses
func (suite *SelectTestSuite) TestSelectWithOrderingAndLimits() {
	suite.T().Logf("Testing SELECT with ordering and limits for %s", suite.dbType)

	// Test 1: Order by age ascending
	var usersByAge []User
	err := suite.db.NewSelect().
		Model(&usersByAge).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(usersByAge, 3)
	suite.True(usersByAge[0].Age <= usersByAge[1].Age)
	suite.True(usersByAge[1].Age <= usersByAge[2].Age)

	// Test 2: Order by age descending
	var usersByAgeDesc []User
	err = suite.db.NewSelect().
		Model(&usersByAgeDesc).
		OrderByDesc("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(usersByAgeDesc, 3)
	suite.True(usersByAgeDesc[0].Age >= usersByAgeDesc[1].Age)
	suite.True(usersByAgeDesc[1].Age >= usersByAgeDesc[2].Age)

	// Test 3: Limit results
	var limitedUsers []User
	err = suite.db.NewSelect().
		Model(&limitedUsers).
		OrderBy("age").
		Limit(2).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(limitedUsers, 2, "Should return only 2 users")

	// Test 4: Order by multiple columns
	var postsByStatusAndViews []Post
	err = suite.db.NewSelect().
		Model(&postsByStatusAndViews).
		OrderBy("status").
		OrderByDesc("view_count").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(postsByStatusAndViews) > 0, "Should have posts ordered by status and view count")

	// Test 5: Offset and Limit (pagination)
	var paginatedUsers []User
	err = suite.db.NewSelect().
		Model(&paginatedUsers).
		OrderBy("age").
		Offset(1).
		Limit(2).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(paginatedUsers, 2, "Should return 2 users with offset")

	// Verify offset worked by comparing with full result
	if len(usersByAge) >= 3 {
		suite.Equal(usersByAge[1].Id, paginatedUsers[0].Id, "First paginated user should match second from full result")
	}
}

// TestSelectSpecificColumns tests selecting specific columns
func (suite *SelectTestSuite) TestSelectSpecificColumns() {
	suite.T().Logf("Testing SELECT with specific columns for %s", suite.dbType)

	// Test 1: Select only specific columns
	type UserBasic struct {
		Id    string `bun:"id"`
		Name  string `bun:"name"`
		Email string `bun:"email"`
	}

	var basicUsers []UserBasic
	err := suite.db.NewSelect().
		Model((*User)(nil)).
		Select("id", "name", "email").
		OrderBy("name").
		Scan(suite.ctx, &basicUsers)
	suite.NoError(err)
	suite.Len(basicUsers, 3)

	for _, user := range basicUsers {
		suite.NotEmpty(user.Id, "ID should be populated")
		suite.NotEmpty(user.Name, "Name should be populated")
		suite.NotEmpty(user.Email, "Email should be populated")
	}

	// Test 2: Select with expression (calculated field)
	type UserWithAge struct {
		Name    string `bun:"name"`
		Age     int16  `bun:"age"`
		AgeDesc string `bun:"age_desc"`
	}

	var usersWithAgeDesc []UserWithAge
	err = suite.db.NewSelect().
		Model((*User)(nil)).
		Select("name", "age").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Concat("'Age: '", "age")
		}, "age_desc").
		OrderBy("name").
		Scan(suite.ctx, &usersWithAgeDesc)
	suite.NoError(err)
	suite.Len(usersWithAgeDesc, 3)

	for _, user := range usersWithAgeDesc {
		suite.NotEmpty(user.AgeDesc, "Age description should be calculated")
		suite.Contains(user.AgeDesc, "Age:", "Age description should contain 'Age:'")
	}

	// Test 3: Select distinct values
	var distinctStatuses []string
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Distinct().
		Select("status").
		OrderBy("status").
		Scan(suite.ctx, &distinctStatuses)
	suite.NoError(err)
	suite.True(len(distinctStatuses) > 0, "Should have distinct statuses")

	// Verify distinctness
	seen := make(map[string]bool)
	for _, status := range distinctStatuses {
		suite.False(seen[status], "Status %s should appear only once", status)
		seen[status] = true
	}
}

// TestSelectWithComplexConditions tests complex WHERE conditions
func (suite *SelectTestSuite) TestSelectWithComplexConditions() {
	suite.T().Logf("Testing SELECT with complex conditions for %s", suite.dbType)

	// Test 1: Nested conditions with groups
	var complexUsers []User
	err := suite.db.NewSelect().
		Model(&complexUsers).
		Where(func(cb ConditionBuilder) {
			cb.Equals("is_active", true).Group(func(innerCb ConditionBuilder) {
				innerCb.LessThan("age", 30).OrGreaterThan("age", 32)
			})
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	// Should match: active users with age < 30 OR age > 32
	for _, user := range complexUsers {
		suite.True(user.IsActive, "User should be active")
		suite.True(user.Age < 30 || user.Age > 32, "User age should be < 30 or > 32")
	}

	// Test 2: NOT conditions
	var notCharlie []User
	err = suite.db.NewSelect().
		Model(&notCharlie).
		Where(func(cb ConditionBuilder) {
			cb.NotEquals("name", "Charlie Brown")
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(notCharlie, 2, "Should have 2 users not named Charlie Brown")

	for _, user := range notCharlie {
		suite.NotEqual("Charlie Brown", user.Name, "User should not be Charlie Brown")
	}

	// Test 3: NULL checks
	var postsWithDescription []Post
	err = suite.db.NewSelect().
		Model(&postsWithDescription).
		Where(func(cb ConditionBuilder) {
			cb.IsNotNull("description")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	for _, post := range postsWithDescription {
		suite.NotNil(post.Description, "Description should not be null")
		suite.NotEmpty(*post.Description, "Description should not be empty")
	}

	// Test 4: Multiple NOT IN conditions
	var filteredPosts []Post
	err = suite.db.NewSelect().
		Model(&filteredPosts).
		Where(func(cb ConditionBuilder) {
			cb.NotIn("status", []string{"draft"}).
				NotContains("title", "Test")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	for _, post := range filteredPosts {
		suite.NotEqual("draft", post.Status, "Post should not be draft")
		suite.NotContains(post.Title, "Test", "Post title should not contain 'Test'")
	}
}

// TestSelectWithWindowFunctions tests SELECT with window functions
func (suite *SelectTestSuite) TestSelectWithWindowFunctions() {
	suite.T().Logf("Testing SELECT with window functions for %s", suite.dbType)

	// Test 1: ROW_NUMBER window function
	type UserWithRowNumber struct {
		Id     string `bun:"id"`
		Name   string `bun:"name"`
		Age    int16  `bun:"age"`
		RowNum int64  `bun:"row_num"`
	}

	var usersWithRowNum []UserWithRowNumber
	err := suite.db.NewSelect().
		Model((*User)(nil)).
		Select("id", "name", "age").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.RowNumber(func(rn RowNumberBuilder) {
				rn.Over().OrderBy("age")
			})
		}, "row_num").
		OrderBy("age").
		Scan(suite.ctx, &usersWithRowNum)
	suite.NoError(err)
	suite.Len(usersWithRowNum, 3)

	// Verify ROW_NUMBER sequence
	for i, user := range usersWithRowNum {
		suite.Equal(int64(i+1), user.RowNum, "ROW_NUMBER should be sequential")
	}

	// Test 2: RANK and DENSE_RANK window functions
	type PostWithRank struct {
		Id        string `bun:"id"`
		Title     string `bun:"title"`
		Status    string `bun:"status"`
		ViewCount int64  `bun:"view_count"`
		Rank      int64  `bun:"rank"`
		DenseRank int64  `bun:"dense_rank"`
	}

	var postsWithRank []PostWithRank
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("id", "title", "status", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Rank(func(r RankBuilder) {
				r.Over().PartitionBy("status").OrderByDesc("view_count")
			})
		}, "rank").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.DenseRank(func(dr DenseRankBuilder) {
				dr.Over().PartitionBy("status").OrderByDesc("view_count")
			})
		}, "dense_rank").
		OrderBy("status").
		OrderByDesc("view_count").
		Scan(suite.ctx, &postsWithRank)
	suite.NoError(err)
	suite.True(len(postsWithRank) > 0, "Should have posts with rank")

	// Verify ranking within partitions
	statusGroups := make(map[string][]PostWithRank)
	for _, post := range postsWithRank {
		statusGroups[post.Status] = append(statusGroups[post.Status], post)
	}

	for status, posts := range statusGroups {
		if len(posts) > 1 {
			// Verify ranks are assigned correctly within each status partition
			suite.True(posts[0].Rank >= 1, "First post in %s partition should have rank >= 1", status)
			suite.True(posts[0].DenseRank >= 1, "First post in %s partition should have dense_rank >= 1", status)
		}
	}

	// Test 3: Aggregate window functions (SUM, AVG, COUNT)
	type UserWithAggregates struct {
		Name         string  `bun:"name"`
		Age          int16   `bun:"age"`
		RunningTotal int64   `bun:"running_total"`
		MovingAvg    float64 `bun:"moving_avg"`
		RunningCount int64   `bun:"running_count"`
	}

	var usersWithAggregates []UserWithAggregates
	err = suite.db.NewSelect().
		Model((*User)(nil)).
		Select("name", "age").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.WSum(func(ws WindowSumBuilder) {
				ws.Column("age").Over().OrderBy("age").Rows().UnboundedPreceding()
			})
		}, "running_total").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.WAvg(func(wa WindowAvgBuilder) {
				wa.Column("age").Over().OrderBy("age").Rows().UnboundedPreceding()
			})
		}, "moving_avg").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.WCount(func(wc WindowCountBuilder) {
				wc.All().Over().OrderBy("age").Rows().UnboundedPreceding()
			})
		}, "running_count").
		OrderBy("age").
		Scan(suite.ctx, &usersWithAggregates)
	suite.NoError(err)
	suite.Len(usersWithAggregates, 3)

	// Verify running totals and counts
	for i, user := range usersWithAggregates {
		suite.Equal(int64(i+1), user.RunningCount, "Running count should increment by 1")
		suite.True(user.RunningTotal > 0, "Running total should be positive")
		suite.True(user.MovingAvg > 0, "Moving average should be positive")
	}

	// Test 4: LAG and LEAD window functions
	type PostWithLagLead struct {
		Title         string `bun:"title"`
		ViewCount     int64  `bun:"view_count"`
		PrevViewCount *int64 `bun:"prev_view_count"`
		NextViewCount *int64 `bun:"next_view_count"`
	}

	var postsWithLagLead []PostWithLagLead
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Lag(func(lb LagBuilder) {
				lb.Column("view_count").Over().OrderBy("view_count")
			})
		}, "prev_view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Lead(func(lb LeadBuilder) {
				lb.Column("view_count").Over().OrderBy("view_count")
			})
		}, "next_view_count").
		OrderBy("view_count").
		Scan(suite.ctx, &postsWithLagLead)
	suite.NoError(err)
	suite.True(len(postsWithLagLead) > 0, "Should have posts with lag/lead")

	// Verify LAG/LEAD behavior
	if len(postsWithLagLead) > 0 {
		// First row should have null prev_view_count
		suite.Nil(postsWithLagLead[0].PrevViewCount, "First row should have null previous value")
		// Last row should have null next_view_count
		lastIdx := len(postsWithLagLead) - 1
		suite.Nil(postsWithLagLead[lastIdx].NextViewCount, "Last row should have null next value")
	}

	// Test 5: FIRST_VALUE and LAST_VALUE window functions
	type PostWithFirstLast struct {
		Title         string `bun:"title"`
		Status        string `bun:"status"`
		ViewCount     int64  `bun:"view_count"`
		FirstInStatus int64  `bun:"first_in_status"`
		LastInStatus  int64  `bun:"last_in_status"`
	}

	var postsWithFirstLast []PostWithFirstLast
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title", "status", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.FirstValue(func(fvb FirstValueBuilder) {
				fvb.Column("view_count").Over().PartitionBy("status").OrderBy("view_count")
			})
		}, "first_in_status").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.LastValue(func(lvb LastValueBuilder) {
				lvb.Column("view_count").Over().PartitionBy("status").OrderBy("view_count").Rows().UnboundedPreceding().And().UnboundedFollowing()
			})
		}, "last_in_status").
		OrderBy("status", "view_count").
		Scan(suite.ctx, &postsWithFirstLast)
	suite.NoError(err)
	suite.True(len(postsWithFirstLast) > 0, "Should have posts with first/last values")

	// Verify FIRST_VALUE and LAST_VALUE behavior
	statusFirstLast := make(map[string][]PostWithFirstLast)
	for _, post := range postsWithFirstLast {
		statusFirstLast[post.Status] = append(statusFirstLast[post.Status], post)
	}

	for status, posts := range statusFirstLast {
		if len(posts) > 1 {
			// All posts in same status should have same first_in_status value
			firstValue := posts[0].FirstInStatus
			lastValue := posts[0].LastInStatus
			for _, post := range posts {
				suite.Equal(firstValue, post.FirstInStatus, "All posts in %s should have same first value", status)
				suite.Equal(lastValue, post.LastInStatus, "All posts in %s should have same last value", status)
			}
		}
	}

	// Test 6: NTILE window function for quartiles
	type UserWithQuartile struct {
		Name     string `bun:"name"`
		Age      int16  `bun:"age"`
		Quartile int64  `bun:"quartile"`
	}

	var usersWithQuartile []UserWithQuartile
	err = suite.db.NewSelect().
		Model((*User)(nil)).
		Select("name", "age").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Ntile(func(nb NtileBuilder) {
				nb.Buckets(4).Over().OrderBy("age")
			})
		}, "quartile").
		OrderBy("age").
		Scan(suite.ctx, &usersWithQuartile)
	suite.NoError(err)
	suite.Len(usersWithQuartile, 3)

	// Verify quartile assignment (with 3 users, we expect quartiles 1, 2, 3 or similar distribution)
	for _, user := range usersWithQuartile {
		suite.True(user.Quartile >= 1 && user.Quartile <= 4, "Quartile should be between 1 and 4")
	}

	// Test 7: Complex window function with multiple partitions and ordering
	type PostAnalytics struct {
		Title           string  `bun:"title"`
		Status          string  `bun:"status"`
		ViewCount       int64   `bun:"view_count"`
		RankInStatus    int64   `bun:"rank_in_status"`
		PercentOfTotal  float64 `bun:"percent_of_total"`
		CumulativeViews int64   `bun:"cumulative_views"`
	}

	var postAnalytics []PostAnalytics
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title", "status", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Rank(func(r RankBuilder) {
				r.Over().PartitionBy("status").OrderByDesc("view_count")
			})
		}, "rank_in_status").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.PercentRank(func(pr PercentRankBuilder) {
				pr.Over().OrderByDesc("view_count")
			})
		}, "percent_of_total").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.WSum(func(ws WindowSumBuilder) {
				ws.Column("view_count").Over().PartitionBy("status").OrderByDesc("view_count").Rows().UnboundedPreceding()
			})
		}, "cumulative_views").
		OrderBy("status").
		OrderByDesc("view_count").
		Scan(suite.ctx, &postAnalytics)
	suite.NoError(err)
	suite.True(len(postAnalytics) > 0, "Should have post analytics")

	// Verify complex analytics calculations
	for _, post := range postAnalytics {
		suite.True(post.RankInStatus >= 1, "Rank should be at least 1")
		suite.True(post.PercentOfTotal >= 0 && post.PercentOfTotal <= 1, "Percent rank should be between 0 and 1")
		suite.True(post.CumulativeViews >= post.ViewCount, "Cumulative views should be at least equal to current view count")
	}
}

// TestSelectWithWindowFunctionsAdvanced covers advanced window features
func (suite *SelectTestSuite) TestSelectWithWindowFunctionsAdvanced() {
	suite.T().Logf("Testing advanced window functions for %s", suite.dbType)

	// Test 1: LAG/LEAD with offset and default value
	type PostWithLagLeadAdvanced struct {
		Title          string `bun:"title"`
		ViewCount      int64  `bun:"view_count"`
		Prev2ViewCount *int64 `bun:"prev2_view_count"`
		Next2ViewCount *int64 `bun:"next2_view_count"`
		Next2OrDefault int64  `bun:"next2_or_default"`
	}

	var advLagLead []PostWithLagLeadAdvanced
	err := suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Lag(func(lb LagBuilder) {
				lb.Column("view_count").Offset(2).Over().OrderBy("view_count")
			})
		}, "prev2_view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Lead(func(lb LeadBuilder) {
				lb.Column("view_count").Offset(2).Over().OrderBy("view_count")
			})
		}, "next2_view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Lead(func(lb LeadBuilder) {
				lb.Column("view_count").Offset(2).DefaultValue(-1).Over().OrderBy("view_count")
			})
		}, "next2_or_default").
		OrderBy("view_count").
		Scan(suite.ctx, &advLagLead)
	suite.NoError(err)
	suite.True(len(advLagLead) > 0, "Should have posts for advanced lag/lead")
	if len(advLagLead) >= 3 {
		// The third row's Prev2 should equal the first row's view_count
		if advLagLead[2].Prev2ViewCount != nil {
			suite.Equal(advLagLead[0].ViewCount, *advLagLead[2].Prev2ViewCount)
		}
	}
	// Default value should apply on the last rows where LEAD overflows
	if len(advLagLead) >= 1 {
		lastIdx := len(advLagLead) - 1
		suite.NotZero(advLagLead[lastIdx].Next2OrDefault)
	}

	// Test 2: Moving average with ROWS BETWEEN 2 PRECEDING AND CURRENT ROW
	type PostWithMovingAvg struct {
		Title     string  `bun:"title"`
		ViewCount int64   `bun:"view_count"`
		MovAvg    float64 `bun:"mov_avg"`
	}

	var movingAvgRows []PostWithMovingAvg
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.WAvg(func(wab WindowAvgBuilder) {
				wab.Column("view_count").Over().OrderBy("view_count").Rows().Preceding(2).And().CurrentRow()
			})
		}, "mov_avg").
		OrderBy("view_count").
		Scan(suite.ctx, &movingAvgRows)
	suite.NoError(err)
	suite.True(len(movingAvgRows) > 0, "Should have moving average values")

	// Test 3: NTH_VALUE with full frame (omit FROM to maximize dialect support)
	type PostWithNthValue struct {
		Status        string `bun:"status"`
		ViewCount     int64  `bun:"view_count"`
		SecondFromEnd int64  `bun:"second_from_end"`
	}

	var nthVals []PostWithNthValue
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("status", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.NthValue(func(nvb NthValueBuilder) {
				nvb.Column("view_count").N(2).Over().PartitionBy("status").OrderBy("view_count").Rows().UnboundedPreceding().And().UnboundedFollowing()
			})
		}, "second_from_end").
		OrderBy("status", "view_count").
		Scan(suite.ctx, &nthVals)
	suite.NoError(err)
	suite.True(len(nthVals) > 0, "Should compute NTH_VALUE from last")
}

// TestSelectWithWindowFunctionsFromClause validates FROM FIRST/LAST placement
func (suite *SelectTestSuite) TestSelectWithWindowFunctionsFromClause() {
	suite.T().Logf("Testing window functions FROM FIRST/LAST for %s", suite.dbType)

	// FROM FIRST/FROM LAST syntax is only supported by Oracle and SQL Server
	// PostgreSQL, MySQL, and SQLite do not support this syntax
	suite.T().Skipf("Skip FROM FIRST/LAST tests for %s - not supported by common databases", suite.dbType)

	// Test 1: NTH_VALUE ... FROM FIRST with full frame
	type PostWithNthFromFirst struct {
		Status string `bun:"status"`
		NthVal int64  `bun:"nth_val"`
	}

	var nthFirst []PostWithNthFromFirst
	err := suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("status").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.NthValue(func(nvb NthValueBuilder) {
				nvb.Column("view_count").N(1).FromFirst().Over().
					PartitionBy("status").OrderBy("view_count").
					Rows().UnboundedPreceding().And().UnboundedFollowing()
			})
		}, "nth_val").
		OrderBy("status").
		Scan(suite.ctx, &nthFirst)
	suite.NoError(err)
	suite.True(len(nthFirst) > 0)

	// Test 2: NTH_VALUE ... FROM LAST with full frame
	type PostWithNthFromLast struct {
		Status string `bun:"status"`
		NthVal int64  `bun:"nth_val"`
	}

	var nthLast []PostWithNthFromLast
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("status").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.NthValue(func(nvb NthValueBuilder) {
				nvb.Column("view_count").N(1).FromLast().Over().
					PartitionBy("status").OrderBy("view_count").
					Rows().UnboundedPreceding().And().UnboundedFollowing()
			})
		}, "nth_val").
		OrderBy("status").
		Scan(suite.ctx, &nthLast)
	suite.NoError(err)
	suite.True(len(nthLast) > 0)
}

// TestSelectWithComplexAggregates tests complex aggregate function features
func (suite *SelectTestSuite) TestSelectWithComplexAggregates() {
	suite.T().Logf("Testing complex aggregate functions for %s", suite.dbType)

	// Test 1: FILTER clause (PostgreSQL, SQLite) vs CASE equivalent (MySQL, Oracle, SQL Server)
	type ConditionalCounts struct {
		TotalPosts      int64 `bun:"total_posts"`
		PublishedPosts  int64 `bun:"published_posts"`
		DraftPosts      int64 `bun:"draft_posts"`
		ReviewPosts     int64 `bun:"review_posts"`
		HighViewPosts   int64 `bun:"high_view_posts"`
		TotalViews      int64 `bun:"total_views"`
		PublishedViews  int64 `bun:"published_views"`
		AvgPublishedAge int64 `bun:"avg_published_age"`
	}

	var conditionalCounts ConditionalCounts
	query := suite.db.NewSelect().
		Model((*Post)(nil)).
		SelectExpr(func(eb ExprBuilder) any {
			return eb.CountAll()
		}, "total_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.SumColumn("view_count")
		}, "total_views")

	// Test FILTER clause - framework should handle database compatibility internally
	query.
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Count(func(cb CountBuilder) {
				cb.All().Filter(func(cb ConditionBuilder) {
					cb.Equals("status", "published")
				})
			})
		}, "published_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Count(func(cb CountBuilder) {
				cb.All().Filter(func(cb ConditionBuilder) {
					cb.Equals("status", "draft")
				})
			})
		}, "draft_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Count(func(cb CountBuilder) {
				cb.All().Filter(func(cb ConditionBuilder) {
					cb.Equals("status", "review")
				})
			})
		}, "review_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Count(func(cb CountBuilder) {
				cb.All().Filter(func(cb ConditionBuilder) {
					cb.GreaterThan("view_count", 80)
				})
			})
		}, "high_view_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Sum(func(sb SumBuilder) {
				sb.Column("view_count").Filter(func(cb ConditionBuilder) {
					cb.Equals("status", "published")
				})
			})
		}, "published_views").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.ToInteger(
				eb.Avg(func(ab AvgBuilder) {
					ab.Column("view_count").Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
				}),
			)
		}, "avg_published_age")

	err := query.Scan(suite.ctx, &conditionalCounts)
	suite.NoError(err)
	suite.True(conditionalCounts.TotalPosts > 0, "Should have posts")
	suite.True(conditionalCounts.PublishedPosts >= 0, "Should count published posts")
	suite.True(conditionalCounts.TotalViews > 0, "Should have total views")

	// Test 2: DISTINCT aggregates
	type DistinctStats struct {
		UniqueStatuses   int64   `bun:"unique_statuses"`
		UniqueCategories int64   `bun:"unique_categories"`
		DistinctUserIds  int64   `bun:"distinct_user_ids"`
		AvgDistinctViews float64 `bun:"avg_distinct_views"`
	}

	var distinctStats DistinctStats
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		SelectExpr(func(eb ExprBuilder) any {
			return eb.CountColumn("status", true) // distinct = true
		}, "unique_statuses").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.CountColumn("category_id", true)
		}, "unique_categories").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.CountColumn("user_id", true)
		}, "distinct_user_ids").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.AvgColumn("view_count", true) // distinct average
		}, "avg_distinct_views").
		Scan(suite.ctx, &distinctStats)
	suite.NoError(err)
	suite.True(distinctStats.UniqueStatuses > 0, "Should have unique statuses")
	suite.True(distinctStats.UniqueCategories > 0, "Should have unique categories")
	suite.True(distinctStats.AvgDistinctViews > 0, "Should have distinct average")

	// Test 3: String aggregation - framework should handle dialect conversion
	type StringAggResult struct {
		StatusList       string `bun:"status_list"`
		OrderedTitleList string `bun:"ordered_title_list"`
	}

	var stringAggResult StringAggResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		SelectExpr(func(eb ExprBuilder) any {
			return eb.StringAgg(func(sab StringAggBuilder) {
				sab.Column("status").Separator(", ").Distinct()
			})
		}, "status_list").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.StringAgg(func(sab StringAggBuilder) {
				sab.Column("title").Separator(" | ").OrderBy("view_count")
			})
		}, "ordered_title_list").
		Scan(suite.ctx, &stringAggResult)
	suite.NoError(err)
	suite.NotEmpty(stringAggResult.StatusList, "Should aggregate distinct statuses")
	suite.NotEmpty(stringAggResult.OrderedTitleList, "Should aggregate ordered titles")
	suite.T().Logf("Status list: %s", stringAggResult.StatusList)
	suite.T().Logf("Ordered title list: %s", stringAggResult.OrderedTitleList)

	// Test 4: Array aggregation - framework should handle database compatibility
	type ArrayAggResult struct {
		ViewCountArray []int64  `bun:"view_count_array,array"`
		OrderedTitles  []string `bun:"ordered_titles,array"`
		UniqueStatuses []string `bun:"unique_statuses,array"`
	}

	var arrayAggResult ArrayAggResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		SelectExpr(func(eb ExprBuilder) any {
			return eb.ArrayAgg(func(aab ArrayAggBuilder) {
				aab.Column("view_count").OrderByDesc("view_count")
			})
		}, "view_count_array").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.ArrayAgg(func(aab ArrayAggBuilder) {
				aab.Column("title").OrderBy("title")
			})
		}, "ordered_titles").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.ArrayAgg(func(aab ArrayAggBuilder) {
				aab.Column("status").Distinct().OrderBy("status")
			})
		}, "unique_statuses").
		Scan(suite.ctx, &arrayAggResult)
	suite.NoError(err)
	suite.True(len(arrayAggResult.ViewCountArray) > 0, "Should have view count array")
	suite.True(len(arrayAggResult.OrderedTitles) > 0, "Should have ordered titles")
	suite.True(len(arrayAggResult.UniqueStatuses) > 0, "Should have unique statuses")

	// Verify ordering (skip for MySQL and SQLite due to database limitations)
	if suite.dbType != constants.DbMySQL && suite.dbType != constants.DbSQLite {
		for i := 1; i < len(arrayAggResult.ViewCountArray); i++ {
			suite.True(arrayAggResult.ViewCountArray[i-1] >= arrayAggResult.ViewCountArray[i],
				"View counts should be in descending order")
		}
	}

	// Test 5: JSON functions - using simple JSON array creation
	type JsonResult struct {
		SimpleArray string `bun:"simple_array"`
	}

	var jsonResult JsonResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONArray(eb.Column("id"), eb.Column("title"), eb.Column("status"))
		}, "simple_array").
		Limit(1).
		Scan(suite.ctx, &jsonResult)
	suite.NoError(err)
	suite.NotEmpty(jsonResult.SimpleArray, "Should have JSON array")
	suite.T().Logf("Simple JSON Array: %s", jsonResult.SimpleArray)

	// Test 6: Complex GROUP BY with multiple conditional aggregates
	type CategoryStats struct {
		CategoryName   string  `bun:"category_name"`
		TotalPosts     int64   `bun:"total_posts"`
		PublishedPosts int64   `bun:"published_posts"`
		DraftPosts     int64   `bun:"draft_posts"`
		TotalViews     int64   `bun:"total_views"`
		PublishedViews int64   `bun:"published_views"`
		AvgViews       float64 `bun:"avg_views"`
		MaxViews       int64   `bun:"max_views"`
		MinViews       int64   `bun:"min_views"`
	}

	var categoryStats []CategoryStats
	categoryQuery := suite.db.NewSelect().
		Model((*Post)(nil)).
		Join((*Category)(nil), func(cb ConditionBuilder) {
			cb.EqualsColumn("c.id", "category_id")
		}).
		SelectAs("c.name", "category_name").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.CountAll()
		}, "total_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.SumColumn("view_count")
		}, "total_views").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.AvgColumn("view_count")
		}, "avg_views").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.MaxColumn("view_count")
		}, "max_views").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.MinColumn("view_count")
		}, "min_views").
		GroupBy("c.id", "c.name").
		OrderBy("c.name")

	// Add conditional aggregates - framework should handle database compatibility
	categoryQuery = categoryQuery.
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Count(func(cb CountBuilder) {
				cb.All().Filter(func(cb ConditionBuilder) {
					cb.Equals("status", "published")
				})
			})
		}, "published_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Count(func(cb CountBuilder) {
				cb.All().Filter(func(cb ConditionBuilder) {
					cb.Equals("status", "draft")
				})
			})
		}, "draft_posts").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Sum(func(sb SumBuilder) {
				sb.Column("p.view_count").Filter(func(cb ConditionBuilder) {
					cb.Equals("status", "published")
				})
			})
		}, "published_views")

	err = categoryQuery.Scan(suite.ctx, &categoryStats)
	suite.NoError(err)
	suite.True(len(categoryStats) > 0, "Should have category statistics")

	for _, stat := range categoryStats {
		suite.NotEmpty(stat.CategoryName, "Category name should not be empty")
		suite.True(stat.TotalPosts > 0, "Each category should have posts")
		suite.True(stat.TotalViews >= 0, "Total views should be non-negative")
		suite.True(stat.PublishedViews >= 0, "Published views should be non-negative")
		suite.True(stat.PublishedPosts >= 0, "Published posts should be non-negative")
		suite.True(stat.DraftPosts >= 0, "Draft posts should be non-negative")
		suite.True(stat.PublishedPosts+stat.DraftPosts <= stat.TotalPosts, "Published + Draft should not exceed total")
		suite.T().Logf("Category %s: %d posts (%d published, %d draft), %d total views (%d published views)",
			stat.CategoryName, stat.TotalPosts, stat.PublishedPosts, stat.DraftPosts,
			stat.TotalViews, stat.PublishedViews)
	}
}

// TestSelectWithJsonFunctions tests JSON functions and operations
func (suite *SelectTestSuite) TestSelectWithJsonFunctions() {
	suite.T().Logf("Testing JSON functions for %s", suite.dbType)

	// Test 1: JSON_OBJECT creation
	type JsonObjectResult struct {
		Id         string `bun:"id"`
		Name       string `bun:"name"`
		UserObject string `bun:"user_object"`
	}

	var jsonObjectResults []JsonObjectResult
	err := suite.db.NewSelect().
		Model((*User)(nil)).
		Select("id", "name").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONObject("user_id", eb.Column("id"), "user_name", eb.Column("name"), "user_age", eb.Column("age"), "is_active", eb.Column("is_active"))
		}, "user_object").
		OrderBy("name").
		Scan(suite.ctx, &jsonObjectResults)
	suite.NoError(err)
	suite.True(len(jsonObjectResults) > 0, "Should have JSON object results")

	for _, result := range jsonObjectResults {
		suite.NotEmpty(result.UserObject, "User object should not be empty")
		suite.Contains(result.UserObject, result.Id, "JSON should contain user ID")
		suite.Contains(result.UserObject, result.Name, "JSON should contain user name")
		suite.T().Logf("User %s JSON: %s", result.Name, result.UserObject)
	}

	// Test 2: JSON_ARRAY creation
	type JsonArrayResult struct {
		PostId    string `bun:"post_id"`
		JsonArray string `bun:"json_array"`
	}

	var jsonArrayResults []JsonArrayResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("id").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONArray(eb.Column("title"), eb.Column("status"), eb.Column("view_count"))
		}, "json_array").
		OrderBy("title").
		Limit(3).
		Scan(suite.ctx, &jsonArrayResults)
	suite.NoError(err)
	suite.True(len(jsonArrayResults) > 0, "Should have JSON array results")

	for _, result := range jsonArrayResults {
		suite.NotEmpty(result.JsonArray, "JSON array should not be empty")
		suite.T().Logf("Post %s JSON Array: %s", result.PostId, result.JsonArray)
	}

	// Test 3: JSON path extraction
	type JsonExtractResult struct {
		Title       string `bun:"title"`
		MetaJson    string `bun:"meta_json"`
		ExtractedId string `bun:"extracted_id"`
	}

	var jsonExtractResults []JsonExtractResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONObject("post_id", eb.Column("id"), "title", eb.Column("title"), "views", eb.Column("view_count"), "status", eb.Column("status"))
		}, "meta_json").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONExtract(eb.JSONObject("id", eb.Column("id")), "$.id")
		}, "extracted_id").
		OrderBy("title").
		Limit(3).
		Scan(suite.ctx, &jsonExtractResults)
	suite.NoError(err)
	suite.True(len(jsonExtractResults) > 0, "Should have JSON extract results")

	for _, result := range jsonExtractResults {
		suite.NotEmpty(result.MetaJson, "Meta JSON should not be empty")
		suite.NotEmpty(result.ExtractedId, "Extracted ID should not be empty")
		suite.T().Logf("Post %s Meta: %s, Extracted ID: %s", result.Title, result.MetaJson, result.ExtractedId)
	}

	// Test 4: JSON validation and type checking
	type JsonValidationResult struct {
		Title    string `bun:"title"`
		JsonData string `bun:"json_data"`
		IsValid  bool   `bun:"is_valid"`
	}

	var jsonValidResults []JsonValidationResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count"), "status", eb.Column("status"))
		}, "json_data").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONValid(eb.JSONObject("test", eb.Column("title")))
		}, "is_valid").
		OrderBy("title").
		Limit(2).
		Scan(suite.ctx, &jsonValidResults)
	suite.NoError(err)
	suite.True(len(jsonValidResults) > 0, "Should have JSON validation results")

	for _, result := range jsonValidResults {
		suite.NotEmpty(result.JsonData, "JSON data should not be empty")
		suite.True(result.IsValid, "JSON should be valid")
		suite.T().Logf("Post %s: Valid: %t, Data: %s",
			result.Title, result.IsValid, result.JsonData)
	}

	// Test 5: JSON modification functions
	type JsonModifyResult struct {
		Title       string `bun:"title"`
		Original    string `bun:"original"`
		WithInsert  string `bun:"with_insert"`
		WithReplace string `bun:"with_replace"`
	}

	var jsonModifyResults []JsonModifyResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count"))
		}, "original").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONInsert(eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count")), "$.status", eb.Column("status"))
		}, "with_insert").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.JSONReplace(eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count")), "$.views", 9999)
		}, "with_replace").
		OrderBy("title").
		Limit(2).
		Scan(suite.ctx, &jsonModifyResults)
	suite.NoError(err)
	suite.True(len(jsonModifyResults) > 0, "Should have JSON modify results")

	for _, result := range jsonModifyResults {
		suite.NotEmpty(result.Original, "Original JSON should not be empty")
		suite.NotEmpty(result.WithInsert, "JSON with insert should not be empty")
		suite.NotEmpty(result.WithReplace, "JSON with replace should not be empty")
		suite.T().Logf("Post %s modifications:", result.Title)
		suite.T().Logf("  Original: %s", result.Original)
		suite.T().Logf("  With Insert: %s", result.WithInsert)
		suite.T().Logf("  With Replace: %s", result.WithReplace)
	}
}

// TestSelectWithDecodeExpressions tests expr.Decode functionality for conditional expressions
func (suite *SelectTestSuite) TestSelectWithDecodeExpressions() {
	suite.T().Logf("Testing Decode expressions for %s", suite.dbType)

	// Test 1: Simple DECODE for status mapping
	type DecodeStatusResult struct {
		Title      string `bun:"title"`
		Status     string `bun:"status"`
		StatusDesc string `bun:"status_desc"`
		StatusPrio int64  `bun:"status_priority"`
	}

	var decodeStatusResults []DecodeStatusResult
	err := suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title", "status").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Decode(eb.Column("status"), "published", "Published Article", "draft", "Draft Article", "review", "Under Review", "Unknown Status")
		}, "status_desc").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Decode(eb.Column("status"), "published", 1, "review", 2, "draft", 3, 99)
		}, "status_priority").
		OrderBy("title").
		Scan(suite.ctx, &decodeStatusResults)
	suite.NoError(err)
	suite.True(len(decodeStatusResults) > 0, "Should have decode status results")

	for _, result := range decodeStatusResults {
		suite.NotEmpty(result.StatusDesc, "Status description should not be empty")
		suite.True(result.StatusPrio > 0, "Status priority should be positive")

		// Verify mapping correctness
		switch result.Status {
		case "published":
			suite.Equal("Published Article", result.StatusDesc)
			suite.Equal(int64(1), result.StatusPrio)
		case "draft":
			suite.Equal("Draft Article", result.StatusDesc)
			suite.Equal(int64(3), result.StatusPrio)
		case "review":
			suite.Equal("Under Review", result.StatusDesc)
			suite.Equal(int64(2), result.StatusPrio)
		default:
			suite.Equal("Unknown Status", result.StatusDesc)
			suite.Equal(int64(99), result.StatusPrio)
		}

		suite.T().Logf("Post %s: %s -> %s (Priority: %d)",
			result.Title, result.Status, result.StatusDesc, result.StatusPrio)
	}

	// Test 2: Simple DECODE with Case expression
	type DecodeSimpleResult struct {
		Title     string `bun:"title"`
		ViewCount int    `bun:"view_count"`
		Category  string `bun:"category"`
	}

	var decodeSimpleResults []DecodeSimpleResult
	err = suite.db.NewSelect().
		Model((*Post)(nil)).
		Select("title", "view_count").
		SelectExpr(func(eb ExprBuilder) any {
			return eb.Case(func(cb CaseBuilder) {
				cb.When(func(cb ConditionBuilder) {
					cb.GreaterThan("view_count", 80)
				}).Then("Popular").
					When(func(cb ConditionBuilder) {
						cb.GreaterThan("view_count", 30)
					}).Then("Moderate").
					Else("Low")
			})
		}, "category").
		OrderBy("title").
		Scan(suite.ctx, &decodeSimpleResults)
	suite.NoError(err)
	suite.True(len(decodeSimpleResults) > 0, "Should have decode results")

	for _, result := range decodeSimpleResults {
		suite.NotEmpty(result.Category, "Category should not be empty")
		suite.T().Logf("Post %s: %d views -> %s", result.Title, result.ViewCount, result.Category)
	}
}
