package orm

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

// WindowFunctionsTestSuite tests window function methods of ExprBuilder including
// ranking functions (RowNumber, Rank, DenseRank, PercentRank, CumeDist, Ntile),
// offset functions (Lag, Lead), value functions (FirstValue, LastValue, NthValue),
// and aggregate window functions (WinCount, WinSum, WinAvg, WinMin, WinMax, WinStringAgg,
// WinArrayAgg, WinStdDev, WinVariance, WinJsonObjectAgg, WinJsonArrayAgg, WinBitOr,
// WinBitAnd, WinBoolOr, WinBoolAnd).
//
// This suite verifies cross-database compatibility for window functions across
// PostgreSQL, MySQL, and SQLite, handling database-specific features appropriately.
type WindowFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestRowNumber tests the ROW_NUMBER window function.
func (suite *WindowFunctionsTestSuite) TestRowNumber() {
	suite.T().Logf("Testing RowNumber function for %s", suite.DbType)

	suite.Run("SequentialRowNumbers", func() {
		type UserWithRowNumber struct {
			Id     string `bun:"id"`
			Name   string `bun:"name"`
			Age    int16  `bun:"age"`
			RowNum int64  `bun:"row_num"`
		}

		var usersWithRowNum []UserWithRowNumber

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name", "age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.RowNumber(func(rn RowNumberBuilder) {
					rn.Over().OrderBy("age")
				})
			}, "row_num").
			OrderBy("age").
			Scan(suite.Ctx, &usersWithRowNum)

		suite.NoError(err, "ROW_NUMBER should work correctly")
		suite.Len(usersWithRowNum, 3, "Should have 3 users")

		// Verify ROW_NUMBER sequence
		for i, user := range usersWithRowNum {
			suite.Equal(int64(i+1), user.RowNum, "ROW_NUMBER should be sequential starting from 1")
		}

		suite.T().Logf("Row numbers verified: %d users with sequential numbers", len(usersWithRowNum))
	})
}

// TestRank tests the RANK window function.
func (suite *WindowFunctionsTestSuite) TestRank() {
	suite.T().Logf("Testing Rank function for %s", suite.DbType)

	suite.Run("RankPartitionedByStatus", func() {
		type PostWithRank struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			Status    string `bun:"status"`
			ViewCount int64  `bun:"view_count"`
			Rank      int64  `bun:"rank"`
		}

		var postsWithRank []PostWithRank

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Rank(func(r RankBuilder) {
					r.Over().PartitionBy("status").OrderByDesc("view_count")
				})
			}, "rank").
			OrderBy("status").
			OrderByDesc("view_count").
			Scan(suite.Ctx, &postsWithRank)

		suite.NoError(err, "RANK should work with partitioning")
		suite.True(len(postsWithRank) > 0, "Should have posts with rank")

		// Verify ranking within partitions
		statusGroups := make(map[string][]PostWithRank)
		for _, post := range postsWithRank {
			statusGroups[post.Status] = append(statusGroups[post.Status], post)
		}

		for status, posts := range statusGroups {
			if len(posts) > 0 {
				suite.True(posts[0].Rank >= 1, "First post in %s partition should have rank >= 1", status)
			}
		}

		suite.T().Logf("Verified ranks for %d posts across %d status groups", len(postsWithRank), len(statusGroups))
	})
}

// TestDenseRank tests the DENSE_RANK window function.
func (suite *WindowFunctionsTestSuite) TestDenseRank() {
	suite.T().Logf("Testing DenseRank function for %s", suite.DbType)

	suite.Run("DenseRankPartitionedByStatus", func() {
		type PostWithDenseRank struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			Status    string `bun:"status"`
			ViewCount int64  `bun:"view_count"`
			Rank      int64  `bun:"rank"`
			DenseRank int64  `bun:"dense_rank"`
		}

		var postsWithRank []PostWithDenseRank

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title", "status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Rank(func(r RankBuilder) {
					r.Over().PartitionBy("status").OrderByDesc("view_count")
				})
			}, "rank").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.DenseRank(func(dr DenseRankBuilder) {
					dr.Over().PartitionBy("status").OrderByDesc("view_count")
				})
			}, "dense_rank").
			OrderBy("status").
			OrderByDesc("view_count").
			Scan(suite.Ctx, &postsWithRank)

		suite.NoError(err, "DENSE_RANK should work with partitioning")
		suite.True(len(postsWithRank) > 0, "Should have posts with dense rank")

		// Verify ranking within partitions
		statusGroups := make(map[string][]PostWithDenseRank)
		for _, post := range postsWithRank {
			statusGroups[post.Status] = append(statusGroups[post.Status], post)
		}

		for status, posts := range statusGroups {
			if len(posts) > 1 {
				suite.True(posts[0].Rank >= 1, "First post in %s partition should have rank >= 1", status)
				suite.True(posts[0].DenseRank >= 1, "First post in %s partition should have dense_rank >= 1", status)
			}
		}

		suite.T().Logf("Verified dense ranks for %d posts across %d status groups", len(postsWithRank), len(statusGroups))
	})
}

