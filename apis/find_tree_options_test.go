package apis_test

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// Test Resources.
type TestCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptions[TestCategory, TestCategorySearch]
}

func NewTestCategoryFindTreeOptionsResource() api.Resource {
	return &TestCategoryFindTreeOptionsResource{
		Resource: api.NewRPCResource("test/category_tree_options"),
		FindTreeOptions: apis.NewFindTreeOptions[TestCategory, TestCategorySearch]().
			Public().
			WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
				LabelColumn: "name",
				ValueColumn: "id",
			}).
			WithIDColumn("id").
			WithParentIDColumn("parent_id"),
	}
}

// Resource with custom field mapping.
type CustomFieldCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptions[TestCategory, TestCategorySearch]
}

func NewCustomFieldCategoryFindTreeOptionsResource() api.Resource {
	return &CustomFieldCategoryFindTreeOptionsResource{
		Resource: api.NewRPCResource("test/category_tree_options_custom"),
		FindTreeOptions: apis.NewFindTreeOptions[TestCategory, TestCategorySearch]().
			Public().
			WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
				LabelColumn:       "code",
				ValueColumn:       "id",
				DescriptionColumn: "description",
			}).
			WithIDColumn("id").
			WithParentIDColumn("parent_id"),
	}
}

// Filtered Tree Options Resource.
type FilteredCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptions[TestCategory, TestCategorySearch]
}

func NewFilteredCategoryFindTreeOptionsResource() api.Resource {
	return &FilteredCategoryFindTreeOptionsResource{
		Resource: api.NewRPCResource("test/category_tree_options_filtered"),
		FindTreeOptions: apis.NewFindTreeOptions[TestCategory, TestCategorySearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				// Only show Books and its children
				cb.Group(func(cb orm.ConditionBuilder) {
					cb.OrEquals("id", "cat002")
					cb.OrEquals("parent_id", "cat002")
				})
			}).
			Public(),
	}
}

// Meta Tree Options Resource.
type MetaCategoryFindTreeOptionsResource struct {
	api.Resource
	apis.FindTreeOptions[TestCategory, TestCategorySearch]
}

func NewMetaCategoryFindTreeOptionsResource() api.Resource {
	return &MetaCategoryFindTreeOptionsResource{
		Resource: api.NewRPCResource("test/category_tree_options_meta"),
		FindTreeOptions: apis.NewFindTreeOptions[TestCategory, TestCategorySearch]().
			Public().
			WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
				LabelColumn: "name",
				ValueColumn: "id",
				MetaColumns: []string{"code", "description"},
			}).
			WithIDColumn("id").
			WithParentIDColumn("parent_id"),
	}
}

// FindTreeOptionsTestSuite tests the FindTreeOptions API functionality
// including basic tree options, custom column mappings, search filters, filter appliers, meta columns, and tree-specific features.
type FindTreeOptionsTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindTreeOptionsTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestCategoryFindTreeOptionsResource,
		NewCustomFieldCategoryFindTreeOptionsResource,
		NewFilteredCategoryFindTreeOptionsResource,
		NewMetaCategoryFindTreeOptionsResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindTreeOptionsTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindTreeOptionsBasic tests basic FindTreeOptions functionality.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsBasic() {
	suite.T().Logf("Testing FindTreeOptions API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_options",
			Action:   "find_tree_options",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	suite.NotNil(body.Data, "Data should not be nil")

	tree := suite.readDataAsSlice(body.Data)
	suite.Len(tree, 3, "Should return 3 root categories")

	// Verify default ordering by id DESC - Clothing (latest) should be first
	first := suite.readDataAsMap(tree[0])
	suite.Equal("Clothing", first["label"], "First category should be Clothing")
	suite.NotEmpty(first["value"], "First option should have value")

	second := suite.readDataAsMap(tree[1])
	suite.Equal("Books", second["label"], "Second category should be Books")

	third := suite.readDataAsMap(tree[2])
	suite.Equal("Electronics", third["label"], "Third category should be Electronics")

	// Check first option (Clothing) has children
	children := suite.readDataAsSlice(first["children"])
	suite.Len(children, 2, "Clothing should have 2 children (Men and Women)")

	// Check child option structure
	childOption := suite.readDataAsMap(children[0])
	suite.NotEmpty(childOption["label"], "Child option should have label")
	suite.NotEmpty(childOption["value"], "Child option should have value")

	suite.T().Logf("Found %d root categories with %d children in first category", len(tree), len(children))
}

// TestFindTreeOptionsWithConfig tests FindTreeOptions with custom config.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithConfig() {
	suite.T().Logf("Testing FindTreeOptions API with custom config for %s", suite.dbType)

	suite.Run("DefaultConfig", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return 3 root categories with default config")

		suite.T().Logf("Found %d root categories with default config (label=name, value=id)", len(tree))
	})

	suite.Run("CustomConfig", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"labelColumn": "code",
				"valueColumn": "id",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return 3 root categories with custom config")

		// Verify code is used as label and ordering by created_at DESC
		first := suite.readDataAsMap(tree[0])
		suite.Equal("clothing", first["label"], "First category should use code 'clothing' as label")

		second := suite.readDataAsMap(tree[1])
		suite.Equal("books", second["label"], "Second category should use code 'books' as label")

		third := suite.readDataAsMap(tree[2])
		suite.Equal("electronics", third["label"], "Third category should use code 'electronics' as label")

		suite.T().Logf("Found %d root categories with custom config (label=code: %s, %s, %s)", len(tree), first["label"], second["label"], third["label"])
	})

	suite.Run("WithDescription", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options_custom",
				Action:   "find_tree_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return 3 root categories with description field")

		// Verify description is included
		electronics := suite.readDataAsMap(tree[0])
		suite.NotEmpty(electronics["description"], "First option should have description field")

		suite.T().Logf("Found %d root categories with description column (description: %v)", len(tree), electronics["description"])
	})
}

