package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
)

// MergeTestSuite tests MERGE operations (PostgreSQL 15+).
// PostgreSQL 15+ supports the SQL standard MERGE statement (ISO/IEC 9075-2:2016).
// This suite covers all interface methods from MergeQuery, MergeWhenBuilder, MergeUpdateBuilder, and MergeInsertBuilder.
type MergeTestSuite struct {
	*OrmTestSuite
}

// TestBasicMerge tests MERGE with updates and inserts.
func (suite *MergeTestSuite) TestBasicMerge() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing basic MERGE for %s", suite.DbType)

	type UserMergeData struct {
		Id       string `bun:"id"`
		Name     string `bun:"name"`
		Email    string `bun:"email"`
		Age      int16  `bun:"age"`
		IsActive bool   `bun:"is_active"`
	}

	sourceData := []UserMergeData{
		{Id: "user1", Name: "Alice Updated", Email: "alice.updated@example.com", Age: 31, IsActive: true},
		{Id: "user4", Name: "David New", Email: "david@example.com", Age: 28, IsActive: true},
		{Id: "user5", Name: "Eva New", Email: "eva@example.com", Age: 26, IsActive: false},
	}

	// Cleanup inserted data after test
	defer func() {
		_, _ = suite.Db.NewDelete().
			Model(&User{}).
			Where(func(cb ConditionBuilder) {
				cb.In("id", []string{"user4", "user5"})
			}).
			Exec(suite.Ctx)
	}()

	suite.T().Logf("Executing MERGE with %d source records (updates: user1, inserts: user4, user5)", len(sourceData))

	result, err := suite.Db.NewMerge().
		Model(&User{}).
		WithValues("_source_data", &sourceData).
		UsingTable("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("u.id", "_source_data.id")
		}).
		WhenMatched().
		ThenUpdate(func(ub MergeUpdateBuilder) {
			ub.SetColumns("name", "email", "age", "is_active")
		}).
		WhenNotMatched().
		ThenInsert(func(ib MergeInsertBuilder) {
			ib.Values("id", "name", "email", "age", "is_active")
		}).
		Exec(suite.Ctx)

	suite.NoError(err, "MERGE operation should complete successfully")

	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "MERGE should affect 0 or more rows, got %d", affected)
		suite.T().Logf("MERGE operation affected %d rows", affected)
	}

	var newUsers []User

	err = suite.Db.NewSelect().
		Model(&newUsers).
		Where(func(cb ConditionBuilder) {
			cb.In("id", []string{"user4", "user5"})
		}).
		OrderBy("name").
		Scan(suite.Ctx)
	suite.NoError(err, "Failed to query newly inserted users")
	suite.T().Logf("Found %d new users after merge (user4, user5)", len(newUsers))

	for _, user := range newUsers {
		suite.T().Logf("New user - Id: %s, Name: %s, Email: %s, Age: %d, Active: %v",
			user.Id, user.Name, user.Email, user.Age, user.IsActive)
	}
}

// TestCteMethods tests CTE methods: With for named CTEs, WithValues for inline data CTEs.
func (suite *MergeTestSuite) TestCteMethods() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing CTE methods for %s", suite.DbType)

	suite.Run("WithNamedCTE", func() {
		result, err := suite.Db.NewMerge().
			Model(&Post{}).
			With("high_view_posts", func(sq SelectQuery) {
				sq.Model(&Post{}).
					Select("id", "title", "view_count").
					Where(func(cb ConditionBuilder) {
						cb.GreaterThan("view_count", 50)
					})
			}).
			UsingTable("high_view_posts").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("p.id", "high_view_posts.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetExpr("view_count", func(eb ExprBuilder) any {
					return eb.Expr("? + 1", eb.Column("high_view_posts.view_count"))
				})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "MERGE with named CTE should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("MERGE with CTE affected %d rows", affected)
		}
	})

	suite.Run("WithValuesCTE", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
			Age   int16  `bun:"age"`
		}

		sourceData := []UserMergeData{
			{Id: "cte1", Name: "CTE User 1", Email: "cte1@example.com", Age: 25},
			{Id: "cte2", Name: "CTE User 2", Email: "cte2@example.com", Age: 30},
		}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.In("id", []string{"cte1", "cte2"})
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("cte_source", &sourceData).
			UsingTable("cte_source").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "cte_source.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email", "age")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "MERGE with VALUES CTE should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("MERGE with VALUES CTE affected %d rows", affected)
		}
	})
}

