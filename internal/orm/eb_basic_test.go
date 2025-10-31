package orm

import (
	"fmt"

	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// BasicExpressionsTestSuite tests basic expression methods of ExprBuilder
// including Column, Null, IsNull, IsNotNull, Literal, Order, Case, Expr, Exprs, ExprsWithSep.
type BasicExpressionsTestSuite struct {
	*OrmTestSuite
}

func (suite *BasicExpressionsTestSuite) TestColumn() {
	suite.T().Logf("Testing Column expression for %s", suite.DbType)

	// Test 1: Simple column reference
	suite.Run("SimpleColumnReference", func() {
		type ColumnResult struct {
			Id    string `bun:"id"`
			Title string `bun:"title"`
		}

		var results []ColumnResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Column("id")
			}, "id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Column("title")
			}, "title").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Column references should work")
		suite.True(len(results) > 0, "Should have column results")

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")
			suite.T().Logf("ID: %s, Title: %s", result.Id, result.Title)
		}
	})

	// Test 2: Table-qualified column reference
	suite.Run("TableQualifiedColumnReference", func() {
		type QualifiedColumnResult struct {
			PostId    string `bun:"post_id"`
			PostTitle string `bun:"post_title"`
			UserName  string `bun:"user_name"`
		}

		var results []QualifiedColumnResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Join((*User)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "p.user_id")
			}).
			SelectAs("p.id", "post_id").
			SelectAs("p.title", "post_title").
			SelectAs("u.name", "user_name").
			OrderBy("p.id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Qualified column references should work")
		suite.True(len(results) > 0, "Should have qualified column results")

		for _, result := range results {
			suite.NotEmpty(result.PostId, "Post ID should not be empty")
			suite.NotEmpty(result.PostTitle, "Post title should not be empty")
			suite.NotEmpty(result.UserName, "User name should not be empty")
			suite.T().Logf("Post: %s - %s, User: %s",
				result.PostId, result.PostTitle, result.UserName)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestNull() {
	suite.T().Logf("Testing Null expression for %s", suite.DbType)

	// Test: NULL value in SELECT
	suite.Run("NullValueInSelect", func() {
		type NullResult struct {
			Id        string  `bun:"id"`
			Title     string  `bun:"title"`
			NullValue *string `bun:"null_value"`
		}

		var results []NullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Null()
			}, "null_value").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NULL expression should work")
		suite.True(len(results) > 0, "Should have NULL results")

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")
			suite.Nil(result.NullValue, "NULL value should be nil")
			suite.T().Logf("ID: %s, Title: %s, NullValue: %v",
				result.Id, result.Title, result.NullValue)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestIsNull() {
	suite.T().Logf("Testing IsNull function for %s", suite.DbType)

	// Test 1: IsNull - check for NULL values
	suite.Run("CheckNullValues", func() {
		type IsNullResult struct {
			Id     string `bun:"id"`
			Status string `bun:"status"`
			IsNull bool   `bun:"is_null"`
		}

		var results []IsNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				// Use CASE to create a NULL value when status = 'active'
				return eb.IsNull(eb.Expr("CASE WHEN ? = 'active' THEN NULL ELSE ? END",
					eb.Column("status"), eb.Column("status")))
			}, "is_null").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "IsNull should work")
		suite.True(len(results) > 0, "Should have IsNull results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Status: %s, IsNull: %v",
				result.Id, result.Status, result.IsNull)
		}
	})

	// Test 2: IsNull in WHERE clause with Coalesce
	suite.Run("IsNullWithCoalesce", func() {
		type NullCoalesceResult struct {
			Id        string `bun:"id"`
			Status    string `bun:"status"`
			SafeValue string `bun:"safe_value"`
		}

		var results []NullCoalesceResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(
					eb.Expr("CASE WHEN ? = 'active' THEN NULL ELSE ? END",
						eb.Column("status"), eb.Column("status")),
					"default",
				)
			}, "safe_value").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Coalesce with NULL checks should work")
		suite.True(len(results) > 0, "Should have Coalesce results")

		for _, result := range results {
			suite.NotEmpty(result.SafeValue, "Safe value should never be empty due to Coalesce")
			suite.T().Logf("ID: %s, Status: %s, SafeValue: %s",
				result.Id, result.Status, result.SafeValue)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestIsNotNull() {
	suite.T().Logf("Testing IsNotNull function for %s", suite.DbType)

	// Test 1: IsNotNull - check for NOT NULL values
	suite.Run("CheckNotNullValues", func() {
		type IsNotNullResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			IsNotNull bool   `bun:"is_not_null"`
		}

		var results []IsNotNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IsNotNull(eb.Column("title"))
			}, "is_not_null").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "IsNotNull should work")
		suite.True(len(results) > 0, "Should have IsNotNull results")

		for _, result := range results {
			suite.True(result.IsNotNull, "Title should not be NULL for existing posts")
			suite.T().Logf("ID: %s, Title: %s, IsNotNull: %v",
				result.Id, result.Title, result.IsNotNull)
		}
	})

	// Test 2: Combined NULL checks
	suite.Run("CombinedNullChecks", func() {
		type NullCheckResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			Status    string `bun:"status"`
			HasTitle  bool   `bun:"has_title"`
			HasStatus bool   `bun:"has_status"`
		}

		var results []NullCheckResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IsNotNull(eb.Column("title"))
			}, "has_title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IsNotNull(eb.Column("status"))
			}, "has_status").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NULL checks in SELECT should work")
		suite.True(len(results) > 0, "Should have NULL check results")

		for _, result := range results {
			suite.NotEmpty(result.Title, "Title should not be empty")
			suite.NotEmpty(result.Status, "Status should not be empty")
			suite.True(result.HasTitle, "Has title should be true")
			suite.True(result.HasStatus, "Has status should be true")
			suite.T().Logf("ID: %s, Title: %s, Status: %s, HasTitle: %v, HasStatus: %v",
				result.Id, result.Title, result.Status, result.HasTitle, result.HasStatus)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestLiteral() {
	suite.T().Logf("Testing Literal expression for %s", suite.DbType)

	// Test: Literal values in expressions
	suite.Run("LiteralValuesInExpressions", func() {
		type LiteralResult struct {
			Id         string `bun:"id"`
			Title      string `bun:"title"`
			LiteralVal string `bun:"literal_val"`
		}

		var results []LiteralResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Literal("test_literal")
			}, "literal_val").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Literal should work")
		suite.True(len(results) > 0, "Should have literal results")

		for _, result := range results {
			suite.Equal("test_literal", result.LiteralVal, "Literal value should match")
			suite.T().Logf("ID: %s, Title: %s, Literal: %s",
				result.Id, result.Title, result.LiteralVal)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestOrder() {
	suite.T().Logf("Testing Order expression for %s", suite.DbType)

	// Test 1: Simple ORDER BY expression
	suite.Run("SimpleOrderBy", func() {
		type OrderResult struct {
			Id   string `bun:"id"`
			Name string `bun:"name"`
			Age  int16  `bun:"age"`
		}

		var results []OrderResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "age").
			OrderByExpr(func(eb ExprBuilder) any {
				return eb.Order(func(ob OrderBuilder) {
					ob.Column("age").Desc()
				})
			}).
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Order should work")
		suite.True(len(results) > 0, "Should have Order results")

		// Verify descending order
		var prevAge int16 = -1
		for i, result := range results {
			suite.T().Logf("ID: %s, Name: %s, Age: %d",
				result.Id, result.Name, result.Age)

			if i > 0 {
				suite.True(result.Age <= prevAge,
					"Results should be ordered by age DESC")
			}

			prevAge = result.Age
		}
	})

	// Test 2: Multiple ORDER BY columns
	suite.Run("MultipleOrderByColumns", func() {
		type MultiOrderResult struct {
			Id       string `bun:"id"`
			IsActive bool   `bun:"is_active"`
			Age      int16  `bun:"age"`
		}

		var results []MultiOrderResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "is_active", "age").
			OrderByExpr(func(eb ExprBuilder) any {
				return eb.Order(func(ob OrderBuilder) {
					ob.Column("is_active").Asc()
				})
			}).
			OrderByExpr(func(eb ExprBuilder) any {
				return eb.Order(func(ob OrderBuilder) {
					ob.Column("age").Desc()
				})
			}).
			Limit(10).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Multiple Order expressions should work")
		suite.True(len(results) > 0, "Should have multi-order results")

		// Log results
		for _, result := range results {
			suite.T().Logf("ID: %s, IsActive: %v, Age: %d",
				result.Id, result.IsActive, result.Age)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestCase() {
	suite.T().Logf("Testing Case expression for %s", suite.DbType)

	// Test 1: Simple CASE expression
	suite.Run("SimpleCaseExpression", func() {
		type CaseResult struct {
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
			Category  string `bun:"category"`
		}

		var results []CaseResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(
						func(cond ConditionBuilder) {
							cond.GreaterThan("view_count", 80)
						}).
						Then("Popular").
						When(func(cond ConditionBuilder) {
							cond.GreaterThan("view_count", 30)
						}).
						Then("Moderate").
						Else("Low")
				})
			}, "category").
			OrderBy("title").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Simple CASE should work")
		suite.True(len(results) > 0, "Should have CASE results")

		for _, result := range results {
			suite.NotEmpty(result.Category, "Category should not be empty")
			suite.T().Logf("Post %s: %d views -> %s",
				result.Title, result.ViewCount, result.Category)
		}
	})

	// Test 2: Nested CASE expression with NULL handling
	suite.Run("NestedCaseWithNullHandling", func() {
		type NestedCaseResult struct {
			Title         string `bun:"title"`
			Status        string `bun:"status"`
			DisplayStatus string `bun:"display_status"`
		}

		var results []NestedCaseResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(
					eb.NullIf(eb.Column("status"), "'draft'"),
					"'Working Draft'",
				)
			}, "display_status").
			OrderBy("title").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Nested CASE with NULL handling should work")
		suite.True(len(results) > 0, "Should have nested CASE results")

		for _, result := range results {
			suite.NotEmpty(result.DisplayStatus, "Display status should not be empty")
			suite.T().Logf("Post %s: Status=%s, DisplayStatus=%s",
				result.Title, result.Status, result.DisplayStatus)
		}
	})

	// Test 3: Complex CASE with multiple conditions
	suite.Run("ComplexCaseWithMultipleConditions", func() {
		type ComplexCaseResult struct {
			Title        string `bun:"title"`
			Status       string `bun:"status"`
			ViewCount    int64  `bun:"view_count"`
			ViewCategory string `bun:"view_category"`
		}

		var results []ComplexCaseResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(
					eb.NullIf(
						eb.Case(func(cb CaseBuilder) {
							cb.When(
								func(cond ConditionBuilder) {
									cond.GreaterThan("view_count", 100)
								}).
								Then("High").
								When(func(cond ConditionBuilder) {
									cond.GreaterThan("view_count", 50)
								}).
								Then("Medium").
								Else("Low")
						}),
						"Low",
					),
					"No Views",
				)
			}, "view_category").
			OrderBy("title").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Complex CASE should work")
		suite.True(len(results) > 0, "Should have complex CASE results")

		for _, result := range results {
			suite.NotEmpty(result.ViewCategory, "View category should not be empty")
			suite.T().Logf("Post %s: Status=%s, Views=%d, Category=%s",
				result.Title, result.Status, result.ViewCount, result.ViewCategory)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestSubQuery() {
	suite.T().Logf("Testing SubQuery expression for %s", suite.DbType)

	// Test 1: Simple subquery in SELECT clause
	suite.Run("SimpleSubQueryInSelect", func() {
		type SubQueryResult struct {
			Id        string  `bun:"id"`
			Title     string  `bun:"title"`
			ViewCount int64   `bun:"view_count"`
			AvgViews  float64 `bun:"avg_views"`
		}

		var results []SubQueryResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SubQuery(func(sq SelectQuery) {
					sq.Model((*Post)(nil)).
						SelectExpr(func(eb ExprBuilder) any {
							return eb.AvgColumn("view_count")
						})
				})
			}, "avg_views").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "SubQuery in SELECT should work")
		suite.True(len(results) > 0, "Should have SubQuery results")

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")
			suite.True(result.AvgViews > 0, "Average views should be positive")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d, AvgViews: %.2f",
				result.Id, result.Title, result.ViewCount, result.AvgViews)
		}
	})

	// Test 2: Subquery in WHERE clause
	suite.Run("SubQueryInWhereClause", func() {
		type FilteredResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []FilteredResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.GreaterThanOrEqualExpr("view_count", func(eb ExprBuilder) any {
					return eb.SubQuery(func(sq SelectQuery) {
						sq.Model((*Post)(nil)).
							SelectExpr(func(eb ExprBuilder) any {
								return eb.AvgColumn("view_count")
							})
					})
				})
			}).
			OrderByDesc("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "SubQuery in WHERE clause should work")
		suite.True(len(results) > 0, "Should have filtered results")

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d (above average)",
				result.Id, result.Title, result.ViewCount)
		}
	})

	// Test 3: Correlated subquery
	suite.Run("CorrelatedSubQuery", func() {
		type CorrelatedResult struct {
			Id        string `bun:"id"`
			Name      string `bun:"name"`
			PostCount int64  `bun:"post_count"`
		}

		var results []CorrelatedResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SubQuery(func(sq SelectQuery) {
					sq.Model((*Post)(nil)).
						SelectExpr(func(eb ExprBuilder) any {
							return eb.CountAll()
						}).
						Where(func(cb ConditionBuilder) {
							cb.EqualsColumn("p.user_id", "u.id")
						})
				})
			}, "post_count").
			OrderBy("name").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Correlated subquery should work")
		suite.True(len(results) > 0, "Should have correlated subquery results")

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.True(result.PostCount >= 0, "Post count should be non-negative")
			suite.T().Logf("ID: %s, Name: %s, PostCount: %d",
				result.Id, result.Name, result.PostCount)
		}
	})

	// Test 4: Subquery with aggregate function
	suite.Run("SubQueryWithAggregateFunction", func() {
		type AggregateSubQueryResult struct {
			Id         string `bun:"id"`
			Title      string `bun:"title"`
			ViewCount  int64  `bun:"view_count"`
			MaxViews   int64  `bun:"max_views"`
			IsTopViews bool   `bun:"is_top_views"`
		}

		var results []AggregateSubQueryResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SubQuery(func(sq SelectQuery) {
					sq.Model((*Post)(nil)).
						SelectExpr(func(eb ExprBuilder) any {
							return eb.MaxColumn("view_count")
						})
				})
			}, "max_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(func(cond ConditionBuilder) {
						cond.EqualsExpr("view_count", func(eb ExprBuilder) any {
							return eb.SubQuery(func(sq SelectQuery) {
								sq.Model((*Post)(nil)).
									SelectExpr(func(eb ExprBuilder) any {
										return eb.MaxColumn("view_count")
									})
							})
						})
					}).
						Then(true).
						Else(false)
				})
			}, "is_top_views").
			OrderByDesc("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "SubQuery with aggregate function should work")
		suite.True(len(results) > 0, "Should have aggregate subquery results")

		var foundTopView bool
		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")
			suite.True(result.MaxViews > 0, "Max views should be positive")
			suite.True(result.ViewCount <= result.MaxViews,
				"View count should not exceed max views")

			if result.IsTopViews {
				foundTopView = true

				suite.Equal(result.ViewCount, result.MaxViews,
					"Top view post should have view_count equal to max_views")
			}

			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d, MaxViews: %d, IsTop: %v",
				result.Id, result.Title, result.ViewCount, result.MaxViews, result.IsTopViews)
		}

		suite.True(foundTopView, "Should find at least one top view post")
	})
}

