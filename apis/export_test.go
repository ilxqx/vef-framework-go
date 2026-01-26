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
	api.P

	Keyword *string `json:"keyword" search:"contains,column=name|email"`
	Status  *string `json:"status"  search:"eq"`
}

// Test Resources for Export

type TestUserExportResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportResource() api.Resource {
	return &TestUserExportResource{
		Resource: api.NewRPCResource("test/user_export"),
		Export:   apis.NewExport[ExportUser, ExportUserSearch]().Public(),
	}
}

type TestUserExportWithOptionsResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithOptionsResource() api.Resource {
	return &TestUserExportWithOptionsResource{
		Resource: api.NewRPCResource("test/user_export_opts"),
		Export: apis.NewExport[ExportUser, ExportUserSearch]().
			Public().
			WithExcelOptions(excel.WithSheetName("用户列表")),
	}
}

type TestUserExportWithFilenameResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithFilenameResource() api.Resource {
	return &TestUserExportWithFilenameResource{
		Resource: api.NewRPCResource("test/user_export_filename"),
		Export: apis.NewExport[ExportUser, ExportUserSearch]().
			Public().
			WithFilenameBuilder(func(search ExportUserSearch, ctx fiber.Ctx) string {
				return "custom_users.xlsx"
			}),
	}
}

type TestUserExportWithPreProcessorResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithPreProcessorResource() api.Resource {
	return &TestUserExportWithPreProcessorResource{
		Resource: api.NewRPCResource("test/user_export_preproc"),
		Export: apis.NewExport[ExportUser, ExportUserSearch]().
			Public().
			WithPreExport(func(models []ExportUser, search ExportUserSearch, ctx fiber.Ctx, db orm.DB) error {
				// Add custom header with count
				ctx.Set("X-Export-Count", string(rune('0'+len(models))))

				return nil
			}),
	}
}

type TestUserExportWithFilterResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportWithFilterResource() api.Resource {
	return &TestUserExportWithFilterResource{
		Resource: api.NewRPCResource("test/user_export_filter"),
		Export: apis.NewExport[ExportUser, ExportUserSearch]().
			WithCondition(func(cb orm.ConditionBuilder) {
				cb.Equals("status", "active")
			}).
			Public(),
	}
}

type TestUserExportCSVResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportCSVResource() api.Resource {
	return &TestUserExportCSVResource{
		Resource: api.NewRPCResource("test/user_export_csv"),
		Export: apis.NewExport[ExportUser, ExportUserSearch]().
			Public().
			WithDefaultFormat(apis.FormatCsv),
	}
}

type TestUserExportCSVWithOptionsResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportCSVWithOptionsResource() api.Resource {
	return &TestUserExportCSVWithOptionsResource{
		Resource: api.NewRPCResource("test/user_export_csv_opts"),
		Export: apis.NewExport[ExportUser, ExportUserSearch]().
			Public().
			WithDefaultFormat(apis.FormatCsv).
			WithCsvOptions(csv.WithExportDelimiter(';')),
	}
}

type TestUserExportCSVWithFilenameResource struct {
	api.Resource
	apis.Export[ExportUser, ExportUserSearch]
}

func NewTestUserExportCSVWithFilenameResource() api.Resource {
	return &TestUserExportCSVWithFilenameResource{
		Resource: api.NewRPCResource("test/user_export_csv_filename"),
		Export: apis.NewExport[ExportUser, ExportUserSearch]().
			Public().
			WithDefaultFormat(apis.FormatCsv).
			WithFilenameBuilder(func(search ExportUserSearch, ctx fiber.Ctx) string {
				return "custom_users.csv"
			}),
	}
}

// ExportTestSuite tests the Export API functionality
// including basic Excel/CSV exports, custom options, filename builders, pre-processors,
// filters, format overrides, and negative cases.
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
	suite.T().Logf("Testing basic Excel export for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType), "Should return Excel content type")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "attachment", "Should have attachment disposition")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "filename=", "Should include filename")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), ".xlsx", "Should have .xlsx extension")

	// Read and verify Excel file
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	// Verify it's a valid Excel file by checking signature
	// Excel files start with PK (ZIP signature)
	suite.Equal(byte('P'), body[0], "Should start with 'P' (ZIP signature)")
	suite.Equal(byte('K'), body[1], "Should have 'K' as second byte (ZIP signature)")
	suite.T().Logf("Successfully exported Excel file with %d bytes", len(body))
}

func (suite *ExportTestSuite) TestExportWithSearchFilter() {
	suite.T().Logf("Testing export with search filters for %s", suite.dbType)

	suite.Run("FilterByStatus", func() {
		status := "active"
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": status,
			},
		})

		suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
		suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			resp.Header.Get(fiber.HeaderContentType), "Should return Excel content type")

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err, "Should read response body successfully")
		suite.NotEmpty(body, "Should return non-empty response body")
		suite.T().Logf("Successfully exported filtered users (status=%s) with %d bytes", status, len(body))
	})

	suite.Run("FilterByKeyword", func() {
		keyword := "Engineer"
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": keyword,
			},
		})

		suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
		suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			resp.Header.Get(fiber.HeaderContentType), "Should return Excel content type")

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err, "Should read response body successfully")
		suite.NotEmpty(body, "Should return non-empty response body")
		suite.T().Logf("Successfully exported filtered users (keyword=%s) with %d bytes", keyword, len(body))
	})
}