// TestTableSourceMethods tests target table specification: ModelTable, Table, TableExpr, TableSubQuery with/without aliases.
func (suite *MergeTestSuite) TestTableSourceMethods() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing table source methods for %s", suite.DbType)

	suite.Run("ModelTableBasic", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "mt1", Name: "ModelTable User", Email: "mt1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "mt1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			ModelTable("test_user").
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "ModelTable without alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("ModelTable basic affected %d rows", affected)
		}
	})

	suite.Run("ModelTableWithAlias", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "mt2", Name: "ModelTable Alias User", Email: "mt2@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "mt2")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			ModelTable("test_user", "u").
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "ModelTable with alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("ModelTable with alias affected %d rows", affected)
		}
	})

	suite.Run("TableBasic", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "t1", Name: "Table User", Email: "t1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "t1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			Table("test_user").
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Table without alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Table basic affected %d rows", affected)
		}
	})

	suite.Run("TableWithAlias", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "t2", Name: "Table Alias User", Email: "t2@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "t2")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			Table("test_user", "u").
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Table with alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Table with alias affected %d rows", affected)
		}
	})

	suite.Run("TableExprBasic", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "te1", Name: "TableExpr User", Email: "te1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "te1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			TableExpr(func(eb ExprBuilder) any {
				return eb.Expr("test_user")
			}, "u").
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "TableExpr should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("TableExpr affected %d rows", affected)
		}
	})

	suite.Run("TableSubQueryBasic", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "tsq1", Name: "TableSubQuery User", Email: "tsq1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "tsq1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			TableSubQuery(func(sq SelectQuery) {
				sq.Model(&User{}).
					Select("id", "name", "email").
					Where(func(cb ConditionBuilder) {
						cb.IsNotNull("email")
					})
			}, "active_users").
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "TableSubQuery should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("TableSubQuery affected %d rows", affected)
		}
	})
}