// TestExists tests the Exists expression.
func (suite *BasicExpressionsTestSuite) TestExists() {
	suite.T().Logf("Testing Exists expression for %s", suite.DbType)

	suite.Run("ExistsInWhereClause", func() {
		type ExistsResult struct {
			Id    string `bun:"id"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		var results []ExistsResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "email").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Exists(func(sq SelectQuery) {
						sq.Model((*Post)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.EqualsColumn("p.user_id", "u.id")
							})
					})
				})
			}).
			OrderBy("name").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "EXISTS in WHERE clause should work")
		suite.True(len(results) > 0, "Should have users with posts")

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.T().Logf("User with posts - ID: %s, Name: %s, Email: %s",
				result.Id, result.Name, result.Email)
		}
	})

	suite.Run("ExistsWithComplexCondition", func() {
		type UserWithHighViewPosts struct {
			Id   string `bun:"id"`
			Name string `bun:"name"`
		}

		var results []UserWithHighViewPosts

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Exists(func(sq SelectQuery) {
						sq.Model((*Post)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.EqualsColumn("p.user_id", "u.id")
								cb.GreaterThan("p.view_count", 50)
							})
					})
				})
			}).
			OrderBy("name").
			Scan(suite.Ctx, &results)

		suite.NoError(err, "EXISTS with complex conditions should work")
		suite.T().Logf("Found %d users with high-view posts", len(results))

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.T().Logf("User with high-view posts - ID: %s, Name: %s",
				result.Id, result.Name)
		}
	})

	suite.Run("ExistsWithAND", func() {
		type Result struct {
			Id    string `bun:"id"`
			Title string `bun:"title"`
		}

		var results []Result

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			Where(func(cb ConditionBuilder) {
				cb.GreaterThan("view_count", 20)
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Exists(func(sq SelectQuery) {
						sq.Model((*User)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.EqualsColumn("u.id", "p.user_id")
								cb.StartsWith("u.name", "A")
							})
					})
				})
			}).
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "EXISTS combined with AND should work")
		suite.T().Logf("Found %d posts", len(results))

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.T().Logf("Post - ID: %s, Title: %s", result.Id, result.Title)
		}
	})
}

// TestNotExists tests the NotExists expression.
func (suite *BasicExpressionsTestSuite) TestNotExists() {
	suite.T().Logf("Testing NotExists expression for %s", suite.DbType)

	suite.Run("NotExistsInWhereClause", func() {
		type NotExistsResult struct {
			Id    string `bun:"id"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		var results []NotExistsResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "email").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotExists(func(sq SelectQuery) {
						sq.Model((*Post)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.EqualsColumn("p.user_id", "u.id")
							})
					})
				})
			}).
			OrderBy("name").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NOT EXISTS in WHERE clause should work")
		suite.T().Logf("Found %d users without posts", len(results))

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.T().Logf("User without posts - ID: %s, Name: %s, Email: %s",
				result.Id, result.Name, result.Email)
		}
	})

	suite.Run("NotExistsWithComplexCondition", func() {
		type UserWithoutHighViewPosts struct {
			Id   string `bun:"id"`
			Name string `bun:"name"`
		}

		var results []UserWithoutHighViewPosts

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotExists(func(sq SelectQuery) {
						sq.Model((*Post)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.EqualsColumn("p.user_id", "u.id")
								cb.GreaterThan("p.view_count", 100)
							})
					})
				})
			}).
			OrderBy("name").
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NOT EXISTS with complex conditions should work")
		suite.T().Logf("Found %d users without high-view posts", len(results))

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.T().Logf("User without high-view posts - ID: %s, Name: %s",
				result.Id, result.Name)
		}
	})

	suite.Run("ExistsAndNotExistsCombined", func() {
		type CombinedResult struct {
			Id   string `bun:"id"`
			Name string `bun:"name"`
		}

		var results []CombinedResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Exists(func(sq SelectQuery) {
						sq.Model((*Post)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.EqualsColumn("p.user_id", "u.id")
							})
					})
				})
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotExists(func(sq SelectQuery) {
						sq.Model((*Post)(nil)).
							Select("id").
							Where(func(cb ConditionBuilder) {
								cb.EqualsColumn("p.user_id", "u.id")
								cb.GreaterThan("p.view_count", 100)
							})
					})
				})
			}).
			OrderBy("name").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "EXISTS and NOT EXISTS combined should work")
		suite.T().Logf("Found %d users with posts but no high-view posts", len(results))

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.T().Logf("User - ID: %s, Name: %s", result.Id, result.Name)
		}
	})
}

