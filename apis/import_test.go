package apis_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v3"
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

// ImportUser is the test model for import tests (uses tabular tags).
type ImportUser struct {
	bun.BaseModel `bun:"table:import_user,alias:iu"`
	orm.Model     `tabular:"-" bun:"extend"`

	Name   string `json:"name"   tabular:"姓名,width=20" bun:",notnull"                  validate:"required"`
	Email  string `json:"email"  tabular:"邮箱,width=25" bun:",notnull"                  validate:"required,email"`
	Age    int    `json:"age"    tabular:"年龄,width=10" bun:",notnull"                  validate:"gte=0,lte=150"`
	Status string `json:"status" tabular:"状态,width=10" bun:",notnull,default:'active'" validate:"required,oneof=active inactive pending"`
}

// ImportUserSearch is the search parameters for ImportUser.
type ImportUserSearch struct {
	api.P
}

// Test Resources for Import

type TestUserImportResource struct {
	api.Resource
	apis.Import[ImportUser]
}

func NewTestUserImportResource() api.Resource {
	return &TestUserImportResource{
		Resource: api.NewRPCResource("test/user_import"),
		Import:   apis.NewImport[ImportUser]().Public(),
	}
}

type TestUserImportWithOptionsResource struct {
	api.Resource
	apis.Import[ImportUser]
}

func NewTestUserImportWithOptionsResource() api.Resource {
	return &TestUserImportWithOptionsResource{
		Resource: api.NewRPCResource("test/user_import_opts"),
		Import: apis.NewImport[ImportUser]().
			Public().
			WithExcelOptions(excel.WithImportSheetName("用户列表")),
	}
}

type TestUserImportWithPreProcessorResource struct {
	api.Resource
	apis.Import[ImportUser]
}

func NewTestUserImportWithPreProcessorResource() api.Resource {
	return &TestUserImportWithPreProcessorResource{
		Resource: api.NewRPCResource("test/user_import_preproc"),
		Import: apis.NewImport[ImportUser]().
			Public().
			WithPreImport(func(models []ImportUser, _ orm.InsertQuery, _ fiber.Ctx, _ orm.DB) error {
				// Pre-process all models - change inactive to pending
				for i := range models {
					if models[i].Status == "inactive" {
						models[i].Status = "pending"
					}
				}

				return nil
			}),
	}
}

type TestUserImportWithPostProcessorResource struct {
	api.Resource
	apis.Import[ImportUser]
}

func NewTestUserImportWithPostProcessorResource() api.Resource {
	return &TestUserImportWithPostProcessorResource{
		Resource: api.NewRPCResource("test/user_import_postproc"),
		Import: apis.NewImport[ImportUser]().
			Public().
			WithPostImport(func(models []ImportUser, ctx fiber.Ctx, _ orm.DB) error {
				// Set custom header with count
				ctx.Set("X-Import-Count", string(rune('0'+len(models))))

				return nil
			}),
	}
}

type TestUserImportCSVResource struct {
	api.Resource
	apis.Import[ImportUser]
}

func NewTestUserImportCSVResource() api.Resource {
	return &TestUserImportCSVResource{
		Resource: api.NewRPCResource("test/user_import_csv"),
		Import: apis.NewImport[ImportUser]().
			Public().
			WithDefaultFormat(apis.FormatCsv),
	}
}

type TestUserImportCSVWithOptionsResource struct {
	api.Resource
	apis.Import[ImportUser]
}

func NewTestUserImportCSVWithOptionsResource() api.Resource {
	return &TestUserImportCSVWithOptionsResource{
		Resource: api.NewRPCResource("test/user_import_csv_opts"),
		Import: apis.NewImport[ImportUser]().
			Public().
			WithDefaultFormat(apis.FormatCsv).
			WithCsvOptions(csv.WithImportDelimiter(';')),
	}
}

// ImportTestSuite tests the Import API functionality
// including basic Excel/CSV imports, validation errors, pre/post processors, format overrides,
// large file handling, and negative cases.
type ImportTestSuite struct {
	BaseSuite
}

// SetupSuite runs once before all tests in the suite.
func (suite *ImportTestSuite) SetupSuite() {
	suite.setupBaseSuite(
		NewTestUserImportResource,
		NewTestUserImportWithOptionsResource,
		NewTestUserImportWithPreProcessorResource,
		NewTestUserImportWithPostProcessorResource,
		NewTestUserImportCSVResource,
		NewTestUserImportCSVWithOptionsResource,
	)
}