// TestPercentRank tests the PERCENT_RANK window function.
func (suite *WindowFunctionsTestSuite) TestPercentRank() {
	suite.T().Logf("Testing PercentRank function for %s", suite.DbType)

	suite.Run("PercentRankByViewCount", func() {
		type PostAnalytics struct {
			Title          string  `bun:"title"`
			Status         string  `bun:"status"`
			ViewCount      int64   `bun:"view_count"`
			RankInStatus   int64   `bun:"rank_in_status"`
			PercentOfTotal float64 `bun:"percent_of_total"`
		}

		var postAnalytics []PostAnalytics

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Rank(func(r RankBuilder) {
					r.Over().PartitionBy("status").OrderByDesc("view_count")
				})
			}, "rank_in_status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.PercentRank(func(pr PercentRankBuilder) {
					pr.Over().OrderByDesc("view_count")
				})
			}, "percent_of_total").
			OrderBy("status").
			OrderByDesc("view_count").
			Scan(suite.Ctx, &postAnalytics)

		suite.NoError(err, "PERCENT_RANK should work correctly")
		suite.True(len(postAnalytics) > 0, "Should have post analytics")

		// Verify percent rank is within valid range
		for _, post := range postAnalytics {
			suite.True(post.RankInStatus >= 1, "Rank should be at least 1")
			suite.True(post.PercentOfTotal >= 0 && post.PercentOfTotal <= 1, "Percent rank should be between 0 and 1")
		}

		suite.T().Logf("Verified percent ranks for %d posts", len(postAnalytics))
	})
}

// TestCumeDist tests the CUME_DIST window function.
func (suite *WindowFunctionsTestSuite) TestCumeDist() {
	suite.T().Logf("Testing CumeDist function for %s", suite.DbType)

	suite.Run("CumeDistByViewCount", func() {
		type CumeDistResult struct {
			Id        string  `bun:"id"`
			ViewCount int64   `bun:"view_count"`
			CumeDist  float64 `bun:"cume_dist"`
		}

		var cumeDistResults []CumeDistResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CumeDist(func(cdb CumeDistBuilder) {
					cdb.Over().OrderBy("view_count")
				})
			}, "cume_dist").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.Ctx, &cumeDistResults)

		suite.NoError(err, "CUME_DIST should work correctly")
		suite.True(len(cumeDistResults) > 0, "Should have cume_dist results")

		for _, result := range cumeDistResults {
			suite.True(result.CumeDist > 0 && result.CumeDist <= 1, "CumeDist should be in (0, 1]")
			suite.T().Logf("ID: %s, ViewCount: %d, CumeDist: %.4f",
				result.Id, result.ViewCount, result.CumeDist)
		}
	})
}

// TestNtile tests the NTILE window function.
func (suite *WindowFunctionsTestSuite) TestNtile() {
	suite.T().Logf("Testing Ntile function for %s", suite.DbType)

	suite.Run("QuartilesUsingNtile", func() {
		type UserWithQuartile struct {
			Name     string `bun:"name"`
			Age      int16  `bun:"age"`
			Quartile int64  `bun:"quartile"`
		}

		var usersWithQuartile []UserWithQuartile

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("name", "age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Ntile(func(nb NtileBuilder) {
					nb.Buckets(4).Over().OrderBy("age")
				})
			}, "quartile").
			OrderBy("age").
			Scan(suite.Ctx, &usersWithQuartile)

		suite.NoError(err, "NTILE should work for quartile distribution")
		suite.Len(usersWithQuartile, 3, "Should have 3 users")

		// Verify quartile assignment
		for _, user := range usersWithQuartile {
			suite.True(user.Quartile >= 1 && user.Quartile <= 4, "Quartile should be between 1 and 4")
		}

		suite.T().Logf("Verified quartiles for %d users", len(usersWithQuartile))
	})
}