// TestNot tests the Not method for negating boolean expressions.
func (suite *BasicExpressionsTestSuite) TestNot() {
	suite.T().Logf("Testing Not expression for %s", suite.DbType)

	// Test 1: NOT with equality condition
	suite.Run("NotWithEqualityCondition", func() {
		type NotResult struct {
			Id     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var results []NotResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Not(eb.Equals(eb.Column("status"), "published"))
				})
			}).
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NOT with equality condition should work")
		suite.True(len(results) > 0, "Should have NOT results")

		for _, result := range results {
			suite.NotEqual("published", result.Status, "Status should not be 'published'")
			suite.T().Logf("ID: %s, Title: %s, Status: %s", result.Id, result.Title, result.Status)
		}
	})

	// Test 2: NOT with comparison condition
	suite.Run("NotWithComparisonCondition", func() {
		type ComparisonNotResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []ComparisonNotResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Not(eb.GreaterThan(eb.Column("view_count"), 50))
				})
			}).
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NOT with comparison condition should work")
		suite.True(len(results) > 0, "Should have comparison NOT results")

		for _, result := range results {
			suite.True(result.ViewCount <= 50, "ViewCount should be <= 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})

	// Test 3: NOT with NULL check
	suite.Run("NotWithNullCheck", func() {
		type NullNotResult struct {
			Id          string `bun:"id"`
			Title       string `bun:"title"`
			Description string `bun:"description"`
		}

		var results []NullNotResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "description").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Not(eb.IsNull(eb.Column("description")))
				})
			}).
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NOT with NULL check should work")
		suite.True(len(results) > 0, "Should have NULL NOT results")

		for _, result := range results {
			suite.NotEmpty(result.Description, "Description should not be empty")
			suite.T().Logf("ID: %s, Title: %s, Description: %s", result.Id, result.Title, result.Description)
		}
	})

	// Test 4: NOT in SELECT expression
	suite.Run("NotInSelectExpression", func() {
		type SelectNotResult struct {
			Id         string `bun:"id"`
			Status     string `bun:"status"`
			IsNotDraft bool   `bun:"is_not_draft"`
		}

		var results []SelectNotResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Not(eb.Equals(eb.Column("status"), "draft"))
			}, "is_not_draft").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NOT in SELECT expression should work")
		suite.True(len(results) > 0, "Should have SELECT NOT results")

		for _, result := range results {
			expected := result.Status != "draft"
			suite.Equal(expected, result.IsNotDraft, "IsNotDraft should match expectation")
			suite.T().Logf("ID: %s, Status: %s, IsNotDraft: %v", result.Id, result.Status, result.IsNotDraft)
		}
	})
}

