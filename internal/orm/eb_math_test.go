package orm

// MathFunctionsTestSuite tests mathematical function methods of ExprBuilder
// including basic math operations, power/root functions, logarithmic functions,
// trigonometric functions, constants, and comparison functions.
type MathFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestAbs tests the Abs function.
func (suite *MathFunctionsTestSuite) TestAbs() {
	suite.T().Logf("Testing Abs function for %s", suite.dbType)

	suite.Run("AbsoluteValue", func() {
		// First get average view count
		var avgViewCount float64

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.AvgColumn("view_count")
			}).
			Scan(suite.ctx, &avgViewCount)
		suite.NoError(err, "Average calculation should work")

		type AbsResult struct {
			Title      string `bun:"title"`
			ViewCount  int64  `bun:"view_count"`
			Difference int64  `bun:"difference"`
			AbsDiff    int64  `bun:"abs_diff"`
		}

		var results []AbsResult

		err = suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? - ?", eb.Column("view_count"), int(avgViewCount))
			}, "difference").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Abs(eb.Expr("? - ?", eb.Column("view_count"), int(avgViewCount)))
			}, "abs_diff").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Abs should work correctly")
		suite.True(len(results) > 0, "Should have abs results")

		for _, result := range results {
			suite.True(result.AbsDiff >= 0, "Absolute value should always be non-negative")

			expectedAbs := result.Difference
			if expectedAbs < 0 {
				expectedAbs = -expectedAbs
			}

			suite.Equal(expectedAbs, result.AbsDiff, "Abs should return absolute value")
			suite.T().Logf("Post %s: ViewCount=%d, Diff=%d, AbsDiff=%d",
				result.Title, result.ViewCount, result.Difference, result.AbsDiff)
		}
	})
}

// TestCeil tests the Ceil function.
func (suite *MathFunctionsTestSuite) TestCeil() {
	suite.T().Logf("Testing Ceil function for %s", suite.dbType)

	suite.Run("CeilDecimalValues", func() {
		type CeilResult struct {
			ViewCount    int64   `bun:"view_count"`
			DecimalValue float64 `bun:"decimal_value"`
			CeiledValue  float64 `bun:"ceiled_value"`
		}

		var results []CeilResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? / 3.0", eb.Column("view_count"))
			}, "decimal_value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Ceil(eb.Expr("? / 3.0", eb.Column("view_count")))
			}, "ceiled_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Ceil should work correctly")
		suite.True(len(results) > 0, "Should have ceil results")

		for _, result := range results {
			suite.True(result.CeiledValue >= result.DecimalValue, "Ceil should be >= original value")
			suite.T().Logf("ViewCount: %d, Decimal: %.2f, Ceiled: %.0f",
				result.ViewCount, result.DecimalValue, result.CeiledValue)
		}
	})
}

// TestFloor tests the Floor function.
func (suite *MathFunctionsTestSuite) TestFloor() {
	suite.T().Logf("Testing Floor function for %s", suite.dbType)

	suite.Run("FloorDecimalValues", func() {
		type FloorResult struct {
			ViewCount    int64   `bun:"view_count"`
			DecimalValue float64 `bun:"decimal_value"`
			FlooredValue float64 `bun:"floored_value"`
		}

		var results []FloorResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? / 3.0", eb.Column("view_count"))
			}, "decimal_value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Floor(eb.Expr("? / 3.0", eb.Column("view_count")))
			}, "floored_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Floor should work correctly")
		suite.True(len(results) > 0, "Should have floor results")

		for _, result := range results {
			suite.True(result.FlooredValue <= result.DecimalValue, "Floor should be <= original value")
			suite.T().Logf("ViewCount: %d, Decimal: %.2f, Floored: %.0f",
				result.ViewCount, result.DecimalValue, result.FlooredValue)
		}
	})
}