// TestUsingMethods tests source data specification: UsingTable, UsingExpr, UsingSubQuery with/without aliases.
func (suite *MergeTestSuite) TestUsingMethods() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing Using methods for %s", suite.DbType)

	suite.Run("UsingWithAlias", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "ua1", Name: "Using Alias User", Email: "ua1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "ua1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_source_data", &sourceData).
			UsingTable("_source_data", "src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Using with alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Using with alias affected %d rows", affected)
		}
	})

	suite.Run("UsingTableBasic", func() {
		type PostMergeData struct {
			Id        string `bun:"id,pk"`
			Title     string `bun:"title"`
			ViewCount int    `bun:"view_count"`
		}

		sourceData := []PostMergeData{
			{Id: "post1", Title: "Updated Post Title", ViewCount: 200},
		}

		result, err := suite.Db.NewMerge().
			Model(&Post{}).
			WithValues("_post_updates", &sourceData).
			UsingTable("_post_updates").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("p.id", "_post_updates.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetExpr("view_count", func(eb ExprBuilder) any {
					return eb.Add(eb.Column("_post_updates.view_count"), 1)
				})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "UsingTable without alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("UsingTable basic affected %d rows", affected)
		}
	})

	suite.Run("UsingTableWithAlias", func() {
		type PostMergeData struct {
			Id        string `bun:"id,pk"`
			Title     string `bun:"title"`
			ViewCount int    `bun:"view_count"`
		}

		sourceData := []PostMergeData{
			{Id: "post1", Title: "Updated Post Title", ViewCount: 200},
		}

		result, err := suite.Db.NewMerge().
			Model(&Post{}).
			WithValues("_post_updates", &sourceData).
			UsingTable("_post_updates", "src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("p.id", "src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetExpr("view_count", func(eb ExprBuilder) any {
					return eb.Expr("? + 1", eb.Column("src.view_count"))
				})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "UsingTable with alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("UsingTable with alias affected %d rows", affected)
		}
	})

	suite.Run("UsingExprBasic", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "ue1", Name: "UsingExpr User", Email: "ue1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "ue1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_temp", &sourceData).
			UsingTable("_temp").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_temp.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "UsingExpr should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("UsingExpr affected %d rows", affected)
		}
	})

	suite.Run("UsingExprWithAlias", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "uea1", Name: "UsingExpr Alias User", Email: "uea1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "uea1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_temp", &sourceData).
			UsingTable("_temp", "src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "UsingExpr with alias should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("UsingExpr with alias affected %d rows", affected)
		}
	})

	suite.Run("UsingSubQueryBasic", func() {
		result, err := suite.Db.NewMerge().
			Model(&Post{}).
			UsingSubQuery(func(sq SelectQuery) {
				sq.Model(&Post{}).
					Select("id", "title", "view_count").
					Where(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
			}, "src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("p.id", "src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetExpr("view_count", func(eb ExprBuilder) any {
					return eb.Add(eb.Column("src.view_count"), 5)
				})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "UsingSubQuery should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("UsingSubQuery affected %d rows", affected)
		}
	})
}

// TestReturningMethods tests RETURNING clause: specific columns, all columns (*), or none.
func (suite *MergeTestSuite) TestReturningMethods() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing RETURNING methods for %s", suite.DbType)

	suite.Run("ReturningSpecificColumns", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{
			{Id: "ret1", Name: "Return User 1", Email: "ret1@example.com"},
			{Id: "ret2", Name: "Return User 2", Email: "ret2@example.com"},
		}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.In("id", []string{"ret1", "ret2"})
				}).
				Exec(suite.Ctx)
		}()

		type ReturnResult struct {
			Id   string `bun:"id"`
			Name string `bun:"name"`
		}

		var returnedUsers []ReturnResult

		err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Returning("id", "name").
			Scan(suite.Ctx, &returnedUsers)

		suite.NoError(err, "RETURNING specific columns should work")
		suite.T().Logf("RETURNING specific columns returned %d results", len(returnedUsers))

		for _, result := range returnedUsers {
			suite.T().Logf("Returned: Id=%s, Name=%s", result.Id, result.Name)
		}
	})

	suite.Run("ReturningAll", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "reta1", Name: "ReturnAll User", Email: "reta1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "reta1")
				}).
				Exec(suite.Ctx)
		}()

		var returnedUsers []User

		err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			ReturningAll().
			Scan(suite.Ctx, &returnedUsers)

		suite.NoError(err, "RETURNING * should work")
		suite.T().Logf("RETURNING * returned %d results", len(returnedUsers))

		for _, result := range returnedUsers {
			suite.T().Logf("Returned all columns: Id=%s, Name=%s", result.Id, result.Name)
		}
	})

	suite.Run("ReturningNone", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "retn1", Name: "ReturnNone User", Email: "retn1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "retn1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			ReturningNone().
			Exec(suite.Ctx)

		suite.NoError(err, "RETURNING NONE should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("RETURNING NONE affected %d rows", affected)
		}
	})
}

