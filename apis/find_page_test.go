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
type TestUserFindPageResource struct {
	api.Resource
	apis.FindPage[TestUser, TestUserSearch]
}

func NewTestUserFindPageResource() api.Resource {
	return &TestUserFindPageResource{
		Resource: api.NewRPCResource("test/user_page"),
		FindPage: apis.NewFindPage[TestUser, TestUserSearch]().Public(),
	}
}

// Processed User Resource - with processor.
type ProcessedUserFindPageResource struct {
	api.Resource
	apis.FindPage[TestUser, TestUserSearch]
}

func NewProcessedUserFindPageResource() api.Resource {
	return &ProcessedUserFindPageResource{
		Resource: api.NewRPCResource("test/user_page_processed"),
		FindPage: apis.NewFindPage[TestUser, TestUserSearch]().
			Public().
			WithProcessor(func(users []TestUser, _ TestUserSearch, _ fiber.Ctx) any {
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
	apis.FindPage[TestUser, TestUserSearch]
}

func NewFilteredUserFindPageResource() api.Resource {
	return &FilteredUserFindPageResource{
		Resource: api.NewRPCResource("test/user_page_filtered"),
		FindPage: apis.NewFindPage[TestUser, TestUserSearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				cb.Equals("status", "active")
			}).
			Public(),
	}
}

// AuditUser User Resource - with audit user names.
type AuditUserTestUserFindPageResource struct {
	api.Resource
	apis.FindPage[TestUser, TestUserSearch]
}

func NewAuditUserTestUserFindPageResource() api.Resource {
	return &AuditUserTestUserFindPageResource{
		Resource: api.NewRPCResource("test/user_page_audit"),
		FindPage: apis.NewFindPage[TestUser, TestUserSearch]().
			WithAuditUserNames((*TestAuditUser)(nil)).
			Public(),
	}
}

// FindPageTestSuite tests the FindPage API functionality
// including basic pagination, search filters, processors, filter appliers, audit user names, and negative cases.
type FindPageTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindPageTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserFindPageResource,
		NewProcessedUserFindPageResource,
		NewFilteredUserFindPageResource,
		NewAuditUserTestUserFindPageResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindPageTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindPageBasic tests basic FindPage functionality.
func (suite *FindPageTestSuite) TestFindPageBasic() {
	suite.T().Logf("Testing FindPage API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page",
			Action:   "find_page",
			Version:  "v1",
		},
		Meta: map[string]any{
			"page": 1,
			"size": 5,
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	suite.NotNil(body.Data, "Data should not be nil")

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(10), page["total"], "Total should be 10")
	suite.Equal(float64(1), page["page"], "Page should be 1")
	suite.Equal(float64(5), page["size"], "Size should be 5")

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 5, "Should return 5 items on first page")

	suite.T().Logf("Found %d items on page %v of %v (size=%v, total=%v)", len(items), page["page"], page["total"], page["size"], page["total"])
}

// TestFindPagePagination tests pagination functionality.
func (suite *FindPageTestSuite) TestFindPagePagination() {
	suite.T().Logf("Testing FindPage API pagination for %s", suite.dbType)

	suite.Run("FirstPage", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "find_page",
				Version:  "v1",
			},
			Meta: map[string]any{
				"page": 1,
				"size": 3,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"], "Total should be 10")
		suite.Equal(float64(1), page["page"], "Page should be 1")
		suite.Equal(float64(3), page["size"], "Size should be 3")

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 3, "Should return 3 items on first page")

		suite.T().Logf("First page: %d items (page=%v, size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
	})

	suite.Run("SecondPage", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "find_page",
				Version:  "v1",
			},
			Meta: map[string]any{
				"page": 2,
				"size": 3,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"], "Total should be 10")
		suite.Equal(float64(2), page["page"], "Page should be 2")
		suite.Equal(float64(3), page["size"], "Size should be 3")

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 3, "Should return 3 items on second page")

		suite.T().Logf("Second page: %d items (page=%v, size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
	})

	suite.Run("LastPage", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "find_page",
				Version:  "v1",
			},
			Meta: map[string]any{
				"page": 4,
				"size": 3,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"], "Total should be 10")
		suite.Equal(float64(4), page["page"], "Page should be 4")

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 1, "Should return 1 item on last page")

		suite.T().Logf("Last page: %d items (page=%v, size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
	})

	suite.Run("EmptyPage", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "find_page",
				Version:  "v1",
			},
			Meta: map[string]any{
				"page": 100,
				"size": 10,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"], "Total should be 10")

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 0, "Should return empty items array for out-of-range page")

		suite.T().Logf("Empty page: %d items (page=%v, size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
	})
}