func (suite *ExportTestSuite) TestExportWithCustomFilename() {
	suite.T().Logf("Testing export with custom filename for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_filename",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "custom_users.xlsx", "Should use custom filename")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")
	suite.T().Logf("Successfully exported with custom filename: %s", resp.Header.Get(fiber.HeaderContentDisposition))
}

func (suite *ExportTestSuite) TestExportWithPreProcessor() {
	suite.T().Logf("Testing export with pre-processor for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_preproc",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.NotEmpty(resp.Header.Get("X-Export-Count"), "Should set X-Export-Count header in pre-processor")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")
	suite.T().Logf("Pre-processor executed successfully with header: %s", resp.Header.Get("X-Export-Count"))
}

func (suite *ExportTestSuite) TestExportWithFilterApplier() {
	suite.T().Logf("Testing export with filter applier for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_filter",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType), "Should return Excel content type")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	// Parse the Excel file to verify only active users are exported
	importer := excel.NewImporterFor[ExportUser]()
	users, _, err := importer.Import(bytes.NewReader(body))
	suite.NoError(err, "Should parse exported Excel file successfully")

	exportedUsers := users.([]ExportUser)
	suite.NotEmpty(exportedUsers, "Should export at least one user")

	// Verify all exported users have status "active"
	for _, user := range exportedUsers {
		suite.Equal("active", user.Status, "Should only export active users")
	}

	suite.T().Logf("Filter applier successfully exported %d active users", len(exportedUsers))
}

func (suite *ExportTestSuite) TestExportEmptyResult() {
	suite.T().Logf("Testing export with empty result for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
		Params: map[string]any{
			"keyword": "NonexistentKeyword12345XYZ",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType), "Should return Excel content type")

	// Even empty export should return a valid Excel file with headers
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	// Verify it's still a valid Excel file
	suite.Equal(byte('P'), body[0], "Should start with 'P' (ZIP signature)")
	suite.Equal(byte('K'), body[1], "Should have 'K' as second byte (ZIP signature)")
	suite.T().Log("Empty result correctly returned valid Excel file with headers")
}

func (suite *ExportTestSuite) TestExportWithOptions() {
	suite.T().Logf("Testing export with options for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_opts",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		resp.Header.Get(fiber.HeaderContentType), "Should return Excel content type")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	// Verify the Excel file can be parsed successfully with custom sheet name
	importer := excel.NewImporterFor[ExportUser](excel.WithImportSheetName("用户列表"))
	users, _, err := importer.Import(bytes.NewReader(body))
	suite.NoError(err, "Should parse Excel file with custom sheet name successfully")

	exportedUsers := users.([]ExportUser)
	suite.NotEmpty(exportedUsers, "Should export at least one user")
	suite.T().Logf("Successfully exported %d users with custom options", len(exportedUsers))
}

func (suite *ExportTestSuite) TestExportNegativeCases() {
	suite.T().Logf("Testing export negative cases for %s", suite.dbType)

	suite.Run("InvalidSearchParameter", func() {
		// Export should handle invalid search parameters gracefully
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"nonexistent_field": "value",
			},
		})

		suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
		suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			resp.Header.Get(fiber.HeaderContentType), "Should return Excel content type")

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err, "Should read response body successfully")
		suite.NotEmpty(body, "Should return non-empty response body")
		suite.T().Log("Invalid search parameter handled gracefully")
	})
}

func (suite *ExportTestSuite) TestExportContentType() {
	suite.T().Logf("Testing export content type for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")

	// Verify correct content type for Excel files
	contentType := resp.Header.Get(fiber.HeaderContentType)
	suite.Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", contentType, "Should return correct Excel content type")

	// Verify Content-Disposition header
	contentDisposition := resp.Header.Get(fiber.HeaderContentDisposition)
	suite.Contains(contentDisposition, "attachment", "Should have attachment disposition")
	suite.Contains(contentDisposition, "filename=", "Should include filename")
	suite.T().Log("Content type and disposition headers verified successfully")
}

func (suite *ExportTestSuite) TestExportResponseHeaders() {
	suite.T().Logf("Testing export response headers for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")

	// Check all required response headers
	suite.NotEmpty(resp.Header.Get(fiber.HeaderContentType), "Should have Content-Type header")
	suite.NotEmpty(resp.Header.Get(fiber.HeaderContentDisposition), "Should have Content-Disposition header")

	// Verify the response body contains data
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")
	suite.Greater(len(body), 100, "Should return reasonably sized Excel file")
	suite.T().Logf("Response headers verified successfully, file size: %d bytes", len(body))
}