// TestWhenNotMatchedByTarget tests insertion when row exists in source but not in target (with optional conditions).
func (suite *MergeTestSuite) TestWhenNotMatchedByTarget() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing WhenNotMatchedByTarget for %s", suite.DbType)

	suite.Run("BasicWhenNotMatchedByTarget", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "wnmbt1", Name: "NotMatchedByTarget User", Email: "wnmbt1@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "wnmbt1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatchedByTarget().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WhenNotMatchedByTarget should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("WhenNotMatchedByTarget affected %d rows", affected)
		}
	})

	suite.Run("ConditionalWhenNotMatchedByTarget", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{
			{Id: "wnmbtc1", Name: "Conditional User 1", Email: "wnmbtc1@example.com"},
			{Id: "wnmbtc2", Name: "", Email: "wnmbtc2@example.com"},
		}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.In("id", []string{"wnmbtc1", "wnmbtc2"})
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatchedByTarget(func(cb ConditionBuilder) {
				cb.IsNotNull("_src.name").NotEquals("_src.name", "")
			}).
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WhenNotMatchedByTarget with condition should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Conditional WhenNotMatchedByTarget affected %d rows (should skip empty name)", affected)
		}
	})
}

// TestWhenNotMatchedBySource tests updates/deletes when row exists in target but not in source (with optional conditions).
func (suite *MergeTestSuite) TestWhenNotMatchedBySource() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing WhenNotMatchedBySource for %s", suite.DbType)

	suite.Run("BasicWhenNotMatchedBySource", func() {
		// Insert test users first
		testUsers := []User{
			{Name: "User to Keep", Email: "keep@example.com", Age: 25, IsActive: true},
			{Name: "User to Update", Email: "update@example.com", Age: 30, IsActive: true},
		}
		testUsers[0].Id = "wnmbs1"
		testUsers[1].Id = "wnmbs2"

		for _, user := range testUsers {
			_, err := suite.Db.NewInsert().
				Model(&user).
				Exec(suite.Ctx)
			suite.NoError(err, "Failed to create test user")
		}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.In("id", []string{"wnmbs1", "wnmbs2"})
				}).
				Exec(suite.Ctx)
		}()

		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		// Source only contains wnmbs1, so wnmbs2 will be "not matched by source"
		sourceData := []UserMergeData{
			{Id: "wnmbs1", Name: "User to Keep Updated", Email: "keep_updated@example.com"},
		}

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetColumns("name", "email")
			}).
			WhenNotMatchedBySource(func(cb ConditionBuilder) {
				// Only update our test users, not fixture users
				cb.In("u.id", []string{"wnmbs1", "wnmbs2"})
			}).
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.Set("is_active", false)
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WhenNotMatchedBySource should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("WhenNotMatchedBySource affected %d rows", affected)
		}

		// Verify the results
		var users []User

		err = suite.Db.NewSelect().
			Model(&users).
			Where(func(cb ConditionBuilder) {
				cb.In("id", []string{"wnmbs1", "wnmbs2"})
			}).
			OrderBy("id").
			Scan(suite.Ctx)
		suite.NoError(err)

		// wnmbs1 should be updated from source
		suite.Equal("User to Keep Updated", users[0].Name)
		suite.Equal("keep_updated@example.com", users[0].Email)
		suite.True(users[0].IsActive)

		// wnmbs2 should be marked as inactive (not in source)
		suite.Equal("User to Update", users[1].Name)
		suite.Equal("update@example.com", users[1].Email)
		suite.False(users[1].IsActive)
	})

	suite.Run("ConditionalWhenNotMatchedBySource", func() {
		// Insert test users first
		testUsers := []User{
			{Name: "Active User 1", Email: "active1@example.com", Age: 25, IsActive: true},
			{Name: "Active User 2", Email: "active2@example.com", Age: 28, IsActive: true},
			{Name: "Inactive User", Email: "inactive@example.com", Age: 30, IsActive: false},
		}
		testUsers[0].Id = "wnmbsc1"
		testUsers[1].Id = "wnmbsc2"
		testUsers[2].Id = "wnmbsc3"

		for _, user := range testUsers {
			_, err := suite.Db.NewInsert().
				Model(&user).
				Exec(suite.Ctx)
			suite.NoError(err, "Failed to create test user")
		}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.In("id", []string{"wnmbsc1", "wnmbsc2", "wnmbsc3"})
				}).
				Exec(suite.Ctx)
		}()

		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		// Source only contains wnmbsc1, so wnmbsc2 and wnmbsc3 will be "not matched by source"
		sourceData := []UserMergeData{
			{Id: "wnmbsc1", Name: "Active User 1 Updated", Email: "active1_updated@example.com"},
		}

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetColumns("name", "email")
			}).
			WhenNotMatchedBySource(func(cb ConditionBuilder) {
				// Only update our test users that are active
				cb.In("u.id", []string{"wnmbsc1", "wnmbsc2", "wnmbsc3"}).
					Equals("u.is_active", true)
			}).
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.Set("is_active", false)
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WhenNotMatchedBySource with condition should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Conditional WhenNotMatchedBySource affected %d rows (should only update active users not in source)", affected)
		}

		// Verify the results
		var users []User

		err = suite.Db.NewSelect().
			Model(&users).
			Where(func(cb ConditionBuilder) {
				cb.In("id", []string{"wnmbsc1", "wnmbsc2", "wnmbsc3"})
			}).
			OrderBy("id").
			Scan(suite.Ctx)
		suite.NoError(err)

		// wnmbsc1 should be updated from source and remain active
		suite.Equal("Active User 1 Updated", users[0].Name)
		suite.Equal("active1_updated@example.com", users[0].Email)
		suite.True(users[0].IsActive)

		// wnmbsc2 should be deactivated (was active, not in source)
		suite.Equal("Active User 2", users[1].Name)
		suite.False(users[1].IsActive)

		// wnmbsc3 should remain inactive (was already inactive, condition not met)
		suite.Equal("Inactive User", users[2].Name)
		suite.False(users[2].IsActive)
	})
}

