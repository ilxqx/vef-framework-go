package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
)

// AggregationFunctionsTestSuite tests aggregate expression methods of ExprBuilder
// including Count, Sum, Avg, Min, Max, StringAgg, ArrayAgg, JsonObjectAgg, JsonArrayAgg,
// BitOr, BitAnd, BoolOr, BoolAnd, StdDev, and Variance functions.
//
// This suite verifies cross-database compatibility for aggregation functions across
// PostgreSQL, MySQL, and SQLite, handling database-specific features appropriately.
type AggregationFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestCount tests the Count aggregate function with various scenarios.
func (suite *AggregationFunctionsTestSuite) TestCount() {
	suite.T().Logf("Testing Count function for %s", suite.dbType)

	suite.Run("ConditionalCountWithFilter", func() {
		// Test Count with FILTER clause (PostgreSQL, SQLite) vs CASE equivalent (MySQL)
		type ConditionalCounts struct {
			TotalPosts     int64 `bun:"total_posts"`
			PublishedPosts int64 `bun:"published_posts"`
			DraftPosts     int64 `bun:"draft_posts"`
			ReviewPosts    int64 `bun:"review_posts"`
			HighViewPosts  int64 `bun:"high_view_posts"`
		}

		var counts ConditionalCounts

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountAll()
			}, "total_posts").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Count(func(cb CountBuilder) {
					cb.All().Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
				})
			}, "published_posts").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Count(func(cb CountBuilder) {
					cb.All().Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "draft")
					})
				})
			}, "draft_posts").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Count(func(cb CountBuilder) {
					cb.All().Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "review")
					})
				})
			}, "review_posts").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Count(func(cb CountBuilder) {
					cb.All().Filter(func(cb ConditionBuilder) {
						cb.GreaterThan("view_count", 80)
					})
				})
			}, "high_view_posts").
			Scan(suite.ctx, &counts)

		suite.NoError(err, "Conditional count aggregation should work")
		suite.True(counts.TotalPosts > 0, "Should have posts")
		suite.True(counts.PublishedPosts >= 0, "Should count published posts")
		suite.True(counts.DraftPosts >= 0, "Should count draft posts")
		suite.True(counts.ReviewPosts >= 0, "Should count review posts")
		suite.True(counts.HighViewPosts >= 0, "Should count high view posts")

		suite.T().Logf("Total: %d, Published: %d, Draft: %d, Review: %d, HighView: %d",
			counts.TotalPosts, counts.PublishedPosts, counts.DraftPosts,
			counts.ReviewPosts, counts.HighViewPosts)
	})
}

// TestCountColumn tests the CountColumn aggregate function.
func (suite *AggregationFunctionsTestSuite) TestCountColumn() {
	suite.T().Logf("Testing CountColumn function for %s", suite.dbType)

	suite.Run("BasicCountColumn", func() {
		type AgeStats struct {
			CountAge int64 `bun:"count_age"`
		}

		var ageStats AgeStats

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("age")
			}, "count_age").
			Scan(suite.ctx, &ageStats)

		suite.NoError(err, "CountColumn should work")
		suite.Equal(int64(3), ageStats.CountAge, "Should count 3 users with age")

		suite.T().Logf("Counted %d age values", ageStats.CountAge)
	})

	suite.Run("DistinctCountColumn", func() {
		type DistinctStats struct {
			UniqueStatuses   int64 `bun:"unique_statuses"`
			UniqueCategories int64 `bun:"unique_categories"`
			DistinctUserIDs  int64 `bun:"distinct_user_ids"`
		}

		var stats DistinctStats

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("status", true) // distinct = true
			}, "unique_statuses").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("category_id", true)
			}, "unique_categories").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountColumn("user_id", true)
			}, "distinct_user_ids").
			Scan(suite.ctx, &stats)

		suite.NoError(err, "DISTINCT CountColumn should work")
		suite.True(stats.UniqueStatuses > 0, "Should have unique statuses")
		suite.True(stats.UniqueCategories > 0, "Should have unique categories")
		suite.True(stats.DistinctUserIDs > 0, "Should have distinct user IDs")

		suite.T().Logf("Unique statuses: %d, categories: %d, users: %d",
			stats.UniqueStatuses, stats.UniqueCategories, stats.DistinctUserIDs)
	})
}

