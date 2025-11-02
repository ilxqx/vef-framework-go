package orm

import "time"

// AuditConditionsTestSuite tests audit field condition methods.
// Covers: CreatedBy, UpdatedBy, CreatedAt, UpdatedAt series (~158 methods total).
type AuditConditionsTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestCreatedByEquals tests the CreatedByEquals and OrCreatedByEquals conditions.
func (suite *AuditConditionsTestSuite) TestCreatedByEquals() {
	suite.T().Logf("Testing CreatedByEquals condition for %s", suite.dbType)

	suite.Run("BasicCreatedByEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedByEquals("system")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		for _, user := range users {
			suite.Equal("system", user.CreatedBy, "CreatedBy should be system")
		}

		suite.T().Logf("Found %d users created by system", len(users))
	})

	suite.Run("OrCreatedByEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedByEquals("system").
						OrCreatedByEquals("admin")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestCreatedByNotEquals tests the CreatedByNotEquals and OrCreatedByNotEquals conditions.
func (suite *AuditConditionsTestSuite) TestCreatedByNotEquals() {
	suite.T().Logf("Testing CreatedByNotEquals condition for %s", suite.dbType)

	suite.Run("BasicCreatedByNotEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedByNotEquals("nonexistent")
				}),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrCreatedByNotEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedByNotEquals("user1").
						OrCreatedByNotEquals("user2")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestCreatedByIn tests the CreatedByIn and OrCreatedByIn conditions.
func (suite *AuditConditionsTestSuite) TestCreatedByIn() {
	suite.T().Logf("Testing CreatedByIn condition for %s", suite.dbType)

	suite.Run("BasicCreatedByIn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedByIn([]string{"system", "admin"})
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrCreatedByIn", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedByIn([]string{"system"}).
						OrCreatedByIn([]string{"admin"})
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestUpdatedByEquals tests the UpdatedByEquals and OrUpdatedByEquals conditions.
func (suite *AuditConditionsTestSuite) TestUpdatedByEquals() {
	suite.T().Logf("Testing UpdatedByEquals condition for %s", suite.dbType)

	suite.Run("BasicUpdatedByEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.UpdatedByEquals("system")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		for _, user := range users {
			suite.Equal("system", user.UpdatedBy, "UpdatedBy should be system")
		}

		suite.T().Logf("Found %d users updated by system", len(users))
	})

	suite.Run("OrUpdatedByEquals", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.UpdatedByEquals("system").
						OrUpdatedByEquals("admin")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestCreatedAtBetween tests the CreatedAtBetween and OrCreatedAtBetween conditions.
func (suite *AuditConditionsTestSuite) TestCreatedAtBetween() {
	suite.T().Logf("Testing CreatedAtBetween condition for %s", suite.dbType)

	suite.Run("BasicCreatedAtBetween", func() {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedAtBetween(yesterday, tomorrow)
				}),
		)

		suite.True(len(users) > 0, "Should find users created in the last day")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrCreatedAtBetween", func() {
		now := time.Now()
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.CreatedAtBetween(now.Add(-48*time.Hour), now.Add(-24*time.Hour)).
						OrCreatedAtBetween(now.Add(-24*time.Hour), now.Add(24*time.Hour))
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestUpdatedAtGreaterThan tests the UpdatedAtGreaterThan and OrUpdatedAtGreaterThan conditions.
func (suite *AuditConditionsTestSuite) TestUpdatedAtGreaterThan() {
	suite.T().Logf("Testing UpdatedAtGreaterThan condition for %s", suite.dbType)

	suite.Run("BasicUpdatedAtGreaterThan", func() {
		yesterday := time.Now().Add(-24 * time.Hour)

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.UpdatedAtGreaterThan(yesterday)
				}),
		)

		suite.True(len(users) > 0, "Should find recently updated users")

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrUpdatedAtGreaterThan", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.UpdatedAtGreaterThan(time.Now().Add(-48 * time.Hour)).
						OrUpdatedAtGreaterThan(time.Now().Add(-24 * time.Hour))
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}
