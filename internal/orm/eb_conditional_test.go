package orm

// ConditionalFunctionsTestSuite tests conditional function methods of ExprBuilder
// including Coalesce (returns first non-null value), NullIf (returns NULL when equal),
// and IfNull (returns default when NULL).
type ConditionalFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestCoalesce tests the Coalesce function.
func (suite *ConditionalFunctionsTestSuite) TestCoalesce() {
	suite.T().Logf("Testing Coalesce function for %s", suite.DbType)

	suite.Run("CoalesceWithDefaults", func() {
		type CoalesceResult struct {
			Name           string  `bun:"name"`
			Description    *string `bun:"description"`
			SafeDesc       string  `bun:"safe_desc"`
			MultiCoalesce  string  `bun:"multi_coalesce"`
			CoalesceNumber int64   `bun:"coalesce_number"`
		}

		var coalesceResults []CoalesceResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectAs("description", "description").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.Column("description"), "'No description available'")
			}, "safe_desc").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.Column("description"), eb.Column("title"), "'Untitled'")
			}, "multi_coalesce").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.Expr("NULL"), eb.Column("view_count"), 0)
			}, "coalesce_number").
			OrderBy("title").
			Limit(5).
			Scan(suite.Ctx, &coalesceResults)

		suite.NoError(err, "Coalesce should work correctly")
		suite.True(len(coalesceResults) > 0, "Should have coalesce results")

		for _, result := range coalesceResults {
			suite.NotEmpty(result.SafeDesc, "Coalesce should provide default value")
			suite.NotEmpty(result.MultiCoalesce, "Multi-argument coalesce should work")
			suite.True(result.CoalesceNumber >= 0, "Coalesce with numbers should work")

			if result.Description == nil {
				suite.Equal("No description available", result.SafeDesc, "Should use default for NULL description")
			} else {
				suite.Equal(*result.Description, result.SafeDesc, "Should use actual description when not NULL")
			}

			suite.T().Logf("Post %s: SafeDesc=%s, MultiCoalesce=%s, CoalesceNumber=%d",
				result.Name, result.SafeDesc, result.MultiCoalesce, result.CoalesceNumber)
		}
	})
}

// TestNullIf tests the NullIf function.
func (suite *ConditionalFunctionsTestSuite) TestNullIf() {
	suite.T().Logf("Testing NullIf function for %s", suite.DbType)

	suite.Run("NullIfEqualityCheck", func() {
		type NullIfResult struct {
			Title      string  `bun:"title"`
			Status     string  `bun:"status"`
			CheckDraft *string `bun:"check_draft"`
			CheckViews *int64  `bun:"check_views"`
		}

		var nullIfResults []NullIfResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NullIf(eb.Column("status"), "'draft'")
			}, "check_draft").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NullIf(eb.Column("view_count"), 0)
			}, "check_views").
			OrderBy("title").
			Limit(5).
			Scan(suite.Ctx, &nullIfResults)

		suite.NoError(err, "NullIf should work correctly")
		suite.True(len(nullIfResults) > 0, "Should have NullIf results")

		for _, result := range nullIfResults {
			if result.Status == "draft" {
				suite.Nil(result.CheckDraft, "NullIf should return NULL when values are equal (draft)")
			} else {
				suite.NotNil(result.CheckDraft, "NullIf should return first argument when values differ")
				suite.Equal(result.Status, *result.CheckDraft, "NullIf should return status when not draft")
			}

			suite.T().Logf("Post %s: Status=%s, CheckDraft=%v, CheckViews=%v",
				result.Title, result.Status, result.CheckDraft, result.CheckViews)
		}
	})
}