// TestCountAll tests the CountAll aggregate function.
func (suite *AggregationFunctionsTestSuite) TestCountAll() {
	suite.T().Logf("Testing CountAll function for %s", suite.dbType)

	suite.Run("CountAllWithGrouping", func() {
		type StatusCount struct {
			Status string `bun:"status"`
			Count  int64  `bun:"post_count"`
		}

		var statusCounts []StatusCount

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountAll()
			}, "post_count").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &statusCounts)

		suite.NoError(err, "CountAll with GROUP BY should work")
		suite.True(len(statusCounts) > 0, "Should have status counts")

		totalCount := int64(0)
		for _, sc := range statusCounts {
			suite.True(sc.Count > 0, "Each status should have at least 1 post")
			totalCount += sc.Count
			suite.T().Logf("Status %s: %d posts", sc.Status, sc.Count)
		}

		// Verify total matches expected post count from fixture
		expectedPosts := 8 // From fixture.yaml
		suite.Equal(int64(expectedPosts), totalCount, "Total count should match fixture posts")
	})
}

// TestSum tests the Sum aggregate function with builder callback.
func (suite *AggregationFunctionsTestSuite) TestSum() {
	suite.T().Logf("Testing Sum function for %s", suite.dbType)

	suite.Run("SumWithFilter", func() {
		type ConditionalSums struct {
			TotalViews     int64 `bun:"total_views"`
			PublishedViews int64 `bun:"published_views"`
		}

		var sums ConditionalSums

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SumColumn("view_count")
			}, "total_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sum(func(sb SumBuilder) {
					sb.Column("view_count").Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
				})
			}, "published_views").
			Scan(suite.ctx, &sums)

		suite.NoError(err, "Sum with Filter should work")
		suite.True(sums.TotalViews > 0, "Should have total views")
		suite.True(sums.PublishedViews >= 0, "Should have published views")
		suite.True(sums.PublishedViews <= sums.TotalViews, "Published views should not exceed total")

		suite.T().Logf("Total views: %d, Published views: %d", sums.TotalViews, sums.PublishedViews)
	})

	suite.Run("SumInComplexGroupBy", func() {
		type CategoryStats struct {
			CategoryName   string  `bun:"category_name"`
			TotalPosts     int64   `bun:"total_posts"`
			PublishedPosts int64   `bun:"published_posts"`
			DraftPosts     int64   `bun:"draft_posts"`
			TotalViews     int64   `bun:"total_views"`
			PublishedViews int64   `bun:"published_views"`
			AvgViews       float64 `bun:"avg_views"`
			MaxViews       int64   `bun:"max_views"`
			MinViews       int64   `bun:"min_views"`
		}

		var stats []CategoryStats

		query := suite.db.NewSelect().
			Model((*Post)(nil)).
			Join((*Category)(nil), func(cb ConditionBuilder) {
				cb.EqualsColumn("c.id", "category_id")
			}).
			SelectAs("c.name", "category_name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountAll()
			}, "total_posts").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SumColumn("view_count")
			}, "total_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}, "avg_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MaxColumn("view_count")
			}, "max_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MinColumn("view_count")
			}, "min_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Count(func(cb CountBuilder) {
					cb.All().Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
				})
			}, "published_posts").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Count(func(cb CountBuilder) {
					cb.All().Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "draft")
					})
				})
			}, "draft_posts").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sum(func(sb SumBuilder) {
					sb.Column("p.view_count").Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
				})
			}, "published_views").
			GroupBy("c.id", "c.name").
			OrderBy("c.name")

		err := query.Scan(suite.ctx, &stats)
		suite.NoError(err, "Complex GROUP BY with multiple aggregates should work")
		suite.True(len(stats) > 0, "Should have category statistics")

		for _, stat := range stats {
			suite.NotEmpty(stat.CategoryName, "Category name should not be empty")
			suite.True(stat.TotalPosts > 0, "Each category should have posts")
			suite.True(stat.TotalViews >= 0, "Total views should be non-negative")
			suite.True(stat.PublishedViews >= 0, "Published views should be non-negative")
			suite.True(stat.PublishedPosts >= 0, "Published posts should be non-negative")
			suite.True(stat.DraftPosts >= 0, "Draft posts should be non-negative")
			suite.True(stat.PublishedPosts+stat.DraftPosts <= stat.TotalPosts,
				"Published + Draft should not exceed total")

			suite.T().Logf("Category %s: %d posts (%d published, %d draft), %d total views (%d published views)",
				stat.CategoryName, stat.TotalPosts, stat.PublishedPosts, stat.DraftPosts,
				stat.TotalViews, stat.PublishedViews)
		}
	})
}

