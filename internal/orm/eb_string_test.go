package orm

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

// StringFunctionsTestSuite tests string manipulation methods of ExprBuilder
// including Concat, ConcatWithSep, SubString, Upper, Lower, Trim, TrimLeft,
// TrimRight, Length, CharLength, Position, Left, Right, Repeat, Replace, and Reverse.
//
// This suite verifies cross-database compatibility for string functions across
// PostgreSQL, MySQL, and SQLite, handling database-specific limitations appropriately.
type StringFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestConcat tests the Concat function.
func (suite *StringFunctionsTestSuite) TestConcat() {
	suite.T().Logf("Testing Concat function for %s", suite.dbType)

	suite.Run("ConcatTitleAndStatus", func() {
		type ConcatResult struct {
			Title          string `bun:"title"`
			Status         string `bun:"status"`
			TitleAndStatus string `bun:"title_and_status"`
			MultiConcat    string `bun:"multi_concat"`
		}

		var concatResults []ConcatResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat(eb.Column("title"), "' - '", eb.Column("status"))
			}, "title_and_status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat("'[', ", eb.Column("status"), "'] '", eb.Column("title"))
			}, "multi_concat").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &concatResults)

		suite.NoError(err, "Concat function should work correctly")
		suite.True(len(concatResults) > 0, "Should have concat results")

		for _, result := range concatResults {
			suite.Contains(result.TitleAndStatus, result.Title, "Concat should include title")
			suite.Contains(result.TitleAndStatus, result.Status, "Concat should include status")
			suite.Contains(result.MultiConcat, result.Title, "Multi concat should include title")
			suite.Contains(result.MultiConcat, result.Status, "Multi concat should include status")
			suite.T().Logf("Post: %s | TitleAndStatus: %s | MultiConcat: %s",
				result.Title, result.TitleAndStatus, result.MultiConcat)
		}
	})
}

// TestConcatWithSep tests the ConcatWithSep function.
func (suite *StringFunctionsTestSuite) TestConcatWithSep() {
	suite.T().Logf("Testing ConcatWithSep function for %s", suite.dbType)

	suite.Run("ConcatWithDashSeparator", func() {
		type ConcatWithSepResult struct {
			Id     string `bun:"id"`
			Joined string `bun:"joined"`
		}

		var concatResults []ConcatWithSepResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ConcatWithSep(" - ", eb.Column("title"), eb.Column("status"))
			}, "joined").
			OrderBy("id").
			Limit(3).
			Scan(suite.ctx, &concatResults)

		suite.NoError(err, "ConcatWithSep should work correctly")
		suite.True(len(concatResults) > 0, "Should have concatenated results")

		for _, result := range concatResults {
			suite.Contains(result.Joined, " - ", "Should contain separator")
			suite.T().Logf("ID: %s, Joined: %s", result.Id, result.Joined)
		}
	})
}

// TestSubString tests the SubString function.
// SubString extracts a substring from a string starting at a 1-based position.
func (suite *StringFunctionsTestSuite) TestSubString() {
	suite.T().Logf("Testing SubString function for %s", suite.dbType)

	suite.Run("ExtractSubstrings", func() {
		type SubstringResult struct {
			Title        string `bun:"title"`
			First5Chars  string `bun:"first5_chars"`
			Middle3Chars string `bun:"middle3_chars"`
		}

		var substringResults []SubstringResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SubString(eb.Column("title"), 1, 5)
			}, "first5_chars").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.SubString(eb.Column("title"), 3, 3)
			}, "middle3_chars").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &substringResults)

		suite.NoError(err, "SubString function should work correctly")
		suite.True(len(substringResults) > 0, "Should have substring results")

		for _, result := range substringResults {
			if len(result.Title) >= 5 {
				suite.True(len(result.First5Chars) <= 5, "First5Chars should be at most 5 characters")
				suite.True(len(result.Middle3Chars) <= 3, "Middle3Chars should be at most 3 characters")
			}

			suite.T().Logf("Title: %s | First5: %s | Middle3: %s",
				result.Title, result.First5Chars, result.Middle3Chars)
		}
	})
}