// TestIfNull tests the IfNull function.
func (suite *ConditionalFunctionsTestSuite) TestIfNull() {
	suite.T().Logf("Testing IfNull function for %s", suite.DbType)

	suite.Run("IfNullWithDefaults", func() {
		type IfNullResult struct {
			Title       string `bun:"title"`
			Description string `bun:"description"`
			ViewCount   int64  `bun:"view_count"`
		}

		var ifNullResults []IfNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IfNull(eb.Column("description"), "'[No description]'")
			}, "description").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IfNull(eb.Expr("NULL"), eb.Column("view_count"))
			}, "view_count").
			OrderBy("title").
			Limit(5).
			Scan(suite.Ctx, &ifNullResults)

		suite.NoError(err, "IfNull should work correctly")
		suite.True(len(ifNullResults) > 0, "Should have IfNull results")

		for _, result := range ifNullResults {
			suite.NotEmpty(result.Description, "IfNull should provide default when NULL")
			suite.True(result.ViewCount >= 0, "IfNull should return view count")

			suite.T().Logf("Post %s: Description=%s, ViewCount=%d",
				result.Title, result.Description, result.ViewCount)
		}
	})
}

// TestCombinedConditionalFunctions tests multiple conditional functions working together.
func (suite *ConditionalFunctionsTestSuite) TestCombinedConditionalFunctions() {
	suite.T().Logf("Testing combined conditional functions for %s", suite.DbType)

	suite.Run("NestedConditionalFunctions", func() {
		type CombinedResult struct {
			Title         string `bun:"title"`
			Status        string `bun:"status"`
			SafeStatus    string `bun:"safe_status"`
			DisplayStatus string `bun:"display_status"`
			DescOrTitle   string `bun:"desc_or_title"`
			ViewCategory  string `bun:"view_category"`
		}

		var combinedResults []CombinedResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IfNull(eb.Column("status"), "'unknown'")
			}, "safe_status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.NullIf(eb.Column("status"), "'draft'"), "'Working Draft'")
			}, "display_status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.Column("description"), eb.Column("title"))
			}, "desc_or_title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(
					eb.NullIf(
						eb.Case(func(cb CaseBuilder) {
							cb.When(func(cond ConditionBuilder) {
								cond.GreaterThan("view_count", 100)
							}).Then("'High'").
								When(func(cond ConditionBuilder) {
									cond.GreaterThan("view_count", 50)
								}).Then("'Medium'").
								Else("'Low'")
						}),
						"'Low'",
					),
					"'No Views'",
				)
			}, "view_category").
			OrderBy("title").
			Limit(5).
			Scan(suite.Ctx, &combinedResults)

		suite.NoError(err, "Combined conditional functions should work")
		suite.True(len(combinedResults) > 0, "Should have combined results")

		for _, result := range combinedResults {
			suite.NotEmpty(result.SafeStatus, "SafeStatus should not be empty")
			suite.NotEmpty(result.DisplayStatus, "DisplayStatus should not be empty")
			suite.NotEmpty(result.DescOrTitle, "DescOrTitle should not be empty")
			suite.NotEmpty(result.ViewCategory, "ViewCategory should not be empty")

			suite.T().Logf("Post %s: Status=%s, DisplayStatus=%s, ViewCategory=%s",
				result.Title, result.Status, result.DisplayStatus, result.ViewCategory)
		}
	})
}