// TestSumColumn tests the SumColumn aggregate function.
func (suite *AggregationFunctionsTestSuite) TestSumColumn() {
	suite.T().Logf("Testing SumColumn function for %s", suite.dbType)

	suite.Run("BasicSumColumn", func() {
		type ViewStats struct {
			TotalViews int64   `bun:"total_views"`
			AvgViews   float64 `bun:"avg_views"`
		}

		var stats ViewStats

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SumColumn("view_count")
			}, "total_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}, "avg_views").
			Scan(suite.ctx, &stats)

		suite.NoError(err, "SumColumn should work")
		suite.True(stats.TotalViews > 0, "Should have total views")
		suite.True(stats.AvgViews > 0, "Should have average views")

		suite.T().Logf("Total views: %d, Average views: %.2f", stats.TotalViews, stats.AvgViews)
	})
}

// TestAvg tests the Avg aggregate function with builder callback.
func (suite *AggregationFunctionsTestSuite) TestAvg() {
	suite.T().Logf("Testing Avg function for %s", suite.dbType)

	suite.Run("AvgWithFilter", func() {
		type ConditionalAvg struct {
			AvgPublishedViews int64 `bun:"avg_published_views"`
		}

		var avg ConditionalAvg

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ToInteger(
					eb.Avg(func(ab AvgBuilder) {
						ab.Column("view_count").Filter(func(cb ConditionBuilder) {
							cb.Equals("status", "published")
						})
					}),
				)
			}, "avg_published_views").
			Scan(suite.ctx, &avg)

		suite.NoError(err, "Avg with Filter should work")
		suite.True(avg.AvgPublishedViews >= 0, "Average should be non-negative")

		suite.T().Logf("Average published views: %d", avg.AvgPublishedViews)
	})
}

// TestAvgColumn tests the AvgColumn aggregate function.
func (suite *AggregationFunctionsTestSuite) TestAvgColumn() {
	suite.T().Logf("Testing AvgColumn function for %s", suite.dbType)

	suite.Run("BasicAvgColumn", func() {
		type AgeStats struct {
			AvgAge float64 `bun:"avg_age"`
		}

		var stats AgeStats

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("age")
			}, "avg_age").
			Scan(suite.ctx, &stats)

		suite.NoError(err, "AvgColumn should work")
		suite.InDelta(30.0, stats.AvgAge, 1.0, "Average age should be around 30")

		suite.T().Logf("Average age: %.2f", stats.AvgAge)
	})

	suite.Run("DistinctAvgColumn", func() {
		type DistinctAvg struct {
			AvgDistinctViews float64 `bun:"avg_distinct_views"`
		}

		var avg DistinctAvg

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count", true) // distinct average
			}, "avg_distinct_views").
			Scan(suite.ctx, &avg)

		suite.NoError(err, "DISTINCT AvgColumn should work")
		suite.True(avg.AvgDistinctViews > 0, "Should have distinct average")

		suite.T().Logf("Distinct average views: %.2f", avg.AvgDistinctViews)
	})
}

// TestMin tests the Min aggregate function with builder callback.
func (suite *AggregationFunctionsTestSuite) TestMin() {
	suite.T().Logf("Testing Min function for %s", suite.dbType)

	suite.Run("MinInCombinedStats", func() {
		type CombinedStats struct {
			Status   string  `bun:"status"`
			MinViews int64   `bun:"min_views"`
			MaxViews int64   `bun:"max_views"`
			AvgViews float64 `bun:"avg_views"`
		}

		var stats []CombinedStats

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MinColumn("view_count")
			}, "min_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MaxColumn("view_count")
			}, "max_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}, "avg_views").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &stats)

		suite.NoError(err, "Min in combined stats should work")
		suite.True(len(stats) > 0, "Should have results")

		for _, stat := range stats {
			suite.True(stat.MinViews >= 0, "Min should be non-negative")
			suite.True(stat.MaxViews >= stat.MinViews, "Max should be >= Min")
			suite.T().Logf("Status: %s, Min: %d, Max: %d, Avg: %.2f",
				stat.Status, stat.MinViews, stat.MaxViews, stat.AvgViews)
		}
	})
}

// TestMinColumn tests the MinColumn aggregate function.
func (suite *AggregationFunctionsTestSuite) TestMinColumn() {
	suite.T().Logf("Testing MinColumn function for %s", suite.dbType)

	suite.Run("BasicMinColumn", func() {
		type AgeStats struct {
			MinAge int16 `bun:"min_age"`
		}

		var stats AgeStats

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MinColumn("age")
			}, "min_age").
			Scan(suite.ctx, &stats)

		suite.NoError(err, "MinColumn should work")
		suite.Equal(int16(25), stats.MinAge, "Min age should be 25")

		suite.T().Logf("Minimum age: %d", stats.MinAge)
	})
}

