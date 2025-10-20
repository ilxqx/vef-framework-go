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
type TestCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptionsApi[TestCategory, TestCategorySearch]
}

func NewTestCategoryFindTreeOptionsResource() api.Resource {
	return &TestCategoryFindTreeOptionsResource{
		Resource: api.NewResource("test/category_tree_options"),
		FindTreeOptionsApi: apis.NewFindTreeOptionsApi[TestCategory, TestCategorySearch]().
			Public().
			ColumnMapping(&apis.TreeOptionColumnMapping{
				OptionColumnMapping: apis.OptionColumnMapping{
					LabelColumn: "name",
					ValueColumn: "id",
				},
				IdColumn:       "id",
				ParentIdColumn: "parent_id",
			}),
	}
}

// Resource with custom field mapping.
type CustomFieldCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptionsApi[TestCategory, TestCategorySearch]
}

func NewCustomFieldCategoryFindTreeOptionsResource() api.Resource {
	return &CustomFieldCategoryFindTreeOptionsResource{
		Resource: api.NewResource("test/category_tree_options_custom"),
		FindTreeOptionsApi: apis.NewFindTreeOptionsApi[TestCategory, TestCategorySearch]().
			Public().
			ColumnMapping(&apis.TreeOptionColumnMapping{
				OptionColumnMapping: apis.OptionColumnMapping{
					LabelColumn:       "code",
					ValueColumn:       "id",
					DescriptionColumn: "description",
				},
				IdColumn:       "id",
				ParentIdColumn: "parent_id",
			}),
	}
}

// Filtered Tree Options Resource.
type FilteredCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptionsApi[TestCategory, TestCategorySearch]
}

func NewFilteredCategoryFindTreeOptionsResource() api.Resource {
	return &FilteredCategoryFindTreeOptionsResource{
		Resource: api.NewResource("test/category_tree_options_filtered"),
		FindTreeOptionsApi: apis.NewFindTreeOptionsApi[TestCategory, TestCategorySearch]().
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

// FindTreeOptionsTestSuite is the test suite for FindTreeOptions Api tests.
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
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_options",
			Action:   "find_tree_options",
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

	// Verify default ordering by created_at DESC - Clothing (latest) should be first
	first := suite.readDataAsMap(tree[0])
	suite.Equal("Clothing", first["label"])
	suite.NotEmpty(first["value"])
	suite.NotEmpty(first["id"])

	second := suite.readDataAsMap(tree[1])
	suite.Equal("Books", second["label"])

	third := suite.readDataAsMap(tree[2])
	suite.Equal("Electronics", third["label"])

	// Check first option (Clothing) has children
	children := suite.readDataAsSlice(first["children"])
	suite.Len(children, 2) // Men and Women

	// Check child option structure
	childOption := suite.readDataAsMap(children[0])
	suite.NotEmpty(childOption["label"])
	suite.NotEmpty(childOption["value"])
	suite.NotEmpty(childOption["parentId"])
}

// TestFindTreeOptionsWithConfig tests FindTreeOptions with custom config.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithConfig() {
	suite.Run("DefaultConfig", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
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
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
				Version:  "v1",
			},
			Params: map[string]any{
				"labelColumn": "code",
				"valueColumn": "id",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3)

		// Verify code is used as label and ordering by created_at DESC
		first := suite.readDataAsMap(tree[0])
		suite.Equal("clothing", first["label"])

		second := suite.readDataAsMap(tree[1])
		suite.Equal("books", second["label"])

		third := suite.readDataAsMap(tree[2])
		suite.Equal("electronics", third["label"])
	})

	suite.Run("WithDescription", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options_custom",
				Action:   "find_tree_options",
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
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
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
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
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
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_options_filtered",
			Action:   "find_tree_options",
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
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
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
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
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