// TestLag tests the LAG window function.
func (suite *WindowFunctionsTestSuite) TestLag() {
	suite.T().Logf("Testing Lag function for %s", suite.DbType)

	suite.Run("LagWithDefaultOffset", func() {
		type PostWithLag struct {
			Title         string `bun:"title"`
			ViewCount     int64  `bun:"view_count"`
			PrevViewCount *int64 `bun:"prev_view_count"`
		}

		var postsWithLag []PostWithLag

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lag(func(lb LagBuilder) {
					lb.Column("view_count").Over().OrderBy("view_count")
				})
			}, "prev_view_count").
			OrderBy("view_count").
			Scan(suite.Ctx, &postsWithLag)

		suite.NoError(err, "LAG should work with default offset")
		suite.True(len(postsWithLag) > 0, "Should have posts with lag")

		// First row should have null prev_view_count
		if len(postsWithLag) > 0 {
			suite.Nil(postsWithLag[0].PrevViewCount, "First row should have null previous value")
		}

		suite.T().Logf("Verified LAG for %d posts", len(postsWithLag))
	})

	suite.Run("LagWithCustomOffset", func() {
		type PostWithLagAdvanced struct {
			Title          string `bun:"title"`
			ViewCount      int64  `bun:"view_count"`
			Prev2ViewCount *int64 `bun:"prev2_view_count"`
		}

		var advLag []PostWithLagAdvanced

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lag(func(lb LagBuilder) {
					lb.Column("view_count").Offset(2).Over().OrderBy("view_count")
				})
			}, "prev2_view_count").
			OrderBy("view_count").
			Scan(suite.Ctx, &advLag)

		suite.NoError(err, "LAG should work with custom offset")
		suite.True(len(advLag) > 0, "Should have posts for advanced lag")

		if len(advLag) >= 3 {
			// The third row's Prev2 should equal the first row's view_count
			if advLag[2].Prev2ViewCount != nil {
				suite.Equal(advLag[0].ViewCount, *advLag[2].Prev2ViewCount, "Third row's LAG(2) should match first row's value")
			}
		}

		suite.T().Logf("Verified LAG with offset 2 for %d posts", len(advLag))
	})
}

// TestLead tests the LEAD window function.
func (suite *WindowFunctionsTestSuite) TestLead() {
	suite.T().Logf("Testing Lead function for %s", suite.DbType)

	suite.Run("LeadWithDefaultOffset", func() {
		type PostWithLead struct {
			Title         string `bun:"title"`
			ViewCount     int64  `bun:"view_count"`
			NextViewCount *int64 `bun:"next_view_count"`
		}

		var postsWithLead []PostWithLead

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lead(func(lb LeadBuilder) {
					lb.Column("view_count").Over().OrderBy("view_count")
				})
			}, "next_view_count").
			OrderBy("view_count").
			Scan(suite.Ctx, &postsWithLead)

		suite.NoError(err, "LEAD should work with default offset")
		suite.True(len(postsWithLead) > 0, "Should have posts with lead")

		// Last row should have null next_view_count
		if len(postsWithLead) > 0 {
			lastIdx := len(postsWithLead) - 1
			suite.Nil(postsWithLead[lastIdx].NextViewCount, "Last row should have null next value")
		}

		suite.T().Logf("Verified LEAD for %d posts", len(postsWithLead))
	})

	suite.Run("LeadWithCustomOffsetAndDefault", func() {
		type PostWithLeadAdvanced struct {
			Title          string `bun:"title"`
			ViewCount      int64  `bun:"view_count"`
			Next2ViewCount *int64 `bun:"next2_view_count"`
			Next2OrDefault int64  `bun:"next2_or_default"`
		}

		var advLead []PostWithLeadAdvanced

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lead(func(lb LeadBuilder) {
					lb.Column("view_count").Offset(2).Over().OrderBy("view_count")
				})
			}, "next2_view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lead(func(lb LeadBuilder) {
					lb.Column("view_count").Offset(2).DefaultValue(-1).Over().OrderBy("view_count")
				})
			}, "next2_or_default").
			OrderBy("view_count").
			Scan(suite.Ctx, &advLead)

		suite.NoError(err, "LEAD should work with custom offset and default value")
		suite.True(len(advLead) > 0, "Should have posts for advanced lead")

		// Default value should apply on the last rows where LEAD overflows
		if len(advLead) >= 1 {
			lastIdx := len(advLead) - 1
			suite.NotZero(advLead[lastIdx].Next2OrDefault, "Default value should be applied when LEAD exceeds bounds")
		}

		suite.T().Logf("Verified LEAD with offset 2 and default for %d posts", len(advLead))
	})
}

// TestFirstValue tests the FIRST_VALUE window function.
func (suite *WindowFunctionsTestSuite) TestFirstValue() {
	suite.T().Logf("Testing FirstValue function for %s", suite.DbType)

	suite.Run("FirstValuePartitionedByStatus", func() {
		type PostWithFirstValue struct {
			Title         string `bun:"title"`
			Status        string `bun:"status"`
			ViewCount     int64  `bun:"view_count"`
			FirstInStatus int64  `bun:"first_in_status"`
		}

		var postsWithFirst []PostWithFirstValue

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.FirstValue(func(fvb FirstValueBuilder) {
					fvb.Column("view_count").Over().PartitionBy("status").OrderBy("view_count")
				})
			}, "first_in_status").
			OrderBy("status", "view_count").
			Scan(suite.Ctx, &postsWithFirst)

		suite.NoError(err, "FIRST_VALUE should work with partitioning")
		suite.True(len(postsWithFirst) > 0, "Should have posts with first value")

		// Verify FIRST_VALUE behavior
		statusFirstValues := make(map[string][]PostWithFirstValue)
		for _, post := range postsWithFirst {
			statusFirstValues[post.Status] = append(statusFirstValues[post.Status], post)
		}

		for status, posts := range statusFirstValues {
			if len(posts) > 1 {
				// All posts in same status should have same first_in_status value
				firstValue := posts[0].FirstInStatus
				for _, post := range posts {
					suite.Equal(firstValue, post.FirstInStatus, "All posts in %s should have same first value", status)
				}
			}
		}

		suite.T().Logf("Verified FIRST_VALUE for %d posts across %d status groups", len(postsWithFirst), len(statusFirstValues))
	})
}

