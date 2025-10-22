package apis_test

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// Test Resources.
type TestUserFindOptionsResource struct {
	api.Resource
	apis.FindOptionsApi[TestUser, TestUserSearch]
}

func NewTestUserFindOptionsResource() api.Resource {
	return &TestUserFindOptionsResource{
		Resource: api.NewResource("test/user_options"),
		FindOptionsApi: apis.NewFindOptionsApi[TestUser, TestUserSearch]().
			Public().
			WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
				LabelColumn: "name",
				ValueColumn: "id",
			}),
	}
}

// Resource with custom field mapping.
type CustomFieldUserFindOptionsResource struct {
	api.Resource
	apis.FindOptionsApi[TestUser, TestUserSearch]
}

func NewCustomFieldUserFindOptionsResource() api.Resource {
	return &CustomFieldUserFindOptionsResource{
		Resource: api.NewResource("test/user_options_custom"),
		FindOptionsApi: apis.NewFindOptionsApi[TestUser, TestUserSearch]().
			Public().
			WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
				LabelColumn:       "email",
				ValueColumn:       "id",
				DescriptionColumn: "description",
			}),
	}
}

// Filtered Options Resource.
type FilteredUserFindOptionsResource struct {
	api.Resource
	apis.FindOptionsApi[TestUser, TestUserSearch]
}

func NewFilteredUserFindOptionsResource() api.Resource {
	return &FilteredUserFindOptionsResource{
		Resource: api.NewResource("test/user_options_filtered"),
		FindOptionsApi: apis.NewFindOptionsApi[TestUser, TestUserSearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				cb.Equals("status", "active")
			}).
			Public(),
	}
}

// FindOptionsTestSuite is the test suite for FindOptions Api tests.
type FindOptionsTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindOptionsTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserFindOptionsResource,
		NewCustomFieldUserFindOptionsResource,
		NewFilteredUserFindOptionsResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindOptionsTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindOptionsBasic tests basic FindOptions functionality.
func (suite *FindOptionsTestSuite) TestFindOptionsBasic() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_options",
			Action:   "find_options",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	suite.NotNil(body.Data)

	options := suite.readDataAsSlice(body.Data)
	suite.Len(options, 10)

	// Check first option structure
	firstOption := suite.readDataAsMap(options[0])
	suite.NotEmpty(firstOption["label"])
	suite.NotEmpty(firstOption["value"])
}

// TestFindOptionsWithConfig tests FindOptions with custom config.
func (suite *FindOptionsTestSuite) TestFindOptionsWithConfig() {
	suite.Run("DefaultConfig", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10)
	})

	suite.Run("CustomConfig", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"labelColumn": "email",
				"valueColumn": "id",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10)

		// Verify email is used as label
		firstOption := suite.readDataAsMap(options[0])
		label, ok := firstOption["label"].(string)
		suite.True(ok)
		suite.Contains(label, "@") // Email should contain @
	})

	suite.Run("WithDescription", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options_custom",
				Action:   "find_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10)

		// Verify description is included
		firstOption := suite.readDataAsMap(options[0])
		suite.NotEmpty(firstOption["description"])
	})
}

// TestFindOptionsWithSearch tests FindOptions with search conditions.
func (suite *FindOptionsTestSuite) TestFindOptionsWithSearch() {
	suite.Run("SearchByStatus", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 7) // 7 active users
	})

	suite.Run("SearchByKeyword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Johnson",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 1) // Only Alice Johnson
	})
}

// TestFindOptionsWithFilterApplier tests FindOptions with filter applier.
func (suite *FindOptionsTestSuite) TestFindOptionsWithFilterApplier() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_options_filtered",
			Action:   "find_options",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	options := suite.readDataAsSlice(body.Data)
	suite.Len(options, 7) // Only active users
}

// TestFindOptionsNegativeCases tests negative scenarios.
func (suite *FindOptionsTestSuite) TestFindOptionsNegativeCases() {
	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "NonexistentKeyword",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 0)
	})

	suite.Run("InvalidFieldName", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"labelColumn": "nonexistent_field",
				"valueColumn": "id",
			},
		})

		// Should return error for invalid field
		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})
}