// TestMax tests the Max aggregate function with builder callback.
func (suite *AggregationFunctionsTestSuite) TestMax() {
	suite.T().Logf("Testing Max function for %s", suite.dbType)

	suite.Run("MaxWithFilter", func() {
		type MaxResult struct {
			MaxPublishedViews int64 `bun:"max_published_views"`
		}

		var result MaxResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Max(func(mb MaxBuilder) {
					mb.Column("view_count").Filter(func(cb ConditionBuilder) {
						cb.Equals("status", "published")
					})
				})
			}, "max_published_views").
			Scan(suite.ctx, &result)

		suite.NoError(err, "Max with Filter should work")
		suite.True(result.MaxPublishedViews >= 0, "Max should be non-negative")

		suite.T().Logf("Max published views: %d", result.MaxPublishedViews)
	})

	suite.Run("MaxInGroupBy", func() {
		type GroupedMax struct {
			Status   string `bun:"status"`
			MaxViews int64  `bun:"max_views"`
		}

		var results []GroupedMax

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MaxColumn("view_count")
			}, "max_views").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "Max in GROUP BY should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.True(result.MaxViews >= 0, "Max should be non-negative")
			suite.T().Logf("Status: %s, Max views: %d", result.Status, result.MaxViews)
		}
	})
}

// TestMaxColumn tests the MaxColumn aggregate function.
func (suite *AggregationFunctionsTestSuite) TestMaxColumn() {
	suite.T().Logf("Testing MaxColumn function for %s", suite.dbType)

	suite.Run("BasicMaxColumn", func() {
		type AgeStats struct {
			MaxAge int16 `bun:"max_age"`
		}

		var stats AgeStats

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MaxColumn("age")
			}, "max_age").
			Scan(suite.ctx, &stats)

		suite.NoError(err, "MaxColumn should work")
		suite.Equal(int16(35), stats.MaxAge, "Max age should be 35")

		suite.T().Logf("Maximum age: %d", stats.MaxAge)
	})
}

// TestStringAgg tests the StringAgg aggregate function.
func (suite *AggregationFunctionsTestSuite) TestStringAgg() {
	suite.T().Logf("Testing StringAgg function for %s", suite.dbType)

	suite.Run("StringAggWithDistinctAndSeparator", func() {
		type StringAggResult struct {
			StatusList       string `bun:"status_list"`
			OrderedTitleList string `bun:"ordered_title_list"`
		}

		var result StringAggResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StringAgg(func(sab StringAggBuilder) {
					sab.Column("status").Separator(", ").Distinct()
				})
			}, "status_list").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StringAgg(func(sab StringAggBuilder) {
					sab.Column("title").Separator(" | ").OrderBy("view_count")
				})
			}, "ordered_title_list").
			Scan(suite.ctx, &result)

		suite.NoError(err, "StringAgg should work")
		suite.NotEmpty(result.StatusList, "Should aggregate distinct statuses")
		suite.NotEmpty(result.OrderedTitleList, "Should aggregate ordered titles")

		suite.T().Logf("Status list: %s", result.StatusList)
		suite.T().Logf("Ordered title list: %s", result.OrderedTitleList)
	})

	suite.Run("StringAggGroupedByStatus", func() {
		type StringAggResult struct {
			Status string `bun:"status"`
			Titles string `bun:"titles"`
		}

		var results []StringAggResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StringAgg(func(sab StringAggBuilder) {
					sab.Column("title").Separator(", ")
				})
			}, "titles").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "StringAgg with GROUP BY should work across all databases")
		suite.True(len(results) > 0, "Should have StringAgg results")

		for _, result := range results {
			suite.NotEmpty(result.Titles, "Aggregated titles should not be empty")
			suite.T().Logf("Status: %s, Titles: %s", result.Status, result.Titles)
		}
	})
}