// TestUpper tests the Upper function.
func (suite *StringFunctionsTestSuite) TestUpper() {
	suite.T().Logf("Testing Upper function for %s", suite.dbType)

	suite.Run("ConvertToUppercase", func() {
		type CaseResult struct {
			Title       string `bun:"title"`
			UpperTitle  string `bun:"upper_title"`
			Status      string `bun:"status"`
			UpperStatus string `bun:"upper_status"`
		}

		var caseResults []CaseResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("title"))
			}, "upper_title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("status"))
			}, "upper_status").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &caseResults)

		suite.NoError(err, "Upper function should work correctly")
		suite.True(len(caseResults) > 0, "Should have case conversion results")

		for _, result := range caseResults {
			suite.Equal(strings.ToUpper(result.Title), result.UpperTitle, "Upper should convert title to uppercase")
			suite.Equal(strings.ToUpper(result.Status), result.UpperStatus, "Upper should convert status to uppercase")
			suite.T().Logf("Title: %s → Upper: %s", result.Title, result.UpperTitle)
		}
	})
}

// TestLower tests the Lower function.
func (suite *StringFunctionsTestSuite) TestLower() {
	suite.T().Logf("Testing Lower function for %s", suite.dbType)

	suite.Run("ConvertToLowercase", func() {
		type CaseResult struct {
			Title       string `bun:"title"`
			LowerTitle  string `bun:"lower_title"`
			Status      string `bun:"status"`
			LowerStatus string `bun:"lower_status"`
		}

		var caseResults []CaseResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lower(eb.Column("title"))
			}, "lower_title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lower(eb.Column("status"))
			}, "lower_status").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &caseResults)

		suite.NoError(err, "Lower function should work correctly")
		suite.True(len(caseResults) > 0, "Should have case conversion results")

		for _, result := range caseResults {
			suite.Equal(strings.ToLower(result.Title), result.LowerTitle, "Lower should convert title to lowercase")
			suite.Equal(strings.ToLower(result.Status), result.LowerStatus, "Lower should convert status to lowercase")
			suite.T().Logf("Title: %s → Lower: %s", result.Title, result.LowerTitle)
		}
	})
}

// TestTrim tests the Trim function.
func (suite *StringFunctionsTestSuite) TestTrim() {
	suite.T().Logf("Testing Trim function for %s", suite.dbType)

	suite.Run("TrimWhitespace", func() {
		type TrimResult struct {
			Status        string `bun:"status"`
			TrimmedStatus string `bun:"trimmed_status"`
		}

		var trimResults []TrimResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Trim(eb.Column("status"))
			}, "trimmed_status").
			OrderBy("status").
			Limit(5).
			Scan(suite.ctx, &trimResults)

		suite.NoError(err, "Trim function should work correctly")
		suite.True(len(trimResults) > 0, "Should have trim results")

		for _, result := range trimResults {
			suite.NotEmpty(result.Status, "Status should not be empty")
			suite.NotEmpty(result.TrimmedStatus, "Trimmed status should not be empty")
			// Since status values don't have leading/trailing spaces, they should be equal
			suite.Equal(result.Status, result.TrimmedStatus, "Trim should preserve non-whitespace text")
			suite.T().Logf("Status: '%s' | Trimmed: '%s'", result.Status, result.TrimmedStatus)
		}
	})
}

// TestTrimLeft tests the TrimLeft function.
func (suite *StringFunctionsTestSuite) TestTrimLeft() {
	suite.T().Logf("Testing TrimLeft function for %s", suite.dbType)

	suite.Run("TrimLeadingWhitespace", func() {
		type TrimResult struct {
			Id          string `bun:"id"`
			Original    string `bun:"original"`
			LeftTrimmed string `bun:"left_trimmed"`
		}

		var trimResults []TrimResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat("   ", eb.Column("status"), "   ")
			}, "original").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.TrimLeft(eb.Concat("   ", eb.Column("status"), "   "))
			}, "left_trimmed").
			Limit(3).
			Scan(suite.ctx, &trimResults)

		suite.NoError(err, "TrimLeft should work correctly")
		suite.True(len(trimResults) > 0, "Should have trim results")

		for _, result := range trimResults {
			suite.Contains(result.Original, "   ", "Original should contain spaces")
			suite.NotEqual(result.Original, result.LeftTrimmed, "Left trimmed should differ from original")
			suite.T().Logf("ID: %s, Original: '%s', LeftTrim: '%s'",
				result.Id, result.Original, result.LeftTrimmed)
		}
	})
}

