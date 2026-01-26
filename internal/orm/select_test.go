package orm

import (
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/page"
)

// SelectTestSuite tests SELECT operations including basic queries, column selection,
// joins, subqueries, ordering, pagination, locking, set operations, and execution methods
// across all databases (PostgreSQL, MySQL, SQLite).
type SelectTestSuite struct {
	*OrmTestSuite
}

// TestCTE tests Common Table Expression methods (With, WithValues, WithRecursive).
func (suite *SelectTestSuite) TestCTE() {
	suite.T().Logf("Testing CTE methods for %s", suite.dbType)

	suite.Run("WithBasicCTE", func() {
		type PostWithUser struct {
			ID       string `bun:"id"`
			Title    string `bun:"title"`
			UserName string `bun:"user_name"`
		}

		var postsWithUsers []PostWithUser

		err := suite.db.NewSelect().
			With("active_users", func(query SelectQuery) {
				query.Model((*User)(nil)).
					Select("id", "name").
					Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active")
					})
			}).
			Model((*Post)(nil)).
			Select("p.id", "p.title").
			SelectAs("u.name", "user_name").
			Join((*User)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "p.user_id")
			}, "u").
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("p.title").
			Limit(3).
			Scan(suite.ctx, &postsWithUsers)

		suite.NoError(err, "WITH clause should work correctly")
		suite.True(len(postsWithUsers) > 0, "Should return posts with user information")

		for _, post := range postsWithUsers {
			suite.NotEmpty(post.ID, "ID should not be empty")
			suite.NotEmpty(post.Title, "Title should not be empty")
			suite.T().Logf("Post: %s by %s", post.Title, post.UserName)
		}
	})

	suite.Run("WithValuesCTE", func() {
		type StatusValue struct {
			Status string `bun:"status"`
		}

		type StatusInfo struct {
			Status string `bun:"status"`
			Count  int64  `bun:"count"`
		}

		statusValues := []StatusValue{
			{Status: "published"},
			{Status: "draft"},
			{Status: "review"},
		}

		var statusCounts []StatusInfo

		err := suite.db.NewSelect().
			WithValues("status_values", &statusValues).
			Select("sv.status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("p.id")
			}, "count").
			Table("status_values", "sv").
			LeftJoin((*Post)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("sv.status", "p.status")
			}).
			GroupBy("sv.status").
			OrderBy("sv.status").
			Scan(suite.ctx, &statusCounts)

		suite.NoError(err, "WITH VALUES should work correctly")
		suite.True(len(statusCounts) > 0, "Should return status counts")

		for _, status := range statusCounts {
			suite.NotEmpty(status.Status, "Status should not be empty")
			suite.True(status.Count >= 0, "Count should be non-negative")
			suite.T().Logf("Status %s: %d posts", status.Status, status.Count)
		}
	})

	suite.Run("WithRecursiveCTE", func() {
		if suite.dbType == constants.SQLite {
			suite.T().Skip("Skipping for SQLite: bun framework bug causes extra parentheses in generated UNION SQL")
		}

		type CommentHierarchy struct {
			ID       string `bun:"id"`
			Content  string `bun:"content"`
			ParentID string `bun:"parent_id"`
			Level    int    `bun:"level"`
		}

		var commentTree []CommentHierarchy

		err := suite.db.NewSelect().
			WithRecursive("comment_tree", func(query SelectQuery) {
				query.Model((*Post)(nil)).
					Select("id", "category_id", "title", "status").
					SelectExpr(func(ExprBuilder) any {
						return 0
					}, "level").
					Where(func(cb ConditionBuilder) {
						cb.IsNull("category_id")
					}).
					UnionAll(func(unionQuery SelectQuery) {
						unionQuery.Model((*Post)(nil)).
							Select("ct.id", "ct.category_id", "ct.title", "ct.status").
							SelectExpr(func(eb ExprBuilder) any {
								return eb.Add(eb.Column("ct.level"), 1)
							}, "level").
							JoinTable("comment_tree", func(cb ConditionBuilder) {
								cb.EqualsColumn("category_id", "ct.id")
							}, "ct")
					})
			}).
			Table("comment_tree").
			OrderBy("level", "id").
			Limit(10).
			Scan(suite.ctx, &commentTree)

		suite.NoError(err, "WITH RECURSIVE should work when supported")

		for _, comment := range commentTree {
			suite.NotEmpty(comment.ID, "Comment ID should not be empty")
			suite.True(comment.Level >= 0, "Level should be non-negative")
			suite.T().Logf("Comment level %d: %s", comment.Level, comment.Content)
		}
	})
}

// TestSelectAll tests SelectAll method.
func (suite *SelectTestSuite) TestSelectAll() {
	suite.T().Logf("Testing SelectAll for %s", suite.dbType)

	suite.Run("SelectAllUsers", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			SelectAll().
			OrderBy("name").
			Limit(3).
			Scan(suite.ctx)

		suite.NoError(err, "SelectAll should work correctly")
		suite.Len(users, 3, "Should return 3 users")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.NotEmpty(user.Name, "Name should be populated")
			suite.NotEmpty(user.Email, "Email should be populated")
			suite.T().Logf("User: ID=%s, Name=%s, Email=%s", user.ID, user.Name, user.Email)
		}
	})
}