// TestArrayAgg tests the ArrayAgg aggregate function.
func (suite *AggregationFunctionsTestSuite) TestArrayAgg() {
	suite.T().Logf("Testing ArrayAgg function for %s", suite.dbType)

	suite.Run("ArrayAggWithOrdering", func() {
		type ArrayAggResult struct {
			ViewCountArray []int64  `bun:"view_count_array,array"`
			OrderedTitles  []string `bun:"ordered_titles,array"`
			UniqueStatuses []string `bun:"unique_statuses,array"`
		}

		var result ArrayAggResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ArrayAgg(func(aab ArrayAggBuilder) {
					aab.Column("view_count").OrderByDesc("view_count")
				})
			}, "view_count_array").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ArrayAgg(func(aab ArrayAggBuilder) {
					aab.Column("title").OrderBy("title")
				})
			}, "ordered_titles").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ArrayAgg(func(aab ArrayAggBuilder) {
					aab.Column("status").Distinct().OrderBy("status")
				})
			}, "unique_statuses").
			Scan(suite.ctx, &result)

		suite.NoError(err, "ArrayAgg should work")
		suite.True(len(result.ViewCountArray) > 0, "Should have view count array")
		suite.True(len(result.OrderedTitles) > 0, "Should have ordered titles")
		suite.True(len(result.UniqueStatuses) > 0, "Should have unique statuses")

		// Verify ordering (only PostgreSQL supports ORDER BY in ARRAY_AGG)
		if suite.dbType == constants.Postgres {
			for i := 1; i < len(result.ViewCountArray); i++ {
				suite.True(result.ViewCountArray[i-1] >= result.ViewCountArray[i],
					"View counts should be in descending order")
			}
		}

		suite.T().Logf("View count array length: %d", len(result.ViewCountArray))
		suite.T().Logf("Ordered titles length: %d", len(result.OrderedTitles))
		suite.T().Logf("Unique statuses: %v", result.UniqueStatuses)
	})

	suite.Run("ArrayAggGroupedByStatus", func() {
		type ArrayAggResult struct {
			Status string `bun:"status"`
			Count  int64  `bun:"count"`
		}

		var results []ArrayAggResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountAll()
			}, "count").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "ArrayAgg supporting query should work across all databases")
		suite.True(len(results) > 0, "Should have ArrayAgg supporting results")

		for _, result := range results {
			suite.True(result.Count > 0, "Count should be positive")
			suite.T().Logf("Status: %s, Count: %d", result.Status, result.Count)
		}
	})
}

// TestJsonObjectAgg tests the JsonObjectAgg aggregate function.
func (suite *AggregationFunctionsTestSuite) TestJsonObjectAgg() {
	suite.T().Logf("Testing JsonObjectAgg function for %s", suite.dbType)

	suite.Run("JsonObjectAggGroupedByStatus", func() {
		type JsonObjectAggResult struct {
			Status     string `bun:"status"`
			StatusMeta string `bun:"status_meta"`
		}

		var results []JsonObjectAggResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObjectAgg(func(joab JSONObjectAggBuilder) {
					joab.KeyColumn("id")
					joab.Column("title")
				})
			}, "status_meta").
			GroupBy("status").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonObjectAgg should work across all databases")
		suite.True(len(results) > 0, "Should have JsonObjectAgg results")

		for _, result := range results {
			suite.NotEmpty(result.StatusMeta, "Aggregated JSON should not be empty")
			suite.T().Logf("Status: %s, Aggregated: %s", result.Status, result.StatusMeta)
		}
	})
}

// TestJsonArrayAgg tests the JsonArrayAgg aggregate function.
// Note: MySQL does not support ORDER BY or DISTINCT in JSON_ARRAYAGG.
func (suite *AggregationFunctionsTestSuite) TestJsonArrayAgg() {
	suite.T().Logf("Testing JsonArrayAgg function for %s", suite.dbType)

	suite.Run("JsonArrayAggBasic", func() {
		type JsonArrayAggResult struct {
			Status     string `bun:"status"`
			TitlesJSON string `bun:"titles_json"`
		}

		var results []JsonArrayAggResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONArrayAgg(func(jaab JSONArrayAggBuilder) {
					jaab.Column("title")
				})
			}, "titles_json").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonArrayAgg should work")
		suite.True(len(results) > 0, "Should have JsonArrayAgg results")

		for _, result := range results {
			suite.NotEmpty(result.TitlesJSON, "Aggregated JSON array should not be empty")
			suite.T().Logf("Status: %s, JSON Array: %s", result.Status, result.TitlesJSON)
		}
	})

	suite.Run("JsonArrayAggWithOrdering", func() {
		if suite.dbType == constants.MySQL {
			suite.T().Skipf("JsonArrayAgg with ORDER BY skipped for %s (MySQL does not support ORDER BY in JSON_ARRAYAGG)", suite.dbType)

			return
		}

		type OrderedJsonArray struct {
			Status     string `bun:"status"`
			TitlesJSON string `bun:"titles_json"`
		}

		var results []OrderedJsonArray

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONArrayAgg(func(jaab JSONArrayAggBuilder) {
					jaab.Column("title").OrderBy("view_count")
				})
			}, "titles_json").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonArrayAgg with ORDER BY should work on supported databases")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.NotEmpty(result.TitlesJSON, "JSON array should not be empty")
			suite.T().Logf("Status: %s, Ordered JSON Array: %s", result.Status, result.TitlesJSON)
		}
	})

	suite.Run("JsonArrayAggWithDistinct", func() {
		if suite.dbType == constants.MySQL {
			suite.T().Skipf("JsonArrayAgg with DISTINCT skipped for %s (MySQL does not support DISTINCT in JSON_ARRAYAGG)", suite.dbType)

			return
		}

		type DistinctStatusJson struct {
			AllStatuses string `bun:"all_statuses"`
		}

		var result DistinctStatusJson

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONArrayAgg(func(jaab JSONArrayAggBuilder) {
					jaab.Column("status").Distinct()
				})
			}, "all_statuses").
			Scan(suite.ctx, &result)

		suite.NoError(err, "JsonArrayAgg with DISTINCT should work on supported databases")
		suite.NotEmpty(result.AllStatuses, "Should have JSON array of statuses")

		suite.T().Logf("All Statuses JSON: %s", result.AllStatuses)
	})
}

