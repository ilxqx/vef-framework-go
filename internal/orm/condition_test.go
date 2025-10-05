package orm

import (
	"time"
)

type ConditionTestSuite struct {
	*ORMTestSuite
}

// TestBasicConditions tests basic condition builders.
func (suite *ConditionTestSuite) TestBasicConditions() {
	suite.T().Logf("Testing basic conditions for %s", suite.dbType)

	// Test 1: Equals condition
	var users []User

	err := suite.db.NewSelect().
		Model(&users).
		Where(func(cb ConditionBuilder) {
			cb.Equals("name", "Alice Johnson")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(users, 1)
	suite.Equal("Alice Johnson", users[0].Name)

	// Test 2: NotEquals condition
	var notAlice []User

	err = suite.db.NewSelect().
		Model(&notAlice).
		Where(func(cb ConditionBuilder) {
			cb.NotEquals("name", "Alice Johnson")
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(notAlice, 2)

	for _, user := range notAlice {
		suite.NotEqual("Alice Johnson", user.Name)
	}

	// Test 3: GreaterThan condition
	var olderUsers []User

	err = suite.db.NewSelect().
		Model(&olderUsers).
		Where(func(cb ConditionBuilder) {
			cb.GreaterThan("age", 30)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(olderUsers, 1)
	suite.Equal("Charlie Brown", olderUsers[0].Name)
	suite.True(olderUsers[0].Age > 30)

	// Test 4: GreaterThanOrEqual condition
	var thirtyPlusUsers []User

	err = suite.db.NewSelect().
		Model(&thirtyPlusUsers).
		Where(func(cb ConditionBuilder) {
			cb.GreaterThanOrEqual("age", 30)
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(thirtyPlusUsers, 2) // Alice (30) and Charlie (35)

	for _, user := range thirtyPlusUsers {
		suite.True(user.Age >= 30)
	}

	// Test 5: LessThan condition
	var youngerUsers []User

	err = suite.db.NewSelect().
		Model(&youngerUsers).
		Where(func(cb ConditionBuilder) {
			cb.LessThan("age", 30)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(youngerUsers, 1)
	suite.Equal("Bob Smith", youngerUsers[0].Name)
	suite.True(youngerUsers[0].Age < 30)

	// Test 6: LessThanOrEqual condition
	var thirtyOrLessUsers []User

	err = suite.db.NewSelect().
		Model(&thirtyOrLessUsers).
		Where(func(cb ConditionBuilder) {
			cb.LessThanOrEqual("age", 30)
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(thirtyOrLessUsers, 2) // Bob (25) and Alice (30)

	for _, user := range thirtyOrLessUsers {
		suite.True(user.Age <= 30)
	}
}

// TestInAndNotInConditions tests IN and NOT IN conditions.
func (suite *ConditionTestSuite) TestInAndNotInConditions() {
	suite.T().Logf("Testing IN and NOT IN conditions for %s", suite.dbType)

	// Test 1: IN condition with string slice
	var specificUsers []User

	err := suite.db.NewSelect().
		Model(&specificUsers).
		Where(func(cb ConditionBuilder) {
			cb.In("email", []string{"alice@example.com", "bob@example.com"})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(specificUsers, 2)

	emails := make([]string, len(specificUsers))
	for i, user := range specificUsers {
		emails[i] = user.Email
	}

	suite.Contains(emails, "alice@example.com")
	suite.Contains(emails, "bob@example.com")

	// Test 2: IN condition with integer slice
	var specificAgeUsers []User

	err = suite.db.NewSelect().
		Model(&specificAgeUsers).
		Where(func(cb ConditionBuilder) {
			cb.In("age", []int16{25, 35})
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(specificAgeUsers, 2)
	suite.Equal(int16(25), specificAgeUsers[0].Age) // Bob
	suite.Equal(int16(35), specificAgeUsers[1].Age) // Charlie

	// Test 3: NOT IN condition
	var excludedUsers []User

	err = suite.db.NewSelect().
		Model(&excludedUsers).
		Where(func(cb ConditionBuilder) {
			cb.NotIn("email", []string{"charlie@example.com"})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(excludedUsers, 2)

	for _, user := range excludedUsers {
		suite.NotEqual("charlie@example.com", user.Email)
	}

	// Test 4: IN with subquery
	var usersWithPosts []User

	err = suite.db.NewSelect().
		Model(&usersWithPosts).
		Where(func(cb ConditionBuilder) {
			cb.InSubQuery("id", func(subquery SelectQuery) {
				subquery.Model((*Post)(nil)).
					Select("user_id").
					Distinct()
			})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.True(len(usersWithPosts) > 0, "Should have users with posts")

	// Test 5: NOT IN with subquery
	var usersWithoutDrafts []User

	err = suite.db.NewSelect().
		Model(&usersWithoutDrafts).
		Where(func(cb ConditionBuilder) {
			cb.NotInSubQuery("id", func(subquery SelectQuery) {
				subquery.Model((*Post)(nil)).
					Select("user_id").
					Where(func(cb ConditionBuilder) {
						cb.Equals("status", "draft")
					})
			})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	// Should have users who don't have any draft posts
}

// TestBetweenConditions tests BETWEEN and NOT BETWEEN conditions.
func (suite *ConditionTestSuite) TestBetweenConditions() {
	suite.T().Logf("Testing BETWEEN conditions for %s", suite.dbType)

	// Test 1: BETWEEN condition with integers
	var middleAgedUsers []User

	err := suite.db.NewSelect().
		Model(&middleAgedUsers).
		Where(func(cb ConditionBuilder) {
			cb.Between("age", 25, 30)
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(middleAgedUsers, 2) // Bob (25) and Alice (30)

	for _, user := range middleAgedUsers {
		suite.True(user.Age >= 25 && user.Age <= 30)
	}

	// Test 2: NOT BETWEEN condition
	var extremeAgeUsers []User

	err = suite.db.NewSelect().
		Model(&extremeAgeUsers).
		Where(func(cb ConditionBuilder) {
			cb.NotBetween("age", 26, 34)
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	// Should find Bob (25) and Charlie (35) but not Alice (30)
	for _, user := range extremeAgeUsers {
		suite.True(user.Age < 26 || user.Age > 34)
	}

	// Test 3: BETWEEN with view counts
	var moderateViewPosts []Post

	err = suite.db.NewSelect().
		Model(&moderateViewPosts).
		Where(func(cb ConditionBuilder) {
			cb.Between("view_count", 50, 100)
		}).
		OrderBy("view_count").
		Scan(suite.ctx)
	suite.NoError(err)

	for _, post := range moderateViewPosts {
		suite.True(post.ViewCount >= 50 && post.ViewCount <= 100)
	}
}

// TestStringConditions tests string-related conditions like LIKE, StartsWith, etc.
func (suite *ConditionTestSuite) TestStringConditions() {
	suite.T().Logf("Testing string conditions for %s", suite.dbType)

	// Test 1: Contains condition (LIKE %value%)
	var johnsonUsers []User

	err := suite.db.NewSelect().
		Model(&johnsonUsers).
		Where(func(cb ConditionBuilder) {
			cb.Contains("name", "Johnson")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(johnsonUsers, 1)
	suite.Equal("Alice Johnson", johnsonUsers[0].Name)

	// Test 2: StartsWith condition (LIKE value%)
	var aliceUsers []User

	err = suite.db.NewSelect().
		Model(&aliceUsers).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("name", "Alice")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(aliceUsers, 1)
	suite.Equal("Alice Johnson", aliceUsers[0].Name)

	// Test 3: EndsWith condition (LIKE %value)
	var brownUsers []User

	err = suite.db.NewSelect().
		Model(&brownUsers).
		Where(func(cb ConditionBuilder) {
			cb.EndsWith("name", "Brown")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(brownUsers, 1)
	suite.Equal("Charlie Brown", brownUsers[0].Name)

	// Test 4: NotContains condition (NOT LIKE %value%)
	var nonJohnsonUsers []User

	err = suite.db.NewSelect().
		Model(&nonJohnsonUsers).
		Where(func(cb ConditionBuilder) {
			cb.NotContains("name", "Johnson")
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(nonJohnsonUsers, 2)

	for _, user := range nonJohnsonUsers {
		suite.NotContains(user.Name, "Johnson")
	}

	// Test 5: ContainsAny condition
	var multiContainUsers []User

	err = suite.db.NewSelect().
		Model(&multiContainUsers).
		Where(func(cb ConditionBuilder) {
			cb.ContainsAny("name", []string{"Alice", "Bob"})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(multiContainUsers, 2) // Alice Johnson and Bob Smith

	// Test 6: StartsWithAny condition
	var multiStartUsers []User

	err = suite.db.NewSelect().
		Model(&multiStartUsers).
		Where(func(cb ConditionBuilder) {
			cb.StartsWithAny("name", []string{"Alice", "Charlie"})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(multiStartUsers, 2) // Alice Johnson and Charlie Brown

	// Test 7: Case insensitive contains
	var caseInsensitiveUsers []User

	err = suite.db.NewSelect().
		Model(&caseInsensitiveUsers).
		Where(func(cb ConditionBuilder) {
			cb.ContainsIgnoreCase("name", "ALICE")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(caseInsensitiveUsers, 1)
	suite.Equal("Alice Johnson", caseInsensitiveUsers[0].Name)
}

// TestNullConditions tests IS NULL and IS NOT NULL conditions.
func (suite *ConditionTestSuite) TestNullConditions() {
	suite.T().Logf("Testing NULL conditions for %s", suite.dbType)

	// Test 1: IS NOT NULL condition
	var postsWithDescription []Post

	err := suite.db.NewSelect().
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

	// Test 2: IS NULL condition
	var postsWithoutDescription []Post

	err = suite.db.NewSelect().
		Model(&postsWithoutDescription).
		Where(func(cb ConditionBuilder) {
			cb.IsNull("description")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	for _, post := range postsWithoutDescription {
		suite.Nil(post.Description, "Description should be null")
	}

	// Test 3: IS NOT NULL on category parent_id
	var subCategories []Category

	err = suite.db.NewSelect().
		Model(&subCategories).
		Where(func(cb ConditionBuilder) {
			cb.IsNotNull("parent_id")
		}).
		Scan(suite.ctx)
	suite.NoError(err)

	for _, category := range subCategories {
		suite.NotNil(category.ParentId, "Parent ID should not be null for subcategories")
	}

	// Test 4: IS NULL on category parent_id (root categories)
	var rootCategories []Category

	err = suite.db.NewSelect().
		Model(&rootCategories).
		Where(func(cb ConditionBuilder) {
			cb.IsNull("parent_id")
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)

	for _, category := range rootCategories {
		suite.Nil(category.ParentId, "Parent ID should be null for root categories")
	}
}

// TestBooleanConditions tests boolean conditions like IsTrue, IsFalse.
func (suite *ConditionTestSuite) TestBooleanConditions() {
	suite.T().Logf("Testing boolean conditions for %s", suite.dbType)

	// Test 1: IsTrue condition
	var activeUsers []User

	err := suite.db.NewSelect().
		Model(&activeUsers).
		Where(func(cb ConditionBuilder) {
			cb.IsTrue("is_active")
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(activeUsers, 2) // Alice and Bob are active

	for _, user := range activeUsers {
		suite.True(user.IsActive)
	}

	// Test 2: IsFalse condition
	var inactiveUsers []User

	err = suite.db.NewSelect().
		Model(&inactiveUsers).
		Where(func(cb ConditionBuilder) {
			cb.IsFalse("is_active")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(inactiveUsers, 1) // Charlie is inactive

	for _, user := range inactiveUsers {
		suite.False(user.IsActive)
	}

	// Test 3: Equals with boolean value (alternative approach)
	var explicitActiveUsers []User

	err = suite.db.NewSelect().
		Model(&explicitActiveUsers).
		Where(func(cb ConditionBuilder) {
			cb.Equals("is_active", true)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(explicitActiveUsers, 2)

	for _, user := range explicitActiveUsers {
		suite.True(user.IsActive)
	}
}

// TestLogicalOperators tests AND, OR conditions and grouping.
func (suite *ConditionTestSuite) TestLogicalOperators() {
	suite.T().Logf("Testing logical operators for %s", suite.dbType)

	// Test 1: Simple AND conditions (chained)
	var activeAdults []User

	err := suite.db.NewSelect().
		Model(&activeAdults).
		Where(func(cb ConditionBuilder) {
			cb.Equals("is_active", true).GreaterThanOrEqual("age", 30)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(activeAdults, 1) // Only Alice (active and 30)
	suite.Equal("Alice Johnson", activeAdults[0].Name)

	// Test 2: OR conditions
	var youngOrOld []User

	err = suite.db.NewSelect().
		Model(&youngOrOld).
		Where(func(cb ConditionBuilder) {
			cb.LessThanOrEqual("age", 25).OrGreaterThanOrEqual("age", 35)
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(youngOrOld, 2) // Bob (25) and Charlie (35)

	// Test 3: Complex grouping (A AND (B OR C))
	var complexCondition []User

	err = suite.db.NewSelect().
		Model(&complexCondition).
		Where(func(cb ConditionBuilder) {
			cb.Equals("is_active", true).Group(func(innerCb ConditionBuilder) {
				innerCb.Equals("age", 25).OrEquals("age", 30)
			})
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(complexCondition, 2) // Bob (active, 25) and Alice (active, 30)

	// Test 4: Nested groups ((A OR B) AND (C OR D))
	var nestedCondition []User

	err = suite.db.NewSelect().
		Model(&nestedCondition).
		Where(func(cb ConditionBuilder) {
			cb.Group(func(innerCb ConditionBuilder) {
				innerCb.Equals("name", "Alice Johnson").OrEquals("name", "Bob Smith")
			}).Group(func(innerCb ConditionBuilder) {
				innerCb.GreaterThan("age", 20).OrLessThan("age", 40)
			})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(nestedCondition, 2) // Alice and Bob (both match name criteria and age criteria)

	// Test 5: Mixed OR and AND with different precedence
	var mixedLogic []User

	err = suite.db.NewSelect().
		Model(&mixedLogic).
		Where(func(cb ConditionBuilder) {
			cb.Equals("is_active", true).
				OrGroup(func(innerCb ConditionBuilder) {
					innerCb.Equals("name", "Charlie Brown").GreaterThan("age", 30)
				})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	// Should match: active users OR (Charlie AND age > 30)
	// Result: Alice, Bob (active), and Charlie (name=Charlie AND age=35>30)
	suite.Len(mixedLogic, 3)
}

// TestColumnComparisons tests comparing columns with each other.
func (suite *ConditionTestSuite) TestColumnComparisons() {
	suite.T().Logf("Testing column comparisons for %s", suite.dbType)

	// Test 1: EqualsColumn - compare two columns
	var usersWithSameCreatedUpdated []User

	err := suite.db.NewSelect().
		Model(&usersWithSameCreatedUpdated).
		Where(func(cb ConditionBuilder) {
			cb.EqualsColumn("created_at", "updated_at")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	// From fixture data, all users have same created_at and updated_at
	suite.Len(usersWithSameCreatedUpdated, 3)

	// Test 2: NotEqualsColumn
	var usersWithDifferentCreatedUpdated []User

	err = suite.db.NewSelect().
		Model(&usersWithDifferentCreatedUpdated).
		Where(func(cb ConditionBuilder) {
			cb.NotEqualsColumn("created_at", "updated_at")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	// Should be empty since all users have same created_at and updated_at in fixture
	suite.Len(usersWithDifferentCreatedUpdated, 0)

	// Test 3: GreaterThanColumn comparison
	// Let's use a more realistic example with posts
	var postsViewCountVsId []Post

	err = suite.db.NewSelect().
		Model(&postsViewCountVsId).
		Where(func(cb ConditionBuilder) {
			// This is contrived but tests the functionality
			cb.GreaterThanColumn("view_count", "view_count") // Always false
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(postsViewCountVsId, 0) // Should be empty as no column is greater than itself
}

// TestExpressionConditions tests conditions with SQL expressions.
func (suite *ConditionTestSuite) TestExpressionConditions() {
	suite.T().Logf("Testing expression conditions for %s", suite.dbType)

	// Test 1: Expression comparing string length using CHAR_LENGTH(name) > CHAR_LENGTH('Alice')
	var usersWithLongNames []User

	err := suite.db.NewSelect().
		Model(&usersWithLongNames).
		Where(func(cb ConditionBuilder) {
			cb.Expr(func(eb ExprBuilder) any {
				return eb.Expr("? > ?", eb.CharLength(eb.Column("name")), eb.CharLength("Alice"))
			})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	// Should find users with names longer than "Alice" (5 characters)
	for _, user := range usersWithLongNames {
		suite.True(len(user.Name) > 5)
	}

	// Test 2: Expr - raw expression condition
	var youngActiveUsers []User

	err = suite.db.NewSelect().
		Model(&youngActiveUsers).
		Where(func(cb ConditionBuilder) {
			cb.Expr(func(eb ExprBuilder) any {
				return eb.Expr("age < ? AND is_active = ?", 30, true)
			})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(youngActiveUsers, 1) // Bob (25, active)

	// Test 3: Complex expression with calculations
	var calculatedCondition []User

	err = suite.db.NewSelect().
		Model(&calculatedCondition).
		Where(func(cb ConditionBuilder) {
			cb.GreaterThanExpr("age", func(eb ExprBuilder) any {
				return eb.Expr("? + ?", 20, 5) // age > 25
			})
		}).
		OrderBy("age").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(calculatedCondition, 2) // Alice (30) and Charlie (35)
}

// TestAuditConditions tests audit-related conditions (created_by, updated_by, etc.)
func (suite *ConditionTestSuite) TestAuditConditions() {
	suite.T().Logf("Testing audit conditions for %s", suite.dbType)

	// Test 1: CreatedByEquals
	var systemCreatedUsers []User

	err := suite.db.NewSelect().
		Model(&systemCreatedUsers).
		Where(func(cb ConditionBuilder) {
			cb.CreatedByEquals("system")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(systemCreatedUsers, 3) // All users created by system in fixture

	for _, user := range systemCreatedUsers {
		suite.Equal("system", user.CreatedBy)
	}

	// Test 2: CreatedByNotEquals
	var nonSystemUsers []User

	err = suite.db.NewSelect().
		Model(&nonSystemUsers).
		Where(func(cb ConditionBuilder) {
			cb.CreatedByNotEquals("system")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(nonSystemUsers, 0) // All users created by system

	// Test 3: UpdatedByEquals
	var systemUpdatedUsers []User

	err = suite.db.NewSelect().
		Model(&systemUpdatedUsers).
		Where(func(cb ConditionBuilder) {
			cb.UpdatedByEquals("system")
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(systemUpdatedUsers, 3) // All users updated by system

	// Test 4: CreatedByIn
	var multiCreatorUsers []User

	err = suite.db.NewSelect().
		Model(&multiCreatorUsers).
		Where(func(cb ConditionBuilder) {
			cb.CreatedByIn([]string{"system", "admin"})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(multiCreatorUsers, 3) // All users created by system (admin not in fixture)
}

// TestTimeConditions tests time-based conditions.
func (suite *ConditionTestSuite) TestTimeConditions() {
	suite.T().Logf("Testing time conditions for %s", suite.dbType)

	// Get current time for comparisons
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	// Test 1: CreatedAtGreaterThan
	var recentUsers []User

	err := suite.db.NewSelect().
		Model(&recentUsers).
		Where(func(cb ConditionBuilder) {
			cb.CreatedAtGreaterThan(pastTime)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	// All fixture users should be created recently (during test setup)
	suite.Len(recentUsers, 3)

	// Test 2: CreatedAtLessThan
	var futureUsers []User

	err = suite.db.NewSelect().
		Model(&futureUsers).
		Where(func(cb ConditionBuilder) {
			cb.CreatedAtLessThan(futureTime)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	// All users should be created before future time
	suite.Len(futureUsers, 3)

	// Test 3: CreatedAtBetween
	var timeBoundUsers []User

	err = suite.db.NewSelect().
		Model(&timeBoundUsers).
		Where(func(cb ConditionBuilder) {
			cb.CreatedAtBetween(pastTime, futureTime)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	// All users should be within this range
	suite.Len(timeBoundUsers, 3)

	// Test 4: UpdatedAtGreaterThan
	var recentlyUpdatedUsers []User

	err = suite.db.NewSelect().
		Model(&recentlyUpdatedUsers).
		Where(func(cb ConditionBuilder) {
			cb.UpdatedAtGreaterThan(pastTime)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(recentlyUpdatedUsers, 3)
}

// TestPrimaryKeyConditions tests PK-related conditions.
func (suite *ConditionTestSuite) TestPrimaryKeyConditions() {
	suite.T().Logf("Testing primary key conditions for %s", suite.dbType)

	// First get a user to test with
	var firstUser User

	err := suite.db.NewSelect().
		Model(&firstUser).
		OrderBy("name").
		Limit(1).
		Scan(suite.ctx)
	suite.NoError(err)

	// Test 1: PKEquals
	var userByPK []User

	err = suite.db.NewSelect().
		Model(&userByPK).
		Where(func(cb ConditionBuilder) {
			cb.PKEquals(firstUser.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(userByPK, 1)
	suite.Equal(firstUser.Id, userByPK[0].Id)

	// Test 2: PKNotEquals
	var otherUsers []User

	err = suite.db.NewSelect().
		Model(&otherUsers).
		Where(func(cb ConditionBuilder) {
			cb.PKNotEquals(firstUser.Id)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(otherUsers, 2) // Should have 2 other users

	for _, user := range otherUsers {
		suite.NotEqual(firstUser.Id, user.Id)
	}

	// Get IDs for IN test
	var allUsers []User

	err = suite.db.NewSelect().
		Model(&allUsers).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)

	// Test 3: PKIn
	var twoUsers []User

	err = suite.db.NewSelect().
		Model(&twoUsers).
		Where(func(cb ConditionBuilder) {
			cb.PKIn([]string{allUsers[0].Id, allUsers[1].Id})
		}).
		OrderBy("name").
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(twoUsers, 2)

	// Test 4: PKNotIn
	var excludedUser []User

	err = suite.db.NewSelect().
		Model(&excludedUser).
		Where(func(cb ConditionBuilder) {
			cb.PKNotIn([]string{allUsers[0].Id, allUsers[1].Id})
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(excludedUser, 1) // Should be the third user
	suite.Equal(allUsers[2].Id, excludedUser[0].Id)
}

// TestApplyConditions tests the Apply and ApplyIf functionality.
func (suite *ConditionTestSuite) TestApplyConditions() {
	suite.T().Logf("Testing Apply conditions for %s", suite.dbType)

	// Define some reusable condition functions
	activeFilter := func(cb ConditionBuilder) {
		cb.Equals("is_active", true)
	}

	ageFilter := func(cb ConditionBuilder) {
		cb.GreaterThan("age", 25)
	}

	// Test 1: Apply single condition
	var activeUsers []User

	err := suite.db.NewSelect().
		Model(&activeUsers).
		Where(func(cb ConditionBuilder) {
			cb.Apply(activeFilter)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(activeUsers, 2) // Alice and Bob are active

	for _, user := range activeUsers {
		suite.True(user.IsActive)
	}

	// Test 2: Apply multiple conditions
	var activeOlderUsers []User

	err = suite.db.NewSelect().
		Model(&activeOlderUsers).
		Where(func(cb ConditionBuilder) {
			cb.Apply(activeFilter, ageFilter)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(activeOlderUsers, 1) // Only Alice (active and age > 25)
	suite.Equal("Alice Johnson", activeOlderUsers[0].Name)

	// Test 3: ApplyIf with true condition
	var conditionalUsers []User

	includeAgeFilter := true
	err = suite.db.NewSelect().
		Model(&conditionalUsers).
		Where(func(cb ConditionBuilder) {
			cb.Apply(activeFilter).ApplyIf(includeAgeFilter, ageFilter)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(conditionalUsers, 1) // Same as above since condition is true

	// Test 4: ApplyIf with false condition
	var conditionalUsers2 []User

	includeAgeFilter = false
	err = suite.db.NewSelect().
		Model(&conditionalUsers2).
		Where(func(cb ConditionBuilder) {
			cb.Apply(activeFilter).ApplyIf(includeAgeFilter, ageFilter)
		}).
		Scan(suite.ctx)
	suite.NoError(err)
	suite.Len(conditionalUsers2, 2) // Both active users since age filter not applied
}
