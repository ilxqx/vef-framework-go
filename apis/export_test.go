package apis_test

import (
	"bytes"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/internal/orm"
)

// ExportUser is the test model for export tests (uses tabular tags).
type ExportUser struct {
	bun.BaseModel `bun:"table:export_user,alias:eu"`
	orm.Model     `tabular:"-" bun:"extend"`

	Name   string `json:"name"   tabular:"姓名,width=20" bun:",notnull"`
	Email  string `json:"email"  tabular:"邮箱,width=25" bun:",notnull"`
	Age    int    `json:"age"    tabular:"年龄,width=10" bun:",notnull"`
	Status string `json:"status" tabular:"状态,width=10" bun:",notnull,default:'active'"`
}

// ExportUserSearch is the search parameters for ExportUser.
type ExportUserSearch struct {
	api.In

	Keyword *string `json:"keyword" search:"contains,column=name|email"`
	Status  *string `json:"status"  search:"eq"`
}

// Test Resources for Export

type TestUserExportResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportResource() api.Resource {
	return &TestUserExportResource{
		Resource:  api.NewResource("test/user_export"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().Public(),
	}
}

type TestUserExportWithOptionsResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithOptionsResource() api.Resource {
	return &TestUserExportWithOptionsResource{
		Resource: api.NewResource("test/user_export_opts"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().
			Public().
			ExcelOptions(excel.WithSheetName("用户列表")),
	}
}

type TestUserExportWithFilenameResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithFilenameResource() api.Resource {
	return &TestUserExportWithFilenameResource{
		Resource: api.NewResource("test/user_export_filename"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().
			Public().
			FilenameBuilder(func(search ExportUserSearch, ctx fiber.Ctx) string {
				return "custom_users.xlsx"
			}),
	}
}

type TestUserExportWithPreProcessorResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithPreProcessorResource() api.Resource {
	return &TestUserExportWithPreProcessorResource{
		Resource: api.NewResource("test/user_export_preproc"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().
			Public().
			PreExport(func(models []ExportUser, search ExportUserSearch, ctx fiber.Ctx, db orm.Db) error {
				// Add custom header with count
				ctx.Set("X-Export-Count", string(rune('0'+len(models))))

				return nil
			}),
	}
}

type TestUserExportWithFilterResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithFilterResource() api.Resource {
	return &TestUserExportWithFilterResource{
		Resource: api.NewResource("test/user_export_filter"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().
			Public().
			FilterApplier(func(search ExportUserSearch, ctx fiber.Ctx) orm.ApplyFunc[orm.ConditionBuilder] {
				return func(cb orm.ConditionBuilder) {
					cb.Equals("status", "active")
				}
			}),
	}
}

type TestUserExportCSVResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportCSVResource() api.Resource {
	return &TestUserExportCSVResource{
		Resource: api.NewResource("test/user_export_csv"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().
			Public().
			Format(apis.FormatCSV),
	}
}

type TestUserExportCSVWithOptionsResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportCSVWithOptionsResource() api.Resource {
	return &TestUserExportCSVWithOptionsResource{
		Resource: api.NewResource("test/user_export_csv_opts"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().
			Public().
			Format(apis.FormatCSV).
			CSVOptions(csv.WithExportDelimiter(';')),
	}
}

type TestUserExportCSVWithFilenameResource struct {
	api.Resource
	apis.ExportAPI[ExportUser, ExportUserSearch]
}

func NewTestUserExportCSVWithFilenameResource() api.Resource {
	return &TestUserExportCSVWithFilenameResource{
		Resource: api.NewResource("test/user_export_csv_filename"),
		ExportAPI: apis.NewExportAPI[ExportUser, ExportUserSearch]().
			Public().
			Format(apis.FormatCSV).
			FilenameBuilder(func(search ExportUserSearch, ctx fiber.Ctx) string {
				return "custom_users.csv"
			}),
	}
}

// ExportTestSuite is the test suite for Export API tests.
type ExportTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *ExportTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserExportResource,
		NewTestUserExportWithOptionsResource,
		NewTestUserExportWithFilenameResource,
		NewTestUserExportWithPreProcessorResource,
		NewTestUserExportWithFilterResource,
		NewTestUserExportCSVResource,
		NewTestUserExportCSVWithOptionsResource,
		NewTestUserExportCSVWithFilenameResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *ExportTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// Export Tests

func (suite *ExportTestSuite) TestExportBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType))
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "attachment")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "filename=")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), ".xlsx")

	// Read and verify Excel file
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	// Verify it's a valid Excel file by checking signature
	// Excel files start with PK (ZIP signature)
	suite.Equal(byte('P'), body[0])
	suite.Equal(byte('K'), body[1])
}

func (suite *ExportTestSuite) TestExportWithSearchFilter() {
	suite.Run("FilterByStatus", func() {
		status := "active"
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": status,
			},
		})

		suite.Equal(200, resp.StatusCode)
		suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)
		suite.NotEmpty(body)
	})

	suite.Run("FilterByKeyword", func() {
		keyword := "Engineer"
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": keyword,
			},
		})

		suite.Equal(200, resp.StatusCode)
		suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)
		suite.NotEmpty(body)
	})
}

func (suite *ExportTestSuite) TestExportWithCustomFilename() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_filename",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "custom_users.xlsx")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)
}

func (suite *ExportTestSuite) TestExportWithPreProcessor() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_preproc",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.NotEmpty(resp.Header.Get("X-Export-Count"))

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)
}