// TestRound tests the Round function.
func (suite *MathFunctionsTestSuite) TestRound() {
	suite.T().Logf("Testing Round function for %s", suite.dbType)

	suite.Run("RoundWithDifferentPrecisions", func() {
		type RoundResult struct {
			ViewCount     int64   `bun:"view_count"`
			RoundedNoPrec float64 `bun:"rounded_no_prec"`
			RoundedPrec1  float64 `bun:"rounded_prec1"`
			RoundedPrec2  float64 `bun:"rounded_prec2"`
		}

		var results []RoundResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Round(eb.Expr("? * 1.456", eb.Column("view_count")))
			}, "rounded_no_prec").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Round(eb.Expr("? * 1.456", eb.Column("view_count")), 1)
			}, "rounded_prec1").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Round(eb.Expr("? * 1.456", eb.Column("view_count")), 2)
			}, "rounded_prec2").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Round should work correctly")
		suite.True(len(results) > 0, "Should have round results")

		for _, result := range results {
			suite.True(result.RoundedNoPrec >= 0, "Rounded value should be non-negative")
			suite.True(result.RoundedPrec1 >= 0, "Rounded value with precision 1 should be non-negative")
			suite.True(result.RoundedPrec2 >= 0, "Rounded value with precision 2 should be non-negative")
			suite.T().Logf("ViewCount: %d, Rounded: %.0f, Prec1: %.1f, Prec2: %.2f",
				result.ViewCount, result.RoundedNoPrec, result.RoundedPrec1, result.RoundedPrec2)
		}
	})
}

// TestTrunc tests the Trunc function.
func (suite *MathFunctionsTestSuite) TestTrunc() {
	suite.T().Logf("Testing Trunc function for %s", suite.dbType)

	suite.Run("TruncateDecimalValues", func() {
		type TruncResult struct {
			ViewCount  int64   `bun:"view_count"`
			Divided    float64 `bun:"divided"`
			TruncValue float64 `bun:"trunc_value"`
		}

		var results []TruncResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? / 3.0", eb.Column("view_count"))
			}, "divided").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Trunc(eb.Expr("? / 3.0", eb.Column("view_count")), 2)
			}, "trunc_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Trunc should work correctly")
		suite.True(len(results) > 0, "Should have trunc results")

		for _, result := range results {
			suite.True(result.TruncValue >= 0, "Truncated value should be non-negative")
			suite.T().Logf("ViewCount: %d, Divided: %.4f, Truncated: %.2f",
				result.ViewCount, result.Divided, result.TruncValue)
		}
	})
}

// TestPower tests the Power function.
func (suite *MathFunctionsTestSuite) TestPower() {
	suite.T().Logf("Testing Power function for %s", suite.dbType)

	suite.Run("PowerOfNumbers", func() {
		type PowerResult struct {
			ViewCount int64   `bun:"view_count"`
			Squared   float64 `bun:"squared"`
			Cubed     float64 `bun:"cubed"`
		}

		var results []PowerResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Power(eb.Column("view_count"), 2)
			}, "squared").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Power(eb.Column("view_count"), 3)
			}, "cubed").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Power should work correctly")
		suite.True(len(results) > 0, "Should have power results")

		for _, result := range results {
			suite.True(result.Squared >= 0, "Squared value should be non-negative")
			suite.True(result.Cubed >= 0, "Cubed value should be non-negative")
			suite.T().Logf("ViewCount: %d, Squared: %.0f, Cubed: %.0f",
				result.ViewCount, result.Squared, result.Cubed)
		}
	})
}

// TestSqrt tests the Sqrt function.
func (suite *MathFunctionsTestSuite) TestSqrt() {
	suite.T().Logf("Testing Sqrt function for %s", suite.dbType)

	suite.Run("SquareRootOfNumbers", func() {
		type SqrtResult struct {
			ViewCount int64   `bun:"view_count"`
			SqrtValue float64 `bun:"sqrt_value"`
		}

		var results []SqrtResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sqrt(eb.Column("view_count"))
			}, "sqrt_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Sqrt should work correctly")
		suite.True(len(results) > 0, "Should have sqrt results")

		for _, result := range results {
			suite.True(result.SqrtValue >= 0, "Square root should be non-negative")
			suite.T().Logf("ViewCount: %d, Sqrt: %.2f", result.ViewCount, result.SqrtValue)
		}
	})
}

// TestExp tests the Exp function.
func (suite *MathFunctionsTestSuite) TestExp() {
	suite.T().Logf("Testing Exp function for %s", suite.dbType)

	suite.Run("ExponentialFunction", func() {
		type ExpResult struct {
			ViewCount int64   `bun:"view_count"`
			ExpValue  float64 `bun:"exp_value"`
		}

		var results []ExpResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Exp(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "exp_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Exp should work correctly")
		suite.True(len(results) > 0, "Should have exp results")

		for _, result := range results {
			suite.True(result.ExpValue > 0, "Exponential should always be positive")
			suite.T().Logf("ViewCount: %d, Exp(vc/100): %.4f",
				result.ViewCount, result.ExpValue)
		}
	})
}