// TestFindPageWithSearch tests FindPage with search conditions.
func (suite *FindPageTestSuite) TestFindPageWithSearch() {
	suite.T().Logf("Testing FindPage API with search filters for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page",
			Action:   "find_page",
			Version:  "v1",
		},
		Meta: map[string]any{
			"page": 1,
			"size": 10,
		},
		Params: map[string]any{
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(7), page["total"], "Total should be 7 active users")

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 7, "Should return 7 active users")

	suite.T().Logf("Found %d active users on page %v (size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
}

// TestFindPageWithProcessor tests FindPage with post-processing.
func (suite *FindPageTestSuite) TestFindPageWithWithProcessor() {
	suite.T().Logf("Testing FindPage API with processor for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page_processed",
			Action:   "find_page",
			Version:  "v1",
		},
		Meta: map[string]any{
			"page": 1,
			"size": 5,
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(10), page["total"], "Total should be 10")

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 5, "Should return 5 processed items")

	firstUser := suite.readDataAsMap(items[0])
	suite.Equal(true, firstUser["processed"], "Processed flag should be true")

	suite.T().Logf("Found %d items with post-processing applied on page %v (size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
}

// TestFindPageWithFilterApplier tests FindPage with filter applier.
func (suite *FindPageTestSuite) TestFindPageWithFilterApplier() {
	suite.T().Logf("Testing FindPage API with filter applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page_filtered",
			Action:   "find_page",
			Version:  "v1",
		},
		Meta: map[string]any{
			"page": 1,
			"size": 10,
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(7), page["total"], "Total should be 7 active users")

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 7, "Should return 7 active users with filter applier")

	suite.T().Logf("Found %d active users with filter applier on page %v (size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
}

// TestFindPageNegativeCases tests negative scenarios.
func (suite *FindPageTestSuite) TestFindPageNegativeCases() {
	suite.T().Logf("Testing FindPage API negative cases for %s", suite.dbType)

	suite.Run("InvalidPageNumber", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "find_page",
				Version:  "v1",
			},
			Meta: map[string]any{
				"page": 0,
				"size": 10,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(1), page["page"], "Page should be normalized to 1")

		suite.T().Logf("Invalid page number 0 was normalized to %v", page["page"])
	})

	suite.Run("InvalidPageSize", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "find_page",
				Version:  "v1",
			},
			Meta: map[string]any{
				"page": 1,
				"size": 0,
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(10), page["total"], "Total should be 10")
		items := suite.readDataAsSlice(page["items"])
		suite.Greater(len(items), 0, "Should return items with default page size")

		suite.T().Logf("Invalid page size 0 returned %d items with default size", len(items))
	})

	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_page",
				Action:   "find_page",
				Version:  "v1",
			},
			Meta: map[string]any{
				"page": 1,
				"size": 10,
			},
			Params: map[string]any{
				"keyword": "NonexistentKeyword",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		page := suite.readDataAsMap(body.Data)
		suite.Equal(float64(0), page["total"], "Total should be 0 for no matching records")

		items := suite.readDataAsSlice(page["items"])
		suite.Len(items, 0, "Should return empty items array")

		suite.T().Logf("No matching records: total=%v, items=%d", page["total"], len(items))
	})
}

// TestFindPageWithAuditUserNames tests FindPage with audit user names populated.
func (suite *FindPageTestSuite) TestFindPageWithAuditUserNames() {
	suite.T().Logf("Testing FindPage API with audit user names for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_page_audit",
			Action:   "find_page",
			Version:  "v1",
		},
		Meta: map[string]any{
			"page": 1,
			"size": 5,
		},
		Params: map[string]any{
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")

	page := suite.readDataAsMap(body.Data)
	suite.Equal(float64(7), page["total"], "Total should be 7 active users")
	suite.Equal(float64(1), page["page"], "Page should be 1")
	suite.Equal(float64(5), page["size"], "Size should be 5")

	items := suite.readDataAsSlice(page["items"])
	suite.Len(items, 5, "Should return 5 items on first page")

	firstUser := suite.readDataAsMap(items[0])
	suite.NotNil(firstUser["createdByName"], "First user should have createdByName")
	suite.NotNil(firstUser["updatedByName"], "First user should have updatedByName")

	for _, u := range items {
		user := suite.readDataAsMap(u)
		suite.NotNil(user["createdByName"], "User %s should have createdByName", user["id"])
		suite.NotNil(user["updatedByName"], "User %s should have updatedByName", user["id"])
		suite.Contains([]string{"John Doe", "Jane Smith", "Michael Johnson", "Sarah Williams"}, user["createdByName"], "createdByName should be from audit users")
		suite.Contains([]string{"John Doe", "Jane Smith", "Michael Johnson", "Sarah Williams"}, user["updatedByName"], "updatedByName should be from audit users")
	}

	suite.T().Logf("Found %d users with audit user names populated on page %v (size=%v, total=%v)", len(items), page["page"], page["size"], page["total"])
}
