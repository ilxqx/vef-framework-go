package apis_test

import (
	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/treebuilder"
)

// Tree builder for TestCategory.
func buildCategoryTree(flatCategories []TestCategory) []TestCategory {
	adapter := treebuilder.Adapter[TestCategory]{
		GetId: func(c TestCategory) string {
			return c.Id
		},
		GetParentId: func(c TestCategory) string {
			return lo.FromPtrOr(c.ParentId, constants.Empty)
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
	apis.FindTreeAPI[TestCategory, TestCategorySearch]
}

func NewTestCategoryFindTreeResource() api.Resource {
	return &TestCategoryFindTreeResource{
		Resource: api.NewResource("test/category_tree"),
		FindTreeAPI: apis.NewFindTreeAPI[TestCategory, TestCategorySearch](buildCategoryTree).
			Public().
			IdColumn("id").
			ParentIdColumn("parent_id"),
	}
}

// Filtered Tree Resource.
type FilteredCategoryFindTreeResource struct {
	api.Resource
	apis.FindTreeAPI[TestCategory, TestCategorySearch]
}

func NewFilteredCategoryFindTreeResource() api.Resource {
	return &FilteredCategoryFindTreeResource{
		Resource: api.NewResource("test/category_tree_filtered"),
		FindTreeAPI: apis.NewFindTreeAPI[TestCategory, TestCategorySearch](buildCategoryTree).
			Public().
			IdColumn("id").
			ParentIdColumn("parent_id").
			FilterApplier(func(search TestCategorySearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder] {
				return func(cb orm.ConditionBuilder) {
					// Only show Electronics and its children
					cb.Group(func(cb orm.ConditionBuilder) {
						cb.OrEquals("id", "cat001")
						cb.OrEquals("parent_id", "cat001")
					})
				}
			}),
	}
}

// Ordered Tree Resource.
type OrderedCategoryFindTreeResource struct {
	api.Resource
	apis.FindTreeAPI[TestCategory, TestCategorySearch]
}

func NewOrderedCategoryFindTreeResource() api.Resource {
	return &OrderedCategoryFindTreeResource{
		Resource: api.NewResource("test/category_tree_ordered"),
		FindTreeAPI: apis.NewFindTreeAPI[TestCategory, TestCategorySearch](buildCategoryTree).
			Public().
			IdColumn("id").
			ParentIdColumn("parent_id").
			SortApplier(func(search TestCategorySearch, ctx fiber.Ctx) orm.ApplyFunc[apis.Sorter] {
				return func(s apis.Sorter) {
					s.OrderBy("sort")
				}
			}),
	}
}

// FindTreeTestSuite is the test suite for FindTree API tests.
type FindTreeTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindTreeTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestCategoryFindTreeResource,
		NewFilteredCategoryFindTreeResource,
		NewOrderedCategoryFindTreeResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindTreeTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindTreeBasic tests basic FindTree functionality.
func (suite *FindTreeTestSuite) TestFindTreeBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree",
			Action:   "findTree",
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
	suite.Equal("Clothing", first["name"])

	second := suite.readDataAsMap(tree[1])
	suite.Equal("Books", second["name"])

	third := suite.readDataAsMap(tree[2])
	suite.Equal("Electronics", third["name"])

	// Check Electronics has children (it's the third item due to DESC ordering)
	electronics := third
	children := suite.readDataAsSlice(electronics["children"])
	suite.Len(children, 2) // Computers and Phones
}

// TestFindTreeWithSearch tests FindTree with search conditions.
func (suite *FindTreeTestSuite) TestFindTreeWithSearch() {
	suite.Run("SearchByCode", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "findTree",
				Version:  "v1",
			},
			Params: map[string]any{
				"code": "electronics",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 1) // Only Electronics

		electronics := suite.readDataAsMap(tree[0])
		suite.Equal("Electronics", electronics["name"])
	})

	suite.Run("SearchByParentId", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "findTree",
				Version:  "v1",
			},
			Params: map[string]any{
				"parentId": "cat001", // Electronics' children
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		// Should return 1 tree with Electronics as root (because recursive CTE finds ancestors)
		suite.Len(tree, 1)

		electronics := suite.readDataAsMap(tree[0])
		suite.Equal("Electronics", electronics["name"])

		// Electronics should have 2 children: Computers and Phones
		children := suite.readDataAsSlice(electronics["children"])
		suite.Len(children, 2)
	})

	suite.Run("SearchByKeyword", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "findTree",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Computer",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.GreaterOrEqual(len(tree), 1) // At least Computers category
	})
}

// TestFindTreeWithFilterApplier tests FindTree with filter applier.
func (suite *FindTreeTestSuite) TestFindTreeWithFilterApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_filtered",
			Action:   "findTree",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	tree := suite.readDataAsSlice(body.Data)
	// Should only return Electronics and its direct children
	suite.Len(tree, 1) // Only Electronics root

	electronics := suite.readDataAsMap(tree[0])
	suite.Equal("Electronics", electronics["name"])

	children := suite.readDataAsSlice(electronics["children"])
	suite.Len(children, 2) // Computers and Phones
}

// TestFindTreeWithSortApplier tests FindTree with sort applier.
func (suite *FindTreeTestSuite) TestFindTreeWithSortApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/category_tree_ordered",
			Action:   "findTree",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	tree := suite.readDataAsSlice(body.Data)
	suite.Len(tree, 3)

	// Verify ordering by sort field
	first := suite.readDataAsMap(tree[0])
	suite.Equal("Electronics", first["name"]) // sort = 1

	second := suite.readDataAsMap(tree[1])
	suite.Equal("Books", second["name"]) // sort = 2

	third := suite.readDataAsMap(tree[2])
	suite.Equal("Clothing", third["name"]) // sort = 3
}

// TestFindTreeNegativeCases tests negative scenarios.
func (suite *FindTreeTestSuite) TestFindTreeNegativeCases() {
	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "findTree",
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

	suite.Run("EmptySearchCriteria", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/category_tree",
				Action:   "findTree",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		tree := suite.readDataAsSlice(body.Data)
		suite.Len(tree, 3) // All root categories
	})
}