// TestCoalesceBoundaryConditions tests Coalesce function with boundary conditions.
func (suite *ConditionalFunctionsTestSuite) TestCoalesceBoundaryConditions() {
	suite.T().Logf("Testing Coalesce boundary conditions for %s", suite.DbType)

	suite.Run("CoalesceSingleArgument", func() {
		type SingleArgResult struct {
			Id          string `bun:"id"`
			Title       string `bun:"title"`
			SingleValue string `bun:"single_value"`
		}

		var results []SingleArgResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.Column("title"), eb.Column("title"))
			}, "single_value").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Coalesce with minimal arguments should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(result.Title, result.SingleValue, "Should return the title value")
			suite.T().Logf("Id: %s, Title: %s, SingleValue: %s", result.Id, result.Title, result.SingleValue)
		}
	})

	suite.Run("CoalesceAllNull", func() {
		type AllNullResult struct {
			Id         string  `bun:"id"`
			Title      string  `bun:"title"`
			AllNullVal *string `bun:"all_null_val"`
		}

		var results []AllNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.Expr("NULL"), eb.Expr("NULL"), eb.Expr("NULL"))
			}, "all_null_val").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Coalesce with all NULL arguments should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.AllNullVal, "All NULL arguments should return NULL")
			suite.T().Logf("Id: %s, Title: %s, AllNullVal: %v", result.Id, result.Title, result.AllNullVal)
		}
	})

	suite.Run("CoalesceManyArguments", func() {
		type ManyArgsResult struct {
			Id         string `bun:"id"`
			Title      string `bun:"title"`
			FinalValue string `bun:"final_value"`
		}

		var results []ManyArgsResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(
					eb.Expr("NULL"),
					eb.Expr("NULL"),
					eb.Expr("NULL"),
					eb.Column("title"),
					"'Fallback'",
				)
			}, "final_value").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Coalesce with many arguments should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(result.Title, result.FinalValue, "Should return first non-NULL value (title)")
			suite.T().Logf("Id: %s, Title: %s, FinalValue: %s", result.Id, result.Title, result.FinalValue)
		}
	})
}

// TestNullIfWithNullArguments tests NullIf function with NULL arguments.
func (suite *ConditionalFunctionsTestSuite) TestNullIfWithNullArguments() {
	suite.T().Logf("Testing NullIf with NULL arguments for %s", suite.DbType)

	suite.Run("NullIfFirstArgumentNull", func() {
		type FirstNullResult struct {
			Id     string  `bun:"id"`
			Title  string  `bun:"title"`
			Result *string `bun:"result"`
		}

		var results []FirstNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NullIf(eb.Expr("NULL"), "'value'")
			}, "result").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NullIf with NULL as first argument should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.Result, "NullIf(NULL, value) should return NULL")
			suite.T().Logf("Id: %s, Title: %s, Result: %v", result.Id, result.Title, result.Result)
		}
	})

	suite.Run("NullIfSecondArgumentNull", func() {
		type SecondNullResult struct {
			Id     string  `bun:"id"`
			Title  string  `bun:"title"`
			Result *string `bun:"result"`
		}

		var results []SecondNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NullIf(eb.Column("status"), eb.Expr("NULL"))
			}, "result").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NullIf with NULL as second argument should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotNil(result.Result, "NullIf(value, NULL) should return first argument (never equal)")
			suite.T().Logf("Id: %s, Title: %s, Result: %v", result.Id, result.Title, result.Result)
		}
	})

	suite.Run("NullIfBothArgumentsNull", func() {
		type BothNullResult struct {
			Id     string  `bun:"id"`
			Title  string  `bun:"title"`
			Result *string `bun:"result"`
		}

		var results []BothNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NullIf(eb.Expr("NULL"), eb.Expr("NULL"))
			}, "result").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "NullIf with both NULL arguments should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.Result, "NullIf(NULL, NULL) should return NULL (considered equal)")
			suite.T().Logf("Id: %s, Title: %s, Result: %v", result.Id, result.Title, result.Result)
		}
	})
}

// TestIfNullWithNullArguments tests IfNull function with NULL arguments.
func (suite *ConditionalFunctionsTestSuite) TestIfNullWithNullArguments() {
	suite.T().Logf("Testing IfNull with NULL arguments for %s", suite.DbType)

	suite.Run("IfNullDefaultValueNull", func() {
		type DefaultNullResult struct {
			Id     string  `bun:"id"`
			Title  string  `bun:"title"`
			Result *string `bun:"result"`
		}

		var results []DefaultNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IfNull(eb.Expr("NULL"), eb.Expr("NULL"))
			}, "result").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "IfNull with NULL default value should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Nil(result.Result, "IfNull(NULL, NULL) should return NULL")
			suite.T().Logf("Id: %s, Title: %s, Result: %v", result.Id, result.Title, result.Result)
		}
	})

	suite.Run("IfNullWithValueAndNullDefault", func() {
		type ValueNullDefaultResult struct {
			Id     string  `bun:"id"`
			Title  string  `bun:"title"`
			Status string  `bun:"status"`
			Result *string `bun:"result"`
		}

		var results []ValueNullDefaultResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IfNull(eb.Column("status"), eb.Expr("NULL"))
			}, "result").
			Where(func(cond ConditionBuilder) {
				cond.IsNotNull("status")
			}).
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "IfNull with value and NULL default should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotNil(result.Result, "IfNull(value, NULL) should return value when not NULL")
			suite.Equal(result.Status, *result.Result, "Should return original status value")
			suite.T().Logf("Id: %s, Title: %s, Status: %s, Result: %v", result.Id, result.Title, result.Status, result.Result)
		}
	})
}