// TestLastValue tests the LAST_VALUE window function.
func (suite *WindowFunctionsTestSuite) TestLastValue() {
	suite.T().Logf("Testing LastValue function for %s", suite.DbType)

	suite.Run("LastValuePartitionedByStatus", func() {
		type PostWithLastValue struct {
			Title        string `bun:"title"`
			Status       string `bun:"status"`
			ViewCount    int64  `bun:"view_count"`
			LastInStatus int64  `bun:"last_in_status"`
		}

		var postsWithLast []PostWithLastValue

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.LastValue(func(lvb LastValueBuilder) {
					lvb.Column("view_count").Over().PartitionBy("status").OrderBy("view_count").Rows().UnboundedPreceding().And().UnboundedFollowing()
				})
			}, "last_in_status").
			OrderBy("status", "view_count").
			Scan(suite.Ctx, &postsWithLast)

		suite.NoError(err, "LAST_VALUE should work with partitioning")
		suite.True(len(postsWithLast) > 0, "Should have posts with last value")

		// Verify LAST_VALUE behavior
		statusLastValues := make(map[string][]PostWithLastValue)
		for _, post := range postsWithLast {
			statusLastValues[post.Status] = append(statusLastValues[post.Status], post)
		}

		for status, posts := range statusLastValues {
			if len(posts) > 1 {
				// All posts in same status should have same last_in_status value
				lastValue := posts[0].LastInStatus
				for _, post := range posts {
					suite.Equal(lastValue, post.LastInStatus, "All posts in %s should have same last value", status)
				}
			}
		}

		suite.T().Logf("Verified LAST_VALUE for %d posts across %d status groups", len(postsWithLast), len(statusLastValues))
	})
}

// TestNthValue tests the NTH_VALUE window function.
func (suite *WindowFunctionsTestSuite) TestNthValue() {
	suite.T().Logf("Testing NthValue function for %s", suite.DbType)

	suite.Run("SecondValueInPartition", func() {
		type PostWithNthValue struct {
			Status        string `bun:"status"`
			ViewCount     int64  `bun:"view_count"`
			SecondFromEnd int64  `bun:"second_from_end"`
		}

		var nthVals []PostWithNthValue

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.NthValue(func(nvb NthValueBuilder) {
					nvb.Column("view_count").N(2).Over().PartitionBy("status").OrderBy("view_count").Rows().UnboundedPreceding().And().UnboundedFollowing()
				})
			}, "second_from_end").
			OrderBy("status", "view_count").
			Scan(suite.Ctx, &nthVals)

		suite.NoError(err, "NTH_VALUE should work with full frame")
		suite.True(len(nthVals) > 0, "Should compute NTH_VALUE")

		suite.T().Logf("Verified NTH_VALUE for %d posts", len(nthVals))
	})
}

// TestWinCount tests the COUNT window function.
func (suite *WindowFunctionsTestSuite) TestWinCount() {
	suite.T().Logf("Testing WinCount function for %s", suite.DbType)

	suite.Run("RunningCount", func() {
		type UserWithRunningCount struct {
			Name         string `bun:"name"`
			Age          int16  `bun:"age"`
			RunningCount int64  `bun:"running_count"`
		}

		var usersWithCount []UserWithRunningCount

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("name", "age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinCount(func(wc WindowCountBuilder) {
					wc.All().Over().OrderBy("age").Rows().UnboundedPreceding()
				})
			}, "running_count").
			OrderBy("age").
			Scan(suite.Ctx, &usersWithCount)

		suite.NoError(err, "WinCount should work for running count")
		suite.Len(usersWithCount, 3, "Should have 3 users")

		// Verify running counts
		for i, user := range usersWithCount {
			suite.Equal(int64(i+1), user.RunningCount, "Running count should increment by 1")
		}

		suite.T().Logf("Verified running count for %d users", len(usersWithCount))
	})
}

