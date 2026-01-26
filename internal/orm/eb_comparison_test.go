package orm

// ComparisonExpressionsTestSuite tests comparison expression methods of ExprBuilder
// including Equals, NotEquals, GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual,
// Between, NotBetween, In, and NotIn.
type ComparisonExpressionsTestSuite struct {
	*OrmTestSuite
}

// TestEquals tests the Equals comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestEquals() {
	suite.T().Logf("Testing Equals comparison for %s", suite.dbType)

	// Test 1: Simple equals with string
	suite.Run("SimpleStringEquals", func() {
		type EqualsResult struct {
			ID     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var results []EqualsResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Equals(eb.Column("status"), "published")
				})
			}).
			OrderBy("id").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Equals with string should work")
		suite.True(len(results) > 0, "Should have Equals results")

		for _, result := range results {
			suite.Equal("published", result.Status, "Status should be 'published'")
			suite.T().Logf("ID: %s, Title: %s, Status: %s", result.ID, result.Title, result.Status)
		}
	})

	// Test 2: Equals with integer
	suite.Run("IntegerEquals", func() {
		type IntegerEqualsResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []IntegerEqualsResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Equals(eb.Column("view_count"), 42)
				})
			}).
			OrderBy("id").
			Scan(suite.ctx, &results)

		suite.NoError(err, "Equals with integer should work")

		for _, result := range results {
			suite.Equal(int64(42), result.ViewCount, "ViewCount should be 42")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.ID, result.Title, result.ViewCount)
		}
	})

	// Test 3: Equals in SELECT clause
	suite.Run("EqualsInSelect", func() {
		type SelectEqualsResult struct {
			ID          string `bun:"id"`
			Status      string `bun:"status"`
			IsPublished bool   `bun:"is_published"`
		}

		var results []SelectEqualsResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Equals(eb.Column("status"), "published")
			}, "is_published").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Equals in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.Status == "published" {
				suite.True(result.IsPublished, "IsPublished should be true for published posts")
			} else {
				suite.False(result.IsPublished, "IsPublished should be false for non-published posts")
			}

			suite.T().Logf("ID: %s, Status: %s, IsPublished: %v", result.ID, result.Status, result.IsPublished)
		}
	})
}

// TestNotEquals tests the NotEquals comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestNotEquals() {
	suite.T().Logf("Testing NotEquals comparison for %s", suite.dbType)

	// Test 1: Simple NotEquals with string
	suite.Run("SimpleStringNotEquals", func() {
		type NotEqualsResult struct {
			ID     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var results []NotEqualsResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotEquals(eb.Column("status"), "draft")
				})
			}).
			OrderBy("id").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotEquals with string should work")
		suite.True(len(results) > 0, "Should have NotEquals results")

		for _, result := range results {
			suite.NotEqual("draft", result.Status, "Status should not be 'draft'")
			suite.T().Logf("ID: %s, Title: %s, Status: %s", result.ID, result.Title, result.Status)
		}
	})

	// Test 2: NotEquals with integer
	suite.Run("IntegerNotEquals", func() {
		type IntegerNotEqualsResult struct {
			ID        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []IntegerNotEqualsResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotEquals(eb.Column("view_count"), 0)
				})
			}).
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotEquals with integer should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotEqual(int64(0), result.ViewCount, "ViewCount should not be 0")
			suite.T().Logf("ID: %s, ViewCount: %d", result.ID, result.ViewCount)
		}
	})

	// Test 3: NotEquals in SELECT clause
	suite.Run("NotEqualsInSelect", func() {
		type SelectNotEqualsResult struct {
			ID       string `bun:"id"`
			Status   string `bun:"status"`
			NotDraft bool   `bun:"not_draft"`
		}

		var results []SelectNotEqualsResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NotEquals(eb.Column("status"), "draft")
			}, "not_draft").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotEquals in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.Status != "draft" {
				suite.True(result.NotDraft, "NotDraft should be true for non-draft posts")
			} else {
				suite.False(result.NotDraft, "NotDraft should be false for draft posts")
			}

			suite.T().Logf("ID: %s, Status: %s, NotDraft: %v", result.ID, result.Status, result.NotDraft)
		}
	})
}