// TestLn tests the Ln function.
func (suite *MathFunctionsTestSuite) TestLn() {
	suite.T().Logf("Testing Ln function for %s", suite.dbType)

	suite.Run("NaturalLogarithm", func() {
		type LnResult struct {
			ViewCount int64   `bun:"view_count"`
			LnValue   float64 `bun:"ln_value"`
		}

		var results []LnResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Ln(eb.Column("view_count"))
			}, "ln_value").
			Where(func(cb ConditionBuilder) {
				cb.GreaterThan("view_count", 0)
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Ln should work correctly")
		suite.True(len(results) > 0, "Should have ln results")

		for _, result := range results {
			suite.True(result.LnValue > 0, "Natural log should be positive for view_count > 1")
			suite.T().Logf("ViewCount: %d, Ln: %.4f", result.ViewCount, result.LnValue)
		}
	})
}

// TestLog tests the Log function.
func (suite *MathFunctionsTestSuite) TestLog() {
	suite.T().Logf("Testing Log function for %s", suite.dbType)

	suite.Run("LogarithmBaseTen", func() {
		type LogResult struct {
			ViewCount int64   `bun:"view_count"`
			LogValue  float64 `bun:"log_value"`
		}

		var results []LogResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Log(10, eb.Column("view_count"))
			}, "log_value").
			Where(func(cb ConditionBuilder) {
				cb.GreaterThan("view_count", 0)
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Log should work correctly")
		suite.True(len(results) > 0, "Should have log results")

		for _, result := range results {
			suite.True(result.LogValue > 0, "Log base 10 should be positive for view_count > 1")
			suite.T().Logf("ViewCount: %d, Log10: %.4f", result.ViewCount, result.LogValue)
		}
	})
}

// TestSin tests the Sin function.
func (suite *MathFunctionsTestSuite) TestSin() {
	suite.T().Logf("Testing Sin function for %s", suite.dbType)

	suite.Run("SineTrigonometric", func() {
		type SinResult struct {
			ViewCount int64   `bun:"view_count"`
			SinValue  float64 `bun:"sin_value"`
		}

		var results []SinResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sin(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "sin_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Sin should work correctly")
		suite.True(len(results) > 0, "Should have sin results")

		for _, result := range results {
			suite.True(result.SinValue >= -1.0 && result.SinValue <= 1.0, "Sin value should be between -1 and 1")
			suite.T().Logf("ViewCount: %d, Sin(vc/100): %.4f",
				result.ViewCount, result.SinValue)
		}
	})
}

// TestCos tests the Cos function.
func (suite *MathFunctionsTestSuite) TestCos() {
	suite.T().Logf("Testing Cos function for %s", suite.dbType)

	suite.Run("CosineTrigonometric", func() {
		type CosResult struct {
			ViewCount int64   `bun:"view_count"`
			CosValue  float64 `bun:"cos_value"`
		}

		var results []CosResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Cos(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "cos_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Cos should work correctly")
		suite.True(len(results) > 0, "Should have cos results")

		for _, result := range results {
			suite.True(result.CosValue >= -1.0 && result.CosValue <= 1.0, "Cos value should be between -1 and 1")
			suite.T().Logf("ViewCount: %d, Cos(vc/100): %.4f",
				result.ViewCount, result.CosValue)
		}
	})
}

// TestTan tests the Tan function.
func (suite *MathFunctionsTestSuite) TestTan() {
	suite.T().Logf("Testing Tan function for %s", suite.dbType)

	suite.Run("TangentTrigonometric", func() {
		type TanResult struct {
			ViewCount int64   `bun:"view_count"`
			TanValue  float64 `bun:"tan_value"`
		}

		var results []TanResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Tan(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "tan_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Tan should work correctly")
		suite.True(len(results) > 0, "Should have tan results")

		for _, result := range results {
			suite.T().Logf("ViewCount: %d, Tan(vc/100): %.4f",
				result.ViewCount, result.TanValue)
		}
	})
}

// TestAsin tests the Asin function.
func (suite *MathFunctionsTestSuite) TestAsin() {
	suite.T().Logf("Testing Asin function for %s", suite.dbType)

	suite.Run("ArcsineInverse", func() {
		type AsinResult struct {
			Value     float64 `bun:"value"`
			AsinValue float64 `bun:"asin_value"`
		}

		var results []AsinResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? / 200.0", eb.Column("view_count"))
			}, "value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Asin(eb.Expr("? / 200.0", eb.Column("view_count")))
			}, "asin_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Asin should work correctly")
		suite.True(len(results) > 0, "Should have asin results")

		for _, result := range results {
			suite.True(result.Value >= -1.0 && result.Value <= 1.0, "Value should be in valid range for asin")
			suite.T().Logf("Value: %.4f, Asin: %.4f", result.Value, result.AsinValue)
		}
	})
}

