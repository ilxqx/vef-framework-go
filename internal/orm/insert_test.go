package orm

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
)

// InsertTestSuite tests INSERT operations following InsertQuery interface method order.
// Tests cover all InsertQuery methods including CTE, table specification, column selection,
// column values, conflict handling, RETURNING clause, Apply functions, bulk operations, and error handling.
type InsertTestSuite struct {
	*OrmTestSuite
}

// TestBasicInsert tests Model and Exec methods with single and bulk inserts.
func (suite *InsertTestSuite) TestBasicInsert() {
	suite.T().Logf("Testing basic INSERT for %s", suite.DbType)

	suite.Run("InsertSingleRecord", func() {
		user := &User{
			Name:     "John Doe",
			Email:    "john@example.com",
			Age:      28,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert single user successfully")
		suite.NotEmpty(user.Id, "User Id should be set after insert")

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "john@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err, "Should retrieve inserted user")
		suite.Equal("John Doe", retrieved.Name)
		suite.Equal("john@example.com", retrieved.Email)

		suite.T().Logf("Inserted user: Id=%s, Name=%s", retrieved.Id, retrieved.Name)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "john@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("InsertMultipleRecords", func() {
		users := []*User{
			{Name: "Jane Smith", Email: "jane@example.com", Age: 26, IsActive: true},
			{Name: "Mike Wilson", Email: "mike@example.com", Age: 31, IsActive: false},
		}

		_, err := suite.Db.NewInsert().
			Model(&users).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert multiple users successfully")

		for _, user := range users {
			suite.NotEmpty(user.Id, "Each user should have an Id set")
		}

		var retrieved []User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.In("email", []any{"jane@example.com", "mike@example.com"})
			}).
			Scan(suite.Ctx)
		suite.NoError(err, "Should retrieve all inserted users")
		suite.Len(retrieved, 2, "Should have inserted 2 users")

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.In("email", []any{"jane@example.com", "mike@example.com"})
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestCTE tests With, WithValues, and WithRecursive methods.
// MySQL does not support CTE in INSERT statements, so these tests are skipped for MySQL.
func (suite *InsertTestSuite) TestCTE() {
	suite.T().Logf("Testing CTE methods for %s", suite.DbType)

	if suite.DbType == constants.DbMySQL {
		suite.T().Skipf("CTE in INSERT not supported on %s", suite.DbType)

		return
	}

	suite.Run("InsertWithSimpleCTE", func() {
		category := &Category{
			Name:        "CTE Category",
			Description: lo.ToPtr("Category created via CTE"),
		}

		_, err := suite.Db.NewInsert().
			With("existing_tech", func(sq SelectQuery) {
				sq.Model((*Category)(nil)).
					Select("name", "description").
					Where(func(cb ConditionBuilder) {
						cb.Equals("name", "Technology")
					})
			}).
			Model(category).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert with CTE")
		suite.NotEmpty(category.Id)

		suite.T().Logf("Inserted via CTE: Id=%s, Name=%s", category.Id, category.Name)

		_, err = suite.Db.NewDelete().
			Model((*Category)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(category.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("InsertWithValuesCTE", func() {
		type TempData struct {
			Name  string
			Email string
		}

		tempData := []TempData{
			{Name: "CTE User", Email: "cte@example.com"},
		}

		user := &User{
			Name:     "CTE User",
			Email:    "cte@example.com",
			Age:      30,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			WithValues("temp_data", &tempData).
			Model(user).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert with VALUES CTE")
		suite.NotEmpty(user.Id)

		suite.T().Logf("Inserted with VALUES CTE: Id=%s", user.Id)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "cte@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("InsertWithRecursiveCTE", func() {
		user := &User{
			Name:     "Recursive User",
			Email:    "recursive@example.com",
			Age:      35,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			WithRecursive("user_hierarchy", func(sq SelectQuery) {
				sq.Model((*User)(nil)).
					Select("id", "name").
					Where(func(cb ConditionBuilder) {
						cb.IsNotNull("id")
					}).
					Limit(1)
			}).
			Model(user).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert with recursive CTE")
		suite.NotEmpty(user.Id)

		suite.T().Logf("Inserted with recursive CTE: Id=%s", user.Id)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "recursive@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestTableSpecification tests table specification methods.
// Note: Table/ModelTable/TableFrom/TableExpr/TableSubQuery are primarily for INSERT...SELECT queries.
// Model() is the standard method for inserting from struct values.
func (suite *InsertTestSuite) TestTableSpecification() {
	suite.T().Logf("Testing table specification methods for %s", suite.DbType)

	suite.Run("ModelTableWithModel", func() {
		user := &User{
			Name:     "ModelTable User",
			Email:    "modeltable@example.com",
			Age:      27,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert using Model (standard approach)")
		suite.NotEmpty(user.Id)

		suite.T().Logf("Inserted with Model: Id=%s", user.Id)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestColumnSelection tests Select and Exclude methods.
// Note: SelectAll and ExcludeAll are less commonly used with Model-based inserts.
func (suite *InsertTestSuite) TestColumnSelection() {
	suite.T().Logf("Testing column selection methods for %s", suite.DbType)

	suite.Run("ExcludeSpecificColumns", func() {
		user := &User{
			Name:     "Exclude User",
			Email:    "exclude@example.com",
			Age:      30,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Exclude("meta").
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert excluding specific columns")
		suite.NotEmpty(user.Id)

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Exclude User", retrieved.Name)
		suite.Equal("exclude@example.com", retrieved.Email)

		suite.T().Logf("Inserted with Exclude: Id=%s, Name=%s", retrieved.Id, retrieved.Name)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestColumnValues tests Column and ColumnExpr methods.
func (suite *InsertTestSuite) TestColumnValues() {
	suite.T().Logf("Testing column value methods for %s", suite.DbType)

	suite.Run("ColumnDirectValue", func() {
		user := &User{
			Name:     "Original Name",
			Email:    "column@example.com",
			Age:      25,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Column("name", "Overridden Name").
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert with column override")
		suite.NotEmpty(user.Id)

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "column@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Overridden Name", retrieved.Name, "Name should be overridden")
		suite.Equal(int16(25), retrieved.Age, "Age should keep model value")

		suite.T().Logf("Column override: Name=%s (overridden from Original Name)", retrieved.Name)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("ColumnExprWithFunction", func() {
		user := &User{
			Name:     "expr user",
			Email:    "expr@example.com",
			Age:      28,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			ColumnExpr("name", func(eb ExprBuilder) any {
				return eb.Upper(eb.Literal("Expr User"))
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert with column expression")
		suite.NotEmpty(user.Id)

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "expr@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("EXPR USER", retrieved.Name, "Name should be uppercased by expression")

		suite.T().Logf("ColumnExpr result: Name=%s", retrieved.Name)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("MultipleColumnOverrides", func() {
		user := &User{
			Name:     "Original",
			Email:    "multi@example.com",
			Age:      20,
			IsActive: false,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Column("name", "Multiple Override").
			Column("age", 35).
			Column("is_active", true).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert with multiple column overrides")
		suite.NotEmpty(user.Id)

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "multi@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Multiple Override", retrieved.Name)
		suite.Equal(int16(35), retrieved.Age)
		suite.True(retrieved.IsActive)

		suite.T().Logf("Multiple overrides: Name=%s, Age=%d, IsActive=%v",
			retrieved.Name, retrieved.Age, retrieved.IsActive)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestConflictHandling tests OnConflict with DO NOTHING and DO UPDATE.
func (suite *InsertTestSuite) TestConflictHandling() {
	suite.T().Logf("Testing conflict handling for %s", suite.DbType)

	suite.Run("OnConflictDoNothing", func() {
		original := &User{
			Name:     "Conflict User",
			Email:    "conflict@example.com",
			Age:      30,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(original).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert original user")

		duplicate := &User{
			Name:     "Duplicate User",
			Email:    "conflict@example.com",
			Age:      25,
			IsActive: false,
		}

		_, err = suite.Db.NewInsert().
			Model(duplicate).
			OnConflict(func(cb ConflictBuilder) {
				cb.Columns("email").DoNothing()
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should handle conflict with DO NOTHING")

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "conflict@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Conflict User", retrieved.Name, "Original name should be unchanged")
		suite.Equal(int16(30), retrieved.Age, "Original age should be unchanged")

		suite.T().Logf("DO NOTHING: Name=%s, Age=%d (unchanged)", retrieved.Name, retrieved.Age)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "conflict@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("OnConflictDoUpdate", func() {
		original := &User{
			Name:     "Update Original",
			Email:    "update-conflict@example.com",
			Age:      30,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(original).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert original user")

		update := &User{
			Name:     "Update Modified",
			Email:    "update-conflict@example.com",
			Age:      35,
			IsActive: false,
		}

		_, err = suite.Db.NewInsert().
			Model(update).
			OnConflict(func(cb ConflictBuilder) {
				cb.Columns("email").DoUpdate().
					Set("name", "Update Modified").
					Set("age", 35).
					Set("is_active", false)
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should handle conflict with DO UPDATE")

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "update-conflict@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Update Modified", retrieved.Name, "Name should be updated")
		suite.Equal(int16(35), retrieved.Age, "Age should be updated")
		suite.False(retrieved.IsActive, "IsActive should be updated")

		suite.T().Logf("DO UPDATE: Name=%s, Age=%d, IsActive=%v",
			retrieved.Name, retrieved.Age, retrieved.IsActive)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "update-conflict@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("OnConflictWithCondition", func() {
		original := &User{
			Name:     "Conditional User",
			Email:    "conditional@example.com",
			Age:      40,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(original).
			Exec(suite.Ctx)
		suite.NoError(err)

		update := &User{
			Name:     "Conditional Update",
			Email:    "conditional@example.com",
			Age:      45,
			IsActive: true,
		}

		_, err = suite.Db.NewInsert().
			Model(update).
			OnConflict(func(cb ConflictBuilder) {
				cb.Columns("email").DoUpdate().
					Set("age", 45).
					Where(func(wcb ConditionBuilder) {
						wcb.GreaterThan("age", 35)
					})
			}).
			Exec(suite.Ctx)
		suite.NoError(err, "Should handle conflict with conditional update")

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "conditional@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal(int16(45), retrieved.Age, "Age should be updated based on condition")

		suite.T().Logf("Conditional UPDATE: Age=%d", retrieved.Age)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "conditional@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestReturning tests Returning, ReturningAll, and ReturningNone.
// RETURNING clause is only supported on PostgreSQL and SQLite.
func (suite *InsertTestSuite) TestReturning() {
	suite.T().Logf("Testing RETURNING clause for %s", suite.DbType)

	if suite.DbType == constants.DbMySQL {
		suite.T().Skipf("RETURNING clause not supported on %s", suite.DbType)

		return
	}

	suite.Run("ReturningSpecificColumns", func() {
		user := &User{
			Name:     "Return User",
			Email:    "return@example.com",
			Age:      29,
			IsActive: true,
		}

		err := suite.Db.NewInsert().
			Model(user).
			Returning("id", "name", "email").
			Scan(suite.Ctx, user)
		suite.NoError(err, "Should insert with RETURNING specific columns")
		suite.NotEmpty(user.Id, "Id should be returned")
		suite.Equal("Return User", user.Name, "Name should be returned")
		suite.Equal("return@example.com", user.Email, "Email should be returned")

		suite.T().Logf("RETURNING columns: Id=%s, Name=%s, Email=%s",
			user.Id, user.Name, user.Email)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("ReturningAllColumns", func() {
		user := &User{
			Name:     "Return All User",
			Email:    "returnall@example.com",
			Age:      32,
			IsActive: true,
		}

		err := suite.Db.NewInsert().
			Model(user).
			ReturningAll().
			Scan(suite.Ctx, user)
		suite.NoError(err, "Should insert with RETURNING all columns")
		suite.NotEmpty(user.Id)
		suite.Equal("Return All User", user.Name)
		suite.Equal(int16(32), user.Age)
		suite.True(user.IsActive)

		suite.T().Logf("RETURNING all: Id=%s, Name=%s, Age=%d",
			user.Id, user.Name, user.Age)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("ReturningNoColumns", func() {
		user := &User{
			Name:     "Return None User",
			Email:    "returnnone@example.com",
			Age:      28,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			ReturningNone().
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert with RETURNING none")
		suite.NotEmpty(user.Id, "Id should still be set by audit handler")

		suite.T().Logf("RETURNING none: Id=%s (set by audit)", user.Id)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestApply tests Apply and ApplyIf methods.
func (suite *InsertTestSuite) TestApply() {
	suite.T().Logf("Testing Apply methods for %s", suite.DbType)

	suite.Run("ApplyUnconditional", func() {
		user := &User{
			Name:     "Apply User",
			Email:    "apply@example.com",
			Age:      27,
			IsActive: true,
		}

		applyFunc := func(q InsertQuery) {
			q.Column("name", "Applied Name")
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Apply(applyFunc).
			Exec(suite.Ctx)
		suite.NoError(err, "Should apply function unconditionally")
		suite.NotEmpty(user.Id)

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "apply@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Applied Name", retrieved.Name, "Name should be modified by Apply")

		suite.T().Logf("Apply result: Name=%s", retrieved.Name)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("ApplyConditional", func() {
		user1 := &User{
			Name:     "Conditional User 1",
			Email:    "cond1@example.com",
			Age:      30,
			IsActive: true,
		}

		applyFunc := func(q InsertQuery) {
			q.Column("name", "Modified Name")
		}

		_, err := suite.Db.NewInsert().
			Model(user1).
			ApplyIf(true, applyFunc).
			Exec(suite.Ctx)
		suite.NoError(err)
		suite.NotEmpty(user1.Id)

		var retrieved1 User

		err = suite.Db.NewSelect().
			Model(&retrieved1).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "cond1@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Modified Name", retrieved1.Name, "ApplyIf(true) should apply function")

		user2 := &User{
			Name:     "Conditional User 2",
			Email:    "cond2@example.com",
			Age:      32,
			IsActive: true,
		}

		_, err = suite.Db.NewInsert().
			Model(user2).
			ApplyIf(false, applyFunc).
			Exec(suite.Ctx)
		suite.NoError(err)
		suite.NotEmpty(user2.Id)

		var retrieved2 User

		err = suite.Db.NewSelect().
			Model(&retrieved2).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "cond2@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Conditional User 2", retrieved2.Name, "ApplyIf(false) should not apply function")

		suite.T().Logf("ApplyIf: true=%s, false=%s", retrieved1.Name, retrieved2.Name)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.In("email", []any{"cond1@example.com", "cond2@example.com"})
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("ApplyMultipleFunctions", func() {
		user := &User{
			Name:     "Multi Apply",
			Email:    "multi-apply@example.com",
			Age:      20,
			IsActive: false,
		}

		fn1 := func(q InsertQuery) {
			q.Column("name", "Step 1")
		}
		fn2 := func(q InsertQuery) {
			q.Column("age", 25)
		}
		fn3 := func(q InsertQuery) {
			q.Column("is_active", true)
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Apply(fn1, fn2, fn3).
			Exec(suite.Ctx)
		suite.NoError(err, "Should apply multiple functions")
		suite.NotEmpty(user.Id)

		var retrieved User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "multi-apply@example.com")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Equal("Step 1", retrieved.Name)
		suite.Equal(int16(25), retrieved.Age)
		suite.True(retrieved.IsActive)

		suite.T().Logf("Multiple Apply: Name=%s, Age=%d, IsActive=%v",
			retrieved.Name, retrieved.Age, retrieved.IsActive)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("ApplyWithNilFunction", func() {
		user := &User{
			Name:     "Nil Apply User",
			Email:    "nil-apply@example.com",
			Age:      28,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Apply(nil).
			Exec(suite.Ctx)
		suite.NoError(err, "Should handle nil function safely")
		suite.NotEmpty(user.Id)

		suite.T().Logf("Nil function handled: Id=%s", user.Id)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.PkEquals(user.Id)
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestBulkInsert tests bulk insert operations.
func (suite *InsertTestSuite) TestBulkInsert() {
	suite.T().Logf("Testing bulk INSERT for %s", suite.DbType)

	suite.Run("LargeBatchInsert", func() {
		batchSize := 10

		users := make([]*User, batchSize)
		for i := range batchSize {
			users[i] = &User{
				Name:     fmt.Sprintf("Batch User %d", i+1),
				Email:    fmt.Sprintf("batch%d@example.com", i+1),
				Age:      int16(20 + i),
				IsActive: i%2 == 0,
			}
		}

		start := time.Now()
		_, err := suite.Db.NewInsert().
			Model(&users).
			Exec(suite.Ctx)
		duration := time.Since(start)

		suite.NoError(err, "Should insert batch users successfully")
		suite.T().Logf("Batch insert of %d users took %v", batchSize, duration)

		for _, user := range users {
			suite.NotEmpty(user.Id, "Each user should have an Id")
		}

		var retrieved []User

		err = suite.Db.NewSelect().
			Model(&retrieved).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "batch")
			}).
			Scan(suite.Ctx)
		suite.NoError(err)
		suite.Len(retrieved, batchSize, "Should have inserted all batch users")

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("email", "batch")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("BulkInsertWithRelatedData", func() {
		users := []*User{
			{Name: "Post Author 1", Email: "author1@bulk.com", Age: 30, IsActive: true},
			{Name: "Post Author 2", Email: "author2@bulk.com", Age: 25, IsActive: true},
		}

		_, err := suite.Db.NewInsert().
			Model(&users).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert bulk users")

		posts := []*Post{
			{
				Title:       "Bulk Post 1",
				Content:     "Content for bulk post 1",
				Description: lo.ToPtr("Description 1"),
				UserId:      users[0].Id,
				CategoryId:  suite.getCategoryId(),
				Status:      "published",
				ViewCount:   100,
			},
			{
				Title:      "Bulk Post 2",
				Content:    "Content for bulk post 2",
				UserId:     users[1].Id,
				CategoryId: suite.getCategoryId(),
				Status:     "draft",
				ViewCount:  0,
			},
		}

		_, err = suite.Db.NewInsert().
			Model(&posts).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert related posts")

		for _, post := range posts {
			suite.NotEmpty(post.Id)
			suite.T().Logf("Bulk post: Id=%s, Title=%s, UserId=%s", post.Id, post.Title, post.UserId)
		}

		_, err = suite.Db.NewDelete().
			Model((*Post)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.StartsWith("title", "Bulk Post")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.In("email", []any{"author1@bulk.com", "author2@bulk.com"})
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})
}

// TestErrorHandling tests error scenarios in insert operations.
func (suite *InsertTestSuite) TestErrorHandling() {
	suite.T().Logf("Testing error handling for %s", suite.DbType)

	suite.Run("UniqueConstraintViolation", func() {
		original := &User{
			Name:     "Original User",
			Email:    "unique@example.com",
			Age:      25,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(original).
			Exec(suite.Ctx)
		suite.NoError(err, "Should insert original user")

		duplicate := &User{
			Name:     "Duplicate User",
			Email:    "unique@example.com",
			Age:      30,
			IsActive: false,
		}

		_, err = suite.Db.NewInsert().
			Model(duplicate).
			Exec(suite.Ctx)
		suite.Error(err, "Insert with duplicate email should fail")

		suite.T().Logf("Unique constraint violation handled correctly")

		_, err = suite.Db.NewDelete().
			Model((*User)(nil)).
			Where(func(cb ConditionBuilder) {
				cb.Equals("email", "unique@example.com")
			}).
			Exec(suite.Ctx)
		suite.NoError(err)
	})

	suite.Run("NullConstraintViolation", func() {
		invalid := &User{
			Name:  "",
			Email: "",
		}

		_, err := suite.Db.NewInsert().
			Model(invalid).
			Column("name", nil).
			Column("email", nil).
			Exec(suite.Ctx)
		suite.Error(err, "Insert with null constraint violation should fail")

		suite.T().Logf("NULL constraint violation handled correctly")
	})

	suite.Run("InvalidDataType", func() {
		user := &User{
			Name:     "Invalid Type User",
			Email:    "invalid-type@example.com",
			Age:      30,
			IsActive: true,
		}

		_, err := suite.Db.NewInsert().
			Model(user).
			Column("age", "not a number").
			Exec(suite.Ctx)
		suite.Error(err, "Insert with invalid data type should fail")

		suite.T().Logf("Invalid data type handled correctly")
	})
}

// getCategoryId returns the first available category Id from fixture data.
func (suite *InsertTestSuite) getCategoryId() string {
	var category Category
	if err := suite.Db.NewSelect().
		Model(&category).
		Limit(1).
		Scan(suite.Ctx); err != nil {
		suite.T().Fatalf("Failed to get category Id: %v", err)
	}

	return category.Id
}