// TestGreaterThan tests the GreaterThan comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestGreaterThan() {
	suite.T().Logf("Testing GreaterThan comparison for %s", suite.dbType)

	// Test 1: Simple GreaterThan
	suite.Run("SimpleGreaterThan", func() {
		type GreaterThanResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []GreaterThanResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThan(eb.Column("view_count"), 50)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "GreaterThan should work")
		suite.True(len(results) > 0, "Should have GreaterThan results")

		for _, result := range results {
			suite.True(result.ViewCount > 50, "ViewCount should be > 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.ID, result.Title, result.ViewCount)
		}
	})

	// Test 2: GreaterThan in SELECT clause
	suite.Run("GreaterThanInSelect", func() {
		type SelectGreaterThanResult struct {
			ID         string `bun:"id"`
			ViewCount  int64  `bun:"view_count"`
			IsHighView bool   `bun:"is_high_view"`
		}

		var results []SelectGreaterThanResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.GreaterThan(eb.Column("view_count"), 80)
			}, "is_high_view").
			OrderBy("view_count").
			Limit(10).
			Scan(suite.ctx, &results)

		suite.NoError(err, "GreaterThan in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.ViewCount > 80 {
				suite.True(result.IsHighView, "IsHighView should be true when ViewCount > 80")
			} else {
				suite.False(result.IsHighView, "IsHighView should be false when ViewCount <= 80")
			}

			suite.T().Logf("ID: %s, ViewCount: %d, IsHighView: %v", result.ID, result.ViewCount, result.IsHighView)
		}
	})
}

// TestGreaterThanOrEqual tests the GreaterThanOrEqual comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestGreaterThanOrEqual() {
	suite.T().Logf("Testing GreaterThanOrEqual comparison for %s", suite.dbType)

	// Test 1: Simple GreaterThanOrEqual
	suite.Run("SimpleGreaterThanOrEqual", func() {
		type GreaterThanOrEqualResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []GreaterThanOrEqualResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThanOrEqual(eb.Column("view_count"), 30)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "GreaterThanOrEqual should work")
		suite.True(len(results) > 0, "Should have GreaterThanOrEqual results")

		for _, result := range results {
			suite.True(result.ViewCount >= 30, "ViewCount should be >= 30")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.ID, result.Title, result.ViewCount)
		}
	})

	// Test 2: GreaterThanOrEqual at boundary
	suite.Run("BoundaryGreaterThanOrEqual", func() {
		type BoundaryResult struct {
			ID        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []BoundaryResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.GreaterThanOrEqual(eb.Column("view_count"), 42)
				})
			}).
			OrderBy("view_count").
			Scan(suite.ctx, &results)

		suite.NoError(err, "GreaterThanOrEqual at boundary should work")

		for _, result := range results {
			suite.True(result.ViewCount >= 42, "ViewCount should be >= 42")
			suite.T().Logf("ID: %s, ViewCount: %d", result.ID, result.ViewCount)
		}
	})
}

// TestLessThan tests the LessThan comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestLessThan() {
	suite.T().Logf("Testing LessThan comparison for %s", suite.dbType)

	// Test 1: Simple LessThan
	suite.Run("SimpleLessThan", func() {
		type LessThanResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []LessThanResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.LessThan(eb.Column("view_count"), 70)
				})
			}).
			OrderByDesc("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "LessThan should work")
		suite.True(len(results) > 0, "Should have LessThan results")

		for _, result := range results {
			suite.True(result.ViewCount < 70, "ViewCount should be < 70")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.ID, result.Title, result.ViewCount)
		}
	})

	// Test 2: LessThan in SELECT clause
	suite.Run("LessThanInSelect", func() {
		type SelectLessThanResult struct {
			ID        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
			IsLowView bool   `bun:"is_low_view"`
		}

		var results []SelectLessThanResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.LessThan(eb.Column("view_count"), 30)
			}, "is_low_view").
			OrderBy("view_count").
			Limit(10).
			Scan(suite.ctx, &results)

		suite.NoError(err, "LessThan in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.ViewCount < 30 {
				suite.True(result.IsLowView, "IsLowView should be true when ViewCount < 30")
			} else {
				suite.False(result.IsLowView, "IsLowView should be false when ViewCount >= 30")
			}

			suite.T().Logf("ID: %s, ViewCount: %d, IsLowView: %v", result.ID, result.ViewCount, result.IsLowView)
		}
	})
}