// TestWinSum tests the SUM window function.
func (suite *WindowFunctionsTestSuite) TestWinSum() {
	suite.T().Logf("Testing WinSum function for %s", suite.DbType)

	suite.Run("RunningTotal", func() {
		type UserWithRunningTotal struct {
			Name         string `bun:"name"`
			Age          int16  `bun:"age"`
			RunningTotal int64  `bun:"running_total"`
		}

		var usersWithSum []UserWithRunningTotal

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("name", "age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinSum(func(ws WindowSumBuilder) {
					ws.Column("age").Over().OrderBy("age").Rows().UnboundedPreceding()
				})
			}, "running_total").
			OrderBy("age").
			Scan(suite.Ctx, &usersWithSum)

		suite.NoError(err, "WinSum should work for running total")
		suite.Len(usersWithSum, 3, "Should have 3 users")

		// Verify running totals
		for _, user := range usersWithSum {
			suite.True(user.RunningTotal > 0, "Running total should be positive")
		}

		suite.T().Logf("Verified running total for %d users", len(usersWithSum))
	})

	suite.Run("CumulativeViewsByStatus", func() {
		type PostWithCumulativeViews struct {
			Title           string `bun:"title"`
			Status          string `bun:"status"`
			ViewCount       int64  `bun:"view_count"`
			CumulativeViews int64  `bun:"cumulative_views"`
		}

		var postsWithCumulative []PostWithCumulativeViews

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinSum(func(ws WindowSumBuilder) {
					ws.Column("view_count").Over().PartitionBy("status").OrderByDesc("view_count").Rows().UnboundedPreceding()
				})
			}, "cumulative_views").
			OrderBy("status").
			OrderByDesc("view_count").
			Scan(suite.Ctx, &postsWithCumulative)

		suite.NoError(err, "WinSum should work with partitioning")
		suite.True(len(postsWithCumulative) > 0, "Should have posts with cumulative views")

		for _, post := range postsWithCumulative {
			suite.True(post.CumulativeViews >= post.ViewCount, "Cumulative views should be at least equal to current view count")
		}

		suite.T().Logf("Verified cumulative views for %d posts", len(postsWithCumulative))
	})
}

// TestWinAvg tests the AVG window function.
func (suite *WindowFunctionsTestSuite) TestWinAvg() {
	suite.T().Logf("Testing WinAvg function for %s", suite.DbType)

	suite.Run("MovingAverage", func() {
		type UserWithMovingAvg struct {
			Name      string  `bun:"name"`
			Age       int16   `bun:"age"`
			MovingAvg float64 `bun:"moving_avg"`
		}

		var usersWithAvg []UserWithMovingAvg

		err := suite.Db.NewSelect().
			Model((*User)(nil)).
			Select("name", "age").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinAvg(func(wa WindowAvgBuilder) {
					wa.Column("age").Over().OrderBy("age").Rows().UnboundedPreceding()
				})
			}, "moving_avg").
			OrderBy("age").
			Scan(suite.Ctx, &usersWithAvg)

		suite.NoError(err, "WinAvg should work for moving average")
		suite.Len(usersWithAvg, 3, "Should have 3 users")

		// Verify moving averages
		for _, user := range usersWithAvg {
			suite.True(user.MovingAvg > 0, "Moving average should be positive")
		}

		suite.T().Logf("Verified moving average for %d users", len(usersWithAvg))
	})

	suite.Run("ThreeRowMovingAverage", func() {
		type PostWithMovingAvg struct {
			Title     string  `bun:"title"`
			ViewCount int64   `bun:"view_count"`
			MovAvg    float64 `bun:"mov_avg"`
		}

		var movingAvgRows []PostWithMovingAvg

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinAvg(func(wab WindowAvgBuilder) {
					wab.Column("view_count").Over().OrderBy("view_count").Rows().Preceding(2).And().CurrentRow()
				})
			}, "mov_avg").
			OrderBy("view_count").
			Scan(suite.Ctx, &movingAvgRows)

		suite.NoError(err, "WinAvg should work with ROWS BETWEEN 2 PRECEDING AND CURRENT ROW")
		suite.True(len(movingAvgRows) > 0, "Should have moving average values")

		suite.T().Logf("Verified 3-row moving average for %d posts", len(movingAvgRows))
	})
}

// TestWinMin tests the MIN window function.
func (suite *WindowFunctionsTestSuite) TestWinMin() {
	suite.T().Logf("Testing WinMin function for %s", suite.DbType)

	suite.Run("MinInStatusPartition", func() {
		type WindowMinResult struct {
			Id          string `bun:"id"`
			ViewCount   int64  `bun:"view_count"`
			Status      string `bun:"status"`
			MinInStatus int64  `bun:"min_in_status"`
		}

		var windowMinResults []WindowMinResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinMin(func(wmb WindowMinBuilder) {
					wmb.Column("view_count").Over().PartitionBy("status")
				})
			}, "min_in_status").
			OrderBy("status", "id").
			Limit(8).
			Scan(suite.Ctx, &windowMinResults)

		suite.NoError(err, "WinMin should work correctly")
		suite.True(len(windowMinResults) > 0, "Should have window min results")

		for _, result := range windowMinResults {
			suite.True(result.MinInStatus <= result.ViewCount, "Min should be <= current view count")
			suite.T().Logf("ID: %s, Status: %s, ViewCount: %d, MinInStatus: %d",
				result.Id, result.Status, result.ViewCount, result.MinInStatus)
		}
	})
}