// TearDownSuite runs once after all tests in the suite.
func (suite *ImportTestSuite) TearDownSuite() {
	suite.tearDownBaseSuite()
}

// Import Tests

func (suite *ImportTestSuite) TestImportBasic() {
	suite.T().Logf("Testing basic Excel import for %s", suite.dbType)

	// Create test Excel file
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Import User 1", Email: "import1@example.com", Age: 30, Status: "active"},
		{Name: "Import User 2", Email: "import2@example.com", Age: 25, Status: "active"},
		{Name: "Import User 3", Email: "import3@example.com", Age: 28, Status: "inactive"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users to Excel successfully")

	// Create multipart request
	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import.xlsx", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return success response")
	suite.Equal(i18n.T(result.OkMessage), body.Message, "Should return OK message")

	// Verify response data
	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(3), data["total"], "Should import exactly 3 users")
	suite.T().Logf("Successfully imported %v users", data["total"])
}

func (suite *ImportTestSuite) TestImportWithValidationErrors() {
	suite.T().Logf("Testing import with validation errors for %s", suite.dbType)

	// Create test Excel file with invalid data
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Valid User", Email: "valid@example.com", Age: 30, Status: "active"},
		{Name: "Invalid Email", Email: "invalid-email", Age: 25, Status: "active"},     // Invalid email
		{Name: "Invalid Age", Email: "test@example.com", Age: 200, Status: "active"},   // Invalid age > 150
		{Name: "Invalid Status", Email: "test2@example.com", Age: 25, Status: "wrong"}, // Invalid status
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users with invalid data to Excel successfully")

	// Import should detect validation errors
	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_invalid.xlsx", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should return failure response due to validation errors")

	// Verify error data contains validation errors
	data := suite.readDataAsMap(body.Data)
	suite.Require().NotNil(data["errors"], "Should contain errors field in response data")
	errors := suite.readDataAsSlice(data["errors"])
	suite.NotEmpty(errors, "Should contain validation errors in errors array")
	suite.T().Logf("Validation detected %d error(s) as expected", len(errors))
}

func (suite *ImportTestSuite) TestImportWithMissingRequiredFields() {
	suite.T().Logf("Testing import with missing required fields for %s", suite.dbType)

	// Create test Excel file with missing required fields
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "", Email: "noemail@example.com", Age: 30, Status: "active"}, // Missing name
		{Name: "No Email", Email: "", Age: 25, Status: "active"},            // Missing email
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users with missing fields to Excel successfully")

	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_missing.xlsx", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should return failure response due to missing required fields")

	data := suite.readDataAsMap(body.Data)
	suite.Require().NotNil(data["errors"], "Should contain errors field in response data")
	suite.T().Log("Missing required fields correctly rejected")
}

func (suite *ImportTestSuite) TestImportWithPreProcessor() {
	suite.T().Logf("Testing import with pre-processor for %s", suite.dbType)

	// Create test Excel file with users
	// The preprocessor will change "inactive" status to "pending"
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Pre-processed User 1", Email: "preproc1@example.com", Age: 30, Status: "active"},
		{Name: "Pre-processed User 2", Email: "preproc2@example.com", Age: 25, Status: "inactive"}, // Will be changed to "pending" by preprocessor
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users to Excel successfully")

	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_preproc",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_preproc.xlsx", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return success response, got error: %s", body.Message)

	// Verify data was imported
	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(2), data["total"], "Should import exactly 2 users after pre-processing")
	suite.T().Logf("Pre-processor successfully transformed and imported %v users", data["total"])
}

func (suite *ImportTestSuite) TestImportWithPostProcessor() {
	suite.T().Logf("Testing import with post-processor for %s", suite.dbType)

	// Create test Excel file
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Post-processed User 1", Email: "postproc1@example.com", Age: 30, Status: "active"},
		{Name: "Post-processed User 2", Email: "postproc2@example.com", Age: 25, Status: "active"},
		{Name: "Post-processed User 3", Email: "postproc3@example.com", Age: 28, Status: "inactive"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users to Excel successfully")

	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_postproc",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_postproc.xlsx", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	suite.NotEmpty(resp.Header.Get("X-Import-Count"), "Should set X-Import-Count header in post-processor")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return success response")

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(3), data["total"], "Should import exactly 3 users")
	suite.T().Logf("Post-processor executed, imported %v users with header: %s", data["total"], resp.Header.Get("X-Import-Count"))
}