// TestLessThanOrEqual tests the LessThanOrEqual comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestLessThanOrEqual() {
	suite.T().Logf("Testing LessThanOrEqual comparison for %s", suite.dbType)

	// Test 1: Simple LessThanOrEqual
	suite.Run("SimpleLessThanOrEqual", func() {
		type LessThanOrEqualResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []LessThanOrEqualResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.LessThanOrEqual(eb.Column("view_count"), 50)
				})
			}).
			OrderByDesc("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "LessThanOrEqual should work")
		suite.True(len(results) > 0, "Should have LessThanOrEqual results")

		for _, result := range results {
			suite.True(result.ViewCount <= 50, "ViewCount should be <= 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.ID, result.Title, result.ViewCount)
		}
	})

	// Test 2: LessThanOrEqual at boundary
	suite.Run("BoundaryLessThanOrEqual", func() {
		type BoundaryResult struct {
			ID        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []BoundaryResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.LessThanOrEqual(eb.Column("view_count"), 42)
				})
			}).
			OrderByDesc("view_count").
			Scan(suite.ctx, &results)

		suite.NoError(err, "LessThanOrEqual at boundary should work")

		for _, result := range results {
			suite.True(result.ViewCount <= 42, "ViewCount should be <= 42")
			suite.T().Logf("ID: %s, ViewCount: %d", result.ID, result.ViewCount)
		}
	})
}

// TestBetween tests the Between comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestBetween() {
	suite.T().Logf("Testing Between comparison for %s", suite.dbType)

	// Test 1: Simple Between with integers
	suite.Run("SimpleBetweenIntegers", func() {
		type BetweenResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []BetweenResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Between(eb.Column("view_count"), 30, 70)
				})
			}).
			OrderBy("view_count").
			Scan(suite.ctx, &results)

		suite.NoError(err, "Between with integers should work")
		suite.True(len(results) > 0, "Should have Between results")

		for _, result := range results {
			suite.True(result.ViewCount >= 30 && result.ViewCount <= 70, "ViewCount should be between 30 and 70")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.ID, result.Title, result.ViewCount)
		}
	})

	// Test 2: Between at boundaries
	suite.Run("BetweenAtBoundaries", func() {
		type BoundaryBetweenResult struct {
			ID        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []BoundaryBetweenResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.Between(eb.Column("view_count"), 42, 42)
				})
			}).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Between with same boundaries should work")

		for _, result := range results {
			suite.Equal(int64(42), result.ViewCount, "ViewCount should be exactly 42")
			suite.T().Logf("ID: %s, ViewCount: %d", result.ID, result.ViewCount)
		}
	})

	// Test 3: Between in SELECT clause
	suite.Run("BetweenInSelect", func() {
		type SelectBetweenResult struct {
			ID           string `bun:"id"`
			ViewCount    int64  `bun:"view_count"`
			IsMediumView bool   `bun:"is_medium_view"`
		}

		var results []SelectBetweenResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Between(eb.Column("view_count"), 30, 80)
			}, "is_medium_view").
			OrderBy("view_count").
			Limit(10).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Between in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.ViewCount >= 30 && result.ViewCount <= 80 {
				suite.True(result.IsMediumView, "IsMediumView should be true when between 30 and 80")
			} else {
				suite.False(result.IsMediumView, "IsMediumView should be false when outside range")
			}

			suite.T().Logf("ID: %s, ViewCount: %d, IsMediumView: %v", result.ID, result.ViewCount, result.IsMediumView)
		}
	})
}