// TestWinMax tests the MAX window function.
func (suite *WindowFunctionsTestSuite) TestWinMax() {
	suite.T().Logf("Testing WinMax function for %s", suite.DbType)

	suite.Run("MaxInStatusPartition", func() {
		type WindowMaxResult struct {
			Id          string `bun:"id"`
			ViewCount   int64  `bun:"view_count"`
			Status      string `bun:"status"`
			MaxInStatus int64  `bun:"max_in_status"`
		}

		var windowMaxResults []WindowMaxResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinMax(func(wmb WindowMaxBuilder) {
					wmb.Column("view_count").Over().PartitionBy("status")
				})
			}, "max_in_status").
			OrderBy("status", "id").
			Limit(8).
			Scan(suite.Ctx, &windowMaxResults)

		suite.NoError(err, "WinMax should work correctly")
		suite.True(len(windowMaxResults) > 0, "Should have window max results")

		for _, result := range windowMaxResults {
			suite.True(result.MaxInStatus >= result.ViewCount, "Max should be >= current view count")
			suite.T().Logf("ID: %s, Status: %s, ViewCount: %d, MaxInStatus: %d",
				result.Id, result.Status, result.ViewCount, result.MaxInStatus)
		}
	})
}

// TestWinStringAgg tests the STRING_AGG window function.
// Note: MySQL does not support GROUP_CONCAT as a window function.
func (suite *WindowFunctionsTestSuite) TestWinStringAgg() {
	suite.T().Logf("Testing WinStringAgg function for %s", suite.DbType)

	suite.Run("StringAggPartitionedByStatus", func() {
		if suite.DbType == constants.DbMySQL {
			suite.T().Skipf("WinStringAgg skipped for %s (MySQL does not support GROUP_CONCAT as window function)", suite.DbType)

			return
		}

		type WindowStringAggResult struct {
			Id       string `bun:"id"`
			Status   string `bun:"status"`
			TitleAgg string `bun:"title_agg"`
		}

		var windowStringAggResults []WindowStringAggResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinStringAgg(func(wsab WindowStringAggBuilder) {
					wsab.Column("title").Separator(", ").Over().PartitionBy("status")
				})
			}, "title_agg").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowStringAggResults)

		suite.NoError(err, "WinStringAgg should work correctly")
		suite.True(len(windowStringAggResults) > 0, "Should have window string agg results")

		for _, result := range windowStringAggResults {
			suite.True(len(result.TitleAgg) > 0, "Aggregated titles should not be empty")
			suite.T().Logf("ID: %s, Status: %s, TitleAgg: %s",
				result.Id, result.Status, result.TitleAgg)
		}
	})
}

// TestWinArrayAgg tests the ARRAY_AGG window function (PostgreSQL only).
func (suite *WindowFunctionsTestSuite) TestWinArrayAgg() {
	suite.T().Logf("Testing WinArrayAgg function for %s", suite.DbType)

	suite.Run("ArrayAggPartitionedByStatus", func() {
		if suite.DbType != constants.DbPostgres {
			suite.T().Skipf("WinArrayAgg skipped for %s (PostgreSQL only)", suite.DbType)
		}

		type WindowArrayAggResult struct {
			Id         string  `bun:"id"`
			Status     string  `bun:"status"`
			ViewCounts []int64 `bun:"view_counts,array"`
		}

		var windowArrayAggResults []WindowArrayAggResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinArrayAgg(func(waab WindowArrayAggBuilder) {
					waab.Column("view_count").Over().PartitionBy("status")
				})
			}, "view_counts").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowArrayAggResults)

		suite.NoError(err, "WinArrayAgg should work correctly")
		suite.True(len(windowArrayAggResults) > 0, "Should have window array agg results")

		for _, result := range windowArrayAggResults {
			suite.True(len(result.ViewCounts) > 0, "Array should not be empty")
			suite.T().Logf("ID: %s, Status: %s, ViewCounts: %v",
				result.Id, result.Status, result.ViewCounts)
		}
	})
}

// TestWinStdDev tests the STDDEV window function.
// Note: SQLite does not support statistical functions.
func (suite *WindowFunctionsTestSuite) TestWinStdDev() {
	suite.T().Logf("Testing WinStdDev function for %s", suite.DbType)

	suite.Run("StdDevInStatusPartition", func() {
		if suite.DbType == constants.DbSQLite {
			suite.T().Skipf("WinStdDev skipped for %s (SQLite does not support statistical functions)", suite.DbType)

			return
		}

		type WindowStdDevResult struct {
			Id             string  `bun:"id"`
			ViewCount      int64   `bun:"view_count"`
			Status         string  `bun:"status"`
			StdDevInStatus float64 `bun:"stddev_in_status"`
		}

		var windowStdDevResults []WindowStdDevResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinStdDev(func(wsb WindowStdDevBuilder) {
					wsb.Column("view_count").Over().PartitionBy("status")
				})
			}, "stddev_in_status").
			OrderBy("status", "id").
			Limit(8).
			Scan(suite.Ctx, &windowStdDevResults)

		suite.NoError(err, "WinStdDev should work correctly")
		suite.True(len(windowStdDevResults) > 0, "Should have window stddev results")

		for _, result := range windowStdDevResults {
			suite.True(result.StdDevInStatus >= 0, "StdDev should be non-negative")
			suite.T().Logf("ID: %s, Status: %s, ViewCount: %d, StdDevInStatus: %.2f",
				result.Id, result.Status, result.ViewCount, result.StdDevInStatus)
		}
	})
}