// TestConditionalFunctionsSpecialValues tests conditional functions with special values.
func (suite *ConditionalFunctionsTestSuite) TestConditionalFunctionsSpecialValues() {
	suite.T().Logf("Testing conditional functions with special values for %s", suite.DbType)

	suite.Run("EmptyStringVsNull", func() {
		type EmptyStringResult struct {
			Id                string `bun:"id"`
			Title             string `bun:"title"`
			EmptyNotNull      string `bun:"empty_not_null"`
			CoalesceEmpty     string `bun:"coalesce_empty"`
			NullIfEmpty       string `bun:"nullif_empty"`
			IfNullEmptyResult string `bun:"ifnull_empty_result"`
		}

		var results []EmptyStringResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return ""
			}, "empty_not_null").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce("", "default")
			}, "coalesce_empty").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.NullIf("", ""), "was_empty")
			}, "nullif_empty").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IfNull("", "default")
			}, "ifnull_empty_result").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Special value test should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal("", result.EmptyNotNull, "Empty string should not be NULL")
			suite.Equal("", result.CoalesceEmpty, "Coalesce should return empty string (not NULL)")
			suite.Equal("was_empty", result.NullIfEmpty, "NullIf('', '') should return NULL, then Coalesce to default")
			suite.Equal("", result.IfNullEmptyResult, "IfNull should return empty string (not NULL)")

			suite.T().Logf("Id: %s, EmptyNotNull='%s', CoalesceEmpty='%s', NullIfEmpty='%s', IfNullEmptyResult='%s'",
				result.Id, result.EmptyNotNull, result.CoalesceEmpty, result.NullIfEmpty, result.IfNullEmptyResult)
		}
	})

	suite.Run("ZeroVsNull", func() {
		type ZeroVsNullResult struct {
			Id               string `bun:"id"`
			Title            string `bun:"title"`
			ZeroNotNull      int64  `bun:"zero_not_null"`
			CoalesceZero     int64  `bun:"coalesce_zero"`
			NullIfZero       int64  `bun:"nullif_zero"`
			IfNullZeroResult int64  `bun:"ifnull_zero_result"`
		}

		var results []ZeroVsNullResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("0")
			}, "zero_not_null").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(0, 999)
			}, "coalesce_zero").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Coalesce(eb.NullIf(0, 0), 888)
			}, "nullif_zero").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IfNull(0, 777)
			}, "ifnull_zero_result").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &results)

		suite.NoError(err, "Zero vs NULL test should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.Equal(int64(0), result.ZeroNotNull, "Zero should not be NULL")
			suite.Equal(int64(0), result.CoalesceZero, "Coalesce should return 0 (not NULL)")
			suite.Equal(int64(888), result.NullIfZero, "NullIf(0, 0) should return NULL, then Coalesce to 888")
			suite.Equal(int64(0), result.IfNullZeroResult, "IfNull should return 0 (not NULL)")

			suite.T().Logf("Id: %s, ZeroNotNull=%d, CoalesceZero=%d, NullIfZero=%d, IfNullZeroResult=%d",
				result.Id, result.ZeroNotNull, result.CoalesceZero, result.NullIfZero, result.IfNullZeroResult)
		}
	})
}
