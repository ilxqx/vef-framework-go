package orm

// ConditionBuilderTestSuite is the base test suite for all condition builder tests.
// It provides common helper methods and test utilities for condition testing.
type ConditionBuilderTestSuite struct {
	*OrmTestSuite
}

// Helper methods for common test patterns

// assertQueryReturnsUsers executes a query and returns the users for further assertions.
func (suite *ConditionBuilderTestSuite) assertQueryReturnsUsers(query SelectQuery) []User {
	var users []User

	err := query.Scan(suite.ctx, &users)
	suite.NoError(err, "Query should execute successfully")

	return users
}

// assertQueryReturnsPosts executes a query and returns the posts for further assertions.
func (suite *ConditionBuilderTestSuite) assertQueryReturnsPosts(query SelectQuery) []Post {
	var posts []Post

	err := query.Scan(suite.ctx, &posts)
	suite.NoError(err, "Query should execute successfully")

	return posts
}