// TestAcos tests the Acos function.
func (suite *MathFunctionsTestSuite) TestAcos() {
	suite.T().Logf("Testing Acos function for %s", suite.dbType)

	suite.Run("ArccosineInverse", func() {
		type AcosResult struct {
			Value     float64 `bun:"value"`
			AcosValue float64 `bun:"acos_value"`
		}

		var results []AcosResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? / 200.0", eb.Column("view_count"))
			}, "value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Acos(eb.Expr("? / 200.0", eb.Column("view_count")))
			}, "acos_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Acos should work correctly")
		suite.True(len(results) > 0, "Should have acos results")

		for _, result := range results {
			suite.True(result.Value >= -1.0 && result.Value <= 1.0, "Value should be in valid range for acos")
			suite.T().Logf("Value: %.4f, Acos: %.4f", result.Value, result.AcosValue)
		}
	})
}

// TestAtan tests the Atan function.
func (suite *MathFunctionsTestSuite) TestAtan() {
	suite.T().Logf("Testing Atan function for %s", suite.dbType)

	suite.Run("ArctangentInverse", func() {
		type AtanResult struct {
			Value     float64 `bun:"value"`
			AtanValue float64 `bun:"atan_value"`
		}

		var results []AtanResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Expr("? / 200.0", eb.Column("view_count"))
			}, "value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Atan(eb.Expr("? / 200.0", eb.Column("view_count")))
			}, "atan_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Atan should work correctly")
		suite.True(len(results) > 0, "Should have atan results")

		for _, result := range results {
			suite.T().Logf("Value: %.4f, Atan: %.4f", result.Value, result.AtanValue)
		}
	})
}

// TestPi tests the Pi function.
func (suite *MathFunctionsTestSuite) TestPi() {
	suite.T().Logf("Testing Pi function for %s", suite.dbType)

	suite.Run("PiConstant", func() {
		type PiResult struct {
			PiValue       float64 `bun:"pi_value"`
			CircleArea    float64 `bun:"circle_area"`
			Circumference float64 `bun:"circumference"`
		}

		var results []PiResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Pi()
			}, "pi_value").
			SelectExpr(func(eb ExprBuilder) any {
				// Area = π * r²
				return eb.Expr("? * ? * ?",
					eb.Pi(),
					eb.Column("view_count"),
					eb.Column("view_count"))
			}, "circle_area").
			SelectExpr(func(eb ExprBuilder) any {
				// Circumference = 2 * π * r
				return eb.Expr("2 * ? * ?", eb.Pi(), eb.Column("view_count"))
			}, "circumference").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Pi should work correctly")
		suite.True(len(results) > 0, "Should have pi results")

		for _, result := range results {
			suite.InDelta(3.14159, result.PiValue, 0.001, "Pi value should be approximately 3.14159")
			suite.True(result.CircleArea > 0, "Circle area should be positive")
			suite.True(result.Circumference > 0, "Circumference should be positive")
			suite.T().Logf("Pi: %.5f, Area: %.2f, Circumference: %.2f",
				result.PiValue, result.CircleArea, result.Circumference)
		}
	})
}

// TestRandom tests the Random function.
func (suite *MathFunctionsTestSuite) TestRandom() {
	suite.T().Logf("Testing Random function for %s", suite.dbType)

	suite.Run("RandomNumberGeneration", func() {
		type RandomResult struct {
			ID        string  `bun:"id"`
			RandomVal float64 `bun:"random_val"`
		}

		var results []RandomResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Random()
			}, "random_val").
			OrderBy("id").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Random should work correctly")
		suite.True(len(results) > 0, "Should have random results")

		for _, result := range results {
			suite.True(result.RandomVal >= 0 && result.RandomVal < 1, "Random should be in [0, 1)")
			suite.T().Logf("ID: %s, Random: %.6f", result.ID, result.RandomVal)
		}
	})
}