// TestBitOr tests the BitOr aggregate function.
// Note: PostgreSQL and MySQL both support native BIT_OR (bitwise aggregation).
// SQLite simulates it using MAX with CASE for boolean-like operations.
func (suite *AggregationFunctionsTestSuite) TestBitOr() {
	suite.T().Logf("Testing BitOr function for %s", suite.dbType)

	suite.Run("BitOrBasic", func() {
		type BitOrResult struct {
			Status     string `bun:"status"`
			BitOrValue int64  `bun:"bit_or_value"`
		}

		var results []BitOrResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BitOr(func(bob BitOrBuilder) {
					bob.Column("view_count")
				})
			}, "bit_or_value").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "BitOr should work across all databases")
		suite.True(len(results) > 0, "Should have BitOr results")

		for _, result := range results {
			suite.True(result.BitOrValue >= 0, "BitOr result should be non-negative")
			suite.T().Logf("Status: %s, BitOr: %d", result.Status, result.BitOrValue)
		}
	})

	suite.Run("BitOrWithGroupBy", func() {
		type GroupedBitOr struct {
			Status     string `bun:"status"`
			Count      int64  `bun:"count"`
			BitOrValue int64  `bun:"bit_or_value"`
		}

		var results []GroupedBitOr

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountAll()
			}, "count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BitOr(func(bob BitOrBuilder) {
					bob.Column("view_count")
				})
			}, "bit_or_value").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "BitOr with GROUP BY should work")
		suite.True(len(results) > 0, "Should have grouped results")

		for _, result := range results {
			suite.True(result.Count > 0, "Count should be positive")
			suite.True(result.BitOrValue >= 0, "BitOr should be non-negative")
			suite.T().Logf("Status: %s, Count: %d, BitOr: %d",
				result.Status, result.Count, result.BitOrValue)
		}
	})
}

// TestBitAnd tests the BitAnd aggregate function.
// Note: PostgreSQL and MySQL both support native BIT_AND (bitwise aggregation).
// SQLite simulates it using MIN with CASE for boolean-like operations.
func (suite *AggregationFunctionsTestSuite) TestBitAnd() {
	suite.T().Logf("Testing BitAnd function for %s", suite.dbType)

	suite.Run("BitAndBasic", func() {
		type BitAndResult struct {
			Status      string `bun:"status"`
			BitAndValue int64  `bun:"bit_and_value"`
		}

		var results []BitAndResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BitAnd(func(bab BitAndBuilder) {
					bab.Column("view_count")
				})
			}, "bit_and_value").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "BitAnd should work across all databases")
		suite.True(len(results) > 0, "Should have BitAnd results")

		for _, result := range results {
			suite.True(result.BitAndValue >= 0, "BitAnd result should be non-negative")
			suite.T().Logf("Status: %s, BitAnd: %d", result.Status, result.BitAndValue)
		}
	})

	suite.Run("BitAndWithGroupBy", func() {
		type GroupedBitAnd struct {
			Status      string `bun:"status"`
			Count       int64  `bun:"count"`
			BitAndValue int64  `bun:"bit_and_value"`
			BitOrValue  int64  `bun:"bit_or_value"`
		}

		var results []GroupedBitAnd

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountAll()
			}, "count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BitAnd(func(bab BitAndBuilder) {
					bab.Column("view_count")
				})
			}, "bit_and_value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BitOr(func(bob BitOrBuilder) {
					bob.Column("view_count")
				})
			}, "bit_or_value").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "BitAnd and BitOr with GROUP BY should work")
		suite.True(len(results) > 0, "Should have grouped results")

		for _, result := range results {
			suite.True(result.Count > 0, "Count should be positive")
			suite.True(result.BitAndValue >= 0, "BitAnd should be non-negative")
			suite.True(result.BitOrValue >= 0, "BitOr should be non-negative")
			suite.T().Logf("Status: %s, Count: %d, BitAnd: %d, BitOr: %d",
				result.Status, result.Count, result.BitAndValue, result.BitOrValue)
		}
	})
}

