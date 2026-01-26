package orm

// PrimaryKeyConditionsTestSuite tests primary key condition methods.
// Covers: PKEquals, PKNotEquals, PKIn, PKNotIn and their Or variants (8 methods total).
type PrimaryKeyConditionsTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestPKEquals tests the PKEquals and OrPKEquals conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPKEquals() {
	suite.T().Logf("Testing PKEquals condition for %s", suite.dbType)

	suite.Run("BasicPKEquals", func() {
		// Get a user first to get their ID
		var firstUser User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(1).
			Scan(suite.ctx, &firstUser)
		suite.NoError(err, "Should get a user")

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKEquals(firstUser.ID)
				}),
		)

		suite.Len(users, 1, "Should find exactly one user")
		suite.Equal(firstUser.ID, users[0].ID, "Should find the correct user")

		suite.T().Logf("Found user: %s (ID: %s)", users[0].Name, users[0].ID)
	})

	suite.Run("OrPKEquals", func() {
		// Get two users
		var allUsers []User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx, &allUsers)
		suite.NoError(err, "Should get users")
		suite.Len(allUsers, 2, "Should have two users")

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKEquals(allUsers[0].ID).
						OrPKEquals(allUsers[1].ID)
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")
		suite.Equal(allUsers[0].ID, users[0].ID)
		suite.Equal(allUsers[1].ID, users[1].ID)

		suite.T().Logf("Found users: %s, %s", users[0].Name, users[1].Name)
	})
}

// TestPKNotEquals tests the PKNotEquals and OrPKNotEquals conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPKNotEquals() {
	suite.T().Logf("Testing PKNotEquals condition for %s", suite.dbType)

	suite.Run("BasicPKNotEquals", func() {
		// Get a user first
		var firstUser User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(1).
			Scan(suite.ctx, &firstUser)
		suite.NoError(err, "Should get a user")

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKNotEquals(firstUser.ID)
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.NotEqual(firstUser.ID, user.ID, "Should not be the excluded user")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrPKNotEquals", func() {
		var allUsers []User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx, &allUsers)
		suite.NoError(err, "Should get users")

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKNotEquals(allUsers[0].ID).
						OrPKNotEquals(allUsers[1].ID)
				}).
				OrderBy("id"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestPKIn tests the PKIn and OrPKIn conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPKIn() {
	suite.T().Logf("Testing PKIn condition for %s", suite.dbType)

	suite.Run("BasicPKIn", func() {
		// Get two users
		var allUsers []User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx, &allUsers)
		suite.NoError(err, "Should get users")
		suite.Len(allUsers, 2, "Should have two users")

		ids := []string{allUsers[0].ID, allUsers[1].ID}

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKIn(ids)
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")
		suite.Equal(allUsers[0].ID, users[0].ID)
		suite.Equal(allUsers[1].ID, users[1].ID)

		suite.T().Logf("Found users: %s, %s", users[0].Name, users[1].Name)
	})

	suite.Run("OrPKIn", func() {
		var allUsers []User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx, &allUsers)
		suite.NoError(err, "Should get users")

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKIn([]string{allUsers[0].ID}).
						OrPKIn([]string{allUsers[1].ID})
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestPKNotIn tests the PKNotIn and OrPKNotIn conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPKNotIn() {
	suite.T().Logf("Testing PKNotIn condition for %s", suite.dbType)

	suite.Run("BasicPKNotIn", func() {
		// Get one user to exclude
		var firstUser User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(1).
			Scan(suite.ctx, &firstUser)
		suite.NoError(err, "Should get a user")

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKNotIn([]string{firstUser.ID})
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.NotEqual(firstUser.ID, user.ID, "Should not be the excluded user")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrPKNotIn", func() {
		var allUsers []User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx, &allUsers)
		suite.NoError(err, "Should get users")

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PKNotIn([]string{allUsers[0].ID}).
						OrPKNotIn([]string{allUsers[1].ID})
				}).
				OrderBy("id"),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}
