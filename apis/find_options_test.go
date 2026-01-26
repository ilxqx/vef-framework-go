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
	apis.FindOptions[TestUser, TestUserSearch]
}

func NewTestUserFindOptionsResource() api.Resource {
	return &TestUserFindOptionsResource{
		Resource: api.NewRPCResource("test/user_options"),
		FindOptions: apis.NewFindOptions[TestUser, TestUserSearch]().
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
	apis.FindOptions[TestUser, TestUserSearch]
}

func NewCustomFieldUserFindOptionsResource() api.Resource {
	return &CustomFieldUserFindOptionsResource{
		Resource: api.NewRPCResource("test/user_options_custom"),
		FindOptions: apis.NewFindOptions[TestUser, TestUserSearch]().
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
	apis.FindOptions[TestUser, TestUserSearch]
}

func NewFilteredUserFindOptionsResource() api.Resource {
	return &FilteredUserFindOptionsResource{
		Resource: api.NewRPCResource("test/user_options_filtered"),
		FindOptions: apis.NewFindOptions[TestUser, TestUserSearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				cb.Equals("status", "active")
			}).
			Public(),
	}
}

// Meta Options Resource.
type MetaUserFindOptionsResource struct {
	api.Resource
	apis.FindOptions[TestUser, TestUserSearch]
}

func NewMetaUserFindOptionsResource() api.Resource {
	return &MetaUserFindOptionsResource{
		Resource: api.NewRPCResource("test/user_options_meta"),
		FindOptions: apis.NewFindOptions[TestUser, TestUserSearch]().
			Public().
			WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
				LabelColumn: "name",
				ValueColumn: "id",
				MetaColumns: []string{"status", "email"},
			}),
	}
}

// FindOptionsTestSuite tests the FindOptions API functionality
// including basic queries, custom column mappings, search filters, filter appliers, meta columns, and negative cases.
type FindOptionsTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindOptionsTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserFindOptionsResource,
		NewCustomFieldUserFindOptionsResource,
		NewFilteredUserFindOptionsResource,
		NewMetaUserFindOptionsResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindOptionsTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindOptionsBasic tests basic FindOptions functionality.
func (suite *FindOptionsTestSuite) TestFindOptionsBasic() {
	suite.T().Logf("Testing FindOptions API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_options",
			Action:   "find_options",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	suite.NotNil(body.Data, "Data should not be nil")

	options := suite.readDataAsSlice(body.Data)
	suite.Len(options, 10, "Should return 10 options")

	// Check first option structure
	firstOption := suite.readDataAsMap(options[0])
	suite.NotEmpty(firstOption["label"], "First option should have label")
	suite.NotEmpty(firstOption["value"], "First option should have value")

	suite.T().Logf("Found %d options with label=%v, value=%v", len(options), firstOption["label"], firstOption["value"])
}

// TestFindOptionsWithConfig tests FindOptions with custom config.
func (suite *FindOptionsTestSuite) TestFindOptionsWithConfig() {
	suite.T().Logf("Testing FindOptions API with custom config for %s", suite.dbType)

	suite.Run("DefaultConfig", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10, "Should return 10 options with default config")

		suite.T().Logf("Found %d options with default config (label=name, value=id)", len(options))
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10, "Should return 10 options with custom config")

		// Verify email is used as label
		firstOption := suite.readDataAsMap(options[0])
		label, ok := firstOption["label"].(string)
		suite.True(ok, "Label should be a string")
		suite.Contains(label, "@", "Email label should contain @ symbol")

		suite.T().Logf("Found %d options with custom config (label=email: %s)", len(options), label)
	})

	suite.Run("WithDescription", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options_custom",
				Action:   "find_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10, "Should return 10 options with description field")

		// Verify description is included
		firstOption := suite.readDataAsMap(options[0])
		suite.NotEmpty(firstOption["description"], "First option should have description field")

		suite.T().Logf("Found %d options with description column (description: %v)", len(options), firstOption["description"])
	})
}

