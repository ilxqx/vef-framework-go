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
type TestUserFindAllResource struct {
	api.Resource
	apis.FindAllAPI[TestUser, TestUserSearch]
}

func NewTestUserFindAllResource() api.Resource {
	return &TestUserFindAllResource{
		Resource:   api.NewResource("test/user_all"),
		FindAllAPI: apis.NewFindAllAPI[TestUser, TestUserSearch]().Public(),
	}
}

// Processed User Resource - with processor.
type ProcessedUserFindAllResource struct {
	api.Resource
	apis.FindAllAPI[TestUser, TestUserSearch]
}

type ProcessedUserList struct {
	Users     []TestUser `json:"users"`
	Processed bool       `json:"processed"`
}

func NewProcessedUserFindAllResource() api.Resource {
	return &ProcessedUserFindAllResource{
		Resource: api.NewResource("test/user_all_processed"),
		FindAllAPI: apis.NewFindAllAPI[TestUser, TestUserSearch]().
			Public().
			Processor(func(users []TestUser, search TestUserSearch, ctx fiber.Ctx) any {
				return ProcessedUserList{
					Users:     users,
					Processed: true,
				}
			}),
	}
}

// Filtered User Resource - with filter applier.
type FilteredUserFindAllResource struct {
	api.Resource
	apis.FindAllAPI[TestUser, TestUserSearch]
}

func NewFilteredUserFindAllResource() api.Resource {
	return &FilteredUserFindAllResource{
		Resource: api.NewResource("test/user_all_filtered"),
		FindAllAPI: apis.NewFindAllAPI[TestUser, TestUserSearch]().
			Public().
			FilterApplier(func(search TestUserSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder] {
				return func(cb orm.ConditionBuilder) {
					cb.Equals("status", "active")
				}
			}),
	}
}

// Ordered User Resource - with order applier.
type OrderedUserFindAllResource struct {
	api.Resource
	apis.FindAllAPI[TestUser, TestUserSearch]
}

func NewOrderedUserFindAllResource() api.Resource {
	return &OrderedUserFindAllResource{
		Resource: api.NewResource("test/user_all_ordered"),
		FindAllAPI: apis.NewFindAllAPI[TestUser, TestUserSearch]().
			Public().
			SortApplier(func(search TestUserSearch, ctx fiber.Ctx) orm.ApplyFunc[apis.Sorter] {
				return func(s apis.Sorter) {
					s.OrderBy("age")
				}
			}),
	}
}

// AuditUser User Resource - with audit user names.
type AuditUserTestUserFindAllResource struct {
	api.Resource
	apis.FindAllAPI[TestUser, TestUserSearch]
}

func NewAuditUserTestUserFindAllResource() api.Resource {
	return &AuditUserTestUserFindAllResource{
		Resource: api.NewResource("test/user_all_audit"),
		FindAllAPI: apis.NewFindAllAPI[TestUser, TestUserSearch]().
			Public().
			WithAuditUserNames((*TestAuditUser)(nil)),
	}
}

// FindAllTestSuite is the test suite for FindAll API tests.
type FindAllTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindAllTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserFindAllResource,
		NewProcessedUserFindAllResource,
		NewFilteredUserFindAllResource,
		NewOrderedUserFindAllResource,
		NewAuditUserTestUserFindAllResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindAllTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindAllBasic tests basic FindAll functionality.
func (suite *FindAllTestSuite) TestFindAllBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all",
			Action:   "findAll",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	suite.NotNil(body.Data)

	// Should return all 10 users
	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 10)
}

// TestFindAllWithSearchApplier tests FindAll with custom search conditions.
func (suite *FindAllTestSuite) TestFindAllWithSearchApplier() {
	suite.Run("SearchByStatus", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "findAll",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 7) // 7 active users in test data
	})

	suite.Run("SearchByKeyword", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "findAll",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Engineer",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 3) // Alice (Software Engineer), Henry (DevOps Engineer), Ivy (QA Engineer)
	})

	suite.Run("SearchByAgeRange", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "findAll",
				Version:  "v1",
			},
			Params: map[string]any{
				"age": []int{25, 28},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.GreaterOrEqual(len(users), 3) // At least Alice (25), Eve (27), Charlie (28)
	})
}

// TestFindAllWithProcessor tests FindAll with post-processing.
func (suite *FindAllTestSuite) TestFindAllWithProcessor() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_processed",
			Action:   "findAll",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.NotNil(body.Data)

	dataMap := suite.readDataAsMap(body.Data)
	suite.Equal(true, dataMap["processed"])
	suite.NotNil(dataMap["users"])

	users := suite.readDataAsSlice(dataMap["users"])
	suite.Len(users, 10)
}

// TestFindAllWithFilterApplier tests FindAll with filter applier.
func (suite *FindAllTestSuite) TestFindAllWithFilterApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_filtered",
			Action:   "findAll",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.NotNil(body.Data)

	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 7) // Only active users
}

// TestFindAllWithSortApplier tests FindAll with sort applier.
func (suite *FindAllTestSuite) TestFindAllWithSortApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_ordered",
			Action:   "findAll",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.NotNil(body.Data)

	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 10)

	// First user should be youngest (Alice, age 25)
	firstUser := suite.readDataAsMap(users[0])
	suite.Equal(float64(25), firstUser["age"])
}

// TestFindAllNegativeCases tests negative scenarios.
func (suite *FindAllTestSuite) TestFindAllNegativeCases() {
	suite.Run("EmptySearchCriteria", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "findAll",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)
	})

	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "findAll",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "NonexistentKeyword",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 0) // Empty array, not nil
	})
}

// TestFindAllWithAuditUserNames tests FindAll with audit user names populated.
func (suite *FindAllTestSuite) TestFindAllWithAuditUserNames() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_audit",
			Action:   "findAll",
			Version:  "v1",
		},
		Params: map[string]any{
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.NotNil(body.Data)

	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 7) // 7 active users

	// Check first user has audit user names
	firstUser := suite.readDataAsMap(users[0])
	suite.NotNil(firstUser["createdByName"])
	suite.NotNil(firstUser["updatedByName"])

	// Verify all users have audit user names populated
	for _, u := range users {
		user := suite.readDataAsMap(u)
		suite.NotNil(user["createdByName"], "User %s should have createdByName", user["id"])
		suite.NotNil(user["updatedByName"], "User %s should have updatedByName", user["id"])
		// Audit user names should be from TestAuditUser data
		suite.Contains([]string{"John Doe", "Jane Smith", "Michael Johnson", "Sarah Williams"}, user["createdByName"])
		suite.Contains([]string{"John Doe", "Jane Smith", "Michael Johnson", "Sarah Williams"}, user["updatedByName"])
	}
}
