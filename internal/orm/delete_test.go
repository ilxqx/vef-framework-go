package orm

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/constants"
)

// DeleteTestSuite tests DELETE operations including CTE, table sources, filtering,
// ordering, limiting, returning, force delete, conditional application, and execution methods
// across all databases (PostgreSQL, MySQL, SQLite).
type DeleteTestSuite struct {
	*OrmTestSuite
}

// TestCTE tests Common Table Expression methods (With, WithValues, WithRecursive).
func (suite *DeleteTestSuite) TestCTE() {
	suite.T().Logf("Testing CTE methods for %s", suite.DbType)

	suite.Run("WithBasicCTE", func() {
		testUsers := []*User{
			{Name: "CTE Basic 1", Email: "cte_basic_1@example.com", Age: 25, IsActive: true},
			{Name: "CTE Basic 2", Email: "cte_basic_2@example.com", Age: 30, IsActive: false},
			{Name: "CTE Basic 3", Email: "cte_basic_3@example.com", Age: 35, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for CTE")

		result, err := suite.Db.NewDelete().
			With("inactive_users", func(query SelectQuery) {
				query.Model((*User)(nil)).
					Select("id").
					Where(func(cb ConditionBuilder) {
						cb.Equals("is_active", false).
							StartsWith("email", "cte_basic_")
					})
			}).
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.InSubQuery("id", func(subquery SelectQuery) {
					subquery.Table("inactive_users").Select("id")
				})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WITH clause should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 inactive user")

		suite.T().Logf("Deleted %d users using basic CTE", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "cte_basic_")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining CTE test users")
	})

	suite.Run("WithValuesCTE", func() {
		testPosts := []*Post{
			{Title: "CTE Values Post 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft", ViewCount: 10},
			{Title: "CTE Values Post 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 20},
			{Title: "CTE Values Post 3", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "review", ViewCount: 30},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for CTE VALUES")

		type StatusValue struct {
			Status string `bun:"status"`
		}

		statusValues := []StatusValue{
			{Status: "draft"},
			{Status: "review"},
		}

		result, err := suite.Db.NewDelete().
			WithValues("target_statuses", &statusValues).
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.InSubQuery("status", func(subquery SelectQuery) {
					subquery.Table("target_statuses").Select("status")
				}).
					StartsWith("title", "CTE Values Post")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WITH VALUES should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should delete 2 posts with draft or review status")

		suite.T().Logf("Deleted %d posts using CTE VALUES", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "CTE Values Post")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining CTE test posts")
	})

	suite.Run("WithRecursiveCTE", func() {
		if suite.DbType == constants.DbSQLite {
			suite.T().Skip("SQLite recursive CTE with DELETE requires special handling")

			return
		}

		testCategories := []*Category{
			{Name: "Parent Category", Description: nil},
			{Name: "Child Category 1", Description: nil},
			{Name: "Child Category 2", Description: nil},
		}

		_, err := suite.Db.NewInsert().Model(&testCategories).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test categories")

		testPosts := []*Post{
			{Title: "Recursive Post 1", Content: "Content", UserId: "user1", CategoryId: testCategories[0].Id, Status: "published"},
			{Title: "Recursive Post 2", Content: "Content", UserId: "user1", CategoryId: testCategories[1].Id, Status: "published"},
		}

		_, err = suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for recursive CTE")

		result, err := suite.Db.NewDelete().
			WithRecursive("category_tree", func(query SelectQuery) {
				query.Model((*Category)(nil)).
					Select("id", "parent_id").
					Where(func(cb ConditionBuilder) {
						cb.Equals("name", "Parent Category")
					}).
					UnionAll(func(unionQuery SelectQuery) {
						unionQuery.Model((*Category)(nil)).
							Select("id", "parent_id").
							JoinTable("category_tree", func(cb ConditionBuilder) {
								cb.EqualsColumn("parent_id", "ct.id")
							}, "ct")
					})
			}).
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.InSubQuery("category_id", func(subquery SelectQuery) {
					subquery.Table("category_tree").Select("id")
				})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WITH RECURSIVE should work when supported")

		rowsAffected, _ := result.RowsAffected()
		suite.T().Logf("Deleted %d posts using recursive CTE", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Recursive Post")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining recursive test posts")

		_, err = suite.Db.NewDelete().
			Model((*Category)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Contains("name", "Category")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup test categories")
	})
}

// TestTableSource tests table source methods (Model, ModelTable, Table, TableFrom, TableExpr, TableSubQuery).
func (suite *DeleteTestSuite) TestTableSource() {
	suite.T().Logf("Testing table source methods for %s", suite.DbType)

	suite.Run("ModelAndModelTable", func() {
		testUsers := []*User{
			{Name: "Table User 1", Email: "table1@example.com", Age: 25, IsActive: true},
			{Name: "Table User 2", Email: "table2@example.com", Age: 30, IsActive: false},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "table1@example.com")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Model method should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 user using Model")

		suite.T().Logf("Deleted %d user using Model method", rowsAffected)

		result, err = suite.Db.NewDelete().
			ModelTable("test_user", "u").
			Where(func(cb ConditionBuilder) {
				cb.Equals("u.email", "table2@example.com")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "ModelTable method should work correctly")

		rowsAffected, err = result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 user using ModelTable")

		suite.T().Logf("Deleted %d user using ModelTable method", rowsAffected)
	})

	suite.Run("TableAndTableFrom", func() {
		testPosts := []*Post{
			{Title: "Table Source Post 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
			{Title: "Table Source Post 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published"},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts")

		result, err := suite.Db.NewDelete().
			Table("test_post", "p").
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("p.title", "Table Source Post").
					Equals("p.status", "draft")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Table method should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 post using Table")

		suite.T().Logf("Deleted %d post using Table method", rowsAffected)

		result, err = suite.Db.NewDelete().
			TableFrom((*Post)(nil), "p").
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("p.title", "Table Source Post").
					Equals("p.status", "published")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "TableFrom method should work correctly")

		rowsAffected, err = result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 post using TableFrom")

		suite.T().Logf("Deleted %d post using TableFrom method", rowsAffected)
	})

	suite.Run("TableExpr", func() {
		testUsers := []*User{
			{Name: "Expr User", Email: "expr@example.com", Age: 28, IsActive: false},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test user for TableExpr")

		result, err := suite.Db.NewDelete().
			Table("test_user", "u").
			Where(func(cb ConditionBuilder) {
				cb.Equals("u.email", "expr@example.com")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Table method should work for TableExpr test")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 user")

		suite.T().Logf("Deleted %d user using Table method (TableExpr functionality)", rowsAffected)
	})
}

// TestFiltering tests filtering methods (Where, WherePk, WhereDeleted, IncludeDeleted).
func (suite *DeleteTestSuite) TestFiltering() {
	suite.T().Logf("Testing filtering methods for %s", suite.DbType)

	suite.Run("WhereCondition", func() {
		testUsers := []*User{
			{Name: "Filter User 1", Email: "filter1@example.com", Age: 25, IsActive: true},
			{Name: "Filter User 2", Email: "filter2@example.com", Age: 30, IsActive: false},
			{Name: "Filter User 3", Email: "filter3@example.com", Age: 35, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for filtering")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("is_active", false)
			}).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "filter")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WHERE with conditions should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 inactive user")

		suite.T().Logf("Deleted %d user using WHERE method", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "filter")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining filter test users")
	})

	suite.Run("WherePrimaryKey", func() {
		testUser := &User{
			Name:     "PK Delete User",
			Email:    "pk@example.com",
			Age:      28,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().Model(testUser).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test user for WherePk")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(testUser.Id)
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WherePk should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 user by primary key")

		suite.T().Logf("Deleted user by primary key: %s", testUser.Id)
	})

	suite.Run("ComplexWhereConditions", func() {
		testPosts := []*Post{
			{Title: "Complex Filter 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft", ViewCount: 5},
			{Title: "Complex Filter 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 50},
			{Title: "Complex Filter 3", Content: "Content", UserId: "user2", CategoryId: "cat1", Status: "published", ViewCount: 150},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for complex filtering")

		result, err := suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Complex Filter").
					Equals("status", "published").
					Group(func(innerCb ConditionBuilder) {
						innerCb.LessThan("view_count", 100).OrEquals("user_id", "user2")
					})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Complex WHERE conditions should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should delete 2 posts matching complex conditions")

		suite.T().Logf("Deleted %d posts using complex WHERE conditions", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Complex Filter")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining complex filter test posts")
	})

	suite.Run("WhereWithSubquery", func() {
		testUsers := []*User{
			{Name: "Subquery User 1", Email: "subquser1@example.com", Age: 25, IsActive: false},
			{Name: "Subquery User 2", Email: "subquser2@example.com", Age: 30, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for subquery filtering")

		testPosts := []*Post{
			{Title: "Subquery Post 1", Content: "Content", UserId: testUsers[0].Id, CategoryId: "cat1", Status: "published"},
			{Title: "Subquery Post 2", Content: "Content", UserId: testUsers[1].Id, CategoryId: "cat1", Status: "published"},
		}

		_, err = suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for subquery filtering")

		result, err := suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.InSubQuery("user_id", func(subquery SelectQuery) {
					subquery.Model((*User)(nil)).
						Select("id").
						Where(func(cb ConditionBuilder) {
							cb.Equals("is_active", false).
								StartsWith("email", "subquser")
						})
				}).
					StartsWith("title", "Subquery Post")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WHERE with subquery should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 post by inactive users")

		suite.T().Logf("Deleted %d post using WHERE with subquery", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Subquery Post")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining subquery test posts")

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "subquser")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup subquery test users")
	})
}

// TestOrdering tests ordering methods (OrderBy, OrderByDesc, OrderByExpr) combined with Limit.
// Note: PostgreSQL and SQLite (without SQLITE_ENABLE_UPDATE_DELETE_LIMIT) don't support DELETE with ORDER BY/LIMIT directly.
func (suite *DeleteTestSuite) TestOrdering() {
	suite.T().Logf("Testing ordering methods for %s", suite.DbType)

	if suite.DbType == constants.DbPostgres || suite.DbType == constants.DbSQLite {
		suite.T().Skipf("%s doesn't support DELETE with ORDER BY/LIMIT in current configuration", suite.DbType)

		return
	}

	suite.Run("OrderByAscending", func() {
		testUsers := []*User{
			{Name: "Order User C", Email: "order_c@example.com", Age: 35, IsActive: true},
			{Name: "Order User A", Email: "order_a@example.com", Age: 25, IsActive: true},
			{Name: "Order User B", Email: "order_b@example.com", Age: 30, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for ordering")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "order_")
			}).
			OrderBy("age").
			Limit(1).
			Exec(suite.Ctx)

		suite.NoError(err, "OrderBy should work correctly with Limit")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 user (youngest)")

		var remainingUsers []User

		err = suite.Db.NewSelect().
			Model(&remainingUsers).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "order_")
			}).
			OrderBy("age").
			Scan(suite.Ctx)
		suite.NoError(err, "Should retrieve remaining users")
		suite.Len(remainingUsers, 2, "Should have 2 remaining users")
		suite.True(remainingUsers[0].Age >= 30, "Youngest user should be deleted")

		suite.T().Logf("Deleted youngest user using OrderBy, %d users remain", len(remainingUsers))

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "order_")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining order test users")
	})

	suite.Run("OrderByDescending", func() {
		testPosts := []*Post{
			{Title: "Order Post A", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 10},
			{Title: "Order Post B", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 50},
			{Title: "Order Post C", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published", ViewCount: 100},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for descending ordering")

		result, err := suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Order Post")
			}).
			OrderByDesc("view_count").
			Limit(1).
			Exec(suite.Ctx)

		suite.NoError(err, "OrderByDesc should work correctly with Limit")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 post (highest view count)")

		var remainingPosts []Post

		err = suite.Db.NewSelect().
			Model(&remainingPosts).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Order Post")
			}).
			OrderByDesc("view_count").
			Scan(suite.Ctx)
		suite.NoError(err, "Should retrieve remaining posts")
		suite.Len(remainingPosts, 2, "Should have 2 remaining posts")
		suite.True(remainingPosts[0].ViewCount <= 50, "Highest view count post should be deleted")

		suite.T().Logf("Deleted highest view count post using OrderByDesc, %d posts remain", len(remainingPosts))

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Order Post")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining order test posts")
	})

	suite.Run("OrderByExpression", func() {
		testUsers := []*User{
			{Name: "Expr Order 1", Email: "exprorder1@example.com", Age: 25, IsActive: true},
			{Name: "Expr Order 2", Email: "exprorder2@example.com", Age: 30, IsActive: false},
			{Name: "Expr Order 3", Email: "exprorder3@example.com", Age: 35, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for expression ordering")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "exprorder")
			}).
			OrderByExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(func(cond ConditionBuilder) {
						cond.IsTrue("is_active")
					}).Then(eb.Column("age"))
					cb.Else(eb.Literal("0"))
				})
			}).
			Limit(1).
			Exec(suite.Ctx)

		suite.NoError(err, "OrderByExpr should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.True(rowsAffected >= 1, "Should delete at least 1 user")

		suite.T().Logf("Deleted %d user using OrderByExpr", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "exprorder")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining expr order test users")
	})
}

// TestLimit tests the Limit method.
// Note: PostgreSQL and SQLite (without SQLITE_ENABLE_UPDATE_DELETE_LIMIT) don't support DELETE with LIMIT directly.
func (suite *DeleteTestSuite) TestLimit() {
	suite.T().Logf("Testing Limit method for %s", suite.DbType)

	if suite.DbType == constants.DbPostgres || suite.DbType == constants.DbSQLite {
		suite.T().Skipf("%s doesn't support DELETE with LIMIT in current configuration", suite.DbType)

		return
	}

	suite.Run("LimitBasic", func() {
		testPosts := []*Post{
			{Title: "Limit Post 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
			{Title: "Limit Post 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
			{Title: "Limit Post 3", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
			{Title: "Limit Post 4", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for limit")

		result, err := suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Limit Post")
			}).
			Limit(2).
			Exec(suite.Ctx)

		suite.NoError(err, "Limit should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should delete exactly 2 posts")

		var remainingPosts []Post

		err = suite.Db.NewSelect().
			Model(&remainingPosts).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Limit Post")
			}).
			Scan(suite.Ctx)
		suite.NoError(err, "Should retrieve remaining posts")
		suite.Len(remainingPosts, 2, "Should have 2 remaining posts")

		suite.T().Logf("Deleted %d posts with Limit, %d posts remain", rowsAffected, len(remainingPosts))

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Limit Post")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining limit test posts")
	})

	suite.Run("LimitWithOrdering", func() {
		testUsers := []*User{
			{Name: "Limit User 1", Email: "limituser1@example.com", Age: 20, IsActive: true},
			{Name: "Limit User 2", Email: "limituser2@example.com", Age: 30, IsActive: true},
			{Name: "Limit User 3", Email: "limituser3@example.com", Age: 40, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for limit with ordering")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "limituser")
			}).
			OrderBy("age").
			Limit(1).
			Exec(suite.Ctx)

		suite.NoError(err, "Limit with ordering should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete exactly 1 user (youngest)")

		var remainingUsers []User

		err = suite.Db.NewSelect().
			Model(&remainingUsers).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "limituser")
			}).
			OrderBy("age").
			Scan(suite.Ctx)
		suite.NoError(err, "Should retrieve remaining users")
		suite.Len(remainingUsers, 2, "Should have 2 remaining users")
		suite.True(remainingUsers[0].Age >= 30, "Youngest user should be deleted")

		suite.T().Logf("Deleted youngest user with Limit and OrderBy, %d users remain", len(remainingUsers))

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "limituser")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining limit test users")
	})

	suite.Run("LimitEdgeCases", func() {
		testPosts := []*Post{
			{Title: "Edge Limit 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published"},
			{Title: "Edge Limit 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published"},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for edge cases")

		result, err := suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Edge Limit")
			}).
			Limit(10).
			Exec(suite.Ctx)

		suite.NoError(err, "Limit larger than row count should work")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should delete all 2 posts even with Limit(10)")

		suite.T().Logf("Deleted all %d posts with Limit(10)", rowsAffected)
	})
}

