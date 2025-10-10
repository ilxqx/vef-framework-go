package apis_test

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// Test Resources.
type TestUserFindOptionsResource struct {
	api.Resource
	apis.FindOptionsAPI[TestUser, TestUserSearch]
}

func NewTestUserFindOptionsResource() api.Resource {
	return &TestUserFindOptionsResource{
		Resource: api.NewResource("test/user_options"),
		FindOptionsAPI: apis.NewFindOptionsAPI[TestUser, TestUserSearch]().
			Public().
			ColumnMapping(&apis.OptionColumnMapping{
				LabelColumn: "name",
				ValueColumn: "id",
			}),
	}
}

// Resource with custom field mapping.
type CustomFieldUserFindOptionsResource struct {
	api.Resource
	apis.FindOptionsAPI[TestUser, TestUserSearch]
}

func NewCustomFieldUserFindOptionsResource() api.Resource {
	return &CustomFieldUserFindOptionsResource{
		Resource: api.NewResource("test/user_options_custom"),
		FindOptionsAPI: apis.NewFindOptionsAPI[TestUser, TestUserSearch]().
			Public().
			ColumnMapping(&apis.OptionColumnMapping{
				LabelColumn:       "email",
				ValueColumn:       "id",
				DescriptionColumn: "description",
			}),
	}
}

// Filtered Options Resource.
type FilteredUserFindOptionsResource struct {
	api.Resource
	apis.FindOptionsAPI[TestUser, TestUserSearch]
}

func NewFilteredUserFindOptionsResource() api.Resource {
	return &FilteredUserFindOptionsResource{
		Resource: api.NewResource("test/user_options_filtered"),
		FindOptionsAPI: apis.NewFindOptionsAPI[TestUser, TestUserSearch]().
			Public().
			FilterApplier(func(search TestUserSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder] {
				return func(cb orm.ConditionBuilder) {
					cb.Equals("status", "active")
				}
			}),
	}
}

// FindOptionsTestSuite is the test suite for FindOptions API tests.
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
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_options",
			Action:   "findOptions",
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
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "findOptions",
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
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "findOptions",
				Version:  "v1",
			},
			Params: map[string]any{
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
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options_custom",
				Action:   "findOptions",
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
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "findOptions",
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
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "findOptions",
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
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_options_filtered",
			Action:   "findOptions",
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
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "findOptions",
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
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "findOptions",
				Version:  "v1",
			},
			Params: map[string]any{
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