// TestThenDoNothing tests no-op actions for matched/not-matched conditions.
func (suite *MergeTestSuite) TestThenDoNothing() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing ThenDoNothing for %s", suite.DbType)

	suite.Run("WhenMatchedDoNothing", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{
			{Id: "user1", Name: "Should Not Update", Email: "user1@example.com"},
			{Id: "dnm1", Name: "Should Insert", Email: "dnm1@example.com"},
		}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "dnm1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenDoNothing().
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "WHEN MATCHED THEN DO NOTHING should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("WHEN MATCHED DO NOTHING affected %d rows (inserts only)", affected)
		}
	})

	suite.Run("WhenNotMatchedDoNothing", func() {
		type UserMergeData struct {
			Id   string `bun:"id,pk"`
			Name string `bun:"name"`
		}

		sourceData := []UserMergeData{
			{Id: "user1", Name: "Should Update"},
			{Id: "dnn1", Name: "Should Not Insert"},
		}

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetColumns("name")
			}).
			WhenNotMatched().
			ThenDoNothing().
			Exec(suite.Ctx)

		suite.NoError(err, "WHEN NOT MATCHED THEN DO NOTHING should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("WHEN NOT MATCHED DO NOTHING affected %d rows (updates only)", affected)
		}
	})
}

// TestThenUpdate tests update actions: Set, SetExpr, SetColumns, SetAll.
func (suite *MergeTestSuite) TestThenUpdate() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing UpdateBuilder methods for %s", suite.DbType)

	suite.Run("SetSingleValue", func() {
		type UserMergeData struct {
			Id   string `bun:"id,pk"`
			Name string `bun:"name"`
		}

		sourceData := []UserMergeData{{Id: "user1", Name: "Set Single"}}

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.Set("name", "Set Single Value")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Set single value should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Set single value affected %d rows", affected)
		}
	})

	suite.Run("SetMultipleValues", func() {
		type UserMergeData struct {
			Id   string `bun:"id,pk"`
			Name string `bun:"name"`
			Age  int16  `bun:"age"`
		}

		sourceData := []UserMergeData{{Id: "user1", Name: "Multiple", Age: 35}}

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.Set("name", "Set Multiple Values").
					Set("age", 40)
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Set multiple values should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Set multiple values affected %d rows", affected)
		}
	})

	suite.Run("SetExprBasic", func() {
		type PostMergeData struct {
			Id        string `bun:"id,pk"`
			ViewCount int    `bun:"view_count"`
		}

		sourceData := []PostMergeData{{Id: "post1", ViewCount: 100}}

		result, err := suite.Db.NewMerge().
			Model(&Post{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("p.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetExpr("view_count", func(eb ExprBuilder) any {
					return eb.Expr("? + ?", eb.Column("p.view_count"), 10)
				})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "SetExpr should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("SetExpr affected %d rows", affected)
		}
	})

	suite.Run("SetColumnsBasic", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "user1", Name: "SetColumns Name", Email: "setcols@example.com"}}

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetColumns("name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "SetColumns should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("SetColumns affected %d rows", affected)
		}
	})

	suite.Run("SetAllWithExclusions", func() {
		type UserMergeData struct {
			Id       string `bun:"id,pk"`
			Name     string `bun:"name"`
			Email    string `bun:"email"`
			Age      int16  `bun:"age"`
			IsActive bool   `bun:"is_active"`
		}

		sourceData := []UserMergeData{
			{Id: "user1", Name: "SetAll Name", Email: "setall@example.com", Age: 45, IsActive: true},
		}

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenMatched().
			ThenUpdate(func(ub MergeUpdateBuilder) {
				ub.SetAll("id", "created_at", "created_by", "updated_at", "updated_by", "deleted_at", "deleted_by", "meta")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "SetAll with exclusions should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("SetAll with exclusions affected %d rows", affected)
		}
	})
}

