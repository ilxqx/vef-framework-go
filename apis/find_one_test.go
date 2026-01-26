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
type TestUserFindOneResource struct {
	api.Resource
	apis.FindOne[TestUser, TestUserSearch]
}

func NewTestUserFindOneResource() api.Resource {
	return &TestUserFindOneResource{
		Resource: api.NewRPCResource("test/user"),
		FindOne:  apis.NewFindOne[TestUser, TestUserSearch]().Public(),
	}
}

// Processed User Resource - with processor.
type ProcessedUserFindOneResource struct {
	api.Resource
	apis.FindOne[TestUser, TestUserSearch]
}

type ProcessedUser struct {
	TestUser

	Processed bool `json:"processed"`
}

func NewProcessedUserFindOneResource() api.Resource {
	return &ProcessedUserFindOneResource{
		Resource: api.NewRPCResource("test/user_processed"),
		FindOne: apis.NewFindOne[TestUser, TestUserSearch]().
			Public().
			WithProcessor(func(user TestUser, search TestUserSearch, ctx fiber.Ctx) any {
				return ProcessedUser{
					TestUser:  user,
					Processed: true,
				}
			}),
	}
}

// Filtered User Resource - with filter applier.
type FilteredUserFineOneResource struct {
	api.Resource
	apis.FindOne[TestUser, TestUserSearch]
}

func NewFilteredUserFineOneResource() api.Resource {
	return &FilteredUserFineOneResource{
		Resource: api.NewRPCResource("test/user_filtered"),
		FindOne: apis.NewFindOne[TestUser, TestUserSearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				cb.Equals("status", "active").GreaterThan("age", 32)
			}).
			Public(),
	}
}

// Ordered User Resource - with order applier.
type OrderedUserFindOneResource struct {
	api.Resource
	apis.FindOne[TestUser, TestUserSearch]
}

func NewOrderedUserFindOneResource() api.Resource {
	return &OrderedUserFindOneResource{
		Resource: api.NewRPCResource("test/user_ordered"),
		FindOne: apis.NewFindOne[TestUser, TestUserSearch]().
			WithDefaultSort(&sort.OrderSpec{
				Column:    "age",
				Direction: sort.OrderDesc,
			}).
			Public(),
	}
}

// AuditUser User Resource - with audit user names.
type AuditUserTestUserFindOneResource struct {
	api.Resource
	apis.FindOne[TestUser, TestUserSearch]
}

func NewAuditUserTestUserFindOneResource() api.Resource {
	return &AuditUserTestUserFindOneResource{
		Resource: api.NewRPCResource("test/user_audit"),
		FindOne: apis.NewFindOne[TestUser, TestUserSearch]().
			WithAuditUserNames((*TestAuditUser)(nil)).
			Public(),
	}
}

// FindOneTestSuite tests the FindOne API functionality
// including basic queries, search filters, processors, sorting, audit user names, and negative cases.
type FindOneTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *FindOneTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserFindOneResource,
		NewProcessedUserFindOneResource,
		NewFilteredUserFineOneResource,
		NewOrderedUserFindOneResource,
		NewAuditUserTestUserFindOneResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindOneTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindOneBasic tests basic FindOne functionality.
func (suite *FindOneTestSuite) TestFindOneBasic() {
	suite.T().Logf("Testing FindOne API basic functionality for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user",
			Action:   "find_one",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user003",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.Equal(body.Message, i18n.T(result.OkMessage), "Should return OK message")
	suite.NotNil(body.Data, "Data should not be nil")
	suite.Subset(body.Data, map[string]any{
		"id":          "user003",
		"name":        "Charlie Brown",
		"email":       "charlie@example.com",
		"age":         float64(28),
		"status":      "inactive",
		"description": "Designer",
	}, "Should return correct user data")

	suite.T().Logf("Found user: user003 (Charlie Brown)")
}

// TestFindOneNotFound tests FindOne when record doesn't exist.
func (suite *FindOneTestSuite) TestFindOneNotFound() {
	suite.T().Logf("Testing FindOne API record not found for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user",
			Action:   "find_one",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "nonexistent-id",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.Equal(body.Code, result.ErrCodeRecordNotFound, "Should return record not found error code")
	suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound), "Should return record not found message")
	suite.Nil(body.Data, "Data should be nil when record not found")

	suite.T().Logf("Record not found as expected for nonexistent-id")
}

// TestFindOneWithSearchApplier tests FindOne with custom search conditions.
func (suite *FindOneTestSuite) TestFindOneWithSearchApplier() {
	suite.T().Logf("Testing FindOne API with search filters for %s", suite.dbType)

	suite.Run("SearchByKeyword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Johnson",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")
		suite.Subset(body.Data, map[string]any{
			"id":          "user001",
			"name":        "Alice Johnson",
			"email":       "alice@example.com",
			"age":         float64(25),
			"status":      "active",
			"description": "Software Engineer",
		}, "Should return user matching keyword 'Johnson'")

		suite.T().Logf("Found user by keyword: user001 (Alice Johnson)")
	})

	suite.Run("SearchByEmail", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v1",
			},
			Params: map[string]any{
				"email": "grace@example.com",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")
		suite.Subset(body.Data, map[string]any{
			"id":          "user007",
			"name":        "Grace Lee",
			"email":       "grace@example.com",
			"age":         float64(29),
			"status":      "active",
			"description": "UX Researcher",
		}, "Should return user matching email")

		suite.T().Logf("Found user by email: user007 (Grace Lee)")
	})

	suite.Run("SearchByAgeRange", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v1",
			},
			Params: map[string]any{
				"age": []int{33, 34},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")
		suite.Subset(body.Data, map[string]any{
			"id":          "user010",
			"name":        "Jack Taylor",
			"email":       "jack@example.com",
			"age":         float64(33),
			"status":      "active",
			"description": "Team Lead",
		}, "Should return user in age range 33-34")

		suite.T().Logf("Found user by age range: user010 (Jack Taylor, age 33)")
	})

	suite.Run("SearchByMultipleConditions", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v1",
			},
			Params: map[string]any{
				"email":  "ivy@example.com",
				"status": "inactive",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response")
		suite.NotNil(body.Data, "Data should not be nil")
		suite.Subset(body.Data, map[string]any{
			"id":          "user009",
			"name":        "Ivy Chen",
			"email":       "ivy@example.com",
			"age":         float64(26),
			"status":      "inactive",
			"description": "QA Engineer",
		}, "Should return user matching multiple conditions")

		suite.T().Logf("Found user by multiple conditions: user009 (Ivy Chen)")
	})
}