// TestBoolOr tests the BoolOr aggregate function.
func (suite *AggregationFunctionsTestSuite) TestBoolOr() {
	suite.T().Logf("Testing BoolOr function for %s", suite.dbType)

	suite.Run("BoolOrWithBoolAnd", func() {
		type BoolResult struct {
			Status       string `bun:"status"`
			AllHighViews bool   `bun:"all_high_views"`
			AnyHighViews bool   `bun:"any_high_views"`
		}

		var results []BoolResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BoolAnd(func(bab BoolAndBuilder) {
					bab.Expr(eb.Expr("? > 100", eb.Column("view_count")))
				})
			}, "all_high_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BoolOr(func(bob BoolOrBuilder) {
					bob.Expr(eb.Expr("? > 100", eb.Column("view_count")))
				})
			}, "any_high_views").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "BoolOr and BoolAnd should work across all databases")
		suite.True(len(results) > 0, "Should have boolean operation results")

		for _, result := range results {
			suite.T().Logf("Status: %s, AllHighViews: %v, AnyHighViews: %v",
				result.Status, result.AllHighViews, result.AnyHighViews)
		}
	})
}

// TestBoolAnd tests the BoolAnd aggregate function.
func (suite *AggregationFunctionsTestSuite) TestBoolAnd() {
	suite.T().Logf("Testing BoolAnd function for %s", suite.dbType)

	suite.Run("BoolAndBasic", func() {
		type BoolAndResult struct {
			AllMeetCriteria bool `bun:"all_meet_criteria"`
		}

		var result BoolAndResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BoolAnd(func(bab BoolAndBuilder) {
					bab.Expr(eb.Expr("? > 10", eb.Column("view_count")))
				})
			}, "all_meet_criteria").
			Scan(suite.ctx, &result)

		suite.NoError(err, "BoolAnd should work across all databases")
		suite.T().Logf("All posts have > 10 views: %v", result.AllMeetCriteria)
	})

	suite.Run("BoolAndGrouped", func() {
		type GroupedBoolAnd struct {
			Status            string `bun:"status"`
			AllAboveThreshold bool   `bun:"all_above_threshold"`
		}

		var results []GroupedBoolAnd

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.BoolAnd(func(bab BoolAndBuilder) {
					bab.Expr(eb.Expr("? >= 30", eb.Column("view_count")))
				})
			}, "all_above_threshold").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "BoolAnd with GROUP BY should work across all databases")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.T().Logf("Status: %s, All >= 30 views: %v",
				result.Status, result.AllAboveThreshold)
		}
	})
}