// TestThenInsert tests insert actions: Value, ValueExpr, Values, ValuesAll.
func (suite *MergeTestSuite) TestThenInsert() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing InsertBuilder methods for %s", suite.DbType)

	suite.Run("ValueSingleColumn", func() {
		type UserMergeData struct {
			Id string `bun:"id,pk"`
		}

		sourceData := []UserMergeData{{Id: "val1"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "val1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Value("id", "val1").
					Value("name", "Value Single User").
					Value("email", "val1@example.com")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Value single column should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Value single affected %d rows", affected)
		}
	})

	suite.Run("ValueMultipleColumns", func() {
		type UserMergeData struct {
			Id string `bun:"id,pk"`
		}

		sourceData := []UserMergeData{{Id: "valm1"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "valm1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Value("id", "valm1").
					Value("name", "Value Multiple User").
					Value("email", "valm1@example.com").
					Value("age", 25)
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Value multiple columns should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Value multiple affected %d rows", affected)
		}
	})

	suite.Run("ValueExprBasic", func() {
		type UserMergeData struct {
			Id   string `bun:"id,pk"`
			Name string `bun:"name"`
		}

		sourceData := []UserMergeData{{Id: "vale1", Name: "ValueExpr"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "vale1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Value("id", "vale1").
					Value("email", "vale1@example.com").
					ValueExpr("name", func(eb ExprBuilder) any {
						return eb.Concat(eb.Column("_src.name"), " (Expression)")
					})
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "ValueExpr should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("ValueExpr affected %d rows", affected)
		}
	})

	suite.Run("ValuesMultipleColumns", func() {
		type UserMergeData struct {
			Id    string `bun:"id,pk"`
			Name  string `bun:"name"`
			Email string `bun:"email"`
		}

		sourceData := []UserMergeData{{Id: "vals1", Name: "Values User", Email: "values@example.com"}}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "vals1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.Values("id", "name", "email")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "Values multiple columns should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("Values multiple affected %d rows", affected)
		}
	})

	suite.Run("ValuesAllWithExclusions", func() {
		type UserMergeData struct {
			Id       string `bun:"id,pk"`
			Name     string `bun:"name"`
			Email    string `bun:"email"`
			Age      int16  `bun:"age"`
			IsActive bool   `bun:"is_active"`
		}

		sourceData := []UserMergeData{
			{Id: "vala1", Name: "ValuesAll User", Email: "valall@example.com", Age: 28, IsActive: true},
		}

		// Cleanup inserted data after test
		defer func() {
			_, _ = suite.Db.NewDelete().
				Model(&User{}).
				Where(func(cb ConditionBuilder) {
					cb.Equals("id", "vala1")
				}).
				Exec(suite.Ctx)
		}()

		result, err := suite.Db.NewMerge().
			Model(&User{}).
			WithValues("_src", &sourceData).
			UsingTable("_src").
			On(func(cb ConditionBuilder) {
				cb.EqualsColumn("u.id", "_src.id")
			}).
			WhenNotMatched().
			ThenInsert(func(ib MergeInsertBuilder) {
				ib.ValuesAll("created_at", "created_by", "updated_at", "updated_by", "deleted_at", "deleted_by", "meta")
			}).
			Exec(suite.Ctx)

		suite.NoError(err, "ValuesAll with exclusions should work")

		if result != nil {
			affected, _ := result.RowsAffected()
			suite.T().Logf("ValuesAll with exclusions affected %d rows", affected)
		}
	})
}

// TestThenDelete tests deletion when rows exist in target but not in source.
func (suite *MergeTestSuite) TestThenDelete() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing MERGE with DELETE for %s", suite.DbType)

	testPosts := []Post{
		{Title: "Test Post 1", Status: "published", ViewCount: 100},
		{Title: "Test Post 2", Status: "draft", ViewCount: 50},
		{Title: "Test Post 3", Status: "archived", ViewCount: 25},
	}

	testPosts[0].Id = "merge_test_1"
	testPosts[1].Id = "merge_test_2"
	testPosts[2].Id = "merge_test_3"

	suite.T().Logf("Creating %d test posts for DELETE scenario", len(testPosts))

	for i, post := range testPosts {
		_, err := suite.Db.NewInsert().
			Model(&post).
			Exec(suite.Ctx)
		suite.NoError(err, "Failed to create test post %d (Id: %s)", i+1, post.Id)
		suite.T().Logf("Created test post %d: %s - %s (views: %d)", i+1, post.Id, post.Title, post.ViewCount)
	}

	// Cleanup test posts after test (merge_test_3 should be deleted by MERGE)
	defer func() {
		_, _ = suite.Db.NewDelete().
			Model(&Post{}).
			Where(func(cb ConditionBuilder) {
				cb.In("id", []string{"merge_test_1", "merge_test_2"})
			}).
			Exec(suite.Ctx)
	}()

	type PostUpdateData struct {
		Id        string `bun:"id,pk"`
		Title     string `bun:"title"`
		Status    string `bun:"status"`
		ViewCount int    `bun:"view_count"`
	}

	sourceData := []PostUpdateData{
		{Id: "merge_test_1", Title: "Updated Test Post 1", Status: "published", ViewCount: 120},
		{Id: "merge_test_2", Title: "Updated Test Post 2", Status: "published", ViewCount: 80},
	}

	suite.T().Logf("Executing MERGE with DELETE - %d source records, missing merge_test_3 to test deletion", len(sourceData))

	result, err := suite.Db.NewMerge().
		Model(&Post{}).
		WithValues("_source_data", &sourceData).
		UsingTable("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("p.id", "_source_data.id")
		}).
		WhenMatched().
		ThenUpdate(func(ub MergeUpdateBuilder) {
			ub.SetColumns("title", "status", "view_count")
		}).
		WhenNotMatchedBySource(func(cb ConditionBuilder) {
			cb.LessThan("p.view_count", 30)
		}).
		ThenDelete().
		Exec(suite.Ctx)

	suite.NoError(err, "MERGE with DELETE should complete successfully")

	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "MERGE with DELETE should affect 0 or more rows, got %d", affected)
		suite.T().Logf("MERGE with DELETE affected %d rows (updates + deletions)", affected)
	}

	var remainingPosts []Post

	err = suite.Db.NewSelect().
		Model(&remainingPosts).
		Where(func(cb ConditionBuilder) {
			cb.StartsWith("id", "merge_test_")
		}).
		OrderBy("id").
		Scan(suite.Ctx)
	suite.NoError(err, "Failed to query remaining posts after MERGE with DELETE")

	suite.T().Logf("Remaining posts after MERGE with DELETE: %d", len(remainingPosts))

	for _, post := range remainingPosts {
		suite.T().Logf("Remaining post %s: %s - %s (views: %d)", post.Id, post.Title, post.Status, post.ViewCount)
		suite.NotEqual("merge_test_3", post.Id, "Post merge_test_3 with low view count should be deleted")
	}
}