// TestFindOneWithProcessor tests FindOne with post-processing.
func (suite *FindOneTestSuite) TestFindOneWithWithProcessor() {
	suite.T().Logf("Testing FindOne API with processor for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_processed",
			Action:   "find_one",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user001",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")
	suite.Subset(body.Data, map[string]any{
		"id":          "user001",
		"name":        "Alice Johnson",
		"email":       "alice@example.com",
		"age":         float64(25),
		"status":      "active",
		"description": "Software Engineer",
		"processed":   true,
	}, "Should return processed user data")

	suite.T().Logf("Found user with post-processing applied: user001 (processed=true)")
}

// TestFindOneWithFilterApplier tests FindOne with filter applier.
func (suite *FindOneTestSuite) TestFindOneWithFilterApplier() {
	suite.T().Logf("Testing FindOne API with filter applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_filtered",
			Action:   "find_one",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")
	suite.Subset(body.Data, map[string]any{
		"id":          "user010",
		"name":        "Jack Taylor",
		"email":       "jack@example.com",
		"age":         float64(33),
		"status":      "active",
		"description": "Team Lead",
	}, "Should return user matching filter (status=active AND age>32)")

	suite.T().Logf("Found user with filter applier: user010 (Jack Taylor)")
}

// TestFindOneWithSortApplier tests FindOne with sort applier.
func (suite *FindOneTestSuite) TestFindOneWithSortApplier() {
	suite.T().Logf("Testing FindOne API with sort applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_ordered",
			Action:   "find_one",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")
	suite.Subset(body.Data, map[string]any{
		"id":          "user006",
		"name":        "Frank Miller",
		"email":       "frank@example.com",
		"age":         float64(35),
		"status":      "inactive",
		"description": "Sales Manager",
	}, "Should return user with highest age (sorted by age DESC)")

	suite.T().Logf("Found user with sort applier: user006 (Frank Miller, age 35 - oldest)")
}

// TestFindOneNegativeCases tests negative scenarios.
func (suite *FindOneTestSuite) TestFindOneNegativeCases() {
	suite.T().Logf("Testing FindOne API negative cases for %s", suite.dbType)

	suite.Run("InvalidResource", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/nonexistent",
				Action:   "find_one",
				Version:  "v1",
			},
		})

		suite.Equal(404, resp.StatusCode, "Should return 404 for invalid resource")

		suite.T().Logf("Invalid resource returned 404 as expected")
	})

	suite.Run("InvalidAction", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "nonexistentAction",
				Version:  "v1",
			},
		})

		suite.Equal(404, resp.StatusCode, "Should return 404 for invalid action")

		suite.T().Logf("Invalid action returned 404 as expected")
	})

	suite.Run("InvalidVersion", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v999",
			},
		})

		suite.Equal(404, resp.StatusCode, "Should return 404 for invalid version")

		suite.T().Logf("Invalid version returned 404 as expected")
	})

	suite.Run("EmptySearchCriteria", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should return successful response with empty criteria")
		suite.NotNil(body.Data, "Data should not be nil")

		suite.T().Logf("Empty search criteria returned first record")
	})

	suite.Run("InvalidRangeValue", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v1",
			},
			Params: map[string]any{
				"age": []int{30},
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Should work even with invalid range format")

		suite.T().Logf("Invalid range value handled gracefully")
	})

	suite.Run("MultipleConditionsNoMatch", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "find_one",
				Version:  "v1",
			},
			Params: map[string]any{
				"email":  "alice@example.com",
				"status": "inactive",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 status code")
		body := suite.readBody(resp)
		suite.Equal(result.ErrCodeRecordNotFound, body.Code, "Should return record not found error code")
		suite.Nil(body.Data, "Data should be nil when no match found")

		suite.T().Logf("Multiple conflicting conditions returned no match as expected")
	})
}

// TestFindOneWithAuditUserNames tests FindOne with audit user names populated.
func (suite *FindOneTestSuite) TestFindOneWithAuditUserNames() {
	suite.T().Logf("Testing FindOne API with audit user names for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_audit",
			Action:   "find_one",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user001",
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 status code")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return successful response")
	suite.NotNil(body.Data, "Data should not be nil")

	user := suite.readDataAsMap(body.Data)
	suite.Equal("user001", user["id"], "Should return correct user id")
	suite.Equal("Alice Johnson", user["name"], "Should return correct user name")

	suite.NotNil(user["createdByName"], "Should have createdByName populated")
	suite.NotNil(user["updatedByName"], "Should have updatedByName populated")

	suite.Equal("John Doe", user["createdByName"], "Should return correct creator name")
	suite.Equal("Jane Smith", user["updatedByName"], "Should return correct updater name")

	suite.T().Logf("Found user with audit names: user001 (created by: %s, updated by: %s)", user["createdByName"], user["updatedByName"])
}
