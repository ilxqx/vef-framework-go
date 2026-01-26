package apis_test

import (
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/sort"
	"github.com/ilxqx/vef-framework-go/treebuilder"
)

// Tree builder for TestCategory.
func buildCategoryTree(flatCategories []TestCategory) []TestCategory {
	adapter := treebuilder.Adapter[TestCategory]{
		GetID: func(c TestCategory) string {
			return c.ID
		},
		GetParentID: func(c TestCategory) string {
			return lo.FromPtrOr(c.ParentID, constants.Empty)
		},
		SetChildren: func(c *TestCategory, children []TestCategory) {
			c.Children = children
		},
	}

	return treebuilder.Build(flatCategories, adapter)
}

// Test Resources.
type TestCategoryFindTreeResource struct {
	api.Resource
	apis.FindTree[TestCategory, TestCategorySearch]
}

func NewTestCategoryFindTreeResource() api.Resource {
	return &TestCategoryFindTreeResource{
		Resource: api.NewRPCResource("test/category_tree"),
		FindTree: apis.NewFindTree[TestCategory, TestCategorySearch](buildCategoryTree).
			Public().
			WithIDColumn("id").
			WithParentIDColumn("parent_id"),
	}
}

// Filtered Tree Resource.
type FilteredCategoryFindTreeResource struct {
	api.Resource
	apis.FindTree[TestCategory, TestCategorySearch]
}

func NewFilteredCategoryFindTreeResource() api.Resource {
	return &FilteredCategoryFindTreeResource{
		Resource: api.NewRPCResource("test/category_tree_filtered"),
		FindTree: apis.NewFindTree[TestCategory, TestCategorySearch](buildCategoryTree).
			WithCondition(func(cb orm.ConditionBuilder) {
				// Only show Electronics and its children
				cb.Group(func(cb orm.ConditionBuilder) {
					cb.OrEquals("id", "cat001")
					cb.OrEquals("parent_id", "cat001")
				})
			}).
			Public().
			WithIDColumn("id").
			WithParentIDColumn("parent_id"),
	}
}

// Ordered Tree Resource.
type OrderedCategoryFindTreeResource struct {
	api.Resource
	apis.FindTree[TestCategory, TestCategorySearch]
}

func NewOrderedCategoryFindTreeResource() api.Resource {
	return &OrderedCategoryFindTreeResource{
		Resource: api.NewRPCResource("test/category_tree_ordered"),
		FindTree: apis.NewFindTree[TestCategory, TestCategorySearch](buildCategoryTree).
			WithDefaultSort(&sort.OrderSpec{
				Column: "sort",
			}).
			Public().
			WithIDColumn("id").
			WithParentIDColumn("parent_id"),
	}
}

// AuditUser Tree Resource - with audit user names.
type AuditUserCategoryFindTreeResource struct {
	api.Resource
	apis.FindTree[TestCategory, TestCategorySearch]
}

func NewAuditUserCategoryFindTreeResource() api.Resource {
	return &AuditUserCategoryFindTreeResource{
		Resource: api.NewRPCResource("test/category_tree_audit"),
		FindTree: apis.NewFindTree[TestCategory, TestCategorySearch](buildCategoryTree).
			Public().
			WithIDColumn("id").
			WithParentIDColumn("parent_id").
			WithAuditUserNames((*TestAuditUser)(nil)),
	}
}

// FindTreeTestSuite tests the FindTree API functionality
// including basic tree building, search filters, filter appliers, sort appliers, audit user names, and negative cases.
type FindTreeTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindTreeTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestCategoryFindTreeResource,
		NewFilteredCategoryFindTreeResource,
		NewOrderedCategoryFindTreeResource,
		NewAuditUserCategoryFindTreeResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindTreeTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindTreeBasic tests basic FindTree functionality.