// CSV Export Tests

func (suite *ExportTestSuite) TestExportCSVBasic() {
	suite.T().Logf("Testing basic CSV export for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType), "Should return CSV content type")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "attachment", "Should have attachment disposition")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "filename=", "Should include filename")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), ".csv", "Should have .csv extension")

	// Read and verify CSV file
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	// Verify it's a valid CSV file by checking for headers
	content := string(body)
	suite.Contains(content, "姓名", "Should contain Chinese header for Name")
	suite.T().Logf("Successfully exported CSV file with %d bytes", len(body))
}

func (suite *ExportTestSuite) TestExportCSVWithSearchFilter() {
	suite.T().Logf("Testing CSV export with search filters for %s", suite.dbType)

	suite.Run("FilterByStatus", func() {
		status := "active"
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export_csv",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"status": status,
			},
		})

		suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
		suite.Equal("text/csv; charset=utf-8",
			resp.Header.Get(fiber.HeaderContentType), "Should return CSV content type")

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err, "Should read response body successfully")
		suite.NotEmpty(body, "Should return non-empty response body")
		suite.T().Logf("Successfully exported CSV with filter (status=%s)", status)
	})

	suite.Run("FilterByKeyword", func() {
		keyword := "Engineer"
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_export_csv",
				Action:   "export",
				Version:  "v1",
			},
			Params: map[string]any{
				"keyword": keyword,
			},
		})

		suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
		suite.Equal("text/csv; charset=utf-8",
			resp.Header.Get(fiber.HeaderContentType), "Should return CSV content type")

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err, "Should read response body successfully")
		suite.NotEmpty(body, "Should return non-empty response body")
		suite.T().Logf("Successfully exported CSV with filter (keyword=%s)", keyword)
	})
}

func (suite *ExportTestSuite) TestExportCSVWithCustomFilename() {
	suite.T().Logf("Testing CSV export with custom filename for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv_filename",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Contains(resp.Header.Get(fiber.HeaderContentDisposition), "custom_users.csv", "Should use custom filename")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")
	suite.T().Log("Successfully exported CSV with custom filename")
}

func (suite *ExportTestSuite) TestExportCSVWithOptions() {
	suite.T().Logf("Testing CSV export with options for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv_opts",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType), "Should return CSV content type")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	// Verify semicolon delimiter is used by parsing with semicolon delimiter
	importer := csv.NewImporterFor[ExportUser](csv.WithImportDelimiter(';'))
	users, _, err := importer.Import(bytes.NewReader(body))
	suite.NoError(err, "Should parse CSV with semicolon delimiter successfully")

	exportedUsers := users.([]ExportUser)
	suite.NotEmpty(exportedUsers, "Should export at least one user")
	suite.T().Logf("Successfully parsed CSV with custom delimiter, got %d users", len(exportedUsers))
}

func (suite *ExportTestSuite) TestExportCSVEmptyResult() {
	suite.T().Logf("Testing CSV export with empty result for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv",
			Action:   "export",
			Version:  "v1",
		},
		Params: map[string]any{
			"keyword": "NonexistentKeyword12345XYZ",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType), "Should return CSV content type")

	// Even empty export should return a valid CSV file with headers
	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	// Verify it contains headers
	content := string(body)
	suite.Contains(content, "姓名", "Should still contain headers")
	suite.T().Log("Empty result correctly returned valid CSV file with headers")
}

func (suite *ExportTestSuite) TestExportFormatOverride() {
	suite.T().Logf("Testing export format override for %s", suite.dbType)

	// Test format parameter override - use Excel endpoint but override to CSV
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export",
			Action:   "export",
			Version:  "v1",
		},
		Meta: map[string]any{
			"format": "csv",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.Equal("text/csv; charset=utf-8",
		resp.Header.Get(fiber.HeaderContentType), "Should return CSV content type after format override")

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err, "Should read response body successfully")
	suite.NotEmpty(body, "Should return non-empty response body")

	content := string(body)
	suite.Contains(content, "姓名", "Should contain CSV headers")
	suite.T().Log("Format override from Excel to CSV successful")
}

func (suite *ExportTestSuite) TestExportCSVContentType() {
	suite.T().Logf("Testing CSV export content type for %s", suite.dbType)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_export_csv",
			Action:   "export",
			Version:  "v1",
		},
	})

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")

	// Verify correct content type for CSV files
	contentType := resp.Header.Get(fiber.HeaderContentType)
	suite.Equal("text/csv; charset=utf-8", contentType, "Should return correct CSV content type")

	// Verify Content-Disposition header
	contentDisposition := resp.Header.Get(fiber.HeaderContentDisposition)
	suite.Contains(contentDisposition, "attachment", "Should have attachment disposition")
	suite.Contains(contentDisposition, "filename=", "Should include filename")
	suite.T().Log("CSV content type and disposition headers verified successfully")
}
