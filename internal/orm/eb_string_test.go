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
	suite.T().Logf("Testing Concat function for %s", suite.DbType)

	suite.Run("ConcatTitleAndStatus", func() {
		type ConcatResult struct {
			Title          string `bun:"title"`
			Status         string `bun:"status"`
			TitleAndStatus string `bun:"title_and_status"`
			MultiConcat    string `bun:"multi_concat"`
		}

		var concatResults []ConcatResult

		err := suite.Db.NewSelect().
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
			Scan(suite.Ctx, &concatResults)

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
	suite.T().Logf("Testing ConcatWithSep function for %s", suite.DbType)

	suite.Run("ConcatWithDashSeparator", func() {
		type ConcatWithSepResult struct {
			Id     string `bun:"id"`
			Joined string `bun:"joined"`
		}

		var concatResults []ConcatWithSepResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.ConcatWithSep(" - ", eb.Column("title"), eb.Column("status"))
			}, "joined").
			OrderBy("id").
			Limit(3).
			Scan(suite.Ctx, &concatResults)

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
	suite.T().Logf("Testing SubString function for %s", suite.DbType)

	suite.Run("ExtractSubstrings", func() {
		type SubstringResult struct {
			Title        string `bun:"title"`
			First5Chars  string `bun:"first5_chars"`
			Middle3Chars string `bun:"middle3_chars"`
		}

		var substringResults []SubstringResult

		err := suite.Db.NewSelect().
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
			Scan(suite.Ctx, &substringResults)

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
	suite.T().Logf("Testing Upper function for %s", suite.DbType)

	suite.Run("ConvertToUppercase", func() {
		type CaseResult struct {
			Title       string `bun:"title"`
			UpperTitle  string `bun:"upper_title"`
			Status      string `bun:"status"`
			UpperStatus string `bun:"upper_status"`
		}

		var caseResults []CaseResult

		err := suite.Db.NewSelect().
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
			Scan(suite.Ctx, &caseResults)

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
	suite.T().Logf("Testing Lower function for %s", suite.DbType)

	suite.Run("ConvertToLowercase", func() {
		type CaseResult struct {
			Title       string `bun:"title"`
			LowerTitle  string `bun:"lower_title"`
			Status      string `bun:"status"`
			LowerStatus string `bun:"lower_status"`
		}

		var caseResults []CaseResult

		err := suite.Db.NewSelect().
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
			Scan(suite.Ctx, &caseResults)

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
	suite.T().Logf("Testing Trim function for %s", suite.DbType)

	suite.Run("TrimWhitespace", func() {
		type TrimResult struct {
			Status        string `bun:"status"`
			TrimmedStatus string `bun:"trimmed_status"`
		}

		var trimResults []TrimResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Trim(eb.Column("status"))
			}, "trimmed_status").
			OrderBy("status").
			Limit(5).
			Scan(suite.Ctx, &trimResults)

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
	suite.T().Logf("Testing TrimLeft function for %s", suite.DbType)

	suite.Run("TrimLeadingWhitespace", func() {
		type TrimResult struct {
			Id          string `bun:"id"`
			Original    string `bun:"original"`
			LeftTrimmed string `bun:"left_trimmed"`
		}

		var trimResults []TrimResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat("   ", eb.Column("status"), "   ")
			}, "original").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.TrimLeft(eb.Concat("   ", eb.Column("status"), "   "))
			}, "left_trimmed").
			Limit(3).
			Scan(suite.Ctx, &trimResults)

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
	suite.T().Logf("Testing TrimRight function for %s", suite.DbType)

	suite.Run("TrimTrailingWhitespace", func() {
		type TrimResult struct {
			Id           string `bun:"id"`
			Original     string `bun:"original"`
			RightTrimmed string `bun:"right_trimmed"`
		}

		var trimResults []TrimResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Concat("   ", eb.Column("status"), "   ")
			}, "original").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.TrimRight(eb.Concat("   ", eb.Column("status"), "   "))
			}, "right_trimmed").
			Limit(3).
			Scan(suite.Ctx, &trimResults)

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
	suite.T().Logf("Testing Length function for %s", suite.DbType)

	suite.Run("CalculateStringLength", func() {
		type LengthResult struct {
			Title        string `bun:"title"`
			TitleLength  int64  `bun:"title_length"`
			Status       string `bun:"status"`
			StatusLength int64  `bun:"status_length"`
		}

		var lengthResults []LengthResult

		err := suite.Db.NewSelect().
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
			Scan(suite.Ctx, &lengthResults)

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
	suite.T().Logf("Testing CharLength function for %s", suite.DbType)

	suite.Run("CalculateCharacterLength", func() {
		type StringLengthResult struct {
			Title   string `bun:"title"`
			CharLen int64  `bun:"char_len"`
		}

		var lengthResults []StringLengthResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.CharLength(eb.Column("title"))
			}, "char_len").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &lengthResults)

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
	suite.T().Logf("Testing Position function for %s", suite.DbType)

	suite.Run("FindSubstringPosition", func() {
		type PositionResult struct {
			Title    string `bun:"title"`
			Position int64  `bun:"pos"`
		}

		var posResults []PositionResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Position("o", eb.Column("title"))
			}, "pos").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &posResults)

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
	suite.T().Logf("Testing Left function for %s", suite.DbType)

	suite.Run("ExtractLeftmostCharacters", func() {
		type LeftResult struct {
			Title    string `bun:"title"`
			LeftPart string `bun:"left_part"`
		}

		var leftResults []LeftResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Left(eb.Column("title"), 10)
			}, "left_part").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &leftResults)

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
	suite.T().Logf("Testing Right function for %s", suite.DbType)

	suite.Run("ExtractRightmostCharacters", func() {
		type RightResult struct {
			Title     string `bun:"title"`
			RightPart string `bun:"right_part"`
		}

		var rightResults []RightResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Right(eb.Column("title"), 5)
			}, "right_part").
			OrderBy("id").
			Limit(5).
			Scan(suite.Ctx, &rightResults)

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
	suite.T().Logf("Testing Repeat function for %s", suite.DbType)

	suite.Run("RepeatString", func() {
		type RepeatResult struct {
			Id       string `bun:"id"`
			Repeated string `bun:"repeated"`
		}

		var repeatResults []RepeatResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Repeat("*", 5)
			}, "repeated").
			Limit(3).
			Scan(suite.Ctx, &repeatResults)

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
	suite.T().Logf("Testing Replace function for %s", suite.DbType)

	suite.Run("ReplaceSubstring", func() {
		type ReplaceResult struct {
			Status         string `bun:"status"`
			ReplacedStatus string `bun:"replaced_status"`
		}

		var replaceResults []ReplaceResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("status").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Replace(eb.Column("status"), "'draft'", "'DRAFT'")
			}, "replaced_status").
			OrderBy("status").
			Limit(5).
			Scan(suite.Ctx, &replaceResults)

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
	suite.T().Logf("Testing Reverse function for %s", suite.DbType)

	suite.Run("ReverseString", func() {
		if suite.DbType == constants.DbSQLite {
			suite.T().Skipf("Reverse not supported on %s (framework limitation: no simulation provided)", suite.DbType)
		}

		type ReverseResult struct {
			Title    string `bun:"title"`
			Reversed string `bun:"reversed"`
		}

		var reverseResults []ReverseResult

		err := suite.Db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Reverse(eb.Column("title"))
			}, "reversed").
			Limit(3).
			Scan(suite.Ctx, &reverseResults)

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
	suite.T().Logf("Testing combined string functions for %s", suite.DbType)

	suite.Run("NestedStringFunctions", func() {
		type CombinedStringResult struct {
			Title       string `bun:"title"`
			UpperTitle  string `bun:"upper_title"`
			LowerTitle  string `bun:"lower_title"`
			CombinedStr string `bun:"combined_str"`
		}

		var combinedResults []CombinedStringResult

		err := suite.Db.NewSelect().
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
			Scan(suite.Ctx, &combinedResults)

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