// TestWinVariance tests the VARIANCE window function.
// Note: SQLite does not support statistical functions.
func (suite *WindowFunctionsTestSuite) TestWinVariance() {
	suite.T().Logf("Testing WinVariance function for %s", suite.DbType)

	suite.Run("VarianceInStatusPartition", func() {
		if suite.DbType == constants.DbSQLite {
			suite.T().Skipf("WinVariance skipped for %s (SQLite does not support statistical functions)", suite.DbType)

			return
		}

		type WindowVarianceResult struct {
			Id               string  `bun:"id"`
			ViewCount        int64   `bun:"view_count"`
			Status           string  `bun:"status"`
			VarianceInStatus float64 `bun:"variance_in_status"`
		}

		var windowVarianceResults []WindowVarianceResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinVariance(func(wvb WindowVarianceBuilder) {
					wvb.Column("view_count").Over().PartitionBy("status")
				})
			}, "variance_in_status").
			OrderBy("status", "id").
			Limit(8).
			Scan(suite.Ctx, &windowVarianceResults)

		suite.NoError(err, "WinVariance should work correctly")
		suite.True(len(windowVarianceResults) > 0, "Should have window variance results")

		for _, result := range windowVarianceResults {
			suite.True(result.VarianceInStatus >= 0, "Variance should be non-negative")
			suite.T().Logf("ID: %s, Status: %s, ViewCount: %d, VarianceInStatus: %.2f",
				result.Id, result.Status, result.ViewCount, result.VarianceInStatus)
		}
	})
}

// TestWinJsonObjectAgg tests the JSON_OBJECT_AGG window function.
func (suite *WindowFunctionsTestSuite) TestWinJsonObjectAgg() {
	suite.T().Logf("Testing WinJsonObjectAgg function for %s", suite.DbType)

	suite.Run("JsonObjectAggPartitionedByStatus", func() {
		type WindowJsonObjectResult struct {
			Id            string `bun:"id"`
			Status        string `bun:"status"`
			JsonObjectAgg string `bun:"json_object_agg"`
		}

		var windowJsonObjectResults []WindowJsonObjectResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinJsonObjectAgg(func(wjoab WindowJsonObjectAggBuilder) {
					wjoab.KeyColumn("id").Column("title").Over().PartitionBy("status")
				})
			}, "json_object_agg").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowJsonObjectResults)

		suite.NoError(err, "WinJsonObjectAgg should work correctly")
		suite.True(len(windowJsonObjectResults) > 0, "Should have window JSON object agg results")

		for _, result := range windowJsonObjectResults {
			suite.True(len(result.JsonObjectAgg) > 0, "JSON object should not be empty")
			suite.True(strings.HasPrefix(result.JsonObjectAgg, "{"), "Should be a JSON object")
			suite.T().Logf("ID: %s, Status: %s, JsonObjectAgg: %s",
				result.Id, result.Status, result.JsonObjectAgg)
		}
	})
}

// TestWinJsonArrayAgg tests the JSON_ARRAY_AGG window function.
func (suite *WindowFunctionsTestSuite) TestWinJsonArrayAgg() {
	suite.T().Logf("Testing WinJsonArrayAgg function for %s", suite.DbType)

	suite.Run("JsonArrayAggPartitionedByStatus", func() {
		type WindowJsonResult struct {
			Id           string `bun:"id"`
			Status       string `bun:"status"`
			JsonArrayAgg string `bun:"json_array_agg"`
		}

		var windowJsonResults []WindowJsonResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinJsonArrayAgg(func(wjaab WindowJsonArrayAggBuilder) {
					wjaab.Column("title").Over().PartitionBy("status")
				})
			}, "json_array_agg").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowJsonResults)

		suite.NoError(err, "WinJsonArrayAgg should work correctly")
		suite.True(len(windowJsonResults) > 0, "Should have window JSON array agg results")

		for _, result := range windowJsonResults {
			suite.True(len(result.JsonArrayAgg) > 0, "JSON array should not be empty")
			suite.True(strings.HasPrefix(result.JsonArrayAgg, "["), "Should be a JSON array")
			suite.T().Logf("ID: %s, Status: %s, JsonArrayAgg: %s",
				result.Id, result.Status, result.JsonArrayAgg)
		}
	})
}