// TestTrimRight tests the TrimRight function.
func (suite *StringFunctionsTestSuite) TestTrimRight() {
	suite.T().Logf("Testing TrimRight function for %s", suite.dbType)

	suite.Run("TrimTrailingWhitespace", func() {
		type TrimResult struct {
			Id           string `bun:"id"`
			Original     string `bun:"original"`
			RightTrimmed string `bun:"right_trimmed"`
		}

		var trimResults []TrimResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat("   ", eb.Column("status"), "   ")
			}, "original").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.TrimRight(eb.Concat("   ", eb.Column("status"), "   "))
			}, "right_trimmed").
			Limit(3).
			Scan(suite.ctx, &trimResults)

		suite.NoError(err, "TrimRight should work correctly")
		suite.True(len(trimResults) > 0, "Should have trim results")

		for _, result := range trimResults {
			suite.Contains(result.Original, "   ", "Original should contain spaces")
			suite.NotEqual(result.Original, result.RightTrimmed, "Right trimmed should differ from original")
			suite.T().Logf("ID: %s, Original: '%s', RightTrim: '%s'",
				result.Id, result.Original, result.RightTrimmed)
		}
	})
}

// TestLength tests the Length function.
func (suite *StringFunctionsTestSuite) TestLength() {
	suite.T().Logf("Testing Length function for %s", suite.dbType)

	suite.Run("CalculateStringLength", func() {
		type LengthResult struct {
			Title        string `bun:"title"`
			TitleLength  int64  `bun:"title_length"`
			Status       string `bun:"status"`
			StatusLength int64  `bun:"status_length"`
		}

		var lengthResults []LengthResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Length(eb.Column("title"))
			}, "title_length").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Length(eb.Column("status"))
			}, "status_length").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &lengthResults)

		suite.NoError(err, "Length function should work correctly")
		suite.True(len(lengthResults) > 0, "Should have length results")

		for _, result := range lengthResults {
			suite.Equal(int64(len(result.Title)), result.TitleLength, "Length should match title byte length")
			suite.Equal(int64(len(result.Status)), result.StatusLength, "Length should match status byte length")
			suite.T().Logf("Title: %s (len=%d) | Status: %s (len=%d)",
				result.Title, result.TitleLength, result.Status, result.StatusLength)
		}
	})
}

// TestCharLength tests the CharLength function.
func (suite *StringFunctionsTestSuite) TestCharLength() {
	suite.T().Logf("Testing CharLength function for %s", suite.dbType)

	suite.Run("CalculateCharacterLength", func() {
		type StringLengthResult struct {
			Title   string `bun:"title"`
			CharLen int64  `bun:"char_len"`
		}

		var lengthResults []StringLengthResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CharLength(eb.Column("title"))
			}, "char_len").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &lengthResults)

		suite.NoError(err, "CharLength should work correctly")
		suite.True(len(lengthResults) > 0, "Should have character length results")

		for _, result := range lengthResults {
			suite.True(result.CharLen > 0, "Character length should be positive")
			suite.T().Logf("Title: %s, CharLen: %d", result.Title, result.CharLen)
		}
	})
}

// TestPosition tests the Position function.
// Position finds the position of a substring within a string (1-based, 0 if not found).
func (suite *StringFunctionsTestSuite) TestPosition() {
	suite.T().Logf("Testing Position function for %s", suite.dbType)

	suite.Run("FindSubstringPosition", func() {
		type PositionResult struct {
			Title    string `bun:"title"`
			Position int64  `bun:"pos"`
		}

		var posResults []PositionResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Position("o", eb.Column("title"))
			}, "pos").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &posResults)

		suite.NoError(err, "Position should work correctly")
		suite.True(len(posResults) > 0, "Should have position results")

		for _, result := range posResults {
			suite.True(result.Position >= 0, "Position should be non-negative (0 means not found)")
			suite.T().Logf("Title: %s, Position of 'o': %d", result.Title, result.Position)
		}
	})
}