func (suite *ExportTestSuite) TestExportWithFilterApplier() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_filter",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	// Parse the Excel file to verify only active users are exported
	importer := excel.NewImporterFor[ExportUser]()
	users, _, err := importer.Import(bytes.NewReader(body))
	suite.NoError(err)

	exportedUsers := users.([]ExportUser)
	suite.NotEmpty(exportedUsers)

	// Verify all exported users have status "active"
	for _, user := range exportedUsers {
		suite.Equal("active", user.Status, "Filter should only export active users")
	}
}

func (suite *ExportTestSuite) TestExportEmptyResult() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
		Params: map[string]any{
			"keyword": "NonexistentKeyword12345XYZ",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType))

	// Even empty export should return a valid Excel file with headers
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	// Verify it's still a valid Excel file
	suite.Equal(byte('P'), body[0])
	suite.Equal(byte('K'), body[1])
}

func (suite *ExportTestSuite) TestExportWithOptions() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_opts",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	// Verify the Excel file can be parsed successfully with custom sheet name
	importer := excel.NewImporterFor[ExportUser](excel.WithImportSheetName("用户列表"))
	users, _, err := importer.Import(bytes.NewReader(body))
	suite.NoError(err)

	exportedUsers := users.([]ExportUser)
	suite.NotEmpty(exportedUsers)
}

func (suite *ExportTestSuite) TestExportNegativeCases() {
	suite.Run("InvalidSearchParameter", func() {
		// Export should handle invalid search parameters gracefully
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"nonexistent_field": "value",
			},
		})

		suite.Equal(200, resp.StatusCode)
		suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)
		suite.NotEmpty(body)
	})
}

func (suite *ExportTestSuite) TestExportContentType() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)

	// Verify correct content type for Excel files
	contentType := resp.Header.Get(fiber.HeaderContentType)
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", contentType)

	// Verify Content-Disposition header
	contentDisposition := resp.Header.Get(fiber.HeaderContentDisposition)
	suite.Contains(contentDisposition, "attachment")
	suite.Contains(contentDisposition, "filename=")
}

func (suite *ExportTestSuite) TestExportResponseHeaders() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)

	// Check all required response headers
	suite.NotEmpty(resp.Header.Get(fiber.HeaderContentType))
	suite.NotEmpty(resp.Header.Get(fiber.HeaderContentDisposition))

	// Verify the response body contains data
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)
	suite.Greater(len(body), 100) // Excel file should be reasonably sized
}

// CSV Export Tests

func (suite *ExportTestSuite) TestExportCSVBasic() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType))
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "attachment")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "filename=")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), ".csv")

	// Read and verify CSV file
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	// Verify it's a valid CSV file by checking for headers
	content := string(body)
	suite.Contains(content, "姓名") // Should contain Chinese header for Name
}

func (suite *ExportTestSuite) TestExportCSVWithSearchFilter() {
	suite.Run("FilterByStatus", func() {
		status := "active"
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export_csv",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": status,
			},
		})

		suite.Equal(200, resp.StatusCode)
		suite.Equal("text/csv; charset=utf-8",
			resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)
		suite.NotEmpty(body)
	})

	suite.Run("FilterByKeyword", func() {
		keyword := "Engineer"
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export_csv",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": keyword,
			},
		})

		suite.Equal(200, resp.StatusCode)
		suite.Equal("text/csv; charset=utf-8",
			resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)
		suite.NotEmpty(body)
	})
}

func (suite *ExportTestSuite) TestExportCSVWithCustomFilename() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv_filename",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "custom_users.csv")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)
}

func (suite *ExportTestSuite) TestExportCSVWithOptions() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv_opts",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	// Verify semicolon delimiter is used by parsing with semicolon delimiter
	importer := csv.NewImporterFor[ExportUser](csv.WithImportDelimiter(';'))
	users, _, err := importer.Import(bytes.NewReader(body))
	suite.NoError(err)

	exportedUsers := users.([]ExportUser)
	suite.NotEmpty(exportedUsers, "Should successfully parse CSV with semicolon delimiter")
}

func (suite *ExportTestSuite) TestExportCSVEmptyResult() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv",
			Action:   "export",
			Version:  "v1",
		},
		Params: map[string]any{
			"keyword": "NonexistentKeyword12345XYZ",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType))

	// Even empty export should return a valid CSV file with headers
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	// Verify it contains headers
	content := string(body)
	suite.Contains(content, "姓名") // Should still have headers
}

func (suite *ExportTestSuite) TestExportFormatOverride() {
	// Test format parameter override - use Excel endpoint but override to CSV
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
		Params: map[string]any{
			"format": "csv",
		},
	})

	suite.Equal(200, resp.StatusCode)
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)
	suite.NotEmpty(body)

	content := string(body)
	suite.Contains(content, "姓名")
}

func (suite *ExportTestSuite) TestExportCSVContentType() {
	resp := suite.makeAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Equal(200, resp.StatusCode)

	// Verify correct content type for CSV files
	contentType := resp.Header.Get(fiber.HeaderContentType)
	suite.Equal("text/csv; charset=utf-8", contentType)

	// Verify Content-Disposition header
	contentDisposition := resp.Header.Get(fiber.HeaderContentDisposition)
	suite.Contains(contentDisposition, "attachment")
	suite.Contains(contentDisposition, "filename=")
}
