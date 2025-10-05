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
type TestUserFindPageResource struct {
	api.Resource
	apis.FindPageAPI[TestUser, TestUserSearch]
}

func NewTestUserFindPageResource() api.Resource {
	return &TestUserFindPageResource{
		Resource:    api.NewResource("test/user_page"),
		FindPageAPI: apis.NewFindPageAPI[TestUser, TestUserSearch]().Public(),
	}
}

// Processed User Resource - with processor.
type ProcessedUserFindPageResource struct {
	api.Resource
	apis.FindPageAPI[TestUser, TestUserSearch]
}

func NewProcessedUserFindPageResource() api.Resource {
	return &ProcessedUserFindPageResource{
		Resource: api.NewResource("test/user_page_processed"),
		FindPageAPI: apis.NewFindPageAPI[TestUser, TestUserSearch]().
			Public().
			Processor(func(users []TestUser, search TestUserSearch, ctx fiber.Ctx) any {
				// Processor must return a slice - convert each user to a processed version
				processed := make([]ProcessedUser, len(users))
				for i, user := range users {
					processed[i] = ProcessedUser{
						TestUser:  user,
						Processed: true,
					}
				}

				return processed
			}),
	}
}

// Filtered User Resource - with filter applier.
type FilteredUserFindPageResource struct {
	api.Resource
	apis.FindPageAPI[TestUser, TestUserSearch]
}

func NewFilteredUserFindPageResource() api.Resource {
	return &FilteredUserFindPageResource{
		Resource: api.NewResource("test/user_page_filtered"),
		FindPageAPI: apis.NewFindPageAPI[TestUser, TestUserSearch]().
			Public().
			FilterApplier(func(search TestUserSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder] {
				return func(cb orm.ConditionBuilder) {
					cb.Equals("status", "active")
				}
			}),
	}
}

// FindPageTestSuite is the test suite for FindPage API tests.
type FindPageTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindPageTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserFindPageResource,
		NewProcessedUserFindPageResource,
		NewFilteredUserFindPageResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindPageTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindPageBasic tests basic FindPage functionality.
func (suite *FindPageTestSuite) TestFindPageBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page",
			Action:   "findPage",
			Version:  "v1",
		},
		Params: map[string]any{
			"page": 1,
			"size": 5,
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	suite.NotNil(body.Data)

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(10), page["total"])
	suite.Equal(float64(1), page["page"])
	suite.Equal(float64(5), page["size"])

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 5)
}

// TestFindPagePagination tests pagination functionality.
func (suite *FindPageTestSuite) TestFindPagePagination() {
	suite.Run("FirstPage", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "findPage",
				Version:  "v1",
			},
			Params: map[string]any{
				"page": 1,
				"size": 3,
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"])
		suite.Equal(float64(1), page["page"])
		suite.Equal(float64(3), page["size"])

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 3)
	})

	suite.Run("SecondPage", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "findPage",
				Version:  "v1",
			},
			Params: map[string]any{
				"page": 2,
				"size": 3,
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"])
		suite.Equal(float64(2), page["page"])
		suite.Equal(float64(3), page["size"])

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 3)
	})

	suite.Run("LastPage", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "findPage",
				Version:  "v1",
			},
			Params: map[string]any{
				"page": 4,
				"size": 3,
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"])
		suite.Equal(float64(4), page["page"])

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 1) // Only 1 record on last page
	})

	suite.Run("EmptyPage", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "findPage",
				Version:  "v1",
			},
			Params: map[string]any{
				"page": 100,
				"size": 10,
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"])

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 0)
	})
}

// TestFindPageWithSearch tests FindPage with search conditions.
func (suite *FindPageTestSuite) TestFindPageWithSearch() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page",
			Action:   "findPage",
			Version:  "v1",
		},
		Params: map[string]any{
			"page":   1,
			"size":   10,
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(7), page["total"]) // 7 active users

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 7)
}

// TestFindPageWithProcessor tests FindPage with post-processing.
func (suite *FindPageTestSuite) TestFindPageWithProcessor() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page_processed",
			Action:   "findPage",
			Version:  "v1",
		},
		Params: map[string]any{
			"page": 1,
			"size": 5,
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(10), page["total"])

	// Processor returned slice of ProcessedUser
	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 5)

	// Check first processed user has the processed flag
	firstUser := suite.readDataAsMap(items[0])
	suite.Equal(true, firstUser["processed"])
}

// TestFindPageWithFilterApplier tests FindPage with filter applier.
func (suite *FindPageTestSuite) TestFindPageWithFilterApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page_filtered",
			Action:   "findPage",
			Version:  "v1",
		},
		Params: map[string]any{
			"page": 1,
			"size": 10,
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(7), page["total"]) // Only active users

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 7)
}

// TestFindPageNegativeCases tests negative scenarios.
func (suite *FindPageTestSuite) TestFindPageNegativeCases() {
	suite.Run("InvalidPageNumber", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "findPage",
				Version:  "v1",
			},
			Params: map[string]any{
				"page": 0, // Should be normalized to 1
				"size": 10,
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(1), page["page"]) // Normalized to 1
	})

	suite.Run("InvalidPageSize", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "findPage",
				Version:  "v1",
			},
			Params: map[string]any{
				"page": 1,
				"size": 0, // Should be normalized to default
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		page := suite.readDataAsMap(body.Data)
		suite.Greater(page["size"], float64(0)) // Should have default size
	})

	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "findPage",
				Version:  "v1",
			},
			Params: map[string]any{
				"page":    1,
				"size":    10,
				"keyword": "NonexistentKeyword",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(0), page["total"])

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 0)
	})
}