// TestLeft tests the Left function.
func (suite *StringFunctionsTestSuite) TestLeft() {
	suite.T().Logf("Testing Left function for %s", suite.dbType)

	suite.Run("ExtractLeftmostCharacters", func() {
		type LeftResult struct {
			Title    string `bun:"title"`
			LeftPart string `bun:"left_part"`
		}

		var leftResults []LeftResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Left(eb.Column("title"), 10)
			}, "left_part").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &leftResults)

		suite.NoError(err, "Left function should work correctly")
		suite.True(len(leftResults) > 0, "Should have left part results")

		for _, result := range leftResults {
			suite.True(len(result.LeftPart) <= 10, "Left part should be at most 10 characters")
			suite.T().Logf("Title: %s, Left(10): %s", result.Title, result.LeftPart)
		}
	})
}

// TestRight tests the Right function.
func (suite *StringFunctionsTestSuite) TestRight() {
	suite.T().Logf("Testing Right function for %s", suite.dbType)

	suite.Run("ExtractRightmostCharacters", func() {
		type RightResult struct {
			Title     string `bun:"title"`
			RightPart string `bun:"right_part"`
		}

		var rightResults []RightResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Right(eb.Column("title"), 5)
			}, "right_part").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &rightResults)

		suite.NoError(err, "Right function should work correctly")
		suite.True(len(rightResults) > 0, "Should have right part results")

		for _, result := range rightResults {
			suite.True(len(result.RightPart) <= 5, "Right part should be at most 5 characters")
			suite.T().Logf("Title: %s, Right(5): %s", result.Title, result.RightPart)
		}
	})
}

// TestRepeat tests the Repeat function.
func (suite *StringFunctionsTestSuite) TestRepeat() {
	suite.T().Logf("Testing Repeat function for %s", suite.dbType)

	suite.Run("RepeatString", func() {
		type RepeatResult struct {
			Id       string `bun:"id"`
			Repeated string `bun:"repeated"`
		}

		var repeatResults []RepeatResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Repeat("*", 5)
			}, "repeated").
			Limit(3).
			Scan(suite.ctx, &repeatResults)

		suite.NoError(err, "Repeat should work correctly")
		suite.True(len(repeatResults) > 0, "Should have repeat results")

		for _, result := range repeatResults {
			suite.Equal("*****", result.Repeated, "Should repeat '*' 5 times")
			suite.T().Logf("ID: %s, Repeated: %s", result.Id, result.Repeated)
		}
	})
}

// TestReplace tests the Replace function.
func (suite *StringFunctionsTestSuite) TestReplace() {
	suite.T().Logf("Testing Replace function for %s", suite.dbType)

	suite.Run("ReplaceSubstring", func() {
		type ReplaceResult struct {
			Status         string `bun:"status"`
			ReplacedStatus string `bun:"replaced_status"`
		}

		var replaceResults []ReplaceResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Replace(eb.Column("status"), "'draft'", "'DRAFT'")
			}, "replaced_status").
			OrderBy("status").
			Limit(5).
			Scan(suite.ctx, &replaceResults)

		suite.NoError(err, "Replace function should work correctly")
		suite.True(len(replaceResults) > 0, "Should have replace results")

		for _, result := range replaceResults {
			suite.NotEmpty(result.Status, "Status should not be empty")
			suite.NotEmpty(result.ReplacedStatus, "Replaced status should not be empty")
			suite.T().Logf("Original: %s | Replaced: %s", result.Status, result.ReplacedStatus)
		}
	})
}

// TestReverse tests the Reverse function.
// Reverse reverses a string (not supported on SQLite).
func (suite *StringFunctionsTestSuite) TestReverse() {
	suite.T().Logf("Testing Reverse function for %s", suite.dbType)

	suite.Run("ReverseString", func() {
		if suite.dbType == constants.DbSQLite {
			suite.T().Skipf("Reverse not supported on %s (framework limitation: no simulation provided)", suite.dbType)
		}

		type ReverseResult struct {
			Title    string `bun:"title"`
			Reversed string `bun:"reversed"`
		}

		var reverseResults []ReverseResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Reverse(eb.Column("title"))
			}, "reversed").
			Limit(3).
			Scan(suite.ctx, &reverseResults)

		suite.NoError(err, "Reverse should work on supported databases")
		suite.True(len(reverseResults) > 0, "Should have reverse results")

		for _, result := range reverseResults {
			suite.NotEmpty(result.Reversed, "Reversed string should not be empty")
			suite.T().Logf("Title: %s, Reversed: %s", result.Title, result.Reversed)
		}
	})
}

