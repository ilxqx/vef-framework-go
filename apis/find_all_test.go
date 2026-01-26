package apis_test

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/sort"
)

// Test Resources.
type TestUserFindAllResource struct {
	api.Resource
	apis.FindAll[TestUser, TestUserSearch]
}

func NewTestUserFindAllResource() api.Resource {
	return &TestUserFindAllResource{
		Resource: api.NewRPCResource("test/user_all"),
		FindAll:  apis.NewFindAll[TestUser, TestUserSearch]().Public(),
	}
}

// Processed User Resource - with processor.
type ProcessedUserFindAllResource struct {
	api.Resource
	apis.FindAll[TestUser, TestUserSearch]
}

type ProcessedUserList struct {
	Users     []TestUser `json:"users"`
	Processed bool       `json:"processed"`
}

func NewProcessedUserFindAllResource() api.Resource {
	return &ProcessedUserFindAllResource{
		Resource: api.NewRPCResource("test/user_all_processed"),
		FindAll: apis.NewFindAll[TestUser, TestUserSearch]().
			Public().
			WithProcessor(func(users []TestUser, search TestUserSearch, ctx fiber.Ctx) any {
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
	apis.FindAll[TestUser, TestUserSearch]
}

func NewFilteredUserFindAllResource() api.Resource {
	return &FilteredUserFindAllResource{
		Resource: api.NewRPCResource("test/user_all_filtered"),
		FindAll: apis.NewFindAll[TestUser, TestUserSearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				cb.Equals("status", "active")
			}).
			Public(),
	}
}

// Ordered User Resource - with order applier.
type OrderedUserFindAllResource struct {
	api.Resource
	apis.FindAll[TestUser, TestUserSearch]
}

func NewOrderedUserFindAllResource() api.Resource {
	return &OrderedUserFindAllResource{
		Resource: api.NewRPCResource("test/user_all_ordered"),
		FindAll: apis.NewFindAll[TestUser, TestUserSearch]().
			WithDefaultSort(&sort.OrderSpec{
				Column: "age",
			}).
			Public(),
	}
}

// AuditUser User Resource - with audit user names.
type AuditUserTestUserFindAllResource struct {
	api.Resource
	apis.FindAll[TestUser, TestUserSearch]
}

func NewAuditUserTestUserFindAllResource() api.Resource {
	return &AuditUserTestUserFindAllResource{
		Resource: api.NewRPCResource("test/user_all_audit"),
		FindAll: apis.NewFindAll[TestUser, TestUserSearch]().
			WithAuditUserNames((*TestAuditUser)(nil)).
			Public(),
	}
}

// NoDefaultSort User Resource - explicitly disable default sorting.
type NoDefaultSortUserFindAllResource struct {
	api.Resource
	apis.FindAll[TestUser, TestUserSearch]
}

func NewNoDefaultSortUserFindAllResource() api.Resource {
	return &NoDefaultSortUserFindAllResource{
		Resource: api.NewRPCResource("test/user_all_no_default_sort"),
		FindAll: apis.NewFindAll[TestUser, TestUserSearch]().
			WithDefaultSort(). // Empty call to disable default sorting
			Public(),
	}
}

// MultipleDefaultSort User Resource - with multiple default sort columns.
type MultipleDefaultSortUserFindAllResource struct {
	api.Resource
	apis.FindAll[TestUser, TestUserSearch]
}

func NewMultipleDefaultSortUserFindAllResource() api.Resource {
	return &MultipleDefaultSortUserFindAllResource{
		Resource: api.NewRPCResource("test/user_all_multi_sort"),
		FindAll: apis.NewFindAll[TestUser, TestUserSearch]().
			WithDefaultSort(
				&sort.OrderSpec{
					Column:    "status",
					Direction: sort.OrderAsc,
				},
				&sort.OrderSpec{
					Column:    "age",
					Direction: sort.OrderDesc,
				},
			).
			Public(),
	}
}

// FindAllTestSuite tests the FindAll API functionality
// including basic queries, search filters, processors, sorting, audit user names, and negative cases.
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
		NewNoDefaultSortUserFindAllResource,
		NewMultipleDefaultSortUserFindAllResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindAllTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindAllBasic tests basic FindAll functionality.
func (suite *FindAllTestSuite) TestFindAllBasic() {
	suite.T().Logf("Testing FindAll API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all",
			Action:   "find_all",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	suite.NotNil(body.Data, "Data should not be nil")

	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 10, "Should return all 10 users")

	suite.T().Logf("Found %d users without filters", len(users))
}

// TestFindAllWithSearchApplier tests FindAll with custom search conditions.
func (suite *FindAllTestSuite) TestFindAllWithSearchApplier() {
	suite.T().Logf("Testing FindAll API with search filters for %s", suite.dbType)

	suite.Run("SearchByStatus", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": "active",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 7, "Should return 7 active users")

		suite.T().Logf("Found %d users with status=active", len(users))
	})

	suite.Run("SearchByKeyword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Engineer",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 3, "Should return 3 users with keyword 'Engineer'")

		suite.T().Logf("Found %d users with keyword=Engineer", len(users))
	})

	suite.Run("SearchByAgeRange", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
				Version:  "v1",
			},
			Params: map[string]any{
				"age": []int{25, 28},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")

		users := suite.readDataAsSlice(body.Data)
		suite.GreaterOrEqual(len(users), 3, "Should return at least 3 users in age range 25-28")

		suite.T().Logf("Found %d users with age range 25-28", len(users))
	})
}