// TestSelectAndSelectAs tests Select and SelectAs methods.
func (suite *SelectTestSuite) TestSelectAndSelectAs() {
	suite.T().Logf("Testing Select and SelectAs for %s", suite.dbType)

	suite.Run("SelectSpecificColumns", func() {
		type UserBasic struct {
			ID    string `bun:"id"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		var users []UserBasic

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "email").
			OrderBy("name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "Select with specific columns should work")
		suite.Len(users, 2, "Should return 2 users")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.NotEmpty(user.Name, "Name should be populated")
			suite.NotEmpty(user.Email, "Email should be populated")
			suite.T().Logf("User: ID=%s, Name=%s, Email=%s", user.ID, user.Name, user.Email)
		}
	})

	suite.Run("SelectWithAlias", func() {
		type PostWithAlias struct {
			ID          string `bun:"id"`
			Title       string `bun:"title"`
			PostStatus  string `bun:"post_status"`
			ViewDisplay string `bun:"view_display"`
		}

		var posts []PostWithAlias

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectAs("status", "post_status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat("Views: ", eb.Column("view_count"))
			}, "view_display").
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "SelectAs should work correctly")
		suite.True(len(posts) > 0, "Should return posts")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.NotEmpty(post.Title, "Title should be populated")
			suite.NotEmpty(post.PostStatus, "Status should be populated")
			suite.Contains(post.ViewDisplay, "Views:", "View display should contain prefix")
			suite.T().Logf("Post: %s (Status: %s, %s)", post.Title, post.PostStatus, post.ViewDisplay)
		}
	})
}

// TestSelectExpr tests SelectExpr method.
func (suite *SelectTestSuite) TestSelectExpr() {
	suite.T().Logf("Testing SelectExpr for %s", suite.dbType)

	suite.Run("SelectExpression", func() {
		type PostWithCalculated struct {
			ID         string `bun:"id"`
			Title      string `bun:"title"`
			StatusDesc string `bun:"status_desc"`
			ViewRange  string `bun:"view_range"`
		}

		var posts []PostWithCalculated

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(func(cond ConditionBuilder) {
						cond.Equals("status", "published")
					}).Then("'Published'")
					cb.When(func(cond ConditionBuilder) {
						cond.Equals("status", "draft")
					}).Then("'Draft'")
					cb.Else("'Other'")
				})
			}, "status_desc").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(func(cond ConditionBuilder) {
						cond.GreaterThan("view_count", 100)
					}).Then(eb.Concat("'High ('", eb.Column("view_count"), "')'"))
					cb.Else(eb.Concat("'Low ('", eb.Column("view_count"), "')'"))
				})
			}, "view_range").
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "SelectExpr should work correctly")
		suite.True(len(posts) > 0, "Should return posts")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.NotEmpty(post.Title, "Title should be populated")
			suite.NotEmpty(post.StatusDesc, "Status description should be calculated")
			suite.NotEmpty(post.ViewRange, "View range should be calculated")
			suite.T().Logf("Post: %s - Status: %s, %s", post.Title, post.StatusDesc, post.ViewRange)
		}
	})

	suite.Run("MultipleSelectExpr", func() {
		type UserWithStats struct {
			ID        string `bun:"id"`
			Name      string `bun:"name"`
			UpperName string `bun:"upper_name"`
			NameLen   int    `bun:"name_len"`
			AgeGroup  string `bun:"age_group"`
		}

		var users []UserWithStats

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("name"))
			}, "upper_name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Length(eb.Column("name"))
			}, "name_len").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(func(cond ConditionBuilder) {
						cond.GreaterThan("age", 30)
					}).Then("'Senior'")
					cb.Else("'Junior'")
				})
			}, "age_group").
			OrderBy("name").
			Limit(3).
			Scan(suite.ctx, &users)

		suite.NoError(err, "Multiple SelectExpr should work correctly")
		suite.True(len(users) > 0, "Should return users")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.NotEmpty(user.Name, "Name should be populated")
			suite.NotEmpty(user.UpperName, "Upper name should be calculated")
			suite.True(user.NameLen > 0, "Name length should be positive")
			suite.NotEmpty(user.AgeGroup, "Age group should be calculated")
			suite.T().Logf("User: %s (%s) - Length: %d, Group: %s",
				user.Name, user.UpperName, user.NameLen, user.AgeGroup)
		}
	})
}

// TestSelectModelColumns tests SelectModelColumns method.
func (suite *SelectTestSuite) TestSelectModelColumns() {
	suite.T().Logf("Testing SelectModelColumns for %s", suite.dbType)

	suite.Run("SelectModelColumnsBasic", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			SelectModelColumns().
			OrderBy("name").
			Limit(2).
			Scan(suite.ctx)

		suite.NoError(err, "SelectModelColumns should work correctly")
		suite.Len(users, 2, "Should return 2 users")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.NotEmpty(user.CreatedAt, "CreatedAt should be populated")
			suite.NotEmpty(user.CreatedBy, "CreatedBy should be populated")
			suite.T().Logf("User: ID=%s, CreatedAt=%s, CreatedBy=%s",
				user.ID, user.CreatedAt, user.CreatedBy)
		}
	})

	suite.Run("SelectModelColumnsWithExpr", func() {
		type UserWithExpr struct {
			ID        string `bun:"id"`
			CreatedAt string `bun:"created_at"`
			CreatedBy string `bun:"created_by"`
			UpdatedAt string `bun:"updated_at"`
			UpdatedBy string `bun:"updated_by"`
			Name      string `bun:"name"`
			NameLen   int    `bun:"name_len"`
		}

		var users []UserWithExpr

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectModelColumns().
			Select("name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Length(eb.Column("name"))
			}, "name_len").
			OrderBy("name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "SelectModelColumns with SelectExpr should work")
		suite.True(len(users) > 0, "Should return users")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.NotEmpty(user.Name, "Name should be populated")
			suite.NotEmpty(user.CreatedAt, "CreatedAt should be populated")
			suite.True(user.NameLen > 0, "Name length should be positive")
			suite.T().Logf("User: ID=%s, Name=%s (len=%d), CreatedAt=%s",
				user.ID, user.Name, user.NameLen, user.CreatedAt)
		}
	})
}

// TestSelectModelPKs tests SelectModelPKs method.
func (suite *SelectTestSuite) TestSelectModelPKs() {
	suite.T().Logf("Testing SelectModelPKs for %s", suite.dbType)

	suite.Run("SelectModelPKsBasic", func() {
		type UserIDOnly struct {
			ID string `bun:"id"`
		}

		var users []UserIDOnly

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectModelPKs().
			OrderBy("id").
			Limit(3).
			Scan(suite.ctx, &users)

		suite.NoError(err, "SelectModelPKs should work correctly")
		suite.Len(users, 3, "Should return 3 users")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.T().Logf("User ID: %s", user.ID)
		}
	})

	suite.Run("SelectModelPKsWithExpr", func() {
		type UserWithExpr struct {
			ID       string `bun:"id"`
			NameDesc string `bun:"name_desc"`
		}

		var users []UserWithExpr

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectModelPKs().
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat("User: ", eb.Column("name"))
			}, "name_desc").
			OrderBy("name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "SelectModelPKs with SelectExpr should work")
		suite.True(len(users) > 0, "Should return users")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.NotEmpty(user.NameDesc, "Name description should be calculated")
			suite.Contains(user.NameDesc, "User:", "Name description should contain prefix")
			suite.T().Logf("User: ID=%s, %s", user.ID, user.NameDesc)
		}
	})
}

// TestExclude tests Exclude and ExcludeAll methods.
func (suite *SelectTestSuite) TestExclude() {
	suite.T().Logf("Testing Exclude methods for %s", suite.dbType)

	suite.Run("ExcludeSpecificColumns", func() {
		type UserWithoutSensitive struct {
			ID        string `bun:"id"`
			Name      string `bun:"name"`
			CreatedAt string `bun:"created_at"`
			CreatedBy string `bun:"created_by"`
			UpdatedAt string `bun:"updated_at"`
			UpdatedBy string `bun:"updated_by"`
		}

		var users []UserWithoutSensitive

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Exclude("name", "email", "age", "is_active", "meta").
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "Exclude should work correctly")
		suite.Len(users, 2, "Should return 2 users")

		for _, user := range users {
			suite.Empty(user.Name, "Name should not be populated")
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.NotEmpty(user.CreatedAt, "CreatedAt should be populated")
			suite.T().Logf("User: ID=%s, CreatedAt=%s", user.ID, user.CreatedAt)
		}
	})

	suite.Run("ExcludeAll", func() {
		type PostWithExpr struct {
			Title         string `bun:"title"`
			StatusDisplay string `bun:"status_display"`
			ViewCategory  string `bun:"view_category"`
		}

		var posts []PostWithExpr

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			ExcludeAll().
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("status"))
			}, "status_display").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(func(cond ConditionBuilder) {
						cond.GreaterThan("view_count", 50)
					}).Then("Popular")
					cb.Else("Normal")
				})
			}, "view_category").
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "ExcludeAll should work correctly")
		suite.True(len(posts) > 0, "Should return posts")

		for _, post := range posts {
			suite.Empty(post.Title, "Title should not be populated")
			suite.NotEmpty(post.StatusDisplay, "Status display should be calculated")
			suite.NotEmpty(post.ViewCategory, "View category should be calculated")
			suite.T().Logf("Post: Status=%s, Category=%s", post.StatusDisplay, post.ViewCategory)
		}
	})
}

// TestSelectMutualExclusivity tests that base column selection methods are mutually exclusive.
func (suite *SelectTestSuite) TestSelectMutualExclusivity() {
	suite.T().Logf("Testing Select method mutual exclusivity for %s", suite.dbType)

	suite.Run("SelectAllOverridesSelect", func() {
		var users1 []User

		err := suite.db.NewSelect().
			Model(&users1).
			Select("name").
			SelectAll().
			Scan(suite.ctx)

		suite.NoError(err, "SelectAll should override Select")
		suite.True(len(users1) > 0, "Should return results")
		suite.NotEmpty(users1[0].Email, "Email should be populated when SelectAll is used")

		suite.T().Logf("SelectAll overrode Select: got %d users with all columns", len(users1))
	})

	suite.Run("SelectOverridesSelectAll", func() {
		type UserNameOnly struct {
			Name string `bun:"name"`
		}

		var users2 []UserNameOnly

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectAll().
			Select("name").
			Scan(suite.ctx, &users2)

		suite.NoError(err, "Select should override SelectAll")
		suite.True(len(users2) > 0, "Should return results")
		suite.NotEmpty(users2[0].Name, "Name should be populated")

		suite.T().Logf("Select overrode SelectAll: got %d users with name only", len(users2))
	})

	suite.Run("SelectModelColumnsOverridesSelectAll", func() {
		var users3 []User

		err := suite.db.NewSelect().
			Model(&users3).
			SelectAll().
			SelectModelColumns().
			Scan(suite.ctx)

		suite.NoError(err, "SelectModelColumns should override SelectAll")
		suite.True(len(users3) > 0, "Should return results")

		suite.T().Logf("SelectModelColumns overrode SelectAll: got %d users", len(users3))
	})

	suite.Run("SelectAllOverridesSelectModelColumns", func() {
		var users4 []User

		err := suite.db.NewSelect().
			Model(&users4).
			SelectModelColumns().
			SelectAll().
			Scan(suite.ctx)

		suite.NoError(err, "SelectAll should override SelectModelColumns")
		suite.True(len(users4) > 0, "Should return results")

		suite.T().Logf("SelectAll overrode SelectModelColumns: got %d users", len(users4))
	})

	suite.Run("SelectModelPKsOverridesSelectModelColumns", func() {
		type UserIDOnly struct {
			ID string `bun:"id,pk"`
		}

		var users5 []UserIDOnly

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectModelColumns().
			SelectModelPKs().
			Scan(suite.ctx, &users5)

		suite.NoError(err, "SelectModelPKs should override SelectModelColumns")
		suite.True(len(users5) > 0, "Should return results")
		suite.NotEmpty(users5[0].ID, "ID should be populated")

		suite.T().Logf("SelectModelPKs overrode SelectModelColumns: got %d users with ID only", len(users5))
	})

	suite.Run("SelectModelColumnsOverridesSelectModelPKs", func() {
		var users6 []User

		err := suite.db.NewSelect().
			Model(&users6).
			SelectModelPKs().
			SelectModelColumns().
			Scan(suite.ctx)

		suite.NoError(err, "SelectModelColumns should override SelectModelPKs")
		suite.True(len(users6) > 0, "Should return results")
		suite.NotEmpty(users6[0].Name, "Name should be populated when SelectModelColumns is used")

		suite.T().Logf("SelectModelColumns overrode SelectModelPKs: got %d users", len(users6))
	})
}

// TestSelectExprCumulative tests that SelectExpr is cumulative and works with any base selection.
func (suite *SelectTestSuite) TestSelectExprCumulative() {
	suite.T().Logf("Testing SelectExpr cumulative behavior for %s", suite.dbType)

	suite.Run("SelectExprWithSelectAll", func() {
		type UserWithComputed struct {
			ID       string `bun:"id"`
			Name     string `bun:"name"`
			Email    string `bun:"email"`
			Age      int16  `bun:"age"`
			AgeGroup string `bun:"age_group"`
		}

		var users1 []UserWithComputed

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectAll().
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(func(cond ConditionBuilder) {
						cond.GreaterThan("age", 30)
					}).Then("'senior'")
					cb.Else("'junior'")
				})
			}, "age_group").
			OrderBy("name").
			Scan(suite.ctx, &users1)

		suite.NoError(err, "SelectExpr should work with SelectAll")
		suite.True(len(users1) > 0, "Should return results")
		suite.NotEmpty(users1[0].Name, "Name should be populated")
		suite.NotEmpty(users1[0].AgeGroup, "Computed age_group should be populated")

		for _, user := range users1 {
			suite.T().Logf("User: %s, Age=%d, AgeGroup=%s", user.Name, user.Age, user.AgeGroup)
		}
	})

	suite.Run("SelectExprWithSelect", func() {
		type UserWithRowNum struct {
			Name   string `bun:"name"`
			Email  string `bun:"email"`
			RowNum int    `bun:"row_num"`
		}

		var users2 []UserWithRowNum

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("name", "email").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.RowNumber(func(rb RowNumberBuilder) {
					rb.Over().OrderBy("name")
				})
			}, "row_num").
			OrderBy("name").
			Scan(suite.ctx, &users2)

		suite.NoError(err, "SelectExpr should work with Select")
		suite.True(len(users2) > 0, "Should return results")
		suite.NotEmpty(users2[0].Name, "Name should be populated")
		suite.True(users2[0].RowNum > 0, "Row number should be populated")

		for _, user := range users2 {
			suite.T().Logf("User: %s, RowNum=%d", user.Name, user.RowNum)
		}
	})

	suite.Run("MultipleSelectExprCallsCumulative", func() {
		type UserWithMultipleComputed struct {
			Name      string `bun:"name"`
			UpperName string `bun:"upper_name"`
			NameLen   int    `bun:"name_len"`
		}

		var users3 []UserWithMultipleComputed

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("name"))
			}, "upper_name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Length(eb.Column("name"))
			}, "name_len").
			OrderBy("name").
			Scan(suite.ctx, &users3)

		suite.NoError(err, "Multiple SelectExpr should be cumulative")
		suite.True(len(users3) > 0, "Should return results")
		suite.NotEmpty(users3[0].Name, "Name should be populated")
		suite.NotEmpty(users3[0].UpperName, "Upper name should be populated")
		suite.True(users3[0].NameLen > 0, "Name length should be populated")

		for _, user := range users3 {
			suite.T().Logf("User: %s, UpperName=%s, NameLen=%d", user.Name, user.UpperName, user.NameLen)
		}
	})

	suite.Run("SelectExprPreservedWhenSwitchingBaseSelection", func() {
		type UserAllWithComputed struct {
			ID       string `bun:"id"`
			Name     string `bun:"name"`
			Email    string `bun:"email"`
			Age      int16  `bun:"age"`
			IsActive bool   `bun:"is_active"`
			RowNum   int    `bun:"row_num"`
		}

		var users4 []UserAllWithComputed

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.RowNumber(func(rb RowNumberBuilder) {
					rb.Over().OrderBy("name")
				})
			}, "row_num").
			SelectAll().
			OrderBy("name").
			Scan(suite.ctx, &users4)

		suite.NoError(err, "SelectExpr should be preserved when switching to SelectAll")
		suite.True(len(users4) > 0, "Should return results")
		suite.NotEmpty(users4[0].Name, "Name should be populated")
		suite.NotEmpty(users4[0].Email, "Email should be populated (SelectAll)")
		suite.True(users4[0].RowNum > 0, "Row number should still be populated (SelectExpr preserved)")

		for _, user := range users4 {
			suite.T().Logf("User: %s, Email=%s, RowNum=%d", user.Name, user.Email, user.RowNum)
		}
	})

	suite.Run("SelectExprWithSelectModelColumns", func() {
		type UserModelWithTotal struct {
			ID         string `bun:"id"`
			CreatedAt  string `bun:"created_at"`
			CreatedBy  string `bun:"created_by"`
			UpdatedAt  string `bun:"updated_at"`
			UpdatedBy  string `bun:"updated_by"`
			Name       string `bun:"name"`
			Email      string `bun:"email"`
			Age        int16  `bun:"age"`
			IsActive   bool   `bun:"is_active"`
			Meta       string `bun:"meta"`
			TotalCount int64  `bun:"total_count"`
		}

		var users5 []UserModelWithTotal

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectModelColumns().
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinCount(func(wcb WindowCountBuilder) {
					wcb.All().Over()
				})
			}, "total_count").
			OrderBy("name").
			Scan(suite.ctx, &users5)

		suite.NoError(err, "SelectExpr should work with SelectModelColumns")
		suite.True(len(users5) > 0, "Should return results")
		suite.NotEmpty(users5[0].Name, "Name should be populated")
		suite.True(users5[0].TotalCount > 0, "Total count should be populated")

		for _, user := range users5 {
			suite.T().Logf("User: %s, TotalCount=%d", user.Name, user.TotalCount)
		}
	})
}

// TestSelectIdempotency tests that SelectModelColumns and SelectModelPKs are idempotent.
func (suite *SelectTestSuite) TestSelectIdempotency() {
	suite.T().Logf("Testing Select method idempotency for %s", suite.dbType)

	suite.Run("MultipleSelectModelColumnsCalls", func() {
		var users1 []User

		err := suite.db.NewSelect().
			Model(&users1).
			SelectModelColumns().
			SelectModelColumns().
			SelectModelColumns().
			Scan(suite.ctx)

		suite.NoError(err, "Multiple SelectModelColumns should not cause errors")
		suite.True(len(users1) > 0, "Should return results")
		suite.NotEmpty(users1[0].Name, "Name should be populated")

		suite.T().Logf("Multiple SelectModelColumns calls: got %d users", len(users1))
	})

	suite.Run("MultipleSelectModelPKsCalls", func() {
		type UserIDOnly struct {
			ID string `bun:"id,pk"`
		}

		var users2 []UserIDOnly

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectModelPKs().
			SelectModelPKs().
			SelectModelPKs().
			Scan(suite.ctx, &users2)

		suite.NoError(err, "Multiple SelectModelPKs should not cause errors")
		suite.True(len(users2) > 0, "Should return results")
		suite.NotEmpty(users2[0].ID, "ID should be populated")

		suite.T().Logf("Multiple SelectModelPKs calls: got %d users", len(users2))
	})

	suite.Run("MultipleSelectAllCalls", func() {
		var users3 []User

		err := suite.db.NewSelect().
			Model(&users3).
			SelectAll().
			SelectAll().
			SelectAll().
			Scan(suite.ctx)

		suite.NoError(err, "Multiple SelectAll should not cause errors")
		suite.True(len(users3) > 0, "Should return results")
		suite.NotEmpty(users3[0].Email, "All columns should be populated")

		suite.T().Logf("Multiple SelectAll calls: got %d users", len(users3))
	})
}

// TestDistinct tests Distinct, DistinctOnColumns, and DistinctOnExpr methods.
func (suite *SelectTestSuite) TestDistinct() {
	suite.T().Logf("Testing Distinct methods for %s", suite.dbType)

	suite.Run("BasicDistinct", func() {
		type DistinctStatus struct {
			Status string `bun:"status"`
		}

		var statuses []DistinctStatus

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Distinct().
			Select("status").
			OrderBy("status").
			Scan(suite.ctx, &statuses)

		suite.NoError(err, "DISTINCT should work correctly")
		suite.True(len(statuses) > 0, "Should return distinct statuses")

		for _, status := range statuses {
			suite.NotEmpty(status.Status, "Status should not be empty")
			suite.T().Logf("Distinct status: %s", status.Status)
		}
	})

	suite.Run("DistinctOnColumns", func() {
		// DISTINCT ON is PostgreSQL-specific, not supported by MySQL or SQLite
		if suite.dbType != "postgres" {
			suite.T().Skipf("DISTINCT ON test skipped for %s", suite.dbType)

			return
		}

		type DistinctPost struct {
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var posts []DistinctPost

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			DistinctOnColumns("title").
			Select("title", "status").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "DISTINCT ON columns should work when supported")
		suite.True(len(posts) > 0, "Should return distinct posts")

		for _, post := range posts {
			suite.NotEmpty(post.Title, "Title should not be empty")
			suite.T().Logf("Distinct post: %s (Status: %s)", post.Title, post.Status)
		}
	})

	suite.Run("DistinctOnExpr", func() {
		// DISTINCT ON is PostgreSQL-specific, not supported by MySQL or SQLite
		if suite.dbType != "postgres" {
			suite.T().Skipf("DISTINCT ON test skipped for %s", suite.dbType)

			return
		}

		type DistinctExpr struct {
			ID         string `bun:"id"`
			Title      string `bun:"title"`
			StatusDesc string `bun:"status_desc"`
		}

		var posts []DistinctExpr

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			DistinctOnExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("title"))
			}).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("status"))
			}, "status_desc").
			OrderByExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("title"))
			}).
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "DISTINCT ON expression should work when supported")
		suite.True(len(posts) > 0, "Should return distinct expression posts")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should not be empty")
			suite.T().Logf("Distinct expr post: %s (Status: %s)", post.Title, post.StatusDesc)
		}
	})
}

// TestModelAndTable tests Model, ModelTable, Table, TableFrom, TableExpr, and TableSubQuery methods.
func (suite *SelectTestSuite) TestModelAndTable() {
	suite.T().Logf("Testing Model and Table methods for %s", suite.dbType)

	suite.Run("ModelAndModelTable", func() {
		type PostFromUserTable struct {
			ID    string `bun:"id"`
			Title string `bun:"title"`
		}

		var posts []PostFromUserTable

		err := suite.db.NewSelect().
			ModelTable("test_user", "u").
			Select("u.id", "u.name").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("u.is_active")
			}).
			OrderBy("u.name").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "ModelTable should work correctly")
		suite.True(len(posts) > 0, "Should return users from specified table")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.T().Logf("User from table: ID=%s, Name=%s", post.ID, post.Title)
		}
	})

	suite.Run("TableAndAlias", func() {
		type TableWithAlias struct {
			ID    string `bun:"id"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		var users []TableWithAlias

		err := suite.db.NewSelect().
			Table("test_user", "u").
			Select("u.id", "u.name", "u.email").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("u.is_active")
			}).
			OrderBy("u.name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "Table with alias should work correctly")
		suite.True(len(users) > 0, "Should return users from table with alias")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.T().Logf("User with alias: ID=%s, Name=%s", user.ID, user.Name)
		}
	})

	suite.Run("TableFrom", func() {
		type UserFromModel struct {
			ID   string `bun:"id"`
			Name string `bun:"name"`
		}

		var users []UserFromModel

		err := suite.db.NewSelect().
			TableFrom((*User)(nil), "u").
			Select("u.id", "u.name").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("u.is_active")
			}).
			OrderBy("u.name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "TableFrom should work correctly")
		suite.True(len(users) > 0, "Should return users from model")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.T().Logf("User from model: ID=%s, Name=%s", user.ID, user.Name)
		}
	})

	suite.Run("TableExpr", func() {
		type ExprTable struct {
			ID   string `bun:"id"`
			Name string `bun:"name"`
		}

		var users []ExprTable

		err := suite.db.NewSelect().
			TableExpr(func(eb ExprBuilder) any {
				return eb.Expr("(SELECT id, name FROM ? WHERE is_active = ?)",
					bun.Name("test_user"),
					true)
			}, "active_users").
			Select("active_users.id", "active_users.name").
			OrderBy("active_users.name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "TableExpr should work when supported")
		suite.True(len(users) > 0, "Should return users from expression table")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.T().Logf("User from expr: ID=%s, Name=%s", user.ID, user.Name)
		}
	})

	suite.Run("TableSubQuery", func() {
		type SubQueryTable struct {
			ID   string `bun:"id"`
			Name string `bun:"name"`
		}

		var users []SubQueryTable

		err := suite.db.NewSelect().
			TableSubQuery(func(query SelectQuery) {
				query.Model((*User)(nil)).
					Select("id", "name").
					Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active")
					})
			}, "active_users").
			Select("active_users.id", "active_users.name").
			OrderBy("active_users.name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "TableSubQuery should work correctly")
		suite.True(len(users) > 0, "Should return users from subquery table")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.T().Logf("User from subquery: ID=%s, Name=%s", user.ID, user.Name)
		}
	})
}

