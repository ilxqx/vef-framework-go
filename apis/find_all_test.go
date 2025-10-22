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
	apis.FindAllApi[TestUser, TestUserSearch]
}

func NewTestUserFindAllResource() api.Resource {
	return &TestUserFindAllResource{
		Resource:   api.NewResource("test/user_all"),
		FindAllApi: apis.NewFindAllApi[TestUser, TestUserSearch]().Public(),
	}
}

// Processed User Resource - with processor.
type ProcessedUserFindAllResource struct {
	api.Resource
	apis.FindAllApi[TestUser, TestUserSearch]
}

type ProcessedUserList struct {
	Users     []TestUser `json:"users"`
	Processed bool       `json:"processed"`
}

func NewProcessedUserFindAllResource() api.Resource {
	return &ProcessedUserFindAllResource{
		Resource: api.NewResource("test/user_all_processed"),
		FindAllApi: apis.NewFindAllApi[TestUser, TestUserSearch]().
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
	apis.FindAllApi[TestUser, TestUserSearch]
}

func NewFilteredUserFindAllResource() api.Resource {
	return &FilteredUserFindAllResource{
		Resource: api.NewResource("test/user_all_filtered"),
		FindAllApi: apis.NewFindAllApi[TestUser, TestUserSearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				cb.Equals("status", "active")
			}).
			Public(),
	}
}

// Ordered User Resource - with order applier.
type OrderedUserFindAllResource struct {
	api.Resource
	apis.FindAllApi[TestUser, TestUserSearch]
}

func NewOrderedUserFindAllResource() api.Resource {
	return &OrderedUserFindAllResource{
		Resource: api.NewResource("test/user_all_ordered"),
		FindAllApi: apis.NewFindAllApi[TestUser, TestUserSearch]().
			WithDefaultSort(&sort.OrderSpec{
				Column: "age",
			}).
			Public(),
	}
}

// AuditUser User Resource - with audit user names.
type AuditUserTestUserFindAllResource struct {
	api.Resource
	apis.FindAllApi[TestUser, TestUserSearch]
}

func NewAuditUserTestUserFindAllResource() api.Resource {
	return &AuditUserTestUserFindAllResource{
		Resource: api.NewResource("test/user_all_audit"),
		FindAllApi: apis.NewFindAllApi[TestUser, TestUserSearch]().
			WithAuditUserNames((*TestAuditUser)(nil)).
			Public(),
	}
}

// NoDefaultSort User Resource - explicitly disable default sorting.
type NoDefaultSortUserFindAllResource struct {
	api.Resource
	apis.FindAllApi[TestUser, TestUserSearch]
}

func NewNoDefaultSortUserFindAllResource() api.Resource {
	return &NoDefaultSortUserFindAllResource{
		Resource: api.NewResource("test/user_all_no_default_sort"),
		FindAllApi: apis.NewFindAllApi[TestUser, TestUserSearch]().
			WithDefaultSort(). // Empty call to disable default sorting
			Public(),
	}
}

// MultipleDefaultSort User Resource - with multiple default sort columns.
type MultipleDefaultSortUserFindAllResource struct {
	api.Resource
	apis.FindAllApi[TestUser, TestUserSearch]
}

func NewMultipleDefaultSortUserFindAllResource() api.Resource {
	return &MultipleDefaultSortUserFindAllResource{
		Resource: api.NewResource("test/user_all_multi_sort"),
		FindAllApi: apis.NewFindAllApi[TestUser, TestUserSearch]().
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

// FindAllTestSuite is the test suite for FindAll Api tests.
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
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all",
			Action:   "find_all",
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 7) // 7 active users in test data
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 3) // Alice (Software Engineer), Henry (DevOps Engineer), Ivy (QA Engineer)
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)

		users := suite.readDataAsSlice(body.Data)
		suite.GreaterOrEqual(len(users), 3) // At least Alice (25), Eve (27), Charlie (28)
	})
}

// TestFindAllWithProcessor tests FindAll with post-processing.
func (suite *FindAllTestSuite) TestFindAllWithWithProcessor() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_processed",
			Action:   "find_all",
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
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_filtered",
			Action:   "find_all",
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
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_all_ordered",
			Action:   "find_all",
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
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
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