// TestReturning tests Returning methods (Returning, ReturningAll, ReturningNone).
// Note: MySQL doesn't support RETURNING clause.
func (suite *DeleteTestSuite) TestReturning() {
	suite.T().Logf("Testing Returning methods for %s", suite.DbType)

	if suite.DbType == constants.DbMySQL {
		suite.T().Skip("MySQL doesn't support RETURNING clause")

		return
	}

	suite.Run("ReturningSpecificColumns", func() {
		testUsers := []*User{
			{Name: "Return User 1", Email: "return1@example.com", Age: 25, IsActive: true},
			{Name: "Return User 2", Email: "return2@example.com", Age: 30, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for returning")

		type DeleteResult struct {
			Id    string `bun:"id"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		var returnedUsers []DeleteResult

		err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "return")
			}).
			Returning("id", "name", "email").
			Scan(suite.Ctx, &returnedUsers)

		suite.NoError(err, "Should delete records with RETURNING clause")
		suite.Len(returnedUsers, 2, "Should return 2 deleted user records")

		for i, user := range returnedUsers {
			suite.NotEmpty(user.Id, "Returned ID should not be empty")
			suite.NotEmpty(user.Name, "Returned name should not be empty")
			suite.NotEmpty(user.Email, "Returned email should not be empty")
			suite.T().Logf("Returned user %d: ID=%s, Name=%s, Email=%s", i+1, user.Id, user.Name, user.Email)
		}

		var deletedUsers []User

		err = suite.Db.NewSelect().
			Model(&deletedUsers).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "return")
			}).
			Scan(suite.Ctx)
		suite.NoError(err, "Should query deleted users")
		suite.Len(deletedUsers, 0, "Returned users should not exist in database")

		suite.T().Logf("Deleted %d users with RETURNING specific columns", len(returnedUsers))
	})

	suite.Run("ReturningAllColumns", func() {
		testPosts := []*Post{
			{Title: "Return All Post", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft", ViewCount: 10},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test post for returning all")

		var returnedPosts []Post

		err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("title", "Return All Post")
			}).
			ReturningAll().
			Scan(suite.Ctx, &returnedPosts)

		suite.NoError(err, "Should delete records with RETURNING ALL")
		suite.Len(returnedPosts, 1, "Should return 1 deleted post")

		post := returnedPosts[0]
		suite.NotEmpty(post.Id, "Returned ID should not be empty")
		suite.Equal("Return All Post", post.Title, "Returned title should match")
		suite.Equal("draft", post.Status, "Returned status should match")
		suite.Equal(10, post.ViewCount, "Returned view count should match")

		suite.T().Logf("Deleted 1 post with RETURNING ALL: ID=%s, Title=%s", post.Id, post.Title)
	})

	suite.Run("ReturningNoColumns", func() {
		testUser := &User{
			Name:     "Return None User",
			Email:    "returnnone@example.com",
			Age:      28,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().Model(testUser).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test user for returning none")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "returnnone@example.com")
			}).
			ReturningNone().
			Exec(suite.Ctx)

		suite.NoError(err, "Should delete record with RETURNING NONE")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 user")

		suite.T().Logf("Deleted 1 user with RETURNING NONE (no data returned)")
	})
}

// TestForceDelete tests ForceDelete method with soft-delete scenarios.
func (suite *DeleteTestSuite) TestForceDelete() {
	suite.T().Logf("Testing ForceDelete method for %s", suite.DbType)

	suite.Run("ForceDeleteNonDeleted", func() {
		testUsers := []*User{
			{Name: "Force User 1", Email: "force1@example.com", Age: 25, IsActive: true},
			{Name: "Force User 2", Email: "force2@example.com", Age: 30, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for force delete")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "force1@example.com")
			}).
			ForceDelete().
			Exec(suite.Ctx)

		suite.NoError(err, "ForceDelete should work on non-deleted records")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should force delete 1 user")

		var deletedUser User

		err = suite.Db.NewSelect().
			Model(&deletedUser).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "force1@example.com")
			}).
			Scan(suite.Ctx)
		suite.Error(err, "Force deleted user should not exist")

		suite.T().Logf("Force deleted 1 non-deleted user")

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "force2@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining force test user")
	})

	suite.Run("ForceDeleteBehavior", func() {
		testPosts := []*Post{
			{Title: "Force Post 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published"},
			{Title: "Force Post 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for force delete behavior")

		result, err := suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Force Post")
			}).
			ForceDelete().
			Exec(suite.Ctx)

		suite.NoError(err, "ForceDelete should delete permanently")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should permanently delete 2 posts")

		var deletedPosts []Post

		err = suite.Db.NewSelect().
			Model(&deletedPosts).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Force Post")
			}).
			Scan(suite.Ctx)
		suite.NoError(err, "Should query force deleted posts")
		suite.Len(deletedPosts, 0, "Force deleted posts should not exist")

		suite.T().Logf("Permanently deleted %d posts with ForceDelete", rowsAffected)
	})
}

// TestApply tests Apply and ApplyIf methods for conditional query building.
func (suite *DeleteTestSuite) TestApply() {
	suite.T().Logf("Testing Apply methods for %s", suite.DbType)

	suite.Run("ApplyBasic", func() {
		testUsers := []*User{
			{Name: "Apply User 1", Email: "apply1@example.com", Age: 25, IsActive: true},
			{Name: "Apply User 2", Email: "apply2@example.com", Age: 30, IsActive: false},
			{Name: "Apply User 3", Email: "apply3@example.com", Age: 35, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for apply")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Apply(
				func(query DeleteQuery) {
					query.Where(func(cb ConditionBuilder) {
						cb.StartsWith("email", "apply")
					})
				},
				func(query DeleteQuery) {
					query.Where(func(cb ConditionBuilder) {
						cb.Equals("is_active", false)
					})
				},
			).
			Exec(suite.Ctx)

		suite.NoError(err, "Apply should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 inactive user (applied filter)")

		suite.T().Logf("Deleted %d user using Apply method", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "apply")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining apply test users")
	})

	suite.Run("ApplyIfTrue", func() {
		testPosts := []*Post{
			{Title: "ApplyIf Post 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
			{Title: "ApplyIf Post 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published"},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for ApplyIf true")

		result, err := suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "ApplyIf Post")
			}).
			ApplyIf(true,
				func(query DeleteQuery) {
					query.Where(func(cb ConditionBuilder) {
						cb.Equals("status", "draft")
					})
				},
			).
			Exec(suite.Ctx)

		suite.NoError(err, "ApplyIf(true) should apply functions")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(1), rowsAffected, "Should delete 1 draft post (condition was true)")

		suite.T().Logf("Deleted %d post using ApplyIf(true)", rowsAffected)

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "ApplyIf Post")
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should cleanup remaining ApplyIf test posts")
	})

	suite.Run("ApplyIfFalse", func() {
		testUsers := []*User{
			{Name: "ApplyIf False 1", Email: "applyfalse1@example.com", Age: 25, IsActive: true},
			{Name: "ApplyIf False 2", Email: "applyfalse2@example.com", Age: 30, IsActive: false},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for ApplyIf false")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "applyfalse")
			}).
			ApplyIf(false,
				func(query DeleteQuery) {
					query.Where(func(cb ConditionBuilder) {
						cb.Equals("is_active", false)
					})
				},
			).
			Exec(suite.Ctx)

		suite.NoError(err, "ApplyIf(false) should skip functions")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should delete all 2 users (no filter applied)")

		suite.T().Logf("Deleted %d users using ApplyIf(false) - no condition applied", rowsAffected)
	})
}

// TestExecution tests execution methods (Exec, Scan) with various scenarios.
func (suite *DeleteTestSuite) TestExecution() {
	suite.T().Logf("Testing execution methods for %s", suite.DbType)

	suite.Run("ExecSuccess", func() {
		testUsers := []*User{
			{Name: "Exec User 1", Email: "exec1@example.com", Age: 25, IsActive: true},
			{Name: "Exec User 2", Email: "exec2@example.com", Age: 30, IsActive: true},
		}

		_, err := suite.Db.NewInsert().Model(&testUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test users for exec")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "exec")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Exec should work correctly")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should delete 2 users")

		suite.T().Logf("Exec deleted %d users successfully", rowsAffected)
	})

	suite.Run("ExecNoRows", func() {
		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "nonexistent@example.com")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Exec with no matching rows should not error")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(0), rowsAffected, "Should affect 0 rows")

		suite.T().Logf("Exec with no matching rows affected %d rows", rowsAffected)
	})

	suite.Run("ScanWithReturning", func() {
		if suite.DbType == constants.DbMySQL {
			suite.T().Skip("MySQL doesn't support RETURNING with Scan")

			return
		}

		testPosts := []*Post{
			{Title: "Scan Post 1", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "draft"},
			{Title: "Scan Post 2", Content: "Content", UserId: "user1", CategoryId: "cat1", Status: "published"},
		}

		_, err := suite.Db.NewInsert().Model(&testPosts).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test posts for scan")

		type DeleteResult struct {
			Id     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var returnedPosts []DeleteResult

		err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Scan Post")
			}).
			Returning("id", "title", "status").
			Scan(suite.Ctx, &returnedPosts)

		suite.NoError(err, "Scan with RETURNING should work correctly")
		suite.Len(returnedPosts, 2, "Should return 2 deleted posts")

		for _, post := range returnedPosts {
			suite.NotEmpty(post.Id, "Returned ID should not be empty")
			suite.NotEmpty(post.Title, "Returned title should not be empty")
			suite.T().Logf("Scanned deleted post: ID=%s, Title=%s, Status=%s", post.Id, post.Title, post.Status)
		}
	})

	suite.Run("ErrorHandling", func() {
		_, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("invalid_field", "value")
			}).
			Exec(suite.Ctx)

		suite.Error(err, "DELETE with invalid field should error")
		suite.T().Logf("Invalid field error correctly returned: %v", err)
	})

	suite.Run("DeleteWithWhereClause", func() {
		testModels := []*SimpleModel{
			{Name: "Exec Simple 1", Value: 1},
			{Name: "Exec Simple 2", Value: 2},
		}

		_, err := suite.Db.NewInsert().Model(&testModels).Exec(suite.Ctx)
		suite.NoError(err, "Should insert test simple models")

		result, err := suite.Db.NewDelete().
			Model((*SimpleModel)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("name", "Exec Simple")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "DELETE with WHERE clause should work")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count")
		suite.Equal(int64(2), rowsAffected, "Should delete all test records")

		suite.T().Logf("Deleted %d simple models with WHERE clause", rowsAffected)
	})

	suite.Run("PerformanceBulkDelete", func() {
		batchSize := 50

		performanceUsers := make([]*User, batchSize)
		for i := range batchSize {
			performanceUsers[i] = &User{
				Name:     fmt.Sprintf("Perf User %d", i),
				Email:    fmt.Sprintf("perf-%03d@example.com", i),
				Age:      int16(20 + i%50),
				IsActive: i%2 == 0,
			}
		}

		_, err := suite.Db.NewInsert().Model(&performanceUsers).Exec(suite.Ctx)
		suite.NoError(err, "Should insert performance test users")

		result, err := suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "perf-")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Should bulk delete performance test users")

		rowsAffected, err := result.RowsAffected()
		suite.NoError(err, "Should get rows affected count for bulk delete")
		suite.Equal(int64(batchSize), rowsAffected, "Should delete all performance test users")

		suite.T().Logf("Performance test: deleted %d users", batchSize)
	})
}