// TestJoins tests all join types and variants.
func (suite *SelectTestSuite) TestJoins() {
	suite.T().Logf("Testing Join methods for %s", suite.dbType)

	suite.Run("BasicInnerJoin", func() {
		type PostWithUser struct {
			ID       string `bun:"id"`
			Title    string `bun:"title"`
			UserName string `bun:"user_name"`
		}

		var posts []PostWithUser

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectAs("u.name", "user_name").
			Join((*User)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "user_id")
			}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "INNER JOIN should work correctly")
		suite.True(len(posts) > 0, "Should return posts with user info")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.T().Logf("Post: %s by %s", post.Title, post.UserName)
		}
	})

	suite.Run("LeftJoin", func() {
		type CategoryWithPostCount struct {
			ID        string `bun:"id"`
			Name      string `bun:"name"`
			PostCount int64  `bun:"post_count"`
		}

		var categories []CategoryWithPostCount

		err := suite.db.NewSelect().
			Model((*Category)(nil)).
			Select("id", "name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("p.id")
			}, "post_count").
			LeftJoin((*Post)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("p.category_id", "id")
			}).
			GroupBy("id", "name").
			OrderBy("name").
			Scan(suite.ctx, &categories)

		suite.NoError(err, "LEFT JOIN should work correctly")
		suite.True(len(categories) > 0, "Should return categories with post counts")

		for _, category := range categories {
			suite.NotEmpty(category.ID, "ID should be populated")
			suite.True(category.PostCount >= 0, "Post count should be non-negative")
			suite.T().Logf("Category: %s (%d posts)", category.Name, category.PostCount)
		}
	})

	suite.Run("RightJoin", func() {
		type UserWithPosts struct {
			ID        string `bun:"id"`
			Name      string `bun:"name"`
			PostID    string `bun:"post_id"`
			PostTitle string `bun:"post_title"`
		}

		var users []UserWithPosts

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			SelectAs("p.id", "post_id").
			SelectAs("p.title", "post_title").
			RightJoin((*Post)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("p.user_id", "id")
			}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("is_active", true)
			}).
			OrderBy("name").
			Limit(3).
			Scan(suite.ctx, &users)

		suite.NoError(err, "RIGHT JOIN should work when supported")
		suite.True(len(users) > 0, "Should return users with posts")

		for _, user := range users {
			suite.T().Logf("User: %s, Post: %s", user.Name, user.PostTitle)
		}
	})

	suite.Run("FullJoin", func() {
		if suite.dbType == constants.MySQL {
			suite.T().Skip("Skipping for MySQL: FULL JOIN not supported (use LEFT JOIN UNION RIGHT JOIN instead)")

			return
		}

		type UserWithPosts struct {
			ID        string `bun:"id"`
			Name      string `bun:"name"`
			PostID    string `bun:"post_id"`
			PostTitle string `bun:"post_title"`
		}

		var users []UserWithPosts

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			SelectAs("p.id", "post_id").
			SelectAs("p.title", "post_title").
			FullJoin((*Post)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("p.user_id", "id")
			}).
			OrderBy("name").
			Limit(2).
			Scan(suite.ctx, &users)

		suite.NoError(err, "FULL JOIN should work when supported")
		suite.True(len(users) > 0, "Should return users with full join")

		for _, user := range users {
			suite.T().Logf("User with FULL JOIN: %s - Post: %s", user.Name, user.PostTitle)
		}
	})

	suite.Run("CrossJoin", func() {
		type UserCategoryCross struct {
			UserID       string `bun:"user_id"`
			UserName     string `bun:"user_name"`
			CategoryID   string `bun:"category_id"`
			CategoryName string `bun:"category_name"`
		}

		var cross []UserCategoryCross

		err := suite.db.NewSelect().
			Table("test_user", "u").
			SelectAs("u.id", "user_id").
			SelectAs("u.name", "user_name").
			SelectAs("c.id", "category_id").
			SelectAs("c.name", "category_name").
			CrossJoinTable("test_category", "c").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("u.is_active")
			}).
			OrderBy("u.name", "c.name").
			Limit(4).
			Scan(suite.ctx, &cross)

		suite.NoError(err, "CROSS JOIN should work when supported")
		suite.True(len(cross) > 0, "Should return cross product")

		for _, item := range cross {
			suite.T().Logf("Cross: %s - %s", item.UserName, item.CategoryName)
		}
	})

	suite.Run("JoinWithTable", func() {
		type PostWithCategory struct {
			ID           string `bun:"id"`
			Title        string `bun:"title"`
			CategoryName string `bun:"category_name"`
		}

		var posts []PostWithCategory

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectAs("c.name", "category_name").
			JoinTable("test_category", func(cb ConditionBuilder) {
				cb.EqualsColumn("c.id", "category_id")
			}, "c").
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "JOIN with table name should work correctly")
		suite.True(len(posts) > 0, "Should return posts with categories")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.T().Logf("Post: %s in %s", post.Title, post.CategoryName)
		}
	})

	suite.Run("JoinWithSubQuery", func() {
		type PostWithActiveUser struct {
			ID       string `bun:"id"`
			Title    string `bun:"title"`
			UserName string `bun:"user_name"`
		}

		var posts []PostWithActiveUser

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectAs("active_users.name", "user_name").
			JoinSubQuery(
				func(subquery SelectQuery) {
					subquery.Model((*User)(nil)).
						Select("id", "name").
						Where(func(cb ConditionBuilder) {
							cb.IsTrue("is_active")
						})
				},
				func(cb ConditionBuilder) {
					cb.EqualsColumn("active_users.id", "user_id")
				},
				"active_users").
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "JOIN with subquery should work correctly")
		suite.True(len(posts) > 0, "Should return posts with active users")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.T().Logf("Post: %s by %s", post.Title, post.UserName)
		}
	})
}

