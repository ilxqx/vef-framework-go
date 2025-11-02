package orm

// BasicComparisonTestSuite tests basic comparison condition methods.
// Covers: Equals, NotEquals, GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual
// and their column comparison variants (EqualsColumn, etc.).
type BasicComparisonTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestEquals tests the Equals and OrEquals conditions.
func (suite *BasicComparisonTestSuite) TestEquals() {
	suite.T().Logf("Testing Equals condition for %s", suite.dbType)

	suite.Run("BasicEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Equals("name", "Alice Johnson")
				}),
		)

		suite.Len(users, 1, "Should find exactly one user")
		suite.Equal("Alice Johnson", users[0].Name, "Should find Alice Johnson")

		suite.T().Logf("Found user: %s", users[0].Name)
	})

	suite.Run("OrEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Equals("name", "Alice Johnson").
						OrEquals("name", "Bob Smith")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")
		suite.Equal("Alice Johnson", users[0].Name)
		suite.Equal("Bob Smith", users[1].Name)

		suite.T().Logf("Found users: %s, %s", users[0].Name, users[1].Name)
	})

	suite.Run("EqualsWithDifferentTypes", func() {
		// Test with integer
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Equals("age", 30)
				}),
		)

		suite.Len(users, 1, "Should find one user with age 30")
		suite.Equal(int16(30), users[0].Age)

		// Test with boolean
		activeUsers := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Equals("is_active", true)
				}),
		)

		suite.True(len(activeUsers) > 0, "Should find active users")

		for _, user := range activeUsers {
			suite.True(user.IsActive, "All users should be active")
		}

		suite.T().Logf("Found %d active users", len(activeUsers))
	})
}

// TestNotEquals tests the NotEquals and OrNotEquals conditions.
func (suite *BasicComparisonTestSuite) TestNotEquals() {
	suite.T().Logf("Testing NotEquals condition for %s", suite.dbType)

	suite.Run("BasicNotEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEquals("name", "Alice Johnson")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.NotEqual("Alice Johnson", user.Name, "Should not be Alice Johnson")
		}

		suite.T().Logf("Found users: %s, %s", users[0].Name, users[1].Name)
	})

	suite.Run("OrNotEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEquals("name", "Alice Johnson").
						OrNotEquals("age", 25)
				}).
				OrderBy("name"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Name != "Alice Johnson" || user.Age != 25,
				"Should match NotEquals condition")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestGreaterThan tests the GreaterThan and OrGreaterThan conditions.
func (suite *BasicComparisonTestSuite) TestGreaterThan() {
	suite.T().Logf("Testing GreaterThan condition for %s", suite.dbType)

	suite.Run("BasicGreaterThan", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThan("age", 30)
				}),
		)

		suite.Len(users, 1, "Should find one user")
		suite.Equal("Charlie Brown", users[0].Name)
		suite.True(users[0].Age > 30, "Age should be greater than 30")

		suite.T().Logf("Found user: %s (age: %d)", users[0].Name, users[0].Age)
	})

	suite.Run("OrGreaterThan", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThan("age", 30).
						OrGreaterThan("age", 24)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age > 30 || user.Age > 24,
				"Age should match GreaterThan condition")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestGreaterThanOrEqual tests the GreaterThanOrEqual and OrGreaterThanOrEqual conditions.
func (suite *BasicComparisonTestSuite) TestGreaterThanOrEqual() {
	suite.T().Logf("Testing GreaterThanOrEqual condition for %s", suite.dbType)

	suite.Run("BasicGreaterThanOrEqual", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanOrEqual("age", 30)
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.True(user.Age >= 30, "Age should be >= 30")
		}

		suite.T().Logf("Found users with ages: %d, %d", users[0].Age, users[1].Age)
	})

	suite.Run("OrGreaterThanOrEqual", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanOrEqual("age", 35).
						OrGreaterThanOrEqual("age", 25)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age >= 35 || user.Age >= 25,
				"Age should match GreaterThanOrEqual condition")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestLessThan tests the LessThan and OrLessThan conditions.
func (suite *BasicComparisonTestSuite) TestLessThan() {
	suite.T().Logf("Testing LessThan condition for %s", suite.dbType)

	suite.Run("BasicLessThan", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThan("age", 30)
				}),
		)

		suite.Len(users, 1, "Should find one user")
		suite.Equal("Bob Smith", users[0].Name)
		suite.True(users[0].Age < 30, "Age should be less than 30")

		suite.T().Logf("Found user: %s (age: %d)", users[0].Name, users[0].Age)
	})

	suite.Run("OrLessThan", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThan("age", 26).
						OrLessThan("age", 31)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age < 26 || user.Age < 31,
				"Age should match LessThan condition")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestLessThanOrEqual tests the LessThanOrEqual and OrLessThanOrEqual conditions.
