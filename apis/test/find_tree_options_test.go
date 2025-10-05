package test

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// Test Resources.
type TestCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptionsAPI[TestCategory, TestCategorySearch]
}

func NewTestCategoryFindTreeOptionsResource() api.Resource {
	return &TestCategoryFindTreeOptionsResource{
		Resource: api.NewResource("test/category_tree_options"),
		FindTreeOptionsAPI: apis.NewFindTreeOptionsAPI[TestCategory, TestCategorySearch]().
			Public().
			FieldMapping(&apis.TreeOptionFieldMapping{
				OptionFieldMapping: apis.OptionFieldMapping{
					LabelField: "name",
					ValueField: "id",
				},
				IdField:       "id",
				ParentIdField: "parent_id",
			}),
	}
}

// Resource with custom field mapping.
type CustomFieldCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptionsAPI[TestCategory, TestCategorySearch]
}

func NewCustomFieldCategoryFindTreeOptionsResource() api.Resource {
	return &CustomFieldCategoryFindTreeOptionsResource{
		Resource: api.NewResource("test/category_tree_options_custom"),
		FindTreeOptionsAPI: apis.NewFindTreeOptionsAPI[TestCategory, TestCategorySearch]().
			Public().
			FieldMapping(&apis.TreeOptionFieldMapping{
				OptionFieldMapping: apis.OptionFieldMapping{
					LabelField:       "code",
					ValueField:       "id",
					DescriptionField: "description",
				},
				IdField:       "id",
				ParentIdField: "parent_id",
			}),
	}
}

// Filtered Tree Options Resource.
type FilteredCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptionsAPI[TestCategory, TestCategorySearch]
}

func NewFilteredCategoryFindTreeOptionsResource() api.Resource {
	return &FilteredCategoryFindTreeOptionsResource{
		Resource: api.NewResource("test/category_tree_options_filtered"),
		FindTreeOptionsAPI: apis.NewFindTreeOptionsAPI[TestCategory, TestCategorySearch]().
			Public().
			FilterApplier(func(search TestCategorySearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder] {
				return func(cb orm.ConditionBuilder) {
					// Only show Books and its children
					cb.Group(func(cb orm.ConditionBuilder) {
						cb.OrEquals("id", "cat002")
						cb.OrEquals("parent_id", "cat002")
					})
				}
			}),
	}
}

// FindTreeOptionsTestSuite is the test suite for FindTreeOptions API tests.
type FindTreeOptionsTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindTreeOptionsTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestCategoryFindTreeOptionsResource,
		NewCustomFieldCategoryFindTreeOptionsResource,
		NewFilteredCategoryFindTreeOptionsResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindTreeOptionsTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindTreeOptionsBasic tests basic FindTreeOptions functionality.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_options",
			Action:   "findTreeOptions",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	suite.NotNil(body.Data)

	tree := suite.readDataAsSlice(body.Data)
	// Should return 3 root categories
	suite.Len(tree, 3)

	// Verify default ordering by created_at ASC - Electronics (earliest) should be first
	first := suite.readDataAsMap(tree[0])
	suite.Equal("Electronics", first["label"])
	suite.NotEmpty(first["value"])
	suite.NotEmpty(first["id"])

	second := suite.readDataAsMap(tree[1])
	suite.Equal("Books", second["label"])

	third := suite.readDataAsMap(tree[2])
	suite.Equal("Clothing", third["label"])

	// Check first option (Electronics) has children
	children := suite.readDataAsSlice(first["children"])
	suite.Len(children, 2) // Computers and Phones

	// Check child option structure
	computers := suite.readDataAsMap(children[0])
	suite.NotEmpty(computers["label"])
	suite.NotEmpty(computers["value"])
	suite.NotEmpty(computers["parentId"])
}

// TestFindTreeOptionsWithConfig tests FindTreeOptions with custom config.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithConfig() {
	suite.Run("DefaultConfig", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "findTreeOptions",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3)
	})

	suite.Run("CustomConfig", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "findTreeOptions",
				Version:  "v1",
			},
			Params: map[string]any{
				"labelField": "code",
				"valueField": "id",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3)

		// Verify code is used as label and ordering by created_at ASC
		first := suite.readDataAsMap(tree[0])
		suite.Equal("electronics", first["label"])

		second := suite.readDataAsMap(tree[1])
		suite.Equal("books", second["label"])

		third := suite.readDataAsMap(tree[2])
		suite.Equal("clothing", third["label"])
	})

	suite.Run("WithDescription", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options_custom",
				Action:   "findTreeOptions",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3)

		// Verify description is included
		electronics := suite.readDataAsMap(tree[0])
		suite.NotEmpty(electronics["description"])
	})
}

// TestFindTreeOptionsWithSearch tests FindTreeOptions with search conditions.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithSearch() {
	suite.Run("SearchByCode", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "findTreeOptions",
				Version:  "v1",
			},
			Params: map[string]any{
				"code": "books",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 1) // Only Books

		books := suite.readDataAsMap(tree[0])
		suite.Equal("Books", books["label"])
	})

	suite.Run("SearchByKeyword", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "findTreeOptions",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Laptop",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.GreaterOrEqual(len(tree), 1)
	})
}

// TestFindTreeOptionsWithFilterApplier tests FindTreeOptions with filter applier.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithFilterApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_options_filtered",
			Action:   "findTreeOptions",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	tree := suite.readDataAsSlice(body.Data)
	// Should only return Books and its children
	suite.Len(tree, 1) // Only Books root

	books := suite.readDataAsMap(tree[0])
	suite.Equal("Books", books["label"])

	children := suite.readDataAsSlice(books["children"])
	suite.Len(children, 2) // Fiction and Non-Fiction
}

// TestFindTreeOptionsNegativeCases tests negative scenarios.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsNegativeCases() {
	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "findTreeOptions",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "NonexistentCategory",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 0)
	})

	suite.Run("InvalidFieldName", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "findTreeOptions",
				Version:  "v1",
			},
			Params: map[string]any{
				"labelField": "nonexistent_field",
				"valueField": "id",
			},
		})

		// Should return error for invalid field
		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})
}