// TestJoinRelations tests JoinRelations method with RelationSpec.
func (suite *SelectTestSuite) TestJoinRelations() {
	suite.T().Logf("Testing JoinRelations for %s", suite.dbType)

	// Define result struct for JoinRelations tests
	type PostWithUserName struct {
		ID       string `bun:"id"`
		Title    string `bun:"title"`
		UserName string `bun:"user_name"`
	}

	suite.Run("BasicJoinRelations", func() {
		var posts []PostWithUserName

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("p.id", "p.title").
			JoinRelations(&RelationSpec{
				Model:         (*User)(nil),
				Alias:         "u",
				ForeignColumn: "user_id",
				SelectedColumns: []ColumnInfo{
					{Name: "name", Alias: "user_name"},
				},
			}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "JoinRelations should work correctly")
		suite.True(len(posts) > 0, "Should return posts with user names")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.NotEmpty(post.UserName, "User name should be loaded via JoinRelations")
			suite.T().Logf("Post: %s by %s", post.Title, post.UserName)
		}
	})

	suite.Run("JoinRelationsMultiple", func() {
		type PostWithUserAndCategory struct {
			ID           string `bun:"id"`
			Title        string `bun:"title"`
			UserName     string `bun:"user_name"`
			CategoryName string `bun:"category_name"`
		}

		var posts []PostWithUserAndCategory

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("p.id", "p.title").
			JoinRelations(
				&RelationSpec{
					Model:         (*User)(nil),
					Alias:         "u",
					ForeignColumn: "user_id",
					SelectedColumns: []ColumnInfo{
						{Name: "name", Alias: "user_name"},
					},
				},
				&RelationSpec{
					Model:         (*Category)(nil),
					Alias:         "c",
					ForeignColumn: "category_id",
					SelectedColumns: []ColumnInfo{
						{Name: "name", Alias: "category_name"},
					},
				},
			).
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "Multiple JoinRelations should work correctly")
		suite.True(len(posts) > 0, "Should return posts with user and category names")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")
			suite.NotEmpty(post.UserName, "User name should be loaded")
			suite.NotEmpty(post.CategoryName, "Category name should be loaded")
			suite.T().Logf("Post: %s by %s in %s", post.Title, post.UserName, post.CategoryName)
		}
	})

	suite.Run("JoinRelationsWithJoinType", func() {
		type PostWithCategory struct {
			ID           string `bun:"id"`
			Title        string `bun:"title"`
			CategoryName string `bun:"category_name"`
		}

		var posts []PostWithCategory

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("p.id", "p.title").
			JoinRelations(&RelationSpec{
				Model:         (*Category)(nil),
				Alias:         "c",
				JoinType:      JoinInner,
				ForeignColumn: "category_id",
				SelectedColumns: []ColumnInfo{
					{Name: "name", Alias: "category_name"},
				},
			}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "JoinRelations with INNER JOIN should work correctly")
		suite.True(len(posts) > 0, "Should return posts with category names")

		for _, post := range posts {
			suite.NotEmpty(post.CategoryName, "Category name should be loaded with INNER JOIN")
		}
	})

	suite.Run("JoinRelationsWithCustomCondition", func() {
		type PostWithActiveUser struct {
			ID       string `bun:"id"`
			Title    string `bun:"title"`
			UserName string `bun:"user_name"`
		}

		var posts []PostWithActiveUser

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("p.id", "p.title").
			JoinRelations(&RelationSpec{
				Model:         (*User)(nil),
				Alias:         "u",
				ForeignColumn: "user_id",
				SelectedColumns: []ColumnInfo{
					{Name: "name", Alias: "user_name"},
				},
				On: func(cb ConditionBuilder) {
					cb.Equals("u.is_active", true)
				},
			}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &posts)

		suite.NoError(err, "JoinRelations with custom condition should work correctly")
		suite.True(len(posts) > 0, "Should return posts with active users only")
	})

	suite.Run("RelationMethod", func() {
		var posts []Post

		err := suite.db.NewSelect().
			Model(&posts).
			Relation("User").
			Relation("Category").
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx)

		suite.NoError(err, "Relation method should load related objects")
		suite.True(len(posts) > 0, "Should return posts with relations")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")

			if post.User != nil {
				suite.NotEmpty(post.User.Name, "User relation should be loaded")
				suite.T().Logf("Post: %s by %s", post.Title, post.User.Name)
			}

			if post.Category != nil {
				suite.NotEmpty(post.Category.Name, "Category relation should be loaded")
			}
		}
	})

	suite.Run("RelationMethodWithApply", func() {
		var posts []Post

		err := suite.db.NewSelect().
			Model(&posts).
			Relation("User", func(query SelectQuery) {
				// Customize User relation to only select specific columns
				query.Select("id", "name", "email")
			}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx)

		suite.NoError(err, "Relation with apply should work correctly")
		suite.True(len(posts) > 0, "Should return posts with customized user relations")

		for _, post := range posts {
			suite.NotEmpty(post.ID, "ID should be populated")

			if post.User != nil {
				suite.NotEmpty(post.User.Name, "User relation should be loaded with custom select")
				suite.NotEmpty(post.User.Email, "User email should be loaded")
				suite.T().Logf("Post: %s by %s (%s)", post.Title, post.User.Name, post.User.Email)
			}
		}
	})
}