// TestFindAllWithProcessor tests FindAll with post-processing.
func (suite *FindAllTestSuite) TestFindAllWithWithProcessor() {
	suite.T().Logf("Testing FindAll API with processor for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_processed",
			Action:   "find_all",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")

	dataMap := suite.readDataAsMap(body.Data)
	suite.Equal(true, dataMap["processed"], "Processed flag should be true")
	suite.NotNil(dataMap["users"], "Users array should not be nil")

	users := suite.readDataAsSlice(dataMap["users"])
	suite.Len(users, 10, "Should return all 10 users in processed format")

	suite.T().Logf("Found %d users with post-processing applied", len(users))
}

// TestFindAllWithFilterApplier tests FindAll with filter applier.
func (suite *FindAllTestSuite) TestFindAllWithFilterApplier() {
	suite.T().Logf("Testing FindAll API with filter applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_filtered",
			Action:   "find_all",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")

	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 7, "Should return only active users (7 users)")

	suite.T().Logf("Found %d active users with filter applier", len(users))
}

// TestFindAllWithSortApplier tests FindAll with sort applier.
func (suite *FindAllTestSuite) TestFindAllWithSortApplier() {
	suite.T().Logf("Testing FindAll API with sort applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_ordered",
			Action:   "find_all",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")

	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 10, "Should return all 10 users")

	firstUser := suite.readDataAsMap(users[0])
	suite.Equal(float64(25), firstUser["age"], "First user should be youngest (age 25)")

	suite.T().Logf("Found %d users sorted by age ascending, first user age: %.0f", len(users), firstUser["age"])
}

// TestFindAllNegativeCases tests negative scenarios.
func (suite *FindAllTestSuite) TestFindAllNegativeCases() {
	suite.T().Logf("Testing FindAll API negative cases for %s", suite.dbType)

	suite.Run("EmptySearchCriteria", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all users when no search criteria provided")

		suite.T().Logf("Found %d users with empty search criteria", len(users))
	})

	suite.Run("NoMatchingRecords", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "NonexistentKeyword",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 0, "Should return empty array when no records match")

		suite.T().Logf("Found %d users with non-existent keyword (empty array as expected)", len(users))
	})
}