// TestAny tests the Any expression method.
// Note: SQLite does not support ANY/ALL operators natively and simulation is not trivial.
// MySQL supports ANY/ALL but does not allow LIMIT in ANY/ALL subqueries (Error 1235).
func (suite *BasicExpressionsTestSuite) TestAny() {
	if suite.DbType == constants.DbSQLite {
		suite.T().Skipf("Test skipped for %s (ANY operator not supported and cannot be easily simulated)", suite.DbType)

		return
	}

	if suite.DbType == constants.DbMySQL {
		suite.T().Skipf("Test skipped for %s (LIMIT in ANY/ALL subqueries not supported - Error 1235)", suite.DbType)

		return
	}

	suite.T().Logf("Testing Any expression for %s", suite.DbType)

	suite.Run("AnyWithEqualityCondition", func() {
		type AnyResult struct {
			Id    string `bun:"id"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		var results []AnyResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "email").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Equals(
						eb.Column("id"),
						eb.Any(func(sq SelectQuery) {
							sq.Model((*Post)(nil)).
								Select("user_id").
								Where(func(cb ConditionBuilder) {
									cb.Equals("status", "published")
								}).
								Limit(3)
						}),
					)
				})
			}).
			OrderBy("name").
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ANY with equality condition should work")
		suite.True(len(results) > 0, "Should have ANY results")

		for _, result := range results {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.T().Logf("ID: %s, Name: %s, Email: %s", result.Id, result.Name, result.Email)
		}
	})

	suite.Run("AnyWithGreaterThanCondition", func() {
		type ViewCountAnyResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []ViewCountAnyResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThan(
						eb.Column("view_count"),
						eb.Any(func(sq SelectQuery) {
							sq.Model((*Post)(nil)).
								SelectExpr(func(eb ExprBuilder) any {
									return eb.Literal(50)
								}).
								Where(func(cb ConditionBuilder) {
									cb.Equals("status", "published")
								}).
								Limit(1)
						}),
					)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ANY with greater than condition should work")
		suite.True(len(results) > 0, "Should have view count ANY results")

		for _, result := range results {
			suite.True(result.ViewCount > 50, "ViewCount should be > 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})
}

// TestAll tests the All expression method.
// Note: SQLite does not support ANY/ALL operators natively and simulation is not trivial.
// MySQL supports ANY/ALL but does not allow LIMIT in ANY/ALL subqueries (Error 1235).
func (suite *BasicExpressionsTestSuite) TestAll() {
	if suite.DbType == constants.DbSQLite {
		suite.T().Skipf("Test skipped for %s (ALL operator not supported and cannot be easily simulated)", suite.DbType)

		return
	}

	if suite.DbType == constants.DbMySQL {
		suite.T().Skipf("Test skipped for %s (LIMIT in ANY/ALL subqueries not supported - Error 1235)", suite.DbType)

		return
	}

	suite.T().Logf("Testing All expression for %s", suite.DbType)

	suite.Run("AllWithEqualityCondition", func() {
		type AllResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []AllResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThanOrEqual(
						eb.Column("view_count"),
						eb.All(func(sq SelectQuery) {
							sq.Model((*Post)(nil)).
								SelectExpr(func(eb ExprBuilder) any {
									return eb.Literal(10)
								}).
								Where(func(cb ConditionBuilder) {
									cb.Equals("status", "draft")
								}).
								Limit(1)
						}),
					)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ALL with greater than or equal condition should work")
		suite.True(len(results) > 0, "Should have ALL results")

		for _, result := range results {
			suite.True(result.ViewCount >= 10, "ViewCount should be >= 10")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})

	suite.Run("AllWithLessThanCondition", func() {
		type LessThanAllResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []LessThanAllResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.LessThan(
						eb.Column("view_count"),
						eb.All(func(sq SelectQuery) {
							sq.Model((*Post)(nil)).
								SelectExpr(func(eb ExprBuilder) any {
									return eb.Literal(200)
								}).
								Where(func(cb ConditionBuilder) {
									cb.Equals("status", "published")
								}).
								Limit(1)
						}),
					)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ALL with less than condition should work")
		suite.True(len(results) > 0, "Should have less than ALL results")

		for _, result := range results {
			suite.True(result.ViewCount < 200, "ViewCount should be < 200")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})
}

// TestArithmeticOperators tests the Add, Subtract, Multiply, Divide, and Paren methods.
func (suite *BasicExpressionsTestSuite) TestArithmeticOperators() {
	suite.T().Logf("Testing arithmetic operators for %s", suite.DbType)

	suite.Run("BasicArithmeticOperations", func() {
		type ArithmeticResult struct {
			Id         string  `bun:"id"`
			ViewCount  int64   `bun:"view_count"`
			Added      int64   `bun:"added"`
			Subtracted int64   `bun:"subtracted"`
			Multiplied int64   `bun:"multiplied"`
			Divided    float64 `bun:"divided"`
			Complex    int64   `bun:"complex"`
		}

		var results []ArithmeticResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Add(eb.Column("view_count"), 10)
			}, "added").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Subtract(eb.Column("view_count"), 5)
			}, "subtracted").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Multiply(eb.Column("view_count"), 2)
			}, "multiplied").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Divide(eb.Column("view_count"), 2)
			}, "divided").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Subtract(
					eb.Multiply(
						eb.Paren(eb.Add(eb.Column("view_count"), 10)),
						2,
					),
					5,
				)
			}, "complex").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Arithmetic operators should work")
		suite.True(len(results) > 0, "Should have arithmetic results")

		for _, result := range results {
			suite.Equal(result.ViewCount+10, result.Added)
			suite.Equal(result.ViewCount-5, result.Subtracted)
			suite.Equal(result.ViewCount*2, result.Multiplied)
			suite.InDelta(float64(result.ViewCount)/2.0, result.Divided, 0.01)

			expected := (result.ViewCount+10)*2 - 5
			suite.Equal(expected, result.Complex)

			suite.T().Logf("ViewCount: %d, Complex: ((%d + 10) * 2) - 5 = %d",
				result.ViewCount, result.ViewCount, result.Complex)
		}
	})
}

// TestComparisonOperators tests the Equals, NotEquals, GreaterThan, GreaterThanOrEqual, LessThan, and LessThanOrEqual methods.
func (suite *BasicExpressionsTestSuite) TestComparisonOperators() {
	suite.T().Logf("Testing comparison operators for %s", suite.DbType)

	// Test 1: Equals operator
	suite.Run("EqualsOperator", func() {
		type EqualsResult struct {
			Id     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var results []EqualsResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Equals(eb.Column("status"), "published")
				})
			}).
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Equals operator should work")
		suite.True(len(results) > 0, "Should have Equals results")

		for _, result := range results {
			suite.Equal("published", result.Status, "Status should be 'published'")
			suite.T().Logf("ID: %s, Title: %s, Status: %s", result.Id, result.Title, result.Status)
		}
	})

	// Test 2: NotEquals operator
	suite.Run("NotEqualsOperator", func() {
		type NotEqualsResult struct {
			Id     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var results []NotEqualsResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotEquals(eb.Column("status"), "draft")
				})
			}).
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NotEquals operator should work")
		suite.True(len(results) > 0, "Should have NotEquals results")

		for _, result := range results {
			suite.NotEqual("draft", result.Status, "Status should not be 'draft'")
			suite.T().Logf("ID: %s, Title: %s, Status: %s", result.Id, result.Title, result.Status)
		}
	})

	// Test 3: GreaterThan operator
	suite.Run("GreaterThanOperator", func() {
		type GreaterThanResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []GreaterThanResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThan(eb.Column("view_count"), 50)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "GreaterThan operator should work")
		suite.True(len(results) > 0, "Should have GreaterThan results")

		for _, result := range results {
			suite.True(result.ViewCount > 50, "ViewCount should be > 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})

	// Test 4: GreaterThanOrEqual operator
	suite.Run("GreaterThanOrEqualOperator", func() {
		type GreaterThanOrEqualResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []GreaterThanOrEqualResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThanOrEqual(eb.Column("view_count"), 30)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "GreaterThanOrEqual operator should work")
		suite.True(len(results) > 0, "Should have GreaterThanOrEqual results")

		for _, result := range results {
			suite.True(result.ViewCount >= 30, "ViewCount should be >= 30")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})

	// Test 5: LessThan operator
	suite.Run("LessThanOperator", func() {
		type LessThanResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []LessThanResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.LessThan(eb.Column("view_count"), 70)
				})
			}).
			OrderByDesc("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "LessThan operator should work")
		suite.True(len(results) > 0, "Should have LessThan results")

		for _, result := range results {
			suite.True(result.ViewCount < 70, "ViewCount should be < 70")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})

	// Test 6: LessThanOrEqual operator
	suite.Run("LessThanOrEqualOperator", func() {
		type LessThanOrEqualResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []LessThanOrEqualResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.LessThanOrEqual(eb.Column("view_count"), 50)
				})
			}).
			OrderByDesc("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "LessThanOrEqual operator should work")
		suite.True(len(results) > 0, "Should have LessThanOrEqual results")

		for _, result := range results {
			suite.True(result.ViewCount <= 50, "ViewCount should be <= 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.Id, result.Title, result.ViewCount)
		}
	})

	// Test 7: Combined comparison operators in SELECT
	suite.Run("CombinedComparisonInSelect", func() {
		type CombinedComparisonResult struct {
			Id           string `bun:"id"`
			ViewCount    int64  `bun:"view_count"`
			IsHighView   bool   `bun:"is_high_view"`
			IsMediumView bool   `bun:"is_medium_view"`
			IsLowView    bool   `bun:"is_low_view"`
		}

		var results []CombinedComparisonResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.GreaterThan(eb.Column("view_count"), 80)
			}, "is_high_view").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? AND ?",
					eb.GreaterThan(eb.Column("view_count"), 30),
					eb.LessThanOrEqual(eb.Column("view_count"), 80))
			}, "is_medium_view").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.LessThanOrEqual(eb.Column("view_count"), 30)
			}, "is_low_view").
			OrderBy("view_count").
			Limit(10).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Combined comparison operators should work")
		suite.True(len(results) > 0, "Should have combined comparison results")

		for _, result := range results {
			// Verify categorization logic
			if result.ViewCount > 80 {
				suite.True(result.IsHighView, "Should be high view")
				suite.False(result.IsMediumView, "Should not be medium view")
				suite.False(result.IsLowView, "Should not be low view")
			} else if result.ViewCount > 30 && result.ViewCount <= 80 {
				suite.False(result.IsHighView, "Should not be high view")
				suite.True(result.IsMediumView, "Should be medium view")
				suite.False(result.IsLowView, "Should not be low view")
			} else {
				suite.False(result.IsHighView, "Should not be high view")
				suite.False(result.IsMediumView, "Should not be medium view")
				suite.True(result.IsLowView, "Should be low view")
			}

			suite.T().Logf("ID: %s, ViewCount: %d, High: %v, Medium: %v, Low: %v",
				result.Id, result.ViewCount, result.IsHighView, result.IsMediumView, result.IsLowView)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestExpr() {
	suite.T().Logf("Testing Expr function for %s", suite.DbType)

	// Test 1: Simple arithmetic expression
	suite.Run("SimpleArithmeticExpression", func() {
		type ExprResult struct {
			Id         string  `bun:"id"`
			ViewCount  int64   `bun:"view_count"`
			Doubled    int64   `bun:"doubled"`
			Multiplied int64   `bun:"multiplied"`
			Divided    float64 `bun:"divided"`
		}

		var results []ExprResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? * 2", eb.Column("view_count"))
			}, "doubled").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? * ?", eb.Column("view_count"), 3)
			}, "multiplied").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? / 2.0", eb.Column("view_count"))
			}, "divided").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Arithmetic expressions should work")
		suite.True(len(results) > 0, "Should have expression results")

		for _, result := range results {
			suite.Equal(result.ViewCount*2, result.Doubled, "Doubled should equal view_count * 2")
			suite.Equal(result.ViewCount*3, result.Multiplied, "Multiplied should equal view_count * 3")
			suite.InDelta(float64(result.ViewCount)/2.0, result.Divided, 0.1,
				"Divided should equal view_count / 2")
			suite.T().Logf("ID: %s, ViewCount: %d, Doubled: %d, Multiplied: %d, Divided: %.1f",
				result.Id, result.ViewCount, result.Doubled, result.Multiplied, result.Divided)
		}
	})

	// Test 2: String concatenation expression
	suite.Run("StringConcatenationExpression", func() {
		type StringExprResult struct {
			Id              string `bun:"id"`
			Title           string `bun:"title"`
			Status          string `bun:"status"`
			TitleWithStatus string `bun:"title_with_status"`
		}

		var results []StringExprResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat(eb.Column("title"), "' - '", eb.Column("status"))
			}, "title_with_status").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "String concatenation expressions should work")
		suite.True(len(results) > 0, "Should have string expression results")

		for _, result := range results {
			suite.Contains(result.TitleWithStatus, result.Title,
				"Concatenated string should contain title")
			suite.Contains(result.TitleWithStatus, result.Status,
				"Concatenated string should contain status")
			suite.T().Logf("ID: %s, Title: %s, Status: %s, Combined: %s",
				result.Id, result.Title, result.Status, result.TitleWithStatus)
		}
	})

	// Test 3: Complex expression with functions
	suite.Run("ComplexExpressionWithFunctions", func() {
		type ComplexExprResult struct {
			Id            string  `bun:"id"`
			ViewCount     int64   `bun:"view_count"`
			AbsDifference int64   `bun:"abs_difference"`
			RoundedAvg    float64 `bun:"rounded_avg"`
		}

		var results []ComplexExprResult

		avgViewCount := float64(0)
		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}).
			Scan(suite.Ctx, &avgViewCount)
		suite.NoError(err)

		err = suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Abs(eb.Expr("? - ?", eb.Column("view_count"), int(avgViewCount)))
			}, "abs_difference").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Round(eb.Expr("? / 10.0", eb.Column("view_count")), 1)
			}, "rounded_avg").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Complex expressions with functions should work")
		suite.True(len(results) > 0, "Should have complex expression results")

		for _, result := range results {
			suite.True(result.AbsDifference >= 0, "Absolute difference should be non-negative")
			suite.True(result.RoundedAvg >= 0, "Rounded average should be non-negative")
			suite.T().Logf("ID: %s, ViewCount: %d, AbsDiff: %d, RoundedAvg: %.1f",
				result.Id, result.ViewCount, result.AbsDifference, result.RoundedAvg)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestExprs() {
	suite.T().Logf("Testing Exprs function for %s", suite.DbType)

	// Test: Combine multiple expressions with comma separator
	suite.Run("CombineMultipleExpressions", func() {
		type ExprsResult struct {
			Id       string `bun:"id"`
			Name     string `bun:"name"`
			Email    string `bun:"email"`
			Combined string `bun:"combined"`
		}

		var results []ExprsResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "email").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Exprs(
					eb.Column("name"),
					eb.Column("email"),
					eb.Column("age"),
				)
			}, "combined").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Exprs should work")
		suite.True(len(results) > 0, "Should have Exprs results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Name: %s, Email: %s, Combined: %s",
				result.Id, result.Name, result.Email, result.Combined)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestExprsWithSep() {
	suite.T().Logf("Testing ExprsWithSep function for %s", suite.DbType)

	// Test 1: Combine boolean expressions with OR
	suite.Run("CombineConditionsWithOR", func() {
		type ExprsWithSepResult struct {
			Id        string `bun:"id"`
			Name      string `bun:"name"`
			Age       int    `bun:"age"`
			IsActive  bool   `bun:"is_active"`
			MatchesOR bool   `bun:"matches_or"`
		}

		var results []ExprsWithSepResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "age", "is_active").
			SelectExpr(func(eb ExprBuilder) any {
				// Combine conditions: age > 28 OR is_active = false
				return eb.ExprsWithSep(
					" OR ",
					eb.Expr("age > ?", 28),
					eb.Expr("is_active = ?", false),
				)
			}, "matches_or").
			OrderBy("name").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ExprsWithSep with OR should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Name: %s, Age: %d, IsActive: %t, MatchesOR: %t",
				result.Id, result.Name, result.Age, result.IsActive, result.MatchesOR)
			// Verify: should be true if age > 28 OR is_active = false
			expected := result.Age > 28 || !result.IsActive
			suite.Equal(expected, result.MatchesOR, "OR condition should match expected result")
		}
	})

	// Test 2: Combine arithmetic expressions with addition
	suite.Run("CombineArithmeticWithAddition", func() {
		type ArithmeticResult struct {
			Id    string `bun:"id"`
			Name  string `bun:"name"`
			Age   int    `bun:"age"`
			Total int    `bun:"total"`
		}

		var results []ArithmeticResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "age").
			SelectExpr(func(eb ExprBuilder) any {
				// Combine: age + 10 + 5
				return eb.ExprsWithSep(
					" + ",
					eb.Column("age"),
					eb.Literal(10),
					eb.Literal(5),
				)
			}, "total").
			OrderBy("name").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ExprsWithSep with addition should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Name: %s, Age: %d, Total: %d",
				result.Id, result.Name, result.Age, result.Total)
			// Verify: total should be age + 10 + 5
			expected := result.Age + 15
			suite.Equal(expected, result.Total, "Addition should calculate correctly")
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestExprByDialect() {
	suite.T().Logf("Testing ExprByDialect function for %s", suite.DbType)

	// Test: Cross-database string concatenation
	// Different databases use different syntax for string concatenation:
	// - SQLite: uses || operator
	// - PostgreSQL/MySQL: use CONCAT function
	suite.Run("CrossDatabaseStringConcatenation", func() {
		type ConcatResult struct {
			Id       string `bun:"id"`
			Name     string `bun:"name"`
			Email    string `bun:"email"`
			FullInfo string `bun:"full_info"`
		}

		var results []ConcatResult

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "email").
			SelectExpr(func(eb ExprBuilder) any {
				// Use ExprByDialect to handle different database syntaxes
				return eb.ExprByDialect(DialectExprs{
					SQLite: func() schema.QueryAppender {
						// SQLite uses || for concatenation
						return eb.Expr("? || ' <' || ? || '>'",
							eb.Column("name"),
							eb.Column("email"))
					},
					Default: func() schema.QueryAppender {
						// PostgreSQL and MySQL use CONCAT
						return eb.Expr("CONCAT(?, ' <', ?, '>')",
							eb.Column("name"),
							eb.Column("email"))
					},
				})
			}, "full_info").
			OrderBy("name").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ExprByDialect should work for cross-database expressions")
		suite.True(len(results) > 0, "Should have ExprByDialect results")

		for _, result := range results {
			suite.NotEmpty(result.FullInfo, "Full info should not be empty")
			suite.Contains(result.FullInfo, result.Name, "Full info should contain name")
			suite.Contains(result.FullInfo, result.Email, "Full info should contain email")
			suite.Contains(result.FullInfo, "<", "Full info should contain opening bracket")
			suite.Contains(result.FullInfo, ">", "Full info should contain closing bracket")

			suite.T().Logf("Id: %s, Name: %s, Email: %s, FullInfo: %s",
				result.Id, result.Name, result.Email, result.FullInfo)
		}
	})

	// Test: Database-specific case expression
	suite.Run("DatabaseSpecificCaseExpression", func() {
		type CaseResult struct {
			Id        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
			Category  string `bun:"category"`
		}

		var results []CaseResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				// All databases support CASE expression the same way
				// But we test ExprByDialect with a common expression
				return eb.ExprByDialect(DialectExprs{
					Default: func() schema.QueryAppender {
						return eb.Case(func(cb CaseBuilder) {
							cb.When(
								func(cond ConditionBuilder) {
									cond.GreaterThan("view_count", 80)
								}).
								Then("High").
								When(func(cond ConditionBuilder) {
									cond.GreaterThan("view_count", 30)
								}).
								Then("Medium").
								Else("Low")
						})
					},
				})
			}, "category").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "ExprByDialect should work with CASE expressions")
		suite.True(len(results) > 0, "Should have case expression results")

		for _, result := range results {
			suite.NotEmpty(result.Category, "Category should not be empty")
			suite.Contains([]string{"High", "Medium", "Low"}, result.Category,
				"Category should be one of: High, Medium, Low")

			suite.T().Logf("Id: %s, ViewCount: %d, Category: %s",
				result.Id, result.ViewCount, result.Category)
		}
	})
}

func (suite *BasicExpressionsTestSuite) TestExecByDialect() {
	suite.T().Logf("Testing ExecByDialect function for %s", suite.DbType)

	// Test: Execute database-specific side effects
	suite.Run("DatabaseSpecificSideEffects", func() {
		// Track which database callback was executed
		var executed string

		// Create a simple query to test with
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)

		// Get the expression builder from the query
		eb := query.ExprBuilder()

		// Execute database-specific callbacks
		eb.ExecByDialect(DialectExecs{
			Postgres: func() {
				executed = "postgres"
			},
			MySQL: func() {
				executed = "mysql"
			},
			SQLite: func() {
				executed = "sqlite"
			},
		})

		// Verify the correct callback was executed based on current database
		switch suite.DbType {
		case constants.DbPostgres:
			suite.Equal("postgres", executed, "Postgres callback should be executed")
		case constants.DbMySQL:
			suite.Equal("mysql", executed, "MySQL callback should be executed")
		case constants.DbSQLite:
			suite.Equal("sqlite", executed, "SQLite callback should be executed")
		}

		suite.T().Logf("Executed callback for: %s", executed)
	})

	// Test: Default callback when database-specific callback is not provided
	suite.Run("DefaultCallback", func() {
		var executed string

		// Create a simple query to test with
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Only provide default callback
		eb.ExecByDialect(DialectExecs{
			Default: func() {
				executed = "default"
			},
		})

		// Default callback should be executed for all databases
		suite.Equal("default", executed, "Default callback should be executed")
		suite.T().Logf("Default callback executed for %s", suite.DbType)
	})
}

func (suite *BasicExpressionsTestSuite) TestExecByDialectWithErr() {
	suite.T().Logf("Testing ExecByDialectWithErr function for %s", suite.DbType)

	// Test: Execute database-specific callbacks with error handling
	suite.Run("DatabaseSpecificWithError", func() {
		var executed string

		// Create a simple query to test with
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Execute database-specific callbacks that return no error
		err := eb.ExecByDialectWithErr(DialectExecsWithErr{
			Postgres: func() error {
				executed = "postgres"

				return nil
			},
			MySQL: func() error {
				executed = "mysql"

				return nil
			},
			SQLite: func() error {
				executed = "sqlite"

				return nil
			},
		})

		suite.NoError(err, "ExecByDialectWithErr should not return error on success")

		// Verify the correct callback was executed
		switch suite.DbType {
		case constants.DbPostgres:
			suite.Equal("postgres", executed, "Postgres callback should be executed")
		case constants.DbMySQL:
			suite.Equal("mysql", executed, "MySQL callback should be executed")
		case constants.DbSQLite:
			suite.Equal("sqlite", executed, "SQLite callback should be executed")
		}

		suite.T().Logf("Executed callback with success for: %s", executed)
	})

	// Test: Error handling in database-specific callbacks
	suite.Run("ErrorHandling", func() {
		// Create a simple query to test with
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Execute callbacks that return an error
		testErr := fmt.Errorf("test error from %s", suite.DbType)
		err := eb.ExecByDialectWithErr(DialectExecsWithErr{
			Postgres: func() error {
				return testErr
			},
			MySQL: func() error {
				return testErr
			},
			SQLite: func() error {
				return testErr
			},
		})

		suite.Error(err, "ExecByDialectWithErr should return error when callback fails")
		suite.Equal(testErr, err, "Should return the exact error from callback")
		suite.T().Logf("Error handling works correctly for %s: %v", suite.DbType, err)
	})

	// Test: Default callback with error
	suite.Run("DefaultCallbackWithError", func() {
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Test successful default callback
		err := eb.ExecByDialectWithErr(DialectExecsWithErr{
			Default: func() error {
				return nil
			},
		})
		suite.NoError(err, "Default callback should work without error")

		// Test default callback returning error
		testErr := fmt.Errorf("default callback error")
		err = eb.ExecByDialectWithErr(DialectExecsWithErr{
			Default: func() error {
				return testErr
			},
		})
		suite.Error(err, "Default callback error should be returned")
		suite.Equal(testErr, err, "Should return the exact error from default callback")
		suite.T().Logf("Default callback error handling works for %s", suite.DbType)
	})
}

func (suite *BasicExpressionsTestSuite) TestFragmentByDialect() {
	suite.T().Logf("Testing FragmentByDialect function for %s", suite.DbType)

	// Test: Generate database-specific query fragments
	suite.Run("DatabaseSpecificFragments", func() {
		// Create a simple query to test with
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Generate database-specific fragments
		fragment, err := eb.FragmentByDialect(DialectFragments{
			Postgres: func() ([]byte, error) {
				return []byte("/* PostgreSQL fragment */"), nil
			},
			MySQL: func() ([]byte, error) {
				return []byte("/* MySQL fragment */"), nil
			},
			SQLite: func() ([]byte, error) {
				return []byte("/* SQLite fragment */"), nil
			},
		})

		suite.NoError(err, "FragmentByDialect should not return error")
		suite.NotNil(fragment, "Fragment should not be nil")

		// Verify the correct fragment was generated
		fragmentStr := string(fragment)
		switch suite.DbType {
		case constants.DbPostgres:
			suite.Contains(fragmentStr, "PostgreSQL", "Should contain PostgreSQL fragment")
		case constants.DbMySQL:
			suite.Contains(fragmentStr, "MySQL", "Should contain MySQL fragment")
		case constants.DbSQLite:
			suite.Contains(fragmentStr, "SQLite", "Should contain SQLite fragment")
		}

		suite.T().Logf("Generated fragment for %s: %s", suite.DbType, fragmentStr)
	})

	// Test: Error handling in fragment generation
	suite.Run("FragmentErrorHandling", func() {
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Generate fragment that returns an error
		testErr := fmt.Errorf("fragment generation error")
		fragment, err := eb.FragmentByDialect(DialectFragments{
			Postgres: func() ([]byte, error) {
				return nil, testErr
			},
			MySQL: func() ([]byte, error) {
				return nil, testErr
			},
			SQLite: func() ([]byte, error) {
				return nil, testErr
			},
		})

		suite.Error(err, "FragmentByDialect should return error when callback fails")
		suite.Equal(testErr, err, "Should return the exact error from callback")
		suite.Nil(fragment, "Fragment should be nil on error")
		suite.T().Logf("Error handling works for fragment generation: %v", err)
	})

	// Test: Default fragment callback
	suite.Run("DefaultFragmentCallback", func() {
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Use default callback
		defaultFragment := []byte("/* Default fragment */")
		fragment, err := eb.FragmentByDialect(DialectFragments{
			Default: func() ([]byte, error) {
				return defaultFragment, nil
			},
		})

		suite.NoError(err, "Default fragment callback should work")
		suite.Equal(defaultFragment, fragment, "Should return default fragment")
		suite.T().Logf("Default fragment works for %s: %s", suite.DbType, string(fragment))
	})

	// Test: Empty fragment (nil bytes)
	suite.Run("EmptyFragment", func() {
		query := suite.Db.NewSelect().Model((*Post)(nil)).Limit(1)
		eb := query.ExprBuilder()

		// Return empty/nil fragment
		fragment, err := eb.FragmentByDialect(DialectFragments{
			Default: func() ([]byte, error) {
				return nil, nil
			},
		})

		suite.NoError(err, "Empty fragment should not cause error")
		suite.Nil(fragment, "Fragment should be nil when callback returns nil")
		suite.T().Logf("Empty fragment handling works for %s", suite.DbType)
	})
}