// TestWhere tests Where, WherePK, WhereDeleted, and IncludeDeleted methods.
func (suite *SelectTestSuite) TestWhere() {
	suite.T().Logf("Testing Where methods for %s", suite.dbType)

	suite.Run("BasicWhere", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active").
					GreaterThan("age", 25)
			}).
			OrderBy("name").
			Limit(3).
			Scan(suite.ctx)

		suite.NoError(err, "WHERE with conditions should work correctly")
		suite.True(len(users) > 0, "Should return active users older than 25")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active")
			suite.True(user.Age > 25, "User age should be greater than 25")
			suite.T().Logf("User: %s (age=%d, active=%t)", user.Name, user.Age, user.IsActive)
		}
	})

	suite.Run("WherePK", func() {
		var firstUser User

		err := suite.db.NewSelect().
			Model(&firstUser).
			OrderBy("name").
			Limit(1).
			Scan(suite.ctx)
		suite.NoError(err, "Should fetch first user")

		var user User

		user.ID = firstUser.ID
		err = suite.db.NewSelect().
			Model(&user).
			WherePK().
			Scan(suite.ctx)

		suite.NoError(err, "WHERE PK should work correctly")
		suite.Equal(firstUser.ID, user.ID, "Should find user by primary key")
		suite.Equal(firstUser.Email, user.Email, "Should match email")

		suite.T().Logf("Found user by PK: %s (%s)", user.Name, user.Email)
	})
}

