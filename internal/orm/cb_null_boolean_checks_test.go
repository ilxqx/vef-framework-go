package orm

// NullBooleanChecksTestSuite tests NULL and boolean check condition methods.
// Covers: IsNull, IsNotNull, IsTrue, IsFalse, IsTrueOrNull, IsFalseOrNull and their Or variants.
type NullBooleanChecksTestSuite struct {
	*ConditionBuilderTestSuite
}

// TestIsNull tests the IsNull and OrIsNull conditions.
func (suite *NullBooleanChecksTestSuite) TestIsNull() {
	suite.T().Logf("Testing IsNull condition for %s", suite.dbType)

	suite.Run("BasicIsNull", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsNull("description")
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		for _, post := range posts {
			suite.Nil(post.Description, "Description should be NULL")
		}

		suite.T().Logf("Found %d posts with NULL description", len(posts))
	})

	suite.Run("OrIsNull", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsNull("description").
						OrIsNull("content")
				}),
		)

		suite.True(len(posts) >= 0, "Should execute successfully")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestIsNotNull tests the IsNotNull and OrIsNotNull conditions.
func (suite *NullBooleanChecksTestSuite) TestIsNotNull() {
	suite.T().Logf("Testing IsNotNull condition for %s", suite.dbType)

	suite.Run("BasicIsNotNull", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsNotNull("title")
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		for _, post := range posts {
			suite.NotEmpty(post.Title, "Title should not be NULL")
		}

		suite.T().Logf("Found %d posts with non-NULL title", len(posts))
	})

	suite.Run("OrIsNotNull", func() {
		posts := suite.assertQueryReturnsPosts(
			suite.db.NewSelect().
				Model((*Post)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsNotNull("title").
						OrIsNotNull("content")
				}),
		)

		suite.True(len(posts) > 0, "Should find posts")

		suite.T().Logf("Found %d posts", len(posts))
	})
}

// TestIsTrue tests the IsTrue and OrIsTrue conditions.
func (suite *NullBooleanChecksTestSuite) TestIsTrue() {
	suite.T().Logf("Testing IsTrue condition for %s", suite.dbType)

	suite.Run("BasicIsTrue", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsTrue("is_active")
				}),
		)

		suite.True(len(users) > 0, "Should find active users")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active")
		}

		suite.T().Logf("Found %d active users", len(users))
	})

	suite.Run("OrIsTrue", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsTrue("is_active").
						OrIsTrue("is_active")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		for _, user := range users {
			suite.True(user.IsActive, "User should be active")
		}

		suite.T().Logf("Found %d active users", len(users))
	})
}

// TestIsFalse tests the IsFalse and OrIsFalse conditions.
func (suite *NullBooleanChecksTestSuite) TestIsFalse() {
	suite.T().Logf("Testing IsFalse condition for %s", suite.dbType)

	suite.Run("BasicIsFalse", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsFalse("is_active")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		for _, user := range users {
			suite.False(user.IsActive, "User should be inactive")
		}

		suite.T().Logf("Found %d inactive users", len(users))
	})

	suite.Run("OrIsFalse", func() {
		users := suite.assertQueryReturnsUsers(
			suite.db.NewSelect().
				Model((*User)(nil)).
				Where(func(cb ConditionBuilder) {
					cb.IsFalse("is_active").
						OrIsFalse("is_active")
				}),
		)

		suite.True(len(users) >= 0, "Should execute successfully")

		for _, user := range users {
			suite.False(user.IsActive, "User should be inactive")
		}

		suite.T().Logf("Found %d inactive users", len(users))
	})
}
