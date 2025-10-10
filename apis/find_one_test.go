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
type TestUserFindOneResource struct {
	api.Resource
	apis.FindOneAPI[TestUser, TestUserSearch]
}

func NewTestUserFindOneResource() api.Resource {
	return &TestUserFindOneResource{
		Resource:   api.NewResource("test/user"),
		FindOneAPI: apis.NewFindOneAPI[TestUser, TestUserSearch]().Public(),
	}
}

// Processed User Resource - with processor.
type ProcessedUserFindOneResource struct {
	api.Resource
	apis.FindOneAPI[TestUser, TestUserSearch]
}

type ProcessedUser struct {
	TestUser

	Processed bool `json:"processed"`
}

func NewProcessedUserFindOneResource() api.Resource {
	return &ProcessedUserFindOneResource{
		Resource: api.NewResource("test/user_processed"),
		FindOneAPI: apis.NewFindOneAPI[TestUser, TestUserSearch]().
			Public().
			Processor(func(user TestUser, search TestUserSearch, ctx fiber.Ctx) any {
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
	apis.FindOneAPI[TestUser, TestUserSearch]
}

func NewFilteredUserFineOneResource() api.Resource {
	return &FilteredUserFineOneResource{
		Resource: api.NewResource("test/user_filtered"),
		FindOneAPI: apis.NewFindOneAPI[TestUser, TestUserSearch]().
			Public().
			FilterApplier(func(search TestUserSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder] {
				return func(cb orm.ConditionBuilder) {
					cb.Equals("status", "active").GreaterThan("age", 32)
				}
			}),
	}
}

// Ordered User Resource - with order applier.
type OrderedUserFindOneResource struct {
	api.Resource
	apis.FindOneAPI[TestUser, TestUserSearch]
}

func NewOrderedUserFindOneResource() api.Resource {
	return &OrderedUserFindOneResource{
		Resource: api.NewResource("test/user_ordered"),
		FindOneAPI: apis.NewFindOneAPI[TestUser, TestUserSearch]().
			Public().
			SortApplier(func(search TestUserSearch, ctx fiber.Ctx) orm.ApplyFunc[apis.Sorter] {
				return func(s apis.Sorter) {
					s.OrderByDesc("age")
				}
			}),
	}
}

// FindOneTestSuite is the test suite for FindOne API tests.
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
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *FindOneTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// TestFindOneBasic tests basic FindOne functionality.
func (suite *FindOneTestSuite) TestFindOneBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user",
			Action:   "findOne",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user003",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(body.Message, i18n.T(result.OkMessage))
	suite.NotNil(body.Data)
	suite.Subset(body.Data, map[string]any{
		"id":          "user003",
		"name":        "Charlie Brown",
		"email":       "charlie@example.com",
		"age":         float64(28),
		"status":      "inactive",
		"description": "Designer",
	})
}

// TestFindOneNotFound tests FindOne when record doesn't exist.
func (suite *FindOneTestSuite) TestFindOneNotFound() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user",
			Action:   "findOne",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "nonexistent-id",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.Equal(body.Code, result.ErrCodeRecordNotFound)
	suite.Equal(body.Message, i18n.T(result.ErrMessageRecordNotFound))
	suite.Nil(body.Data)
}

// TestFindOneWithSearchApplier tests FindOne with custom search conditions.
func (suite *FindOneTestSuite) TestFindOneWithSearchApplier() {
	suite.Run("SearchByKeyword", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": "Johnson",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)
		suite.Subset(body.Data, map[string]any{
			"id":          "user001",
			"name":        "Alice Johnson",
			"email":       "alice@example.com",
			"age":         float64(25),
			"status":      "active",
			"description": "Software Engineer",
		})
	})

	suite.Run("SearchByEmail", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v1",
			},
			Params: map[string]any{
				"email": "grace@example.com",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)
		suite.Subset(body.Data, map[string]any{
			"id":          "user007",
			"name":        "Grace Lee",
			"email":       "grace@example.com",
			"age":         float64(29),
			"status":      "active",
			"description": "UX Researcher",
		})
	})

	suite.Run("SearchByAgeRange", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v1",
			},
			Params: map[string]any{
				"age": []int{33, 34},
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)
		suite.Subset(body.Data, map[string]any{
			"id":          "user010",
			"name":        "Jack Taylor",
			"email":       "jack@example.com",
			"age":         float64(33),
			"status":      "active",
			"description": "Team Lead",
		})
	})

	suite.Run("SearchByMultipleConditions", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v1",
			},
			Params: map[string]any{
				"email":  "ivy@example.com",
				"status": "inactive",
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)
		suite.Subset(body.Data, map[string]any{
			"id":          "user009",
			"name":        "Ivy Chen",
			"email":       "ivy@example.com",
			"age":         float64(26),
			"status":      "inactive",
			"description": "QA Engineer",
		})
	})
}

// TestFindOneWithProcessor tests FindOne with post-processing.
func (suite *FindOneTestSuite) TestFindOneWithProcessor() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_processed",
			Action:   "findOne",
			Version:  "v1",
		},
		Params: map[string]any{
			"id": "user001",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.NotNil(body.Data)
	suite.Subset(body.Data, map[string]any{
		"id":          "user001",
		"name":        "Alice Johnson",
		"email":       "alice@example.com",
		"age":         float64(25),
		"status":      "active",
		"description": "Software Engineer",
		"processed":   true,
	})
}

// TestFindOneWithFilterApplier tests FindOne with filter applier.
func (suite *FindOneTestSuite) TestFindOneWithFilterApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_filtered",
			Action:   "findOne",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.NotNil(body.Data)
	suite.Subset(body.Data, map[string]any{
		"id":          "user010",
		"name":        "Jack Taylor",
		"email":       "jack@example.com",
		"age":         float64(33),
		"status":      "active",
		"description": "Team Lead",
	})
}

// TestFindOneWithSortApplier tests FindOne with sort applier.
func (suite *FindOneTestSuite) TestFindOneWithSortApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_ordered",
			Action:   "findOne",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.NotNil(body.Data)
	suite.Subset(body.Data, map[string]any{
		"id":          "user006",
		"name":        "Frank Miller",
		"email":       "frank@example.com",
		"age":         float64(35),
		"status":      "inactive",
		"description": "Sales Manager",
	})
}

// TestFindOneNegativeCases tests negative scenarios.
func (suite *FindOneTestSuite) TestFindOneNegativeCases() {
	suite.Run("InvalidResource", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/nonexistent",
				Action:   "findOne",
				Version:  "v1",
			},
		})

		suite.Equal(404, resp.StatusCode)
	})

	suite.Run("InvalidAction", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "nonexistentAction",
				Version:  "v1",
			},
		})

		suite.Equal(404, resp.StatusCode)
	})

	suite.Run("InvalidVersion", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v999",
			},
		})

		suite.Equal(404, resp.StatusCode)
	})

	suite.Run("EmptySearchCriteria", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.True(body.IsOk())
		suite.NotNil(body.Data)
	})

	suite.Run("InvalidRangeValue", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v1",
			},
			Params: map[string]any{
				"age": []int{30}, // Invalid range - only one value
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		// Should still work, just ignore invalid range
		suite.True(body.IsOk())
	})

	suite.Run("MultipleConditionsNoMatch", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user",
				Action:   "findOne",
				Version:  "v1",
			},
			Params: map[string]any{
				"email":  "alice@example.com",
				"status": "inactive", // Alice is active, not inactive
			},
		})

		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.Equal(result.ErrCodeRecordNotFound, body.Code)
		suite.Nil(body.Data)
	})
}