// TestGroupByAndHaving tests GroupBy, GroupByExpr, and Having methods.
func (suite *SelectTestSuite) TestGroupByAndHaving() {
	suite.T().Logf("Testing GroupBy and Having methods for %s", suite.dbType)

	suite.Run("BasicGroupBy", func() {
		type UserCount struct {
			Age   int16 `bun:"age"`
			Count int64 `bun:"count"`
		}

		var userCounts []UserCount

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("id")
			}, "count").
			GroupBy("age").
			OrderBy("age").
			Scan(suite.ctx, &userCounts)

		suite.NoError(err, "GROUP BY should work correctly")
		suite.True(len(userCounts) > 0, "Should return user counts by age")

		for _, uc := range userCounts {
			suite.True(uc.Count > 0, "Count should be positive")
			suite.T().Logf("Age %d: %d users", uc.Age, uc.Count)
		}
	})

	suite.Run("GroupByExpr", func() {
		type UserByAgeGroup struct {
			AgeGroup string `bun:"age_group"`
			Count    int64  `bun:"count"`
		}

		var ageGroups []UserByAgeGroup

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(
						func(cond ConditionBuilder) {
							cond.GreaterThan("age", 30)
						}).
						Then("Senior").
						Else("Junior")
				})
			}, "age_group").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("id")
			}, "count").
			GroupByExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(
						func(cond ConditionBuilder) {
							cond.GreaterThan("age", 30)
						}).
						Then("Senior").
						Else("Junior")
				})
			}).
			OrderByExpr(func(ExprBuilder) any {
				// Use positional reference to the first select column (age_group)
				return 1
			}).
			Scan(suite.ctx, &ageGroups)

		suite.NoError(err, "GROUP BY expression should work correctly")
		suite.True(len(ageGroups) > 0, "Should return users by age group")

		for _, group := range ageGroups {
			suite.True(group.Count > 0, "Count should be positive")
			suite.T().Logf("Age group %s: %d users", group.AgeGroup, group.Count)
		}
	})

	suite.Run("Having", func() {
		type AgeStats struct {
			Age   int16 `bun:"age"`
			Count int64 `bun:"count"`
		}

		var ageStats []AgeStats

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("id")
			}, "count").
			GroupBy("age").
			Having(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThanOrEqual(eb.CountColumn("id"), 1)
				})
			}).
			OrderBy("age").
			Scan(suite.ctx, &ageStats)

		suite.NoError(err, "HAVING should work correctly")
		suite.True(len(ageStats) > 0, "Should return ages with users")

		for _, age := range ageStats {
			suite.True(age.Count >= 1, "Count should be at least 1")
			suite.T().Logf("Age %d: %d users", age.Age, age.Count)
		}
	})
}