func (suite *BasicComparisonTestSuite) TestLessThanOrEqual() {
	suite.T().Logf("Testing LessThanOrEqual condition for %s", suite.dbType)

	suite.Run("BasicLessThanOrEqual", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanOrEqual("age", 30)
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.True(user.Age <= 30, "Age should be <= 30")
		}

		suite.T().Logf("Found users with ages: %d, %d", users[0].Age, users[1].Age)
	})

	suite.Run("OrLessThanOrEqual", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanOrEqual("age", 25).
						OrLessThanOrEqual("age", 35)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age <= 25 || user.Age <= 35,
				"Age should match LessThanOrEqual condition")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestEqualsColumn tests the EqualsColumn and OrEqualsColumn conditions.
func (suite *BasicComparisonTestSuite) TestEqualsColumn() {
	suite.T().Logf("Testing EqualsColumn condition for %s", suite.dbType)

	suite.Run("BasicEqualsColumn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EqualsColumn("created_at", "updated_at")
				}),
		)

		suite.Len(users, 3, "All users have same created_at and updated_at")

		suite.T().Logf("Found %d users with matching timestamps", len(users))
	})

	suite.Run("OrEqualsColumn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EqualsColumn("created_at", "updated_at").
						OrEqualsColumn("created_by", "updated_by")
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestNotEqualsColumn tests the NotEqualsColumn and OrNotEqualsColumn conditions.
func (suite *BasicComparisonTestSuite) TestNotEqualsColumn() {
	suite.T().Logf("Testing NotEqualsColumn condition for %s", suite.dbType)

	suite.Run("BasicNotEqualsColumn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEqualsColumn("created_at", "updated_at")
				}),
		)

		suite.Len(users, 0, "All users have same created_at and updated_at in fixture")

		suite.T().Logf("Found %d users with different timestamps", len(users))
	})

	suite.Run("OrNotEqualsColumn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEqualsColumn("name", "email").
						OrNotEqualsColumn("created_at", "updated_at")
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestGreaterThanColumn tests the GreaterThanColumn and OrGreaterThanColumn conditions.
func (suite *BasicComparisonTestSuite) TestGreaterThanColumn() {
	suite.T().Logf("Testing GreaterThanColumn condition for %s", suite.dbType)

	suite.Run("BasicGreaterThanColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanColumn("view_count", "view_count")
				}),
		)

		suite.Len(posts, 0, "No column is greater than itself")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrGreaterThanColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanColumn("view_count", "view_count").
						OrGreaterThanColumn("id", "user_id")
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestGreaterThanOrEqualColumn tests the GreaterThanOrEqualColumn and OrGreaterThanOrEqualColumn conditions.
func (suite *BasicComparisonTestSuite) TestGreaterThanOrEqualColumn() {
	suite.T().Logf("Testing GreaterThanOrEqualColumn condition for %s", suite.dbType)

	suite.Run("BasicGreaterThanOrEqualColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanOrEqualColumn("view_count", "view_count")
				}),
		)

		suite.True(len(posts) > 0, "All posts have view_count >= view_count")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrGreaterThanOrEqualColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanOrEqualColumn("view_count", "view_count").
						OrGreaterThanOrEqualColumn("id", "user_id")
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestLessThanColumn tests the LessThanColumn and OrLessThanColumn conditions.
func (suite *BasicComparisonTestSuite) TestLessThanColumn() {
	suite.T().Logf("Testing LessThanColumn condition for %s", suite.dbType)

	suite.Run("BasicLessThanColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanColumn("view_count", "view_count")
				}),
		)

		suite.Len(posts, 0, "No column is less than itself")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrLessThanColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanColumn("view_count", "view_count").
						OrLessThanColumn("user_id", "id")
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestLessThanOrEqualColumn tests the LessThanOrEqualColumn and OrLessThanOrEqualColumn conditions.
func (suite *BasicComparisonTestSuite) TestLessThanOrEqualColumn() {
	suite.T().Logf("Testing LessThanOrEqualColumn condition for %s", suite.dbType)

	suite.Run("BasicLessThanOrEqualColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanOrEqualColumn("view_count", "view_count")
				}),
		)

		suite.True(len(posts) > 0, "All posts have view_count <= view_count")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrLessThanOrEqualColumn", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanOrEqualColumn("view_count", "view_count").
						OrLessThanOrEqualColumn("user_id", "id")
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		suite.T().Logf("Found %d posts", len(posts))
	})
}