// TestNotBetween tests the NotBetween comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestNotBetween() {
	suite.T().Logf("Testing NotBetween comparison for %s", suite.dbType)

	// Test 1: Simple NotBetween
	suite.Run("SimpleNotBetween", func() {
		type NotBetweenResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []NotBetweenResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotBetween(eb.Column("view_count"), 30, 70)
				})
			}).
			OrderBy("view_count").
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotBetween should work")
		suite.True(len(results) > 0, "Should have NotBetween results")

		for _, result := range results {
			suite.True(result.ViewCount < 30 || result.ViewCount > 70, "ViewCount should be outside 30-70 range")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d", result.ID, result.Title, result.ViewCount)
		}
	})

	// Test 2: NotBetween in SELECT clause
	suite.Run("NotBetweenInSelect", func() {
		type SelectNotBetweenResult struct {
			ID            string `bun:"id"`
			ViewCount     int64  `bun:"view_count"`
			IsExtremeView bool   `bun:"is_extreme_view"`
		}

		var results []SelectNotBetweenResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NotBetween(eb.Column("view_count"), 30, 80)
			}, "is_extreme_view").
			OrderBy("view_count").
			Limit(10).
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotBetween in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.ViewCount < 30 || result.ViewCount > 80 {
				suite.True(result.IsExtremeView, "IsExtremeView should be true when outside 30-80 range")
			} else {
				suite.False(result.IsExtremeView, "IsExtremeView should be false when in range")
			}

			suite.T().Logf("ID: %s, ViewCount: %d, IsExtremeView: %v", result.ID, result.ViewCount, result.IsExtremeView)
		}
	})
}

// TestIn tests the In comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestIn() {
	suite.T().Logf("Testing In comparison for %s", suite.dbType)

	// Test 1: Simple In with strings
	suite.Run("SimpleInStrings", func() {
		type InResult struct {
			ID     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var results []InResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.In(eb.Column("status"), "published", "review")
				})
			}).
			OrderBy("id").
			Scan(suite.ctx, &results)

		suite.NoError(err, "In with strings should work")
		suite.True(len(results) > 0, "Should have In results")

		for _, result := range results {
			suite.True(result.Status == "published" || result.Status == "review", "Status should be in allowed list")
			suite.T().Logf("ID: %s, Title: %s, Status: %s", result.ID, result.Title, result.Status)
		}
	})

	// Test 2: In with integers
	suite.Run("InIntegers", func() {
		type IntegerInResult struct {
			ID        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []IntegerInResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.In(eb.Column("view_count"), 23, 42, 85, 96)
				})
			}).
			OrderBy("view_count").
			Scan(suite.ctx, &results)

		suite.NoError(err, "In with integers should work")

		for _, result := range results {
			suite.True(
				result.ViewCount == 23 || result.ViewCount == 42 || result.ViewCount == 85 || result.ViewCount == 96,
				"ViewCount should be in allowed list")
			suite.T().Logf("ID: %s, ViewCount: %d", result.ID, result.ViewCount)
		}
	})

	// Test 3: In with single value
	suite.Run("InSingleValue", func() {
		type SingleInResult struct {
			ID     string `bun:"id"`
			Status string `bun:"status"`
		}

		var results []SingleInResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.In(eb.Column("status"), "draft")
				})
			}).
			OrderBy("id").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "In with single value should work")

		for _, result := range results {
			suite.Equal("draft", result.Status, "Status should be 'draft'")
			suite.T().Logf("ID: %s, Status: %s", result.ID, result.Status)
		}
	})

	// Test 4: In in SELECT clause
	suite.Run("InInSelect", func() {
		type SelectInResult struct {
			ID             string `bun:"id"`
			Status         string `bun:"status"`
			IsActiveStatus bool   `bun:"is_active_status"`
		}

		var results []SelectInResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.In(eb.Column("status"), "published", "review")
			}, "is_active_status").
			OrderBy("id").
			Limit(10).
			Scan(suite.ctx, &results)

		suite.NoError(err, "In in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.Status == "published" || result.Status == "review" {
				suite.True(result.IsActiveStatus, "IsActiveStatus should be true for active statuses")
			} else {
				suite.False(result.IsActiveStatus, "IsActiveStatus should be false for non-active statuses")
			}

			suite.T().Logf("ID: %s, Status: %s, IsActiveStatus: %v", result.ID, result.Status, result.IsActiveStatus)
		}
	})
}