// TestCombinedStringFunctions tests using multiple string functions together.
// This verifies that string functions can be nested and combined.
func (suite *StringFunctionsTestSuite) TestCombinedStringFunctions() {
	suite.T().Logf("Testing combined string functions for %s", suite.dbType)

	suite.Run("NestedStringFunctions", func() {
		type CombinedStringResult struct {
			Title       string `bun:"title"`
			UpperTitle  string `bun:"upper_title"`
			LowerTitle  string `bun:"lower_title"`
			CombinedStr string `bun:"combined_str"`
		}

		var combinedResults []CombinedStringResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Upper(eb.Column("title"))
			}, "upper_title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Lower(eb.Column("title"))
			}, "lower_title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat(
					eb.Upper(eb.SubString(eb.Column("title"), 1, 3)),
					"...",
					eb.Lower(eb.SubString(eb.Column("title"), 1, 3)),
				)
			}, "combined_str").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &combinedResults)

		suite.NoError(err, "Combined string functions should work correctly")
		suite.True(len(combinedResults) > 0, "Should have combined string results")

		for _, result := range combinedResults {
			suite.NotEmpty(result.UpperTitle, "Upper title should not be empty")
			suite.NotEmpty(result.LowerTitle, "Lower title should not be empty")
			suite.NotEmpty(result.CombinedStr, "Combined string should not be empty")
			suite.T().Logf("Original: %s | Upper: %s | Lower: %s | Combined: %s",
				result.Title, result.UpperTitle, result.LowerTitle, result.CombinedStr)
		}
	})
}

// TestContains tests the Contains function (case-sensitive substring check).
func (suite *StringFunctionsTestSuite) TestContains() {
	suite.T().Logf("Testing Contains function for %s", suite.dbType)

	suite.Run("ContainsSubstring", func() {
		type ContainsResult struct {
			Id            string `bun:"id"`
			Title         string `bun:"title"`
			ContainsPost  bool   `bun:"contains_post"`
			ContainsGuide bool   `bun:"contains_guide"`
		}

		var containsResults []ContainsResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Contains(eb.Column("title"), "'Post'")
			}, "contains_post").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Contains(eb.Column("title"), "'Guide'")
			}, "contains_guide").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &containsResults)

		suite.NoError(err, "Contains function should work correctly")
		suite.True(len(containsResults) > 0, "Should have contains results")

		for _, result := range containsResults {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")

			// Verify case-sensitive matching
			if strings.Contains(result.Title, "Post") {
				suite.True(result.ContainsPost, "Should contain 'Post' when present")
			} else {
				suite.False(result.ContainsPost, "Should not contain 'Post' when absent")
			}

			if strings.Contains(result.Title, "Guide") {
				suite.True(result.ContainsGuide, "Should contain 'Guide' when present")
			} else {
				suite.False(result.ContainsGuide, "Should not contain 'Guide' when absent")
			}

			suite.T().Logf("ID: %s, Title: %s, ContainsPost: %v, ContainsGuide: %v",
				result.Id, result.Title, result.ContainsPost, result.ContainsGuide)
		}
	})
}