// TestFindTreeOptionsWithSearch tests FindTreeOptions with search conditions.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithSearch() {
	suite.T().Logf("Testing FindTreeOptions API with search conditions for %s", suite.dbType)

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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 1, "Should return only Books category")

		books := suite.readDataAsMap(tree[0])
		suite.Equal("Books", books["label"], "Category label should be Books")

		suite.T().Logf("Found 1 category matching code 'books': %s", books["label"])
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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.GreaterOrEqual(len(tree), 1, "Should return at least 1 category matching keyword 'Laptop'")

		suite.T().Logf("Found %d categories matching keyword 'Laptop'", len(tree))
	})
}

// TestFindTreeOptionsWithFilterApplier tests FindTreeOptions with filter applier.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithFilterApplier() {
	suite.T().Logf("Testing FindTreeOptions API with filter applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_options_filtered",
			Action:   "find_tree_options",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	tree := suite.readDataAsSlice(body.Data)
	suite.Len(tree, 1, "Should return only Books root (filtered to Books and its children)")

	books := suite.readDataAsMap(tree[0])
	suite.Equal("Books", books["label"], "Root category should be Books")

	children := suite.readDataAsSlice(books["children"])
	suite.Len(children, 2, "Books should have 2 children (Fiction and Non-Fiction)")

	suite.T().Logf("Found 1 filtered tree rooted at Books with %d children", len(children))
}

// TestFindTreeOptionsNegativeCases tests negative scenarios.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsNegativeCases() {
	suite.T().Logf("Testing FindTreeOptions API negative cases for %s", suite.dbType)

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

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 0, "Should return empty tree when no records match")

		suite.T().Logf("No matching records found as expected for keyword 'NonexistentCategory'")
	})

	suite.Run("InvalidFieldName", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
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

// TestFindTreeOptionsWithMeta tests FindTreeOptions with meta columns.
func (suite *FindTreeOptionsTestSuite) TestFindTreeOptionsWithMeta() {
	suite.T().Logf("Testing FindTreeOptions API with meta columns for %s", suite.dbType)

	suite.Run("DefaultMetaColumns", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options_meta",
				Action:   "find_tree_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return 3 root categories with default meta columns")

		// Verify meta field exists and contains expected keys
		firstOption := suite.readDataAsMap(tree[0])
		meta, ok := firstOption["meta"].(map[string]any)
		suite.True(ok, "meta should be a map")
		suite.NotNil(meta, "meta should not be nil")
		suite.Contains(meta, "code", "meta should contain code field")
		suite.Contains(meta, "description", "meta should contain description field")

		suite.T().Logf("Found %d root categories with default meta columns (code, description)", len(tree))
	})

	suite.Run("CustomMetaColumns", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"metaColumns": []string{"code"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return 3 root categories with custom meta columns")

		// Verify meta field contains custom columns
		firstOption := suite.readDataAsMap(tree[0])
		meta, ok := firstOption["meta"].(map[string]any)
		suite.True(ok, "meta should be a map")
		suite.NotNil(meta, "meta should not be nil")
		suite.Contains(meta, "code", "meta should contain code field")

		suite.T().Logf("Found %d root categories with custom meta columns (code)", len(tree))
	})

	suite.Run("MetaColumnsWithAlias", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
				Version:  "v1",
			},
			Meta: map[string]any{
				"metaColumns": []string{"code AS category_code", "description as desc"},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return 3 root categories with aliased meta columns")

		// Verify alias is used in meta field
		firstOption := suite.readDataAsMap(tree[0])
		meta, ok := firstOption["meta"].(map[string]any)
		suite.True(ok, "meta should be a map")
		suite.NotNil(meta, "meta should not be nil")
		suite.Contains(meta, "category_code", "meta should contain category_code field (aliased from code)")
		suite.Contains(meta, "desc", "meta should contain desc field (aliased from description)")
		suite.NotContains(meta, "code", "meta should not contain original code field when aliased")
		suite.NotContains(meta, "description", "meta should not contain original description field when aliased")

		suite.T().Logf("Found %d root categories with aliased meta columns (code AS category_code, description AS desc)", len(tree))
	})

	suite.Run("VerifyMetaInChildren", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options_meta",
				Action:   "find_tree_options",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return 3 root categories")

		// Check that children also have meta field
		firstOption := suite.readDataAsMap(tree[0])
		children := suite.readDataAsSlice(firstOption["children"])
		suite.Greater(len(children), 0, "First category should have children")

		// Verify child has meta field
		childOption := suite.readDataAsMap(children[0])
		childMeta, ok := childOption["meta"].(map[string]any)
		suite.True(ok, "child meta should be a map")
		suite.NotNil(childMeta, "child meta should not be nil")
		suite.Contains(childMeta, "code", "child meta should contain code field")
		suite.Contains(childMeta, "description", "child meta should contain description field")

		suite.T().Logf("Verified meta columns in %d root categories and their %d children", len(tree), len(children))
	})

	suite.Run("InvalidMetaColumn", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree_options",
				Action:   "find_tree_options",
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