func (suite *ImportTestSuite) TestImportEmptyFile() {
	suite.T().Logf("Testing import with empty Excel file for %s", suite.dbType)

	// Create empty Excel file (with headers but no data rows)
	exporter := excel.NewExporterFor[ImportUser]()

	var testUsers []ImportUser

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export empty user list to Excel successfully")

	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_empty.xlsx", buf.Bytes())

	// Note: Excel importer returns an error when there are no data rows
	// This is the expected behavior - empty files are rejected
	// Status can be either 500 (error during processing) or 200 with error body
	if resp.StatusCode == 500 {
		suite.T().Log("Empty file correctly rejected with 500 status")
	} else {
		suite.Equal(200, resp.StatusCode, "Should return HTTP 200 status")
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should return error response for empty file")
		suite.T().Log("Empty file correctly rejected with error response")
	}
}

func (suite *ImportTestSuite) TestImportLargeFile() {
	suite.T().Logf("Testing import with large Excel file for %s", suite.dbType)

	// Create large test file with many rows
	exporter := excel.NewExporterFor[ImportUser]()

	testUsers := make([]ImportUser, 100)
	for i := range testUsers {
		testUsers[i] = ImportUser{
			Name:   "Bulk User " + string(rune('A'+i%26)),
			Email:  "bulkuser" + string(rune('0'+i%10)) + "@example.com",
			Age:    20 + (i % 50),
			Status: []string{"active", "inactive"}[i%2],
		}
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export 100 test users to Excel successfully")

	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_large.xlsx", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return success response")

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(100), data["total"], "Should import exactly 100 users")
	suite.T().Logf("Successfully imported large file with %v users", data["total"])
}

func (suite *ImportTestSuite) TestImportNegativeCases() {
	suite.T().Logf("Testing import negative cases for %s", suite.dbType)

	suite.Run("MissingFile", func() {
		// Try to import without providing a file
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_import",
				Action:   "import",
				Version:  "v1",
			},
		})

		// Request should fail with status 500 or 200 with error
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Should fail when importing without file")
		suite.T().Log("Missing file correctly rejected")
	})

	suite.Run("InvalidFileFormat", func() {
		// Try to import a non-Excel file
		resp := suite.makeMultipartApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_import",
				Action:   "import",
				Version:  "v1",
			},
		}, "test.txt", []byte("This is not an Excel file"))

		// Should return error (either 500 or 200 with error body)
		if resp.StatusCode == 200 {
			body := suite.readBody(resp)
			suite.False(body.IsOk(), "Should reject invalid file format")
			suite.T().Log("Invalid file format correctly rejected with error response")
		} else {
			// 500 error is also acceptable for invalid file format
			suite.NotEqual(200, resp.StatusCode, "Should return error status for invalid format")
			suite.T().Log("Invalid file format correctly rejected with error status")
		}
	})

	suite.Run("JSONRequest", func() {
		// Import requires multipart/form-data, not JSON
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_import",
				Action:   "import",
				Version:  "v1",
			},
			Params: map[string]any{
				"file": "some-file.xlsx",
			},
		})

		// Should fail because no file was provided or wrong content type
		if resp.StatusCode == 200 {
			body := suite.readBody(resp)
			suite.False(body.IsOk(), "Should fail for JSON request without multipart file")
			suite.T().Log("JSON request correctly rejected with error response")
		} else {
			// Error status is also acceptable
			suite.NotEqual(200, resp.StatusCode, "Should return error status for JSON request")
			suite.T().Log("JSON request correctly rejected with error status")
		}
	})

	suite.Run("CorruptedExcelFile", func() {
		// Try to import corrupted Excel file
		corruptedData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10} // Invalid Excel data

		resp := suite.makeMultipartApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_import",
				Action:   "import",
				Version:  "v1",
			},
		}, "corrupted.xlsx", corruptedData)

		// Should return error (either 500 or 200 with error body)
		if resp.StatusCode == 200 {
			body := suite.readBody(resp)
			suite.False(body.IsOk(), "Should reject corrupted Excel file")
			suite.T().Log("Corrupted file correctly rejected with error response")
		} else {
			// 500 error is also acceptable for corrupted file
			suite.NotEqual(200, resp.StatusCode, "Should return error status for corrupted file")
			suite.T().Log("Corrupted file correctly rejected with error status")
		}
	})
}

// CSV Import Tests