// TestOrderBy tests OrderBy, OrderByDesc, and OrderByExpr methods.
func (suite *SelectTestSuite) TestOrderBy() {
	suite.T().Logf("Testing OrderBy methods for %s", suite.dbType)

	suite.Run("OrderByColumns", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			Select("id", "name", "age").
			OrderBy("age").
			OrderByDesc("name").
			Limit(5).
			Scan(suite.ctx)

		suite.NoError(err, "ORDER BY should work correctly")
		suite.Len(users, 3, "Should return 3 users")

		// Verify ordering by age ascending
		for i := 1; i < len(users); i++ {
			suite.True(users[i-1].Age <= users[i].Age,
				"Users should be ordered by age ascending")
		}

		for _, user := range users {
			suite.T().Logf("User: %s (age=%d)", user.Name, user.Age)
		}
	})

	suite.Run("OrderByExpr", func() {
		type UserWithComputedOrder struct {
			ID       string `bun:"id"`
			Name     string `bun:"name"`
			Age      int16  `bun:"age"`
			OrderKey int    `bun:"order_key"`
		}

		var users []UserWithComputedOrder

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(
						func(cond ConditionBuilder) {
							cond.IsTrue("is_active")
						}).
						Then(eb.Add(eb.Column("age"), 100)).
						Else(eb.Column("age"))
				})
			}, "order_key").
			OrderByExpr(func(eb ExprBuilder) any {
				return eb.Case(func(cb CaseBuilder) {
					cb.When(
						func(cond ConditionBuilder) {
							cond.IsTrue("is_active")
						}).
						Then(eb.Add(eb.Column("age"), 100)).
						Else(eb.Column("age"))
				})
			}).
			Limit(5).
			Scan(suite.ctx, &users)

		suite.NoError(err, "ORDER BY expression should work correctly")
		suite.True(len(users) > 0, "Should return users with computed ordering")

		for _, user := range users {
			suite.NotEmpty(user.ID, "ID should be populated")
			suite.T().Logf("User: %s (age=%d, order_key=%d)", user.Name, user.Age, user.OrderKey)
		}
	})
}

// TestPagination tests Limit, Offset, and Paginate methods.
func (suite *SelectTestSuite) TestPagination() {
	suite.T().Logf("Testing Pagination methods for %s", suite.dbType)

	suite.Run("LimitAndOffset", func() {
		// Get total count first
		totalCount, err := suite.db.NewSelect().
			Model((*User)(nil)).
			Count(suite.ctx)
		suite.NoError(err, "Should count total users")

		// Get first page
		var page1 []User

		err = suite.db.NewSelect().
			Model(&page1).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx)

		suite.NoError(err, "First page should work")
		suite.Len(page1, 2, "First page should have 2 users")

		// Get second page
		var page2 []User

		err = suite.db.NewSelect().
			Model(&page2).
			OrderBy("id").
			Limit(2).
			Offset(2).
			Scan(suite.ctx)

		suite.NoError(err, "Second page should work")
		suite.True(len(page2) > 0, "Second page should have users")

		// Verify no overlap
		if len(page1) > 0 && len(page2) > 0 {
			suite.NotEqual(page1[0].ID, page2[0].ID, "Pages should not overlap")
		}

		suite.T().Logf("Total: %d, Page1: %d users, Page2: %d users",
			totalCount, len(page1), len(page2))
	})

	suite.Run("Paginate", func() {
		// Create pageable request for page 2, size 2
		pageable := page.Pageable{
			Page: 2,
			Size: 2,
		}

		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			OrderBy("id").
			Paginate(pageable).
			Scan(suite.ctx)

		suite.NoError(err, "Paginate should work correctly")
		suite.True(len(users) > 0, "Should return paginated results")

		suite.T().Logf("Page %d (size %d): %d users",
			pageable.Page, pageable.Size, len(users))
	})
}

// TestLocking tests ForShare and ForUpdate methods.
func (suite *SelectTestSuite) TestLocking() {
	suite.T().Logf("Testing Locking methods for %s", suite.dbType)

	// SQLite doesn't support row-level locking (FOR SHARE/FOR UPDATE)
	if suite.dbType == constants.SQLite {
		suite.T().Skip("Skipping for SQLite: row-level locking (FOR SHARE/FOR UPDATE) not supported, uses database-level locking instead")

		return
	}

	suite.Run("ForShare", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			ForShare().
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			Limit(2).
			Scan(suite.ctx)

		suite.NoError(err, "FOR SHARE should work when supported")
		suite.True(len(users) > 0, "Should return users with share lock")

		for _, user := range users {
			suite.T().Logf("User with share lock: %s", user.Name)
		}
	})

	suite.Run("ForUpdate", func() {
		var posts []Post

		err := suite.db.NewSelect().
			Model(&posts).
			ForUpdate().
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "draft")
			}).
			OrderBy("id").
			Limit(1).
			Scan(suite.ctx)

		suite.NoError(err, "FOR UPDATE should work when supported")
		suite.True(len(posts) > 0, "Should return posts with update lock")

		for _, post := range posts {
			suite.T().Logf("Post with update lock: %s", post.Title)
		}
	})

	suite.Run("ForUpdateNoWait", func() {
		var posts []Post

		err := suite.db.NewSelect().
			Model(&posts).
			ForUpdateNoWait().
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "draft")
			}).
			OrderBy("id").
			Limit(1).
			Scan(suite.ctx)

		suite.NoError(err, "FOR UPDATE NOWAIT should work when supported")
		suite.True(len(posts) > 0, "Should return posts with NOWAIT lock")

		for _, post := range posts {
			suite.T().Logf("Post with NOWAIT lock: %s", post.Title)
		}
	})

	suite.Run("ForUpdateSkipLocked", func() {
		var posts []Post

		err := suite.db.NewSelect().
			Model(&posts).
			ForUpdateSkipLocked().
			Where(func(cb ConditionBuilder) {
				cb.Equals("status", "published")
			}).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx)

		suite.NoError(err, "FOR UPDATE SKIP LOCKED should work when supported")
		suite.True(len(posts) > 0, "Should return posts with SKIP LOCKED option")

		for _, post := range posts {
			suite.T().Logf("Post with SKIP LOCKED: %s", post.Title)
		}
	})
}