// TestFindAllDefaultSorting tests default sorting behavior.
func (suite *FindAllTestSuite) TestFindAllDefaultSorting() {
	suite.Run("DefaultSortByPrimaryKey", func() {
		// TestUserFindAllResource has no explicit default sort, should sort by id DESC
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)

		// First user should have the highest id (user010)
		firstUser := suite.readDataAsMap(users[0])
		suite.Equal("user010", firstUser["id"])

		// Last user should have the lowest id (user001)
		lastUser := suite.readDataAsMap(users[len(users)-1])
		suite.Equal("user001", lastUser["id"])
	})

	suite.Run("CustomDefaultSort", func() {
		// OrderedUserFindAllResource has default sort by age ASC
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_ordered",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)

		// First user should be youngest (Alice, age 25)
		firstUser := suite.readDataAsMap(users[0])
		suite.Equal(float64(25), firstUser["age"])
		suite.Equal("Alice Johnson", firstUser["name"])

		// Last user should be oldest (Frank, age 35)
		lastUser := suite.readDataAsMap(users[len(users)-1])
		suite.Equal(float64(35), lastUser["age"])
		suite.Equal("Frank Miller", lastUser["name"])
	})

	suite.Run("DisableDefaultSort", func() {
		// NoDefaultSortUserFindAllResource explicitly disables default sorting
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_no_default_sort",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)
		// Without sorting, order is database-dependent, just verify we got all users
	})

	suite.Run("MultipleDefaultSortColumns", func() {
		// MultipleDefaultSortUserFindAllResource sorts by status ASC, then age DESC
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_all_multi_sort",
				Action:   "find_all",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)

		// First user should be active with highest age (Frank is inactive, so Diana age 32)
		firstUser := suite.readDataAsMap(users[0])
		suite.Equal("active", firstUser["status"])
		// Among active users, should be sorted by age DESC

		// Verify sorting: all active users should come before inactive users
		var lastActiveIndex int
		for i, u := range users {
			user := suite.readDataAsMap(u)
			if user["status"] == "active" {
				lastActiveIndex = i
			}
		}

		// Check that all active users come before inactive users
		for i := 0; i <= lastActiveIndex; i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("active", user["status"])
		}

		// Check that inactive users come after active users
		for i := lastActiveIndex + 1; i < len(users); i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("inactive", user["status"])
		}
	})
}

// TestFindAllRequestSortOverride tests that request-specified sorting overrides default sorting.
func (suite *FindAllTestSuite) TestFindAllRequestSortOverride() {
	suite.Run("OverrideDefaultSortWithRequestSort", func() {
		// OrderedUserFindAllResource has default sort by age ASC
		// Override with name DESC via request
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)

		// First user should have name starting with highest letter
		firstUser := suite.readDataAsMap(users[0])
		firstName := firstUser["name"].(string)

		// Last user should have name starting with lowest letter
		lastUser := suite.readDataAsMap(users[len(users)-1])
		lastName := lastUser["name"].(string)

		// Verify descending order
		suite.True(firstName > lastName, "First name %s should be > last name %s", firstName, lastName)
	})

	suite.Run("OverrideWithMultipleSortColumns", func() {
		// Override default sort with multiple columns: status ASC, name ASC
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)

		// Verify all active users come before inactive users
		var lastActiveIndex int
		for i, u := range users {
			user := suite.readDataAsMap(u)
			if user["status"] == "active" {
				lastActiveIndex = i
			}
		}

		for i := 0; i <= lastActiveIndex; i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("active", user["status"])
		}

		for i := lastActiveIndex + 1; i < len(users); i++ {
			user := suite.readDataAsMap(users[i])
			suite.Equal("inactive", user["status"])
		}
	})

	suite.Run("OverrideDisabledDefaultSort", func() {
		// NoDefaultSortUserFindAllResource has no default sort
		// Add sorting via request
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

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())

		users := suite.readDataAsSlice(body.Data)
		suite.Len(users, 10)

		// Verify emails are in ascending order
		var prevEmail string
		for i, u := range users {
			user := suite.readDataAsMap(u)

			email := user["email"].(string)
			if i > 0 {
				suite.True(email >= prevEmail, "Email %s should be >= previous email %s", email, prevEmail)
			}

			prevEmail = email
		}
	})
}