// TestStartsWith tests the StartsWith function (case-sensitive prefix check).
func (suite *StringFunctionsTestSuite) TestStartsWith() {
	suite.T().Logf("Testing StartsWith function for %s", suite.dbType)

	suite.Run("StartsWithPrefix", func() {
		type StartsWithResult struct {
			Id          string `bun:"id"`
			Title       string `bun:"title"`
			StartsWithG bool   `bun:"starts_with_g"`
			StartsWithP bool   `bun:"starts_with_p"`
		}

		var startsWithResults []StartsWithResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StartsWith(eb.Column("title"), "'G'")
			}, "starts_with_g").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StartsWith(eb.Column("title"), "'P'")
			}, "starts_with_p").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &startsWithResults)

		suite.NoError(err, "StartsWith function should work correctly")
		suite.True(len(startsWithResults) > 0, "Should have startsWith results")

		for _, result := range startsWithResults {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")

			// Verify case-sensitive prefix matching
			if strings.HasPrefix(result.Title, "G") {
				suite.True(result.StartsWithG, "Should start with 'G' when present")
			} else {
				suite.False(result.StartsWithG, "Should not start with 'G' when absent")
			}

			if strings.HasPrefix(result.Title, "P") {
				suite.True(result.StartsWithP, "Should start with 'P' when present")
			} else {
				suite.False(result.StartsWithP, "Should not start with 'P' when absent")
			}

			suite.T().Logf("ID: %s, Title: %s, StartsWithG: %v, StartsWithP: %v",
				result.Id, result.Title, result.StartsWithG, result.StartsWithP)
		}
	})
}

// TestEndsWith tests the EndsWith function (case-sensitive suffix check).
func (suite *StringFunctionsTestSuite) TestEndsWith() {
	suite.T().Logf("Testing EndsWith function for %s", suite.dbType)

	suite.Run("EndsWithSuffix", func() {
		type EndsWithResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			EndsWithE bool   `bun:"ends_with_e"`
			EndsWithT bool   `bun:"ends_with_t"`
		}

		var endsWithResults []EndsWithResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.EndsWith(eb.Column("title"), "'e'")
			}, "ends_with_e").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.EndsWith(eb.Column("title"), "'t'")
			}, "ends_with_t").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &endsWithResults)

		suite.NoError(err, "EndsWith function should work correctly")
		suite.True(len(endsWithResults) > 0, "Should have endsWith results")

		for _, result := range endsWithResults {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")

			// Verify case-sensitive suffix matching
			if strings.HasSuffix(result.Title, "e") {
				suite.True(result.EndsWithE, "Should end with 'e' when present")
			} else {
				suite.False(result.EndsWithE, "Should not end with 'e' when absent")
			}

			if strings.HasSuffix(result.Title, "t") {
				suite.True(result.EndsWithT, "Should end with 't' when present")
			} else {
				suite.False(result.EndsWithT, "Should not end with 't' when absent")
			}

			suite.T().Logf("ID: %s, Title: %s, EndsWithE: %v, EndsWithT: %v",
				result.Id, result.Title, result.EndsWithE, result.EndsWithT)
		}
	})
}

// TestContainsIgnoreCase tests the ContainsIgnoreCase function (case-insensitive substring check).
func (suite *StringFunctionsTestSuite) TestContainsIgnoreCase() {
	suite.T().Logf("Testing ContainsIgnoreCase function for %s", suite.dbType)

	suite.Run("ContainsSubstringIgnoreCase", func() {
		type ContainsIgnoreCaseResult struct {
			Id            string `bun:"id"`
			Title         string `bun:"title"`
			ContainsPost  bool   `bun:"contains_post"`
			ContainsGuide bool   `bun:"contains_guide"`
		}

		var containsResults []ContainsIgnoreCaseResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ContainsIgnoreCase(eb.Column("title"), "'post'")
			}, "contains_post").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ContainsIgnoreCase(eb.Column("title"), "'GUIDE'")
			}, "contains_guide").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &containsResults)

		suite.NoError(err, "ContainsIgnoreCase function should work correctly")
		suite.True(len(containsResults) > 0, "Should have contains ignore case results")

		for _, result := range containsResults {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")

			// Verify case-insensitive matching
			lowerTitle := strings.ToLower(result.Title)
			if strings.Contains(lowerTitle, "post") {
				suite.True(result.ContainsPost, "Should contain 'post' (case-insensitive) when present")
			} else {
				suite.False(result.ContainsPost, "Should not contain 'post' (case-insensitive) when absent")
			}

			if strings.Contains(lowerTitle, "guide") {
				suite.True(result.ContainsGuide, "Should contain 'guide' (case-insensitive) when present")
			} else {
				suite.False(result.ContainsGuide, "Should not contain 'guide' (case-insensitive) when absent")
			}

			suite.T().Logf("ID: %s, Title: %s, ContainsPost: %v, ContainsGuide: %v",
				result.Id, result.Title, result.ContainsPost, result.ContainsGuide)
		}
	})
}