func (suite *FindTreeTestSuite) TestFindTreeBasic() {
	suite.T().Logf("Testing FindTree API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree",
			Action:   "find_tree",
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

	// Verify default ordering by created_at DESC - Clothing (latest) should be first
	first := suite.readDataAsMap(tree[0])
	suite.Equal("Clothing", first["name"], "First category should be Clothing (latest created_at)")

	second := suite.readDataAsMap(tree[1])
	suite.Equal("Books", second["name"], "Second category should be Books")

	third := suite.readDataAsMap(tree[2])
	suite.Equal("Electronics", third["name"], "Third category should be Electronics (earliest created_at)")

	// Check Electronics has children (it's the third item due to DESC ordering)
	electronics := third
	children := suite.readDataAsSlice(electronics["children"])
	suite.Len(children, 2, "Electronics should have 2 children (Computers and Phones)")

	suite.T().Logf("Found %d root categories with default ordering (DESC by created_at)", len(tree))
}

// TestFindTreeWithSearch tests FindTree with search conditions.
func (suite *FindTreeTestSuite) TestFindTreeWithSearch() {
	suite.T().Logf("Testing FindTree API with search filters for %s", suite.dbType)

	suite.Run("SearchByCode", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "find_tree",
				Version:  "v1",
			},
			Params: map[string]any{
				"code": "electronics",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 1, "Should return only Electronics category")

		electronics := suite.readDataAsMap(tree[0])
		suite.Equal("Electronics", electronics["name"], "Category name should be Electronics")

		suite.T().Logf("Found 1 category matching code 'electronics': %s", electronics["name"])
	})

	suite.Run("SearchByParentID", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "find_tree",
				Version:  "v1",
			},
			Params: map[string]any{
				"parentId": "cat001", // Electronics' children
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 1, "Should return 1 tree with Electronics as root (recursive CTE finds ancestors)")

		electronics := suite.readDataAsMap(tree[0])
		suite.Equal("Electronics", electronics["name"], "Root should be Electronics")

		// Electronics should have 2 children: Computers and Phones
		children := suite.readDataAsSlice(electronics["children"])
		suite.Len(children, 2, "Electronics should have 2 children (Computers and Phones)")

		suite.T().Logf("Found tree rooted at Electronics with %d children", len(children))
	})

	suite.Run("SearchByKeyword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "find_tree",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Computer",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.GreaterOrEqual(len(tree), 1, "Should return at least 1 category matching keyword 'Computer'")

		suite.T().Logf("Found %d categories matching keyword 'Computer'", len(tree))
	})
}

// TestFindTreeWithFilterApplier tests FindTree with filter applier.
func (suite *FindTreeTestSuite) TestFindTreeWithFilterApplier() {
	suite.T().Logf("Testing FindTree API with filter applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_filtered",
			Action:   "find_tree",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	tree := suite.readDataAsSlice(body.Data)
	suite.Len(tree, 1, "Should return only Electronics root (filtered to Electronics and its children)")

	electronics := suite.readDataAsMap(tree[0])
	suite.Equal("Electronics", electronics["name"], "Root category should be Electronics")

	children := suite.readDataAsSlice(electronics["children"])
	suite.Len(children, 2, "Electronics should have 2 children (Computers and Phones)")

	suite.T().Logf("Found 1 filtered tree rooted at Electronics with %d children", len(children))
}

// TestFindTreeWithSortApplier tests FindTree with sort applier.
func (suite *FindTreeTestSuite) TestFindTreeWithSortApplier() {
	suite.T().Logf("Testing FindTree API with sort applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_ordered",
			Action:   "find_tree",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	tree := suite.readDataAsSlice(body.Data)
	suite.Len(tree, 3, "Should return 3 root categories")

	// Verify ordering by sort field
	first := suite.readDataAsMap(tree[0])
	suite.Equal("Electronics", first["name"], "First category should be Electronics (sort=1)")

	second := suite.readDataAsMap(tree[1])
	suite.Equal("Books", second["name"], "Second category should be Books (sort=2)")

	third := suite.readDataAsMap(tree[2])
	suite.Equal("Clothing", third["name"], "Third category should be Clothing (sort=3)")

	suite.T().Logf("Found 3 root categories ordered by sort field: Electronics, Books, Clothing")
}

// TestFindTreeNegativeCases tests negative scenarios.
func (suite *FindTreeTestSuite) TestFindTreeNegativeCases() {
	suite.T().Logf("Testing FindTree API negative cases for %s", suite.dbType)

	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "find_tree",
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

	suite.Run("EmptySearchCriteria", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "find_tree",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3, "Should return all 3 root categories with empty search criteria")

		suite.T().Logf("Empty search criteria returned %d root categories", len(tree))
	})
}

// TestFindTreeWithAuditUserNames tests FindTree with audit user names populated.
func (suite *FindTreeTestSuite) TestFindTreeWithAuditUserNames() {
	suite.T().Logf("Testing FindTree API with audit user names for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_audit",
			Action:   "find_tree",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")

	tree := suite.readDataAsSlice(body.Data)
	suite.Len(tree, 3, "Should return 3 root categories")

	// Verify all categories have audit user names (they were all created/updated by 'test' user initially)
	// But our fixture data has created_by and updated_by set to 'test', not actual audit user IDs
	// So we'll verify the structure is correct
	rootCategoriesWithAudit := 0

	childCategoriesWithAudit := 0
	for _, c := range tree {
		category := suite.readDataAsMap(c)
		suite.NotNil(category["createdByName"], "Category %s should have createdByName", category["id"])
		suite.NotNil(category["updatedByName"], "Category %s should have updatedByName", category["id"])

		rootCategoriesWithAudit++

		// Check children if they exist
		if category["children"] != nil {
			children := suite.readDataAsSlice(category["children"])
			for _, ch := range children {
				child := suite.readDataAsMap(ch)
				suite.NotNil(child["createdByName"], "Child category %s should have createdByName", child["id"])
				suite.NotNil(child["updatedByName"], "Child category %s should have updatedByName", child["id"])

				childCategoriesWithAudit++
			}
		}
	}

	suite.T().Logf("Verified audit user names for %d root categories and %d child categories", rootCategoriesWithAudit, childCategoriesWithAudit)
}