// TestFindOptionsWithSearch tests FindOptions with search conditions.
func (suite *FindOptionsTestSuite) TestFindOptionsWithSearch() {
	suite.T().Logf("Testing FindOptions API with search conditions for %s", suite.dbType)

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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 7, "Should return 7 active users")

		suite.T().Logf("Found %d options filtered by status=active", len(options))
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 1, "Should return only Alice Johnson")

		firstOption := suite.readDataAsMap(options[0])
		suite.T().Logf("Found %d option matching keyword 'Johnson' (label=%v)", len(options), firstOption["label"])
	})
}

// TestFindOptionsWithFilterApplier tests FindOptions with filter applier.
func (suite *FindOptionsTestSuite) TestFindOptionsWithFilterApplier() {
	suite.T().Logf("Testing FindOptions API with filter applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_options_filtered",
			Action:   "find_options",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	options := suite.readDataAsSlice(body.Data)
	suite.Len(options, 7, "Should return only active users filtered by condition")

	suite.T().Logf("Found %d options filtered by condition (status=active)", len(options))
}

// TestFindOptionsNegativeCases tests negative scenarios.
func (suite *FindOptionsTestSuite) TestFindOptionsNegativeCases() {
	suite.T().Logf("Testing FindOptions API negative cases for %s", suite.dbType)

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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 0, "Should return empty options when no records match")

		suite.T().Logf("No matching records found as expected for keyword 'NonexistentKeyword'")
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when label column does not exist")

		suite.T().Logf("Validation failed as expected for invalid label column 'nonexistent_field'")
	})
}

// TestFindOptionsWithMeta tests FindOptions with meta columns.
func (suite *FindOptionsTestSuite) TestFindOptionsWithMeta() {
	suite.T().Logf("Testing FindOptions API with meta columns for %s", suite.dbType)

	suite.Run("DefaultMetaColumns", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options_meta",
				Action:   "find_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10, "Should return 10 options with default meta columns")

		// Verify meta field exists and contains expected keys
		firstOption := suite.readDataAsMap(options[0])
		meta, ok := firstOption["meta"].(map[string]any)
		suite.True(ok, "meta should be a map")
		suite.NotNil(meta, "meta should not be nil")
		suite.Contains(meta, "status", "meta should contain status field")
		suite.Contains(meta, "email", "meta should contain email field")

		suite.T().Logf("Found %d options with default meta columns (status, email)", len(options))
	})

	suite.Run("CustomMetaColumns", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"metaColumns": []string{"status", "description"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10, "Should return 10 options with custom meta columns")

		// Verify meta field contains custom columns
		firstOption := suite.readDataAsMap(options[0])
		meta, ok := firstOption["meta"].(map[string]any)
		suite.True(ok, "meta should be a map")
		suite.NotNil(meta, "meta should not be nil")
		suite.Contains(meta, "status", "meta should contain status field")
		suite.Contains(meta, "description", "meta should contain description field")

		suite.T().Logf("Found %d options with custom meta columns (status, description)", len(options))
	})

	suite.Run("MetaColumnsWithAlias", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"metaColumns": []string{"status", "email AS contact"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		options := suite.readDataAsSlice(body.Data)
		suite.Len(options, 10, "Should return 10 options with aliased meta columns")

		// Verify alias is used in meta field
		firstOption := suite.readDataAsMap(options[0])
		meta, ok := firstOption["meta"].(map[string]any)
		suite.True(ok, "meta should be a map")
		suite.NotNil(meta, "meta should not be nil")
		suite.Contains(meta, "status", "meta should contain status field")
		suite.Contains(meta, "contact", "meta should contain contact field (aliased from email)")
		suite.NotContains(meta, "email", "meta should not contain original email field when aliased")

		suite.T().Logf("Found %d options with aliased meta columns (status, email AS contact)", len(options))
	})

	suite.Run("InvalidMetaColumn", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_options",
				Action:   "find_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"metaColumns": []string{"nonexistent_field"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when meta column does not exist")

		suite.T().Logf("Validation failed as expected for invalid meta column 'nonexistent_field'")
	})
}