// TestStartsWithIgnoreCase tests the StartsWithIgnoreCase function (case-insensitive prefix check).
func (suite *StringFunctionsTestSuite) TestStartsWithIgnoreCase() {
	suite.T().Logf("Testing StartsWithIgnoreCase function for %s", suite.dbType)

	suite.Run("StartsWithPrefixIgnoreCase", func() {
		type StartsWithIgnoreCaseResult struct {
			Id          string `bun:"id"`
			Title       string `bun:"title"`
			StartsWithG bool   `bun:"starts_with_g"`
			StartsWithP bool   `bun:"starts_with_p"`
		}

		var startsWithResults []StartsWithIgnoreCaseResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StartsWithIgnoreCase(eb.Column("title"), "'g'")
			}, "starts_with_g").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.StartsWithIgnoreCase(eb.Column("title"), "'p'")
			}, "starts_with_p").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &startsWithResults)

		suite.NoError(err, "StartsWithIgnoreCase function should work correctly")
		suite.True(len(startsWithResults) > 0, "Should have startsWithIgnoreCase results")

		for _, result := range startsWithResults {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")

			// Verify case-insensitive prefix matching
			lowerTitle := strings.ToLower(result.Title)
			if strings.HasPrefix(lowerTitle, "g") {
				suite.True(result.StartsWithG, "Should start with 'g' (case-insensitive) when present")
			} else {
				suite.False(result.StartsWithG, "Should not start with 'g' (case-insensitive) when absent")
			}

			if strings.HasPrefix(lowerTitle, "p") {
				suite.True(result.StartsWithP, "Should start with 'p' (case-insensitive) when present")
			} else {
				suite.False(result.StartsWithP, "Should not start with 'p' (case-insensitive) when absent")
			}

			suite.T().Logf("ID: %s, Title: %s, StartsWithG: %v, StartsWithP: %v",
				result.Id, result.Title, result.StartsWithG, result.StartsWithP)
		}
	})
}

// TestEndsWithIgnoreCase tests the EndsWithIgnoreCase function (case-insensitive suffix check).
func (suite *StringFunctionsTestSuite) TestEndsWithIgnoreCase() {
	suite.T().Logf("Testing EndsWithIgnoreCase function for %s", suite.dbType)

	suite.Run("EndsWithSuffixIgnoreCase", func() {
		type EndsWithIgnoreCaseResult struct {
			Id        string `bun:"id"`
			Title     string `bun:"title"`
			EndsWithE bool   `bun:"ends_with_e"`
			EndsWithT bool   `bun:"ends_with_t"`
		}

		var endsWithResults []EndsWithIgnoreCaseResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id", "title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.EndsWithIgnoreCase(eb.Column("title"), "'E'")
			}, "ends_with_e").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.EndsWithIgnoreCase(eb.Column("title"), "'T'")
			}, "ends_with_t").
			OrderBy("id").
			Limit(5).
			Scan(suite.ctx, &endsWithResults)

		suite.NoError(err, "EndsWithIgnoreCase function should work correctly")
		suite.True(len(endsWithResults) > 0, "Should have endsWithIgnoreCase results")

		for _, result := range endsWithResults {
			suite.NotEmpty(result.Id, "ID should not be empty")
			suite.NotEmpty(result.Title, "Title should not be empty")

			// Verify case-insensitive suffix matching
			lowerTitle := strings.ToLower(result.Title)
			if strings.HasSuffix(lowerTitle, "e") {
				suite.True(result.EndsWithE, "Should end with 'e' (case-insensitive) when present")
			} else {
				suite.False(result.EndsWithE, "Should not end with 'e' (case-insensitive) when absent")
			}

			if strings.HasSuffix(lowerTitle, "t") {
				suite.True(result.EndsWithT, "Should end with 't' (case-insensitive) when present")
			} else {
				suite.False(result.EndsWithT, "Should not end with 't' (case-insensitive) when absent")
			}

			suite.T().Logf("ID: %s, Title: %s, EndsWithE: %v, EndsWithT: %v",
				result.Id, result.Title, result.EndsWithE, result.EndsWithT)
		}
	})
}