// TestWinBitOr tests the BIT_OR window function.
// WinBitOr performs bitwise OR within a window frame.
// Note: PostgreSQL and MySQL support native BIT_OR.
// SQLite simulates it using MAX with CASE for boolean-like operations.
func (suite *WindowFunctionsTestSuite) TestWinBitOr() {
	suite.T().Logf("Testing WinBitOr function for %s", suite.DbType)

	suite.Run("BitOrInStatusPartition", func() {
		type WindowBitResult struct {
			Id          string `bun:"id"`
			ViewCount   int64  `bun:"view_count"`
			Status      string `bun:"status"`
			BitOrResult int64  `bun:"bit_or_result"`
		}

		var windowBitResults []WindowBitResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinBitOr(func(wbob WindowBitOrBuilder) {
					wbob.Column("view_count").Over().PartitionBy("status")
				})
			}, "bit_or_result").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowBitResults)

		suite.NoError(err, "WinBitOr should work correctly")
		suite.True(len(windowBitResults) > 0, "Should have window bit OR results")

		for _, result := range windowBitResults {
			suite.True(result.BitOrResult >= 0, "BitOr should be non-negative")
			suite.T().Logf("ID: %s, Status: %s, ViewCount: %d, BitOr: %d",
				result.Id, result.Status, result.ViewCount, result.BitOrResult)
		}
	})
}

// TestWinBitAnd tests the BIT_AND window function.
// WinBitAnd performs bitwise AND within a window frame.
// Note: PostgreSQL and MySQL support native BIT_AND.
// SQLite simulates it using MIN with CASE for boolean-like operations.
func (suite *WindowFunctionsTestSuite) TestWinBitAnd() {
	suite.T().Logf("Testing WinBitAnd function for %s", suite.DbType)

	suite.Run("BitAndInStatusPartition", func() {
		type WindowBitResult struct {
			Id           string `bun:"id"`
			ViewCount    int64  `bun:"view_count"`
			Status       string `bun:"status"`
			BitAndResult int64  `bun:"bit_and_result"`
		}

		var windowBitResults []WindowBitResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "view_count", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinBitAnd(func(wbab WindowBitAndBuilder) {
					wbab.Column("view_count").Over().PartitionBy("status")
				})
			}, "bit_and_result").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowBitResults)

		suite.NoError(err, "WinBitAnd should work correctly")
		suite.True(len(windowBitResults) > 0, "Should have window bit AND results")

		for _, result := range windowBitResults {
			suite.True(result.BitAndResult >= 0, "BitAnd should be non-negative")
			suite.T().Logf("ID: %s, Status: %s, ViewCount: %d, BitAnd: %d",
				result.Id, result.Status, result.ViewCount, result.BitAndResult)
		}
	})
}

// TestWinBoolOr tests the BOOL_OR window function.
// WinBoolOr performs boolean OR within a window frame.
// Framework uses BOOL_OR (PostgreSQL), MAX+CASE simulation (MySQL/SQLite).
func (suite *WindowFunctionsTestSuite) TestWinBoolOr() {
	suite.T().Logf("Testing WinBoolOr function for %s", suite.DbType)

	suite.Run("BoolOrInStatusPartition", func() {
		type WindowBoolResult struct {
			Id           string `bun:"id"`
			Status       string `bun:"status"`
			BoolOrResult bool   `bun:"bool_or_result"`
		}

		var windowBoolResults []WindowBoolResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinBoolOr(func(wbob WindowBoolOrBuilder) {
					wbob.Expr(eb.Expr("? > 100", eb.Column("view_count"))).Over().PartitionBy("status")
				})
			}, "bool_or_result").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowBoolResults)

		suite.NoError(err, "WinBoolOr should work correctly")
		suite.True(len(windowBoolResults) > 0, "Should have window bool OR results")

		for _, result := range windowBoolResults {
			suite.T().Logf("ID: %s, Status: %s, BoolOr: %v",
				result.Id, result.Status, result.BoolOrResult)
		}
	})
}

// TestWinBoolAnd tests the BOOL_AND window function.
// WinBoolAnd performs boolean AND within a window frame.
// Framework uses BOOL_AND (PostgreSQL), MIN+CASE simulation (MySQL/SQLite).
func (suite *WindowFunctionsTestSuite) TestWinBoolAnd() {
	suite.T().Logf("Testing WinBoolAnd function for %s", suite.DbType)

	suite.Run("BoolAndInStatusPartition", func() {
		type WindowBoolResult struct {
			Id            string `bun:"id"`
			Status        string `bun:"status"`
			BoolAndResult bool   `bun:"bool_and_result"`
		}

		var windowBoolResults []WindowBoolResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.WinBoolAnd(func(wbab WindowBoolAndBuilder) {
					wbab.Expr(eb.Expr("? > 100", eb.Column("view_count"))).Over().PartitionBy("status")
				})
			}, "bool_and_result").
			OrderBy("status", "id").
			Limit(5).
			Scan(suite.Ctx, &windowBoolResults)

		suite.NoError(err, "WinBoolAnd should work correctly")
		suite.True(len(windowBoolResults) > 0, "Should have window bool AND results")

		for _, result := range windowBoolResults {
			suite.T().Logf("ID: %s, Status: %s, BoolAnd: %v",
				result.Id, result.Status, result.BoolAndResult)
		}
	})
}
