package orm

// StringOperationsTestSuite tests string operation condition methods.
// Covers: Contains, StartsWith, EndsWith, ContainsAny, StartsWithAny, EndsWithAny
// and their case-insensitive variants (ContainsIgnoreCase, StartsWithIgnoreCase, etc.)
type StringOperationsTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestContains tests the Contains and OrContains conditions.
func (suite *StringOperationsTestSuite) TestContains() {
	suite.T().Logf("Testing Contains condition for %s", suite.dbType)

	suite.Run("BasicContains", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Contains("name", "Alice")
				}),
		)

		suite.Len(users, 1, "Should find one user")
		suite.Contains(users[0].Name, "Alice", "Name should contain Alice")

		suite.T().Logf("Found user: %s", users[0].Name)
	})

	suite.Run("OrContains", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Contains("name", "Alice").
						OrContains("name", "Bob")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestStartsWith tests the StartsWith and OrStartsWith conditions.
func (suite *StringOperationsTestSuite) TestStartsWith() {
	suite.T().Logf("Testing StartsWith condition for %s", suite.dbType)

	suite.Run("BasicStartsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.StartsWith("name", "Alice")
				}),
		)

		suite.Len(users, 1, "Should find one user")
		suite.True(len(users[0].Name) >= 5 && users[0].Name[:5] == "Alice",
			"Name should start with Alice")

		suite.T().Logf("Found user: %s", users[0].Name)
	})

	suite.Run("OrStartsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.StartsWith("name", "Alice").
						OrStartsWith("name", "Bob")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestEndsWith tests the EndsWith and OrEndsWith conditions.
func (suite *StringOperationsTestSuite) TestEndsWith() {
	suite.T().Logf("Testing EndsWith condition for %s", suite.dbType)

	suite.Run("BasicEndsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EndsWith("name", "Johnson")
				}),
		)

		suite.Len(users, 1, "Should find one user")
		suite.True(len(users[0].Name) >= 7 && users[0].Name[len(users[0].Name)-7:] == "Johnson",
			"Name should end with Johnson")

		suite.T().Logf("Found user: %s", users[0].Name)
	})

	suite.Run("OrEndsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EndsWith("name", "Johnson").
						OrEndsWith("name", "Smith")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestContainsIgnoreCase tests the ContainsIgnoreCase and OrContainsIgnoreCase conditions (case-insensitive).
func (suite *StringOperationsTestSuite) TestContainsIgnoreCase() {
	suite.T().Logf("Testing ContainsIgnoreCase condition for %s", suite.dbType)

	suite.Run("BasicIContains", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ContainsIgnoreCase("name", "alice")
				}),
		)

		suite.Len(users, 1, "Should find one user (case-insensitive)")

		suite.T().Logf("Found user: %s", users[0].Name)
	})

	suite.Run("OrIContains", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ContainsIgnoreCase("name", "alice").
						OrContainsIgnoreCase("name", "bob")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users (case-insensitive)")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestStartsWithIgnoreCase tests the StartsWithIgnoreCase and OrStartsWithIgnoreCase conditions (case-insensitive).
func (suite *StringOperationsTestSuite) TestStartsWithIgnoreCase() {
	suite.T().Logf("Testing StartsWithIgnoreCase condition for %s", suite.dbType)

	suite.Run("BasicIStartsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.StartsWithIgnoreCase("name", "alice")
				}),
		)

		suite.Len(users, 1, "Should find one user (case-insensitive)")

		suite.T().Logf("Found user: %s", users[0].Name)
	})

	suite.Run("OrIStartsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.StartsWithIgnoreCase("name", "alice").
						OrStartsWithIgnoreCase("name", "bob")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users (case-insensitive)")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestEndsWithIgnoreCase tests the EndsWithIgnoreCase and OrEndsWithIgnoreCase conditions (case-insensitive).
func (suite *StringOperationsTestSuite) TestEndsWithIgnoreCase() {
	suite.T().Logf("Testing EndsWithIgnoreCase condition for %s", suite.dbType)

	suite.Run("BasicIEndsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EndsWithIgnoreCase("name", "johnson")
				}),
		)

		suite.Len(users, 1, "Should find one user (case-insensitive)")

		suite.T().Logf("Found user: %s", users[0].Name)
	})

	suite.Run("OrIEndsWith", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EndsWithIgnoreCase("name", "johnson").
						OrEndsWithIgnoreCase("name", "smith")
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users (case-insensitive)")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestContainsAny tests the ContainsAny and OrContainsAny conditions.
func (suite *StringOperationsTestSuite) TestContainsAny() {
	suite.T().Logf("Testing ContainsAny condition for %s", suite.dbType)

	suite.Run("BasicContainsAny", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ContainsAny("name", []string{"Alice", "Bob"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrContainsAny", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.ContainsAny("name", []string{"Alice"}).
						OrContainsAny("name", []string{"Charlie"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestStartsWithAny tests the StartsWithAny and OrStartsWithAny conditions.
func (suite *StringOperationsTestSuite) TestStartsWithAny() {
	suite.T().Logf("Testing StartsWithAny condition for %s", suite.dbType)

	suite.Run("BasicStartsWithAny", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.StartsWithAny("name", []string{"Alice", "Bob"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrStartsWithAny", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.StartsWithAny("name", []string{"Alice"}).
						OrStartsWithAny("name", []string{"Charlie"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestEndsWithAny tests the EndsWithAny and OrEndsWithAny conditions.
func (suite *StringOperationsTestSuite) TestEndsWithAny() {
	suite.T().Logf("Testing EndsWithAny condition for %s", suite.dbType)

	suite.Run("BasicEndsWithAny", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EndsWithAny("name", []string{"Johnson", "Smith"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrEndsWithAny", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EndsWithAny("name", []string{"Johnson"}).
						OrEndsWithAny("name", []string{"Brown"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}