// TestSign tests the Sign function.
func (suite *MathFunctionsTestSuite) TestSign() {
	suite.T().Logf("Testing Sign function for %s", suite.dbType)

	suite.Run("SignFunction", func() {
		type SignResult struct {
			ViewCount int64   `bun:"view_count"`
			SignValue float64 `bun:"sign_value"`
		}

		var results []SignResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sign(eb.Expr("? - 50", eb.Column("view_count")))
			}, "sign_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Sign should work correctly")
		suite.True(len(results) > 0, "Should have sign results")

		for _, result := range results {
			suite.Contains([]float64{-1.0, 0.0, 1.0}, result.SignValue, "Sign should be -1, 0, or 1")
			suite.T().Logf("ViewCount: %d, Sign(vc-50): %.0f", result.ViewCount, result.SignValue)
		}
	})
}

// TestMod tests the Mod function.
func (suite *MathFunctionsTestSuite) TestMod() {
	suite.T().Logf("Testing Mod function for %s", suite.dbType)

	suite.Run("ModuloOperation", func() {
		type ModResult struct {
			ViewCount int64 `bun:"view_count"`
			Mod5      int64 `bun:"mod5"`
			Mod10     int64 `bun:"mod10"`
		}

		var results []ModResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Mod(eb.Column("view_count"), 5)
			}, "mod5").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Mod(eb.Column("view_count"), 10)
			}, "mod10").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Mod should work correctly")
		suite.True(len(results) > 0, "Should have mod results")

		for _, result := range results {
			suite.True(result.Mod5 >= 0 && result.Mod5 < 5, "Mod 5 should be between 0 and 4")
			suite.True(result.Mod10 >= 0 && result.Mod10 < 10, "Mod 10 should be between 0 and 9")
			suite.T().Logf("ViewCount: %d, Mod5: %d, Mod10: %d",
				result.ViewCount, result.Mod5, result.Mod10)
		}
	})
}

// TestGreatest tests the Greatest function.
func (suite *MathFunctionsTestSuite) TestGreatest() {
	suite.T().Logf("Testing Greatest function for %s", suite.dbType)

	suite.Run("GreatestValue", func() {
		type GreatestResult struct {
			ViewCount int64 `bun:"view_count"`
			Greatest  int64 `bun:"greatest"`
		}

		var results []GreatestResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Greatest(eb.Column("view_count"), 20, 30, 40)
			}, "greatest").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Greatest should work correctly")
		suite.True(len(results) > 0, "Should have greatest results")

		for _, result := range results {
			suite.True(result.Greatest >= result.ViewCount, "Greatest should be >= view_count")
			suite.T().Logf("ViewCount: %d, Greatest: %d", result.ViewCount, result.Greatest)
		}
	})
}

// TestLeast tests the Least function.
func (suite *MathFunctionsTestSuite) TestLeast() {
	suite.T().Logf("Testing Least function for %s", suite.dbType)

	suite.Run("LeastValue", func() {
		type LeastResult struct {
			ViewCount int64 `bun:"view_count"`
			Least     int64 `bun:"least"`
		}

		var results []LeastResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Least(eb.Column("view_count"), 20, 30, 40)
			}, "least").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Least should work correctly")
		suite.True(len(results) > 0, "Should have least results")

		for _, result := range results {
			suite.True(result.Least <= result.ViewCount, "Least should be <= view_count")
			suite.T().Logf("ViewCount: %d, Least: %d", result.ViewCount, result.Least)
		}
	})
}

