package orm

// RangeSetOperationsTestSuite tests range and set operation condition methods.
// Covers: Between, NotBetween, BetweenExpr, NotBetweenExpr, In, NotIn, InExpr, NotInExpr.
type RangeSetOperationsTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestBetween tests the Between and OrBetween conditions.
func (suite *RangeSetOperationsTestSuite) TestBetween() {
	suite.T().Logf("Testing Between condition for %s", suite.dbType)

	suite.Run("BasicBetween", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Between("age", 25, 30)
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.True(user.Age >= 25 && user.Age <= 30, "Age should be between 25 and 30")
		}

		suite.T().Logf("Found users with ages: %d, %d", users[0].Age, users[1].Age)
	})

	suite.Run("OrBetween", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Between("age", 25, 26).
						OrBetween("age", 34, 36)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True((user.Age >= 25 && user.Age <= 26) || (user.Age >= 34 && user.Age <= 36),
				"Age should match Between conditions")
		}

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestNotBetween tests the NotBetween and OrNotBetween conditions.
func (suite *RangeSetOperationsTestSuite) TestNotBetween() {
	suite.T().Logf("Testing NotBetween condition for %s", suite.dbType)

	suite.Run("BasicNotBetween", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotBetween("age", 26, 34)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age < 26 || user.Age > 34, "Age should not be between 26 and 34")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrNotBetween", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotBetween("age", 26, 29).
						OrNotBetween("age", 31, 34)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestBetweenExpr tests the BetweenExpr and OrBetweenExpr conditions.
func (suite *RangeSetOperationsTestSuite) TestBetweenExpr() {
	suite.T().Logf("Testing BetweenExpr condition for %s", suite.dbType)

	suite.Run("BasicBetweenExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.BetweenExpr("age",
						func(eb ExprBuilder) any {
							return eb.Expr("?", 25)
						},
						func(eb ExprBuilder) any {
							return eb.Expr("?", 30)
						},
					)
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.True(user.Age >= 25 && user.Age <= 30, "Age should be between 25 and 30")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrBetweenExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.BetweenExpr("age",
						func(eb ExprBuilder) any { return eb.Expr("?", 25) },
						func(eb ExprBuilder) any { return eb.Expr("?", 26) },
					).OrBetweenExpr("age",
						func(eb ExprBuilder) any { return eb.Expr("?", 34) },
						func(eb ExprBuilder) any { return eb.Expr("?", 36) },
					)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestNotBetweenExpr tests the NotBetweenExpr and OrNotBetweenExpr conditions.
func (suite *RangeSetOperationsTestSuite) TestNotBetweenExpr() {
	suite.T().Logf("Testing NotBetweenExpr condition for %s", suite.dbType)

	suite.Run("BasicNotBetweenExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotBetweenExpr("age",
						func(eb ExprBuilder) any { return eb.Expr("?", 26) },
						func(eb ExprBuilder) any { return eb.Expr("?", 34) },
					)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age < 26 || user.Age > 34, "Age should not be between 26 and 34")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrNotBetweenExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotBetweenExpr("age",
						func(eb ExprBuilder) any { return eb.Expr("?", 26) },
						func(eb ExprBuilder) any { return eb.Expr("?", 29) },
					).OrNotBetweenExpr("age",
						func(eb ExprBuilder) any { return eb.Expr("?", 31) },
						func(eb ExprBuilder) any { return eb.Expr("?", 34) },
					)
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestIn tests the In and OrIn conditions.
func (suite *RangeSetOperationsTestSuite) TestIn() {
	suite.T().Logf("Testing In condition for %s", suite.dbType)

	suite.Run("BasicInWithStrings", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.In("email", []string{"alice@example.com", "bob@example.com"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")
		emails := []string{users[0].Email, users[1].Email}
		suite.Contains(emails, "alice@example.com")
		suite.Contains(emails, "bob@example.com")

		suite.T().Logf("Found users: %s, %s", users[0].Name, users[1].Name)
	})

	suite.Run("BasicInWithIntegers", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.In("age", []int16{25, 35})
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")
		suite.Equal(int16(25), users[0].Age)
		suite.Equal(int16(35), users[1].Age)

		suite.T().Logf("Found users with ages: %d, %d", users[0].Age, users[1].Age)
	})

	suite.Run("OrIn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.In("age", []int16{25}).
						OrIn("age", []int16{35})
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestNotIn tests the NotIn and OrNotIn conditions.
func (suite *RangeSetOperationsTestSuite) TestNotIn() {
	suite.T().Logf("Testing NotIn condition for %s", suite.dbType)

	suite.Run("BasicNotIn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotIn("email", []string{"charlie@example.com"})
				}).
				OrderBy("name"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.NotEqual("charlie@example.com", user.Email)
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrNotIn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotIn("age", []int16{25}).
						OrNotIn("age", []int16{30})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestInExpr tests the InExpr and OrInExpr conditions.
func (suite *RangeSetOperationsTestSuite) TestInExpr() {
	suite.T().Logf("Testing InExpr condition for %s", suite.dbType)

	suite.Run("BasicInExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.InExpr("age", func(eb ExprBuilder) any {
						return eb.Expr("?, ?", 25, 35)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrInExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.InExpr("age", func(eb ExprBuilder) any {
						return eb.Expr("?", 25)
					}).OrInExpr("age", func(eb ExprBuilder) any {
						return eb.Expr("?", 35)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestNotInExpr tests the NotInExpr and OrNotInExpr conditions.
func (suite *RangeSetOperationsTestSuite) TestNotInExpr() {
	suite.T().Logf("Testing NotInExpr condition for %s", suite.dbType)

	suite.Run("BasicNotInExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotInExpr("age", func(eb ExprBuilder) any {
						return eb.Expr("(?)", 30)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.NotEqual(int16(30), user.Age)
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrNotInExpr", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotInExpr("age", func(eb ExprBuilder) any {
						return eb.Expr("(?)", 25)
					}).OrNotInExpr("age", func(eb ExprBuilder) any {
						return eb.Expr("(?)", 30)
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}