// TestNotIn tests the NotIn comparison operator.
func (suite *ComparisonExpressionsTestSuite) TestNotIn() {
	suite.T().Logf("Testing NotIn comparison for %s", suite.dbType)

	// Test 1: Simple NotIn with strings
	suite.Run("SimpleNotInStrings", func() {
		type NotInResult struct {
			ID     string `bun:"id"`
			Title  string `bun:"title"`
			Status string `bun:"status"`
		}

		var results []NotInResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotIn(eb.Column("status"), "draft", "archived")
				})
			}).
			OrderBy("id").
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotIn with strings should work")
		suite.True(len(results) > 0, "Should have NotIn results")

		for _, result := range results {
			suite.True(result.Status != "draft" && result.Status != "archived", "Status should not be in excluded list")
			suite.T().Logf("ID: %s, Title: %s, Status: %s", result.ID, result.Title, result.Status)
		}
	})

	// Test 2: NotIn with integers
	suite.Run("NotInIntegers", func() {
		type IntegerNotInResult struct {
			ID        string `bun:"id"`
			ViewCount int64  `bun:"view_count"`
		}

		var results []IntegerNotInResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.NotIn(eb.Column("view_count"), 0, 10, 20)
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotIn with integers should work")

		for _, result := range results {
			suite.True(
				result.ViewCount != 0 && result.ViewCount != 10 && result.ViewCount != 20,
				"ViewCount should not be in excluded list")
			suite.T().Logf("ID: %s, ViewCount: %d", result.ID, result.ViewCount)
		}
	})

	// Test 3: NotIn in SELECT clause
	suite.Run("NotInInSelect", func() {
		type SelectNotInResult struct {
			ID            string `bun:"id"`
			Status        string `bun:"status"`
			IsNotExcluded bool   `bun:"is_not_excluded"`
		}

		var results []SelectNotInResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NotIn(eb.Column("status"), "draft", "archived")
			}, "is_not_excluded").
			OrderBy("id").
			Limit(10).
			Scan(suite.ctx, &results)

		suite.NoError(err, "NotIn in SELECT should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			if result.Status != "draft" && result.Status != "archived" {
				suite.True(result.IsNotExcluded, "IsNotExcluded should be true for non-excluded statuses")
			} else {
				suite.False(result.IsNotExcluded, "IsNotExcluded should be false for excluded statuses")
			}

			suite.T().Logf("ID: %s, Status: %s, IsNotExcluded: %v", result.ID, result.Status, result.IsNotExcluded)
		}
	})
}

// TestCombinedComparisons tests using multiple comparison operators together.
func (suite *ComparisonExpressionsTestSuite) TestCombinedComparisons() {
	suite.T().Logf("Testing combined comparison operators for %s", suite.dbType)

	// Test 1: Combined comparisons in SELECT
	suite.Run("CombinedComparisonsInSelect", func() {
		type CombinedResult struct {
			ID           string `bun:"id"`
			ViewCount    int64  `bun:"view_count"`
			IsHighView   bool   `bun:"is_high_view"`
			IsMediumView bool   `bun:"is_medium_view"`
			IsLowView    bool   `bun:"is_low_view"`
		}

		var results []CombinedResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.GreaterThan(eb.Column("view_count"), 80)
			}, "is_high_view").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Between(eb.Column("view_count"), 30, 80)
			}, "is_medium_view").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.LessThan(eb.Column("view_count"), 30)
			}, "is_low_view").
			OrderBy("view_count").
			Limit(10).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined comparison operators should work")
		suite.True(len(results) > 0, "Should have combined comparison results")

		for _, result := range results {
			// Verify categorization logic
			if result.ViewCount > 80 {
				suite.True(result.IsHighView, "Should be high view")
				suite.False(result.IsMediumView, "Should not be medium view")
				suite.False(result.IsLowView, "Should not be low view")
			} else if result.ViewCount >= 30 && result.ViewCount <= 80 {
				suite.False(result.IsHighView, "Should not be high view")
				suite.True(result.IsMediumView, "Should be medium view")
				suite.False(result.IsLowView, "Should not be low view")
			} else {
				suite.False(result.IsHighView, "Should not be high view")
				suite.False(result.IsMediumView, "Should not be medium view")
				suite.True(result.IsLowView, "Should be low view")
			}

			suite.T().Logf("ID: %s, ViewCount: %d, High: %v, Medium: %v, Low: %v",
				result.ID, result.ViewCount, result.IsHighView, result.IsMediumView, result.IsLowView)
		}
	})
}

