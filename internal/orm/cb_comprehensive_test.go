package orm

// ConditionComprehensiveTestSuite tests complex condition scenarios and comprehensive patterns.
// Covers: ApplyIf, complex combinations, edge cases, and real-world usage patterns.
type ConditionComprehensiveTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestApplyIf tests the ApplyIf method for conditional application of conditions.
func (suite *ConditionComprehensiveTestSuite) TestApplyIf() {
	suite.T().Logf("Testing ApplyIf method for %s", suite.DbType)

	suite.Run("BasicApplyIf", func() {
		applyCondition := true

		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ApplyIf(applyCondition, func(cb ConditionBuilder) {
						cb.Equals("is_active", true)
					})
				}),
		)

		suite.True(len(users) > 0, "Should find active users")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("ApplyNotApplied", func() {
		applyCondition := false

		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ApplyIf(applyCondition, func(cb ConditionBuilder) {
						cb.Equals("is_active", false)
					})
				}),
		)

		suite.Len(users, 3, "Should find all users (condition not applied)")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("MultipleApply", func() {
		applyAge := true
		applyActive := false

		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ApplyIf(applyAge, func(cb ConditionBuilder) {
						cb.GreaterThan("age", 25)
					}).ApplyIf(applyActive, func(cb ConditionBuilder) {
						cb.Equals("is_active", true)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age > 25, "Age should be greater than 25")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestComplexConditionCombinations tests complex real-world condition scenarios.
func (suite *ConditionComprehensiveTestSuite) TestComplexConditionCombinations() {
	suite.T().Logf("Testing complex condition combinations for %s", suite.DbType)

	suite.Run("SearchWithMultipleFilters", func() {
		// Simulate a search with multiple optional filters
		searchName := "Alice"
		minAge := 20
		maxAge := 40
		isActive := true

		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ApplyIf(searchName != "", func(cb ConditionBuilder) {
						cb.Contains("name", searchName)
					}).ApplyIf(minAge > 0, func(cb ConditionBuilder) {
						cb.GreaterThanOrEqual("age", minAge)
					}).ApplyIf(maxAge > 0, func(cb ConditionBuilder) {
						cb.LessThanOrEqual("age", maxAge)
					}).Equals("is_active", isActive)
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("ComplexOrConditionsWithGrouping", func() {
		// (name LIKE '%Alice%' OR name LIKE '%Bob%') AND (age > 20 AND age < 40) AND is_active = true
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.OrGroup(func(cb ConditionBuilder) {
						cb.Contains("name", "Alice").
							OrContains("name", "Bob")
					}).Group(func(cb ConditionBuilder) {
						cb.GreaterThan("age", 20).
							LessThan("age", 40)
					}).Equals("is_active", true)
				}).
				OrderBy("name"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("SubqueryWithComplexConditions", func() {
		// Find posts by active users with age > 25
		posts := suite.assertQueryReturnsPosts(
			suite.Db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.InSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("is_active", true).
									GreaterThan("age", 25)
							})
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestEdgeCases tests edge cases and boundary conditions.
func (suite *ConditionComprehensiveTestSuite) TestEdgeCases() {
	suite.T().Logf("Testing edge cases for %s", suite.DbType)

	suite.Run("EmptyInList", func() {
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.In("age", []int16{})
				}),
		)

		suite.Len(users, 0, "Should find no users with empty IN list")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("NullComparisons", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.Db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsNull("description").
						OrIsNotNull("description")
				}),
		)

		suite.True(len(posts) > 0, "Should find all posts")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("ChainedOrConditions", func() {
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Equals("age", 25).
						OrEquals("age", 30).
						OrEquals("age", 35)
				}).
				OrderBy("age"),
		)

		suite.Len(users, 3, "Should find three users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("MixedAndOrConditions", func() {
		// age = 25 OR (age = 30 AND is_active = true)
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Equals("age", 25).
						OrGroup(func(cb ConditionBuilder) {
							cb.Equals("age", 30).
								Equals("is_active", true)
						})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestPerformanceScenarios tests performance-related scenarios.
func (suite *ConditionComprehensiveTestSuite) TestPerformanceScenarios() {
	suite.T().Logf("Testing performance scenarios for %s", suite.DbType)

	suite.Run("ManyConditions", func() {
		// Test with many conditions to ensure no performance degradation
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsNotNull("id").
						IsNotNull("name").
						IsNotNull("email").
						IsNotNull("age").
						IsNotNull("is_active").
						IsNotNull("created_at").
						IsNotNull("updated_at").
						IsNotNull("created_by").
						IsNotNull("updated_by")
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("DeeplyNestedConditions", func() {
		// Test deeply nested conditions
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Group(func(cb ConditionBuilder) {
							cb.Group(func(cb ConditionBuilder) {
								cb.Equals("is_active", true)
							})
						})
					})
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}
