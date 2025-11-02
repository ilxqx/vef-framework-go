package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
)

// SubqueryOperationsTestSuite tests subquery operation condition methods.
// Covers: InSubQuery, NotInSubQuery, EqualsSubQuery, NotEqualsSubQuery, GreaterThanSubQuery, etc.
// Also covers: Any, All, Exists, NotExists variants.
type SubqueryOperationsTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestInSubQuery tests the InSubQuery and OrInSubQuery conditions.
func (suite *SubqueryOperationsTestSuite) TestInSubQuery() {
	suite.T().Logf("Testing InSubQuery condition for %s", suite.dbType)

	suite.Run("BasicInSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.InSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("is_active", true)
							})
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts from active users")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrInSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.InSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("age", 30)
							})
					}).OrInSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("age", 35)
							})
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestNotInSubQuery tests the NotInSubQuery and OrNotInSubQuery conditions.
func (suite *SubqueryOperationsTestSuite) TestNotInSubQuery() {
	suite.T().Logf("Testing NotInSubQuery condition for %s", suite.dbType)

	suite.Run("BasicNotInSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotInSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("is_active", false)
							})
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts not from inactive users")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrNotInSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotInSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("age", 25)
							})
					}).OrNotInSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("age", 30)
							})
					})
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestEqualsSubQuery tests the EqualsSubQuery and OrEqualsSubQuery conditions.
func (suite *SubqueryOperationsTestSuite) TestEqualsSubQuery() {
	suite.T().Logf("Testing EqualsSubQuery condition for %s", suite.dbType)

	suite.Run("BasicEqualsSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EqualsSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("name", "Alice Johnson")
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts by Alice")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrEqualsSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EqualsSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("name", "Alice Johnson")
							}).
							Limit(1)
					}).OrEqualsSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("name", "Bob Smith")
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestNotEqualsSubQuery tests the NotEqualsSubQuery and OrNotEqualsSubQuery conditions.
func (suite *SubqueryOperationsTestSuite) TestNotEqualsSubQuery() {
	suite.T().Logf("Testing NotEqualsSubQuery condition for %s", suite.dbType)

	suite.Run("BasicNotEqualsSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEqualsSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("name", "Alice Johnson")
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrNotEqualsSubQuery", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEqualsSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("name", "Alice Johnson")
							}).
							Limit(1)
					}).OrNotEqualsSubQuery("user_id", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.Equals("name", "Bob Smith")
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestGreaterThanSubQuery tests the GreaterThanSubQuery and OrGreaterThanSubQuery conditions.
func (suite *SubqueryOperationsTestSuite) TestGreaterThanSubQuery() {
	suite.T().Logf("Testing GreaterThanSubQuery condition for %s", suite.dbType)

	suite.Run("BasicGreaterThanSubQuery", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanSubQuery("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 25)
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age > 25, "Age should be greater than 25")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrGreaterThanSubQuery", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.GreaterThanSubQuery("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 40)
							}).
							Limit(1)
					}).OrGreaterThanSubQuery("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 24)
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestLessThanSubQuery tests the LessThanSubQuery and OrLessThanSubQuery conditions.
func (suite *SubqueryOperationsTestSuite) TestLessThanSubQuery() {
	suite.T().Logf("Testing LessThanSubQuery condition for %s", suite.dbType)

	suite.Run("BasicLessThanSubQuery", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanSubQuery("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 30)
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		for _, user := range users {
			suite.True(user.Age < 30, "Age should be less than 30")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrLessThanSubQuery", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.LessThanSubQuery("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 26)
							}).
							Limit(1)
					}).OrLessThanSubQuery("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 31)
							}).
							Limit(1)
					})
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestEqualsAll tests the EqualsAll and OrEqualsAll conditions.
// Note: SQLite does not support the ALL operator in subqueries (SQL standard feature).
func (suite *SubqueryOperationsTestSuite) TestEqualsAll() {
	suite.T().Logf("Testing EqualsAll condition for %s", suite.dbType)

	// Skip on SQLite - ALL operator not supported
	if suite.dbType == constants.DbSQLite {
		suite.T().Skipf("ALL operator not supported on %s (SQL standard feature)", suite.dbType)

		return
	}

	suite.Run("BasicEqualsAll", func() {
		// Find users whose age equals all ages in a subquery returning a single value
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EqualsAll("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 30)
							})
					})
				}),
		)

		suite.Len(users, 1, "Should find one user with age 30")
		suite.Equal(int16(30), users[0].Age)

		suite.T().Logf("Found user: %s (age: %d)", users[0].Name, users[0].Age)
	})

	suite.Run("OrEqualsAll", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.EqualsAll("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 25)
							})
					}).OrEqualsAll("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 35)
							})
					})
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestNotEqualsAll tests the NotEqualsAll and OrNotEqualsAll conditions.
// Note: SQLite does not support the ALL operator in subqueries (SQL standard feature).
func (suite *SubqueryOperationsTestSuite) TestNotEqualsAll() {
	suite.T().Logf("Testing NotEqualsAll condition for %s", suite.dbType)

	// Skip on SQLite - ALL operator not supported
	if suite.dbType == constants.DbSQLite {
		suite.T().Skipf("ALL operator not supported on %s (SQL standard feature)", suite.dbType)

		return
	}

	suite.Run("BasicNotEqualsAll", func() {
		// Find users whose age does not equal all ages in a subquery
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEqualsAll("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 30)
							})
					})
				}).
				OrderBy("age"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.NotEqual(int16(30), user.Age, "Age should not be 30")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrNotEqualsAll", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.NotEqualsAll("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 25)
							})
					}).OrNotEqualsAll("age", func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Expr("?", 35)
							})
					})
				}).
				OrderBy("age"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestExists tests the Exists and OrExists conditions using Expr with ExprBuilder.
func (suite *SubqueryOperationsTestSuite) TestExists() {
	suite.T().Logf("Testing Exists condition for %s", suite.dbType)

	suite.Run("BasicExists", func() {
		// Find posts where the author exists and is active
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Expr(func(eb ExprBuilder) any {
						return eb.Exists(func(sq SelectQuery) {
							sq.Model((*User)(nil)).
								Where(func(cb ConditionBuilder) {
									cb.Equals("is_active", true).
										EqualsColumn("id", "p.user_id")
								})
						})
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts with active authors")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrExists", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Expr(func(eb ExprBuilder) any {
						return eb.Exists(func(sq SelectQuery) {
							sq.Model((*User)(nil)).
								Where(func(cb ConditionBuilder) {
									cb.Equals("age", 30).
										EqualsColumn("id", "p.user_id")
								})
						})
					}).OrExpr(func(eb ExprBuilder) any {
						return eb.Exists(func(sq SelectQuery) {
							sq.Model((*User)(nil)).
								Where(func(cb ConditionBuilder) {
									cb.Equals("age", 35).
										EqualsColumn("id", "p.user_id")
								})
						})
					})
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestNotExists tests the NotExists and OrNotExists conditions using Expr with ExprBuilder.
func (suite *SubqueryOperationsTestSuite) TestNotExists() {
	suite.T().Logf("Testing NotExists condition for %s", suite.dbType)

	suite.Run("BasicNotExists", func() {
		// Find posts where there's no corresponding category
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Expr(func(eb ExprBuilder) any {
						return eb.NotExists(func(sq SelectQuery) {
							sq.Model((*Category)(nil)).
								Where(func(cb ConditionBuilder) {
									cb.EqualsColumn("id", "p.category_id").
										Equals("name", "NonExistent")
								})
						})
					})
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})

	suite.Run("OrNotExists", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.Expr(func(eb ExprBuilder) any {
						return eb.NotExists(func(sq SelectQuery) {
							sq.Model((*Category)(nil)).
								Where(func(cb ConditionBuilder) {
									cb.EqualsColumn("id", "p.category_id").
										Equals("name", "Category1")
								})
						})
					}).OrExpr(func(eb ExprBuilder) any {
						return eb.NotExists(func(sq SelectQuery) {
							sq.Model((*Category)(nil)).
								Where(func(cb ConditionBuilder) {
									cb.EqualsColumn("id", "p.category_id").
										Equals("name", "Category2")
								})
						})
					})
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})
}