// TestStdDev tests the StdDev (standard deviation) aggregate function.
func (suite *AggregationFunctionsTestSuite) TestStdDev() {
	suite.T().Logf("Testing StdDev function for %s", suite.dbType)

	suite.Run("BasicStdDev", func() {
		if suite.dbType == constants.SQLite {
			suite.T().Skipf("StdDev skipped for %s (SQLite does not support statistical functions)", suite.dbType)
		}

		type StdDevResult struct {
			Status   string  `bun:"status"`
			AvgViews float64 `bun:"avg_views"`
			StdDev   float64 `bun:"std_dev"`
		}

		var results []StdDevResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}, "avg_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StdDev(func(sb StdDevBuilder) {
					sb.Column("view_count")
				})
			}, "std_dev").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "StdDev should work on supported databases")
		suite.True(len(results) > 0, "Should have StdDev results")

		for _, result := range results {
			suite.True(result.AvgViews >= 0, "Average should be non-negative")
			suite.True(result.StdDev >= 0, "Standard deviation should be non-negative")
			suite.T().Logf("Status: %s, AvgViews: %.2f, StdDev: %.2f",
				result.Status, result.AvgViews, result.StdDev)
		}
	})

	suite.Run("CombinedStatisticalFunctions", func() {
		if suite.dbType == constants.SQLite {
			suite.T().Skipf("StdDev skipped for %s (SQLite does not support statistical functions)", suite.dbType)
		}

		type CombinedStats struct {
			Status   string  `bun:"status"`
			Count    int64   `bun:"count"`
			AvgViews float64 `bun:"avg_views"`
			MinViews int64   `bun:"min_views"`
			MaxViews int64   `bun:"max_views"`
			StdDev   float64 `bun:"std_dev"`
		}

		var results []CombinedStats

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CountAll()
			}, "count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}, "avg_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MinColumn("view_count")
			}, "min_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.MaxColumn("view_count")
			}, "max_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StdDev(func(sb StdDevBuilder) {
					sb.Column("view_count")
				})
			}, "std_dev").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined statistical functions should work")
		suite.True(len(results) > 0, "Should have combined stats results")

		for _, result := range results {
			suite.True(result.Count > 0, "Count should be positive")
			suite.True(result.AvgViews >= 0, "Average should be non-negative")
			suite.True(result.MinViews >= 0, "Min should be non-negative")
			suite.True(result.MaxViews >= result.MinViews, "Max should be >= Min")
			suite.True(result.StdDev >= 0, "StdDev should be non-negative")

			suite.T().Logf("Status: %s, Count: %d, Avg: %.2f, Min: %d, Max: %d, StdDev: %.2f",
				result.Status, result.Count, result.AvgViews, result.MinViews,
				result.MaxViews, result.StdDev)
		}
	})
}

// TestVariance tests the Variance aggregate function.
func (suite *AggregationFunctionsTestSuite) TestVariance() {
	suite.T().Logf("Testing Variance function for %s", suite.dbType)

	suite.Run("BasicVariance", func() {
		if suite.dbType == constants.SQLite {
			suite.T().Skipf("Variance skipped for %s (SQLite not supported)", suite.dbType)
		}

		type VarianceResult struct {
			Status   string  `bun:"status"`
			AvgViews float64 `bun:"avg_views"`
			Variance float64 `bun:"variance"`
		}

		var results []VarianceResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}, "avg_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Variance(func(vb VarianceBuilder) {
					vb.Column("view_count")
				})
			}, "variance").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "Variance should work on supported databases")
		suite.True(len(results) > 0, "Should have Variance results")

		for _, result := range results {
			suite.True(result.AvgViews >= 0, "Average should be non-negative")
			suite.True(result.Variance >= 0, "Variance should be non-negative")
			suite.T().Logf("Status: %s, AvgViews: %.2f, Variance: %.2f",
				result.Status, result.AvgViews, result.Variance)
		}
	})

	suite.Run("VarianceWithPopulationAndSample", func() {
		if suite.dbType == constants.SQLite {
			suite.T().Skipf("Variance with modes skipped for %s (SQLite not supported)", suite.dbType)
		}

		type VarianceModes struct {
			Status     string  `bun:"status"`
			VarPop     float64 `bun:"var_pop"`
			VarSamp    float64 `bun:"var_samp"`
			StdDevPop  float64 `bun:"stddev_pop"`
			StdDevSamp float64 `bun:"stddev_samp"`
		}

		var results []VarianceModes

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Variance(func(vb VarianceBuilder) {
					vb.Column("view_count").Population()
				})
			}, "var_pop").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Variance(func(vb VarianceBuilder) {
					vb.Column("view_count").Sample()
				})
			}, "var_samp").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StdDev(func(sb StdDevBuilder) {
					sb.Column("view_count").Population()
				})
			}, "stddev_pop").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StdDev(func(sb StdDevBuilder) {
					sb.Column("view_count").Sample()
				})
			}, "stddev_samp").
			GroupBy("status").
			OrderBy("status").
			Scan(suite.ctx, &results)

		suite.NoError(err, "Variance and StdDev with population/sample modes should work")
		suite.True(len(results) > 0, "Should have results")

		for _, result := range results {
			suite.True(result.VarPop >= 0, "Population variance should be non-negative")
			suite.True(result.VarSamp >= 0, "Sample variance should be non-negative")
			suite.True(result.StdDevPop >= 0, "Population stddev should be non-negative")
			suite.True(result.StdDevSamp >= 0, "Sample stddev should be non-negative")

			suite.T().Logf("Status: %s, VarPop: %.2f, VarSamp: %.2f, StdDevPop: %.2f, StdDevSamp: %.2f",
				result.Status, result.VarPop, result.VarSamp, result.StdDevPop, result.StdDevSamp)
		}
	})
}