// TestSetOperations tests Union, Intersect, and Except methods.
func (suite *SelectTestSuite) TestSetOperations() {
	suite.T().Logf("Testing Set Operations for %s", suite.dbType)

	if suite.dbType == constants.SQLite {
		suite.T().Skip("Skipping for SQLite: bun framework bug causes extra parentheses in generated set operation SQL, resulting in syntax errors")

		return
	}

	suite.Run("Union", func() {
		type CombinedResult struct {
			Name string `bun:"name"`
			Type string `bun:"type"`
		}

		var results []CombinedResult

		err := suite.db.NewSelect().
			Table("test_user").
			Select("name").
			SelectExpr(func(ExprBuilder) any {
				return "user"
			}, "type").
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			Union(func(query SelectQuery) {
				query.Table("test_category").
					Select("name").
					SelectExpr(func(ExprBuilder) any {
						return "category"
					}, "type")
			}).
			OrderBy("name").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "UNION should work correctly")
		suite.True(len(results) > 0, "Should return combined results")

		for _, result := range results {
			suite.NotEmpty(result.Name, "Name should not be empty")
			suite.NotEmpty(result.Type, "Type should not be empty")
			suite.T().Logf("Combined: %s (%s)", result.Name, result.Type)
		}
	})

	suite.Run("UnionAll", func() {
		type CombinedResult struct {
			Name string `bun:"name"`
			Type string `bun:"type"`
		}

		var results []CombinedResult

		err := suite.db.NewSelect().
			Table("test_user").
			Select("name").
			SelectExpr(func(ExprBuilder) any {
				return "user"
			}, "type").
			Limit(1).
			UnionAll(func(query SelectQuery) {
				query.Table("test_category").
					Select("name").
					SelectExpr(func(ExprBuilder) any {
						return "category"
					}, "type").
					Limit(1)
			}).
			OrderBy("type", "name").
			Scan(suite.ctx, &results)

		suite.NoError(err, "UNION ALL should work correctly")
		suite.True(len(results) > 0, "Should return combined results with duplicates")

		for _, result := range results {
			suite.T().Logf("UNION ALL: %s (%s)", result.Name, result.Type)
		}
	})

	suite.Run("Intersect", func() {
		count, err := suite.db.NewSelect().
			TableSubQuery(func(query SelectQuery) {
				query.Table("test_user").
					Select("name").
					Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active")
					}).
					Intersect(func(query SelectQuery) {
						query.Table("test_user").
							Select("name").
							Where(func(cb ConditionBuilder) {
								cb.GreaterThan("age", 25)
							})
					})
			}, "t").
			Count(suite.ctx)

		suite.NoError(err, "INTERSECT should work when supported")
		suite.True(count >= 0, "Count should be non-negative")
		suite.T().Logf("INTERSECT count: %d", count)
	})

	suite.Run("Except", func() {
		count, err := suite.db.NewSelect().
			TableSubQuery(func(query SelectQuery) {
				query.Table("test_user").
					Select("name").
					Except(func(query SelectQuery) {
						query.Table("test_user").
							Select("name").
							Where(func(cb ConditionBuilder) {
								cb.IsTrue("is_active")
							})
					})
			}, "t").
			Count(suite.ctx)

		suite.NoError(err, "EXCEPT should work when supported")
		suite.True(count >= 0, "Count should be non-negative")
		suite.T().Logf("EXCEPT count: %d", count)
	})
}

// TestApply tests Apply and ApplyIf methods.
func (suite *SelectTestSuite) TestApply() {
	suite.T().Logf("Testing Apply methods for %s", suite.dbType)

	suite.Run("BasicApply", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			Apply(
				func(query SelectQuery) {
					query.Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active")
					})
				},
				func(query SelectQuery) {
					query.OrderBy("name")
				},
			).
			Limit(3).
			Scan(suite.ctx)

		suite.NoError(err, "Apply should work correctly")
		suite.Len(users, 2, "Should return 2 active users")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active (applied filter)")
			suite.T().Logf("Applied user: %s", user.Name)
		}
	})

	suite.Run("ApplyIfTrue", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			ApplyIf(
				true,
				func(query SelectQuery) {
					query.Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active")
					})
				},
				func(query SelectQuery) {
					query.OrderBy("name")
				},
			).
			Limit(2).
			Scan(suite.ctx)

		suite.NoError(err, "ApplyIf(true) should apply functions")
		suite.Len(users, 2, "Should return 2 users")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active (condition was true)")
		}
	})

	suite.Run("ApplyIfFalse", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			ApplyIf(
				false,
				func(query SelectQuery) {
					query.Where(func(cb ConditionBuilder) {
						cb.IsTrue("is_active")
					})
				},
			).
			Limit(2).
			Scan(suite.ctx)

		suite.NoError(err, "ApplyIf(false) should skip functions")
		suite.Len(users, 2, "Should return 2 users (no filter applied)")

		for _, user := range users {
			suite.T().Logf("Non-filtered user: %s (active=%t)", user.Name, user.IsActive)
		}
	})
}

// TestExecution tests Exec, Scan, Rows, ScanAndCount, Count, and Exists methods.
func (suite *SelectTestSuite) TestExecution() {
	suite.T().Logf("Testing Execution methods for %s", suite.dbType)

	suite.Run("BasicScan", func() {
		var users []User

		err := suite.db.NewSelect().
			Model(&users).
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			OrderBy("name").
			Limit(2).
			Scan(suite.ctx)

		suite.NoError(err, "Basic Scan should work")
		suite.Len(users, 2, "Should return 2 users")

		for _, user := range users {
			suite.T().Logf("Scanned user: %s", user.Name)
		}
	})

	suite.Run("Count", func() {
		count, err := suite.db.NewSelect().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			Count(suite.ctx)

		suite.NoError(err, "Count should work")
		suite.True(count > 0, "Should have active users")

		suite.T().Logf("Active user count: %d", count)
	})

	suite.Run("Exists", func() {
		exists, err := suite.db.NewSelect().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Exists(suite.ctx)

		suite.NoError(err, "Exists should work")
		suite.True(exists, "Alice should exist")

		suite.T().Logf("Alice exists: %t", exists)
	})

	suite.Run("ScanAndCount", func() {
		var users []User

		total, err := suite.db.NewSelect().
			Model(&users).
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			OrderBy("name").
			Limit(2).
			ScanAndCount(suite.ctx)

		suite.NoError(err, "ScanAndCount should work")
		suite.Len(users, 2, "Should return 2 users")
		suite.True(total >= int64(len(users)), "Total should be >= page size")

		suite.T().Logf("Page: %d users, Total: %d", len(users), total)
	})

	suite.Run("Rows", func() {
		rows, err := suite.db.NewSelect().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.IsTrue("is_active")
			}).
			OrderBy("name").
			Limit(2).
			Rows(suite.ctx)

		suite.NoError(err, "Rows should work")
		suite.NotNil(rows, "Rows should not be nil")

		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}

		suite.NoError(rows.Err(), "rows iteration should not have errors")
		suite.Equal(2, count, "Should return 2 rows")
		suite.T().Logf("Successfully iterated through %d rows", count)
	})

	suite.Run("Exec", func() {
		// Exec with SELECT is less common but should work
		var result struct {
			Name string `bun:"name"`
		}

		_, err := suite.db.NewSelect().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "alice@example.com")
			}).
			Exec(suite.ctx, &result)

		suite.NoError(err, "Exec should work when supported")
		suite.T().Logf("Exec result: %s", result.Name)
	})
}