// TestCombinedMathFunctions tests multiple math functions working together.
func (suite *MathFunctionsTestSuite) TestCombinedMathFunctions() {
	suite.T().Logf("Testing combined math functions for %s", suite.dbType)

	suite.Run("BasicMathCombination", func() {
		type CombinedBasicResult struct {
			Title        string  `bun:"title"`
			ViewCount    int64   `bun:"view_count"`
			AbsViewCount int64   `bun:"abs_view_count"`
			CeiledViews  float64 `bun:"ceiled_views"`
			FlooredViews float64 `bun:"floored_views"`
			RoundedViews float64 `bun:"rounded_views"`
		}

		var results []CombinedBasicResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Abs(eb.Expr("? - 100", eb.Column("view_count")))
			}, "abs_view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Ceil(eb.Expr("? / 10.0", eb.Column("view_count")))
			}, "ceiled_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Floor(eb.Expr("? / 10.0", eb.Column("view_count")))
			}, "floored_views").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Round(eb.Expr("? / 10.0", eb.Column("view_count")), 1)
			}, "rounded_views").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined basic math functions should work correctly")
		suite.True(len(results) > 0, "Should have combined math results")

		for _, result := range results {
			suite.True(result.AbsViewCount >= 0, "Abs should be non-negative")
			suite.True(result.CeiledViews >= result.FlooredViews, "Ceil should be >= Floor")
			suite.T().Logf("Post %s: ViewCount=%d, Abs=%d, Ceil=%.0f, Floor=%.0f, Round=%.1f",
				result.Title, result.ViewCount, result.AbsViewCount,
				result.CeiledViews, result.FlooredViews, result.RoundedViews)
		}
	})

	suite.Run("AdvancedMathCombination", func() {
		type CombinedAdvancedResult struct {
			Title       string  `bun:"title"`
			ViewCount   int64   `bun:"view_count"`
			PowerResult float64 `bun:"power_result"`
			SqrtResult  float64 `bun:"sqrt_result"`
			ModResult   int64   `bun:"mod_result"`
			SignResult  float64 `bun:"sign_result"`
		}

		var results []CombinedAdvancedResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title", "view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Power(eb.Column("view_count"), 2)
			}, "power_result").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sqrt(eb.Column("view_count"))
			}, "sqrt_result").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Mod(eb.Column("view_count"), 7)
			}, "mod_result").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sign(eb.Expr("? - 50", eb.Column("view_count")))
			}, "sign_result").
			OrderBy("title").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined advanced math functions should work correctly")
		suite.True(len(results) > 0, "Should have combined advanced math results")

		for _, result := range results {
			suite.True(result.PowerResult >= 0, "Power result should be non-negative")
			suite.True(result.SqrtResult >= 0, "Sqrt result should be non-negative")
			suite.True(result.ModResult >= 0 && result.ModResult < 7, "Mod result should be between 0 and 6")
			suite.Contains([]float64{-1.0, 0.0, 1.0}, result.SignResult, "Sign result should be -1, 0, or 1")
			suite.T().Logf("Post %s: ViewCount=%d, Power=%.0f, Sqrt=%.2f, Mod=%d, Sign=%.0f",
				result.Title, result.ViewCount, result.PowerResult, result.SqrtResult,
				result.ModResult, result.SignResult)
		}
	})

	suite.Run("TrigonometricCombination", func() {
		type CombinedTrigResult struct {
			ViewCount int64   `bun:"view_count"`
			SinValue  float64 `bun:"sin_value"`
			CosValue  float64 `bun:"cos_value"`
			TanValue  float64 `bun:"tan_value"`
		}

		var results []CombinedTrigResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sin(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "sin_value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Cos(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "cos_value").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Tan(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "tan_value").
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined trigonometric functions should work correctly")
		suite.True(len(results) > 0, "Should have combined trigonometric results")

		for _, result := range results {
			suite.True(result.SinValue >= -1.0 && result.SinValue <= 1.0, "Sin value should be between -1 and 1")
			suite.True(result.CosValue >= -1.0 && result.CosValue <= 1.0, "Cos value should be between -1 and 1")
			suite.T().Logf("ViewCount: %d, Sin: %.4f, Cos: %.4f, Tan: %.4f",
				result.ViewCount, result.SinValue, result.CosValue, result.TanValue)
		}
	})

	suite.Run("LogarithmicCombination", func() {
		type CombinedLogResult struct {
			ViewCount int64   `bun:"view_count"`
			SinLn     float64 `bun:"sin_ln"`
			CosExp    float64 `bun:"cos_exp"`
			TanLog    float64 `bun:"tan_log"`
		}

		var results []CombinedLogResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("view_count").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Sin(eb.Ln(eb.Column("view_count")))
			}, "sin_ln").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Cos(eb.Expr("? / 100.0", eb.Column("view_count")))
			}, "cos_exp").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.Tan(eb.Log(10, eb.Column("view_count")))
			}, "tan_log").
			Where(func(cb ConditionBuilder) {
				cb.GreaterThan("view_count", 1)
			}).
			OrderBy("view_count").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Combined logarithmic and trigonometric functions should work correctly")
		suite.True(len(results) > 0, "Should have combined logarithmic results")

		for _, result := range results {
			suite.True(result.SinLn >= -1.0 && result.SinLn <= 1.0, "Sin(Ln) should be between -1 and 1")
			suite.True(result.CosExp >= -1.0 && result.CosExp <= 1.0, "Cos should be between -1 and 1")
			suite.T().Logf("ViewCount: %d, Sin(Ln): %.4f, Cos: %.4f, Tan(Log): %.4f",
				result.ViewCount, result.SinLn, result.CosExp, result.TanLog)
		}
	})
}