func (suite *ImportTestSuite) TestImportCSVBasic() {
	suite.T().Logf("Testing basic CSV import for %s", suite.dbType)

	// Create test CSV file
	exporter := csv.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "CSV User 1", Email: "csv1@example.com", Age: 30, Status: "active"},
		{Name: "CSV User 2", Email: "csv2@example.com", Age: 25, Status: "active"},
		{Name: "CSV User 3", Email: "csv3@example.com", Age: 28, Status: "inactive"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users to CSV successfully")

	// Create multipart request
	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_csv",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import.csv", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return success response")
	suite.Equal(i18n.T(result.OkMessage), body.Message, "Should return OK message")

	// Verify response data
	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(3), data["total"], "Should import exactly 3 users from CSV")
	suite.T().Logf("Successfully imported %v users from CSV", data["total"])
}

func (suite *ImportTestSuite) TestImportCSVWithValidationErrors() {
	suite.T().Logf("Testing CSV import with validation errors for %s", suite.dbType)

	// Create test CSV file with invalid data
	exporter := csv.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Valid User", Email: "valid@example.com", Age: 30, Status: "active"},
		{Name: "Invalid Email", Email: "invalid-email", Age: 25, Status: "active"},     // Invalid email
		{Name: "Invalid Age", Email: "test@example.com", Age: 200, Status: "active"},   // Invalid age > 150
		{Name: "Invalid Status", Email: "test2@example.com", Age: 25, Status: "wrong"}, // Invalid status
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users with invalid data to CSV successfully")

	// Import should detect validation errors
	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_csv",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_invalid.csv", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Should return failure response due to validation errors")

	// Verify error data contains validation errors
	data := suite.readDataAsMap(body.Data)
	suite.Require().NotNil(data["errors"], "Should contain errors field in response data")
	errors := suite.readDataAsSlice(data["errors"])
	suite.NotEmpty(errors, "Should contain validation errors in errors array")
	suite.T().Logf("CSV validation detected %d error(s) as expected", len(errors))
}

func (suite *ImportTestSuite) TestImportCSVWithOptions() {
	suite.T().Logf("Testing CSV import with custom delimiter for %s", suite.dbType)

	// Create test CSV file with semicolon delimiter
	exporter := csv.NewExporterFor[ImportUser](csv.WithExportDelimiter(';'))
	testUsers := []ImportUser{
		{Name: "CSV Options User 1", Email: "csvopts1@example.com", Age: 30, Status: "active"},
		{Name: "CSV Options User 2", Email: "csvopts2@example.com", Age: 25, Status: "active"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test users to CSV with semicolon delimiter successfully")

	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_csv_opts",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_opts.csv", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return success response")

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(2), data["total"], "Should import exactly 2 users with custom delimiter")
	suite.T().Logf("Successfully imported %v users with semicolon delimiter", data["total"])
}

func (suite *ImportTestSuite) TestImportFormatOverride() {
	suite.T().Logf("Testing import format override for %s", suite.dbType)

	// Test format parameter override
	exporter := csv.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Format Override User", Email: "override@example.com", Age: 30, Status: "active"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err, "Should export test user to CSV successfully")

	// Use Excel endpoint but override format to CSV via parameter
	resp := suite.makeMultipartApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
		Meta: map[string]any{
			"format": "csv",
		},
	}, "test_import_override.csv", buf.Bytes())

	suite.Require().Equal(200, resp.StatusCode, "Should return HTTP 200 status")
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Should return success response")

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(1), data["total"], "Should import exactly 1 user with format override")
	suite.T().Logf("Successfully imported %v user with format override from Excel to CSV", data["total"])
}

// Helper method for multipart requests.
func (suite *ImportTestSuite) makeMultipartApiRequest(req api.Request, filename string, fileContent []byte) *http.Response {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	// Add Api request fields
	_ = writer.WriteField("resource", req.Resource)
	_ = writer.WriteField("action", req.Action)
	_ = writer.WriteField("version", req.Version)

	// Add params as JSON string if present
	if req.Params != nil {
		paramsJSON, err := encoding.ToJSON(req.Params)
		suite.NoError(err)

		_ = writer.WriteField("params", paramsJSON)
	}

	// Add meta as JSON string if present
	if req.Meta != nil {
		metaJSON, err := encoding.ToJSON(req.Meta)
		suite.NoError(err)

		_ = writer.WriteField("meta", metaJSON)
	}

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	suite.NoError(err)
	_, err = part.Write(fileContent)
	suite.NoError(err)

	err = writer.Close()
	suite.NoError(err)

	httpReq := httptest.NewRequest(fiber.MethodPost, "/api", &buf)
	httpReq.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())

	resp, err := suite.app.Test(httpReq)
	suite.Require().NoError(err)

	return resp
}
