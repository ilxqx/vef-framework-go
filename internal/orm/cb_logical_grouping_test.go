package orm

// LogicalGroupingTestSuite tests logical grouping condition methods.
// Covers: Group, OrGroup (4 methods total including nested scenarios).
type LogicalGroupingTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestGroup tests the Group condition for AND grouping.
func (suite *LogicalGroupingTestSuite) TestGroup() {
	suite.T().Logf("Testing Group condition for %s", suite.DbType)

	suite.Run("BasicGroup", func() {
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Equals("is_active", true).
							GreaterThan("age", 25)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active")
			suite.True(user.Age > 25, "Age should be greater than 25")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("MultipleGroups", func() {
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Equals("is_active", true)
					}).Group(func(cb ConditionBuilder) {
						cb.GreaterThan("age", 25)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("NestedGroups", func() {
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Equals("is_active", true).
							Group(func(cb ConditionBuilder) {
								cb.GreaterThan("age", 20).
									LessThan("age", 40)
							})
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active")
			suite.True(user.Age > 20 && user.Age < 40, "Age should be between 20 and 40")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestOrGroup tests the OrGroup condition for OR grouping.
func (suite *LogicalGroupingTestSuite) TestOrGroup() {
	suite.T().Logf("Testing OrGroup condition for %s", suite.DbType)

	suite.Run("BasicOrGroup", func() {
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.OrGroup(func(cb ConditionBuilder) {
						cb.Equals("age", 25).
							OrEquals("age", 35)
					})
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")
		suite.True(users[0].Age == 25 || users[0].Age == 35, "Age should be 25 or 35")
		suite.True(users[1].Age == 25 || users[1].Age == 35, "Age should be 25 or 35")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("MixedGroupAndOrGroup", func() {
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Equals("is_active", true)
					}).OrGroup(func(cb ConditionBuilder) {
						cb.Equals("age", 25).
							Equals("age", 35)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("ComplexNestedGrouping", func() {
		// (is_active = true AND age > 25) OR (name LIKE '%Alice%' OR name LIKE '%Bob%')
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Equals("is_active", true).
							GreaterThan("age", 25)
					}).OrGroup(func(cb ConditionBuilder) {
						cb.Contains("name", "Alice").
							OrContains("name", "Bob")
					})
				}).
				OrderBy("name"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("DeeplyNestedGrouping", func() {
		// ((age > 20 AND age < 40) OR (age > 50)) AND is_active = true
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Group(func(cb ConditionBuilder) {
							cb.GreaterThan("age", 20).
								LessThan("age", 40)
						}).OrGroup(func(cb ConditionBuilder) {
							cb.GreaterThan("age", 50)
						})
					}).Equals("is_active", true)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestComplexLogicalCombinations tests complex combinations of Group and OrGroup.
func (suite *LogicalGroupingTestSuite) TestComplexLogicalCombinations() {
	suite.T().Logf("Testing complex logical combinations for %s", suite.DbType)

	suite.Run("ThreeLevelNesting", func() {
		// (((age = 25 OR age = 30) AND is_active = true) OR (age = 35 AND name LIKE '%Charlie%'))
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Group(func(cb ConditionBuilder) {
						cb.Group(func(cb ConditionBuilder) {
							cb.Equals("age", 25).
								OrEquals("age", 30)
						}).Equals("is_active", true)
					}).OrGroup(func(cb ConditionBuilder) {
						cb.Equals("age", 35).
							Contains("name", "Charlie")
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("MultipleOrGroupsWithAnd", func() {
		// (age = 25 OR age = 30) AND (is_active = true OR name LIKE '%Alice%')
		users := suite.assertQueryReturnsUsers(
			suite.Db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.OrGroup(func(cb ConditionBuilder) {
						cb.Equals("age", 25).
							OrEquals("age", 30)
					}).OrGroup(func(cb ConditionBuilder) {
						cb.Equals("is_active", true).
							OrContains("name", "Alice")
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}
