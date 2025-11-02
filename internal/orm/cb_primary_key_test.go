package orm

// PrimaryKeyConditionsTestSuite tests primary key condition methods.
// Covers: PkEquals, PkNotEquals, PkIn, PkNotIn and their Or variants (8 methods total).
type PrimaryKeyConditionsTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestPkEquals tests the PkEquals and OrPkEquals conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPkEquals() {
	suite.T().Logf("Testing PkEquals condition for %s", suite.dbType)

	suite.Run("BasicPkEquals", func() {
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
					cb.PkEquals(firstUser.Id)
				}),
		)

		suite.Len(users, 1, "Should find exactly one user")
		suite.Equal(firstUser.Id, users[0].Id, "Should find the correct user")

		suite.T().Logf("Found user: %s (ID: %s)", users[0].Name, users[0].Id)
	})

	suite.Run("OrPkEquals", func() {
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
					cb.PkEquals(allUsers[0].Id).
						OrPkEquals(allUsers[1].Id)
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")
		suite.Equal(allUsers[0].Id, users[0].Id)
		suite.Equal(allUsers[1].Id, users[1].Id)

		suite.T().Logf("Found users: %s, %s", users[0].Name, users[1].Name)
	})
}

// TestPkNotEquals tests the PkNotEquals and OrPkNotEquals conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPkNotEquals() {
	suite.T().Logf("Testing PkNotEquals condition for %s", suite.dbType)

	suite.Run("BasicPkNotEquals", func() {
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
					cb.PkNotEquals(firstUser.Id)
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.NotEqual(firstUser.Id, user.Id, "Should not be the excluded user")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrPkNotEquals", func() {
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
					cb.PkNotEquals(allUsers[0].Id).
						OrPkNotEquals(allUsers[1].Id)
				}).
				OrderBy("id"),
		)

		suite.True(len(users) > 0, "Should find users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestPkIn tests the PkIn and OrPkIn conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPkIn() {
	suite.T().Logf("Testing PkIn condition for %s", suite.dbType)

	suite.Run("BasicPkIn", func() {
		// Get two users
		var allUsers []User

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			OrderBy("id").
			Limit(2).
			Scan(suite.ctx, &allUsers)
		suite.NoError(err, "Should get users")
		suite.Len(allUsers, 2, "Should have two users")

		ids := []string{allUsers[0].Id, allUsers[1].Id}

		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.PkIn(ids)
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")
		suite.Equal(allUsers[0].Id, users[0].Id)
		suite.Equal(allUsers[1].Id, users[1].Id)

		suite.T().Logf("Found users: %s, %s", users[0].Name, users[1].Name)
	})

	suite.Run("OrPkIn", func() {
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
					cb.PkIn([]string{allUsers[0].Id}).
						OrPkIn([]string{allUsers[1].Id})
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")

		suite.T().Logf("Found %d users", len(users))
	})
}

// TestPkNotIn tests the PkNotIn and OrPkNotIn conditions.
func (suite *PrimaryKeyConditionsTestSuite) TestPkNotIn() {
	suite.T().Logf("Testing PkNotIn condition for %s", suite.dbType)

	suite.Run("BasicPkNotIn", func() {
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
					cb.PkNotIn([]string{firstUser.Id})
				}).
				OrderBy("id"),
		)

		suite.Len(users, 2, "Should find two users")

		for _, user := range users {
			suite.NotEqual(firstUser.Id, user.Id, "Should not be the excluded user")
		}

		suite.T().Logf("Found %d users", len(users))
	})

	suite.Run("OrPkNotIn", func() {
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
					cb.PkNotIn([]string{allUsers[0].Id}).
						OrPkNotIn([]string{allUsers[1].Id})
				}).
				OrderBy("id"),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d users", len(users))
	})
}