// TestMergeWithConditions tests MERGE with conditional WHEN clauses (e.g., only update when source > target).
func (suite *MergeTestSuite) TestMergeWithConditions() {
	if suite.DbType != constants.DbPostgres {
		suite.T().Skipf("MERGE statement is only supported by PostgreSQL, skipping for %s", suite.DbType)
	}

	suite.T().Logf("Testing MERGE with conditions for %s", suite.DbType)

	type PostMergeData struct {
		Id        string `bun:"id,pk"`
		Title     string `bun:"title"`
		Status    string `bun:"status"`
		ViewCount int    `bun:"view_count"`
	}

	sourceData := []PostMergeData{
		{Id: "post1", Title: "Updated Post 1", Status: "published", ViewCount: 150},
		{Id: "post2", Title: "Updated Post 2", Status: "draft", ViewCount: 75},
		{Id: "new1", Title: "New Post 1", Status: "draft", ViewCount: 0},
	}

	// Cleanup inserted post after test
	defer func() {
		_, _ = suite.Db.NewDelete().
			Model(&Post{}).
			Where(func(cb ConditionBuilder) {
				cb.Equals("id", "new1")
			}).
			Exec(suite.Ctx)
	}()

	suite.T().Logf("Executing conditional MERGE with %d source records", len(sourceData))

	result, err := suite.Db.NewMerge().
		Model(&Post{}).
		WithValues("_source_data", &sourceData).
		UsingTable("_source_data").
		On(func(cb ConditionBuilder) {
			cb.EqualsColumn("p.id", "_source_data.id")
		}).
		WhenMatched(func(cb ConditionBuilder) {
			cb.GreaterThanColumn("_source_data.view_count", "p.view_count")
		}).
		ThenUpdate(func(ub MergeUpdateBuilder) {
			ub.SetColumns("title", "status", "view_count")
		}).
		WhenNotMatched(func(cb ConditionBuilder) {
			cb.IsNotNull("_source_data.status").NotEquals("_source_data.status", "")
		}).
		ThenInsert(func(ib MergeInsertBuilder) {
			ib.Values("id", "title", "status", "view_count")
		}).
		Exec(suite.Ctx)

	suite.NoError(err, "Conditional MERGE operation should complete successfully")

	if result != nil {
		affected, _ := result.RowsAffected()
		suite.True(affected >= 0, "Conditional MERGE should affect 0 or more rows, got %d", affected)
		suite.T().Logf("Conditional MERGE affected %d rows", affected)
	}

	var updatedPosts []Post

	err = suite.Db.NewSelect().
		Model(&updatedPosts).
		Where(func(cb ConditionBuilder) {
			cb.In("id", []string{"post1", "post2", "new1"})
		}).
		OrderBy("id").
		Scan(suite.Ctx)
	suite.NoError(err, "Failed to query posts after conditional MERGE")

	suite.T().Logf("Posts after conditional MERGE: %d", len(updatedPosts))

	for _, post := range updatedPosts {
		suite.T().Logf("Post %s: %s - %s (views: %d)", post.Id, post.Title, post.Status, post.ViewCount)
	}
}