// TestFindAllWithAuditUserNames tests FindAll with audit user names populated.
func (suite *FindAllTestSuite) TestFindAllWithAuditUserNames() {
	suite.T().Logf("Testing FindAll API with audit user names for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_audit",
			Action:   "find_all",
			Version:  "v1",
		},
		Params: map[string]any{
			"status": "active",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")

	users := suite.readDataAsSlice(body.Data)
	suite.Len(users, 7, "Should return 7 active users")

	firstUser := suite.readDataAsMap(users[0])
	suite.NotNil(firstUser["createdByName"], "First user should have createdByName")
	suite.NotNil(firstUser["updatedByName"], "First user should have updatedByName")

	for _, u := range users {
		user := suite.readDataAsMap(u)
		suite.NotNil(user["createdByName"], "User %s should have createdByName", user["id"])
		suite.NotNil(user["updatedByName"], "User %s should have updatedByName", user["id"])
		suite.Contains([]string{"John Doe", "Jane Smith", "Michael Johnson", "Sarah Williams"}, user["createdByName"], "createdByName should be from audit users")
		suite.Contains([]string{"John Doe", "Jane Smith", "Michael Johnson", "Sarah Williams"}, user["updatedByName"], "updatedByName should be from audit users")
	}

	suite.T().Logf("Found %d users with audit user names populated", len(users))
}

// TestFindAllDefaultSorting tests default sorting behavior.
func (suite *FindAllTestSuite) TestFindAllDefaultSorting() {
	suite.T().Logf("Testing FindAll API default sorting for %s", suite.dbType)

	suite.Run("DefaultSortByPrimaryKey", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all 10 users")

		firstUser := suite.readDataAsMap(users[0])
		suite.Equal("user010", firstUser["id"], "First user should have highest id (user010) when sorted by id DESC")

		lastUser := suite.readDataAsMap(users[len(users)-1])
		suite.Equal("user001", lastUser["id"], "Last user should have lowest id (user001) when sorted by id DESC")

		suite.T().Logf("Default sort by primary key: first=%s, last=%s", firstUser["id"], lastUser["id"])
	})

	suite.Run("CustomDefaultSort", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_ordered",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all 10 users")

		firstUser := suite.readDataAsMap(users[0])
		suite.Equal(float64(25), firstUser["age"], "First user should be youngest (age 25)")
		suite.Equal("Alice Johnson", firstUser["name"], "First user should be Alice Johnson")

		lastUser := suite.readDataAsMap(users[len(users)-1])
		suite.Equal(float64(35), lastUser["age"], "Last user should be oldest (age 35)")
		suite.Equal("Frank Miller", lastUser["name"], "Last user should be Frank Miller")

		suite.T().Logf("Custom default sort by age: first=%s (age %.0f), last=%s (age %.0f)", firstUser["name"], firstUser["age"], lastUser["name"], lastUser["age"])
	})

	suite.Run("DisableDefaultSort", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_no_default_sort",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all 10 users without sorting")

		suite.T().Logf("Found %d users with default sort disabled (order is database-dependent)", len(users))
	})

	suite.Run("MultipleDefaultSortColumns", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_multi_sort",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all 10 users")

		firstUser := suite.readDataAsMap(users[0])
		suite.Equal("active", firstUser["status"], "First user should be active (sorted by status ASC)")

		var lastActiveIndex int
		for i, u := range users {
			user := suite.readDataAsMap(u)
			if user["status"] == "active" {
				lastActiveIndex = i
			}
		}

		for i := 0; i <= lastActiveIndex; i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("active", user["status"], "All active users should come before inactive users")
		}

		for i := lastActiveIndex + 1; i < len(users); i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("inactive", user["status"], "All inactive users should come after active users")
		}

		suite.T().Logf("Multiple sort columns verified: %d active users, %d inactive users", lastActiveIndex+1, len(users)-lastActiveIndex-1)
	})
}

// TestFindAllRequestSortOverride tests that request-specified sorting overrides default sorting.
func (suite *FindAllTestSuite) TestFindAllRequestSortOverride() {
	suite.T().Logf("Testing FindAll API request sort override for %s", suite.dbType)

	suite.Run("OverrideDefaultSortWithRequestSort", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_ordered",
				Action:   "find_all",
				Version:  "v1",
			},
			Meta: map[string]any{
				"sort": []map[string]any{
					{
						"column":    "name",
						"direction": "desc",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all 10 users")

		firstUser := suite.readDataAsMap(users[0])
		firstName := firstUser["name"].(string)

		lastUser := suite.readDataAsMap(users[len(users)-1])
		lastName := lastUser["name"].(string)

		suite.True(firstName > lastName, "First name %s should be > last name %s in DESC order", firstName, lastName)

		suite.T().Logf("Request sort override: first=%s, last=%s", firstName, lastName)
	})

	suite.Run("OverrideWithMultipleSortColumns", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_ordered",
				Action:   "find_all",
				Version:  "v1",
			},
			Meta: map[string]any{
				"sort": []map[string]any{
					{
						"column":    "status",
						"direction": "asc",
					},
					{
						"column":    "name",
						"direction": "asc",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all 10 users")

		var lastActiveIndex int
		for i, u := range users {
			user := suite.readDataAsMap(u)
			if user["status"] == "active" {
				lastActiveIndex = i
			}
		}

		for i := 0; i <= lastActiveIndex; i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("active", user["status"], "All active users should come before inactive users")
		}

		for i := lastActiveIndex + 1; i < len(users); i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("inactive", user["status"], "All inactive users should come after active users")
		}

		suite.T().Logf("Multiple sort columns override verified: status ASC, name ASC")
	})

	suite.Run("OverrideDisabledDefaultSort", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_no_default_sort",
				Action:   "find_all",
				Version:  "v1",
			},
			Meta: map[string]any{
				"sort": []map[string]any{
					{
						"column":    "email",
						"direction": "asc",
					},
				},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10, "Should return all 10 users")

		var prevEmail string
		for i, u := range users {
			user := suite.readDataAsMap(u)

			email := user["email"].(string)
			if i > 0 {
				suite.True(email >= prevEmail, "Email %s should be >= previous email %s in ASC order", email, prevEmail)
			}

			prevEmail = email
		}

		suite.T().Logf("Request sort applied to resource with disabled default sort: emails sorted ASC")
	})
}