// TestIsTrue tests the IsTrue comparison operator for boolean expressions.
func (suite *ComparisonExpressionsTestSuite) TestIsTrue() {
	suite.T().Logf("Testing IsTrue comparison for %s", suite.dbType)

	suite.Run("IsTrueOnBooleanColumn", func() {
		type IsTrueResult struct {
			ID       string `bun:"id"`
			Name     string `bun:"name"`
			IsActive bool   `bun:"is_active"`
		}

		var results []IsTrueResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "is_active").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.IsTrue(eb.Column("is_active"))
				})
			}).
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "IsTrue on boolean column should work")
		suite.True(len(results) > 0, "Should have IsTrue results")

		for _, result := range results {
			suite.True(result.IsActive, "IsActive should be true")
			suite.T().Logf("ID: %s, Name: %s, IsActive: %v", result.ID, result.Name, result.IsActive)
		}
	})

	suite.Run("IsTrueOnExpression", func() {
		type ExprIsTrueResult struct {
			ID        string `bun:"id"`
			Title     string `bun:"title"`
			ViewCount int64  `bun:"view_count"`
			IsPopular bool   `bun:"is_popular"`
		}

		var results []ExprIsTrueResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IsTrue(eb.GreaterThan(eb.Column("view_count"), 50))
			}, "is_popular").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.IsTrue(eb.GreaterThan(eb.Column("view_count"), 50))
				})
			}).
			OrderByDesc("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "IsTrue on expression should work")
		suite.True(len(results) > 0, "Should have IsTrue expression results")

		for _, result := range results {
			suite.True(result.ViewCount > 50, "ViewCount should be > 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d, IsPopular: %v",
				result.ID, result.Title, result.ViewCount, result.IsPopular)
		}
	})
}

// TestIsFalse tests the IsFalse comparison operator for boolean expressions.
func (suite *ComparisonExpressionsTestSuite) TestIsFalse() {
	suite.T().Logf("Testing IsFalse comparison for %s", suite.dbType)

	suite.Run("IsFalseOnBooleanColumn", func() {
		type IsFalseResult struct {
			ID       string `bun:"id"`
			Name     string `bun:"name"`
			IsActive bool   `bun:"is_active"`
		}

		var results []IsFalseResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "is_active").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.IsFalse(eb.Column("is_active"))
				})
			}).
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "IsFalse on boolean column should work")
		suite.True(len(results) > 0, "Should have IsFalse results")

		for _, result := range results {
			suite.False(result.IsActive, "IsActive should be false")
			suite.T().Logf("ID: %s, Name: %s, IsActive: %v", result.ID, result.Name, result.IsActive)
		}
	})

	suite.Run("IsFalseOnExpression", func() {
		type ExprIsFalseResult struct {
			ID           string `bun:"id"`
			Title        string `bun:"title"`
			ViewCount    int64  `bun:"view_count"`
			IsNotPopular bool   `bun:"is_not_popular"`
		}

		var results []ExprIsFalseResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.IsFalse(eb.GreaterThan(eb.Column("view_count"), 50))
			}, "is_not_popular").
			Where(func(cb ConditionBuilder) {
				cb.Expr(func(eb ExprBuilder) any {
					return eb.IsFalse(eb.GreaterThan(eb.Column("view_count"), 50))
				})
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "IsFalse on expression should work")
		suite.True(len(results) > 0, "Should have IsFalse expression results")

		for _, result := range results {
			suite.False(result.ViewCount > 50, "ViewCount should be <= 50")
			suite.T().Logf("ID: %s, Title: %s, ViewCount: %d, IsNotPopular: %v",
				result.ID, result.Title, result.ViewCount, result.IsNotPopular)
		}
	})
}
