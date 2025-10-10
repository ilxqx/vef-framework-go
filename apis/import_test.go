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
	orm.Model     `bun:"extend"                     tabular:"-"`

	Name   string `json:"name"   tabular:"姓名,width=20" bun:",notnull"                  validate:"required"`
	Email  string `json:"email"  tabular:"邮箱,width=25" bun:",notnull"                  validate:"required,email"`
	Age    int    `json:"age"    tabular:"年龄,width=10" bun:",notnull"                  validate:"gte=0,lte=150"`
	Status string `json:"status" tabular:"状态,width=10" bun:",notnull,default:'active'" validate:"required,oneof=active inactive pending"`
}

// ImportUserSearch is the search parameters for ImportUser.
type ImportUserSearch struct {
	api.In
}

// Test Resources for Import

type TestUserImportResource struct {
	api.Resource
	apis.ImportAPI[ImportUser, ImportUserSearch]
}

func NewTestUserImportResource() api.Resource {
	return &TestUserImportResource{
		Resource:  api.NewResource("test/user_import"),
		ImportAPI: apis.NewImportAPI[ImportUser, ImportUserSearch]().Public(),
	}
}

type TestUserImportWithOptionsResource struct {
	api.Resource
	apis.ImportAPI[ImportUser, ImportUserSearch]
}

func NewTestUserImportWithOptionsResource() api.Resource {
	return &TestUserImportWithOptionsResource{
		Resource: api.NewResource("test/user_import_opts"),
		ImportAPI: apis.NewImportAPI[ImportUser, ImportUserSearch]().
			Public().
			ExcelOptions(excel.WithImportSheetName("用户列表")),
	}
}

type TestUserImportWithPreProcessorResource struct {
	api.Resource
	apis.ImportAPI[ImportUser, ImportUserSearch]
}

func NewTestUserImportWithPreProcessorResource() api.Resource {
	return &TestUserImportWithPreProcessorResource{
		Resource: api.NewResource("test/user_import_preproc"),
		ImportAPI: apis.NewImportAPI[ImportUser, ImportUserSearch]().
			Public().
			PreImport(func(models []ImportUser, search ImportUserSearch, ctx fiber.Ctx, db orm.Db) error {
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
	apis.ImportAPI[ImportUser, ImportUserSearch]
}

func NewTestUserImportWithPostProcessorResource() api.Resource {
	return &TestUserImportWithPostProcessorResource{
		Resource: api.NewResource("test/user_import_postproc"),
		ImportAPI: apis.NewImportAPI[ImportUser, ImportUserSearch]().
			Public().
			PostImport(func(models []ImportUser, search ImportUserSearch, ctx fiber.Ctx, db orm.Db) error {
				// Set custom header with count
				ctx.Set("X-Import-Count", string(rune('0'+len(models))))

				return nil
			}),
	}
}

type TestUserImportCSVResource struct {
	api.Resource
	apis.ImportAPI[ImportUser, ImportUserSearch]
}

func NewTestUserImportCSVResource() api.Resource {
	return &TestUserImportCSVResource{
		Resource: api.NewResource("test/user_import_csv"),
		ImportAPI: apis.NewImportAPI[ImportUser, ImportUserSearch]().
			Public().
			Format(apis.FormatCSV),
	}
}

type TestUserImportCSVWithOptionsResource struct {
	api.Resource
	apis.ImportAPI[ImportUser, ImportUserSearch]
}

func NewTestUserImportCSVWithOptionsResource() api.Resource {
	return &TestUserImportCSVWithOptionsResource{
		Resource: api.NewResource("test/user_import_csv_opts"),
		ImportAPI: apis.NewImportAPI[ImportUser, ImportUserSearch]().
			Public().
			Format(apis.FormatCSV).
			CSVOptions(csv.WithImportDelimiter(';')),
	}
}

// ImportTestSuite is the test suite for Import API tests.
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
	// Create test Excel file
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Import User 1", Email: "import1@example.com", Age: 30, Status: "active"},
		{Name: "Import User 2", Email: "import2@example.com", Age: 25, Status: "active"},
		{Name: "Import User 3", Email: "import3@example.com", Age: 28, Status: "inactive"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	// Create multipart request
	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import.xlsx", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(i18n.T(result.OkMessage), body.Message)

	// Verify response data
	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(3), data["total"])
}

func (suite *ImportTestSuite) TestImportWithValidationErrors() {
	// Create test Excel file with invalid data
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Valid User", Email: "valid@example.com", Age: 30, Status: "active"},
		{Name: "Invalid Email", Email: "invalid-email", Age: 25, Status: "active"},     // Invalid email
		{Name: "Invalid Age", Email: "test@example.com", Age: 200, Status: "active"},   // Invalid age > 150
		{Name: "Invalid Status", Email: "test2@example.com", Age: 25, Status: "wrong"}, // Invalid status
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	// Import should detect validation errors
	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_invalid.xlsx", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.False(body.IsOk())

	// Verify error data contains validation errors
	data := suite.readDataAsMap(body.Data)
	suite.NotNil(data["errors"])
	errors := suite.readDataAsSlice(data["errors"])
	suite.NotEmpty(errors)
}

func (suite *ImportTestSuite) TestImportWithMissingRequiredFields() {
	// Create test Excel file with missing required fields
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "", Email: "noemail@example.com", Age: 30, Status: "active"}, // Missing name
		{Name: "No Email", Email: "", Age: 25, Status: "active"},            // Missing email
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_missing.xlsx", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.False(body.IsOk())

	data := suite.readDataAsMap(body.Data)
	suite.NotNil(data["errors"])
}

func (suite *ImportTestSuite) TestImportWithPreProcessor() {
	// Create test Excel file with users
	// The preprocessor will change "inactive" status to "pending"
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Pre-processed User 1", Email: "preproc1@example.com", Age: 30, Status: "active"},
		{Name: "Pre-processed User 2", Email: "preproc2@example.com", Age: 25, Status: "inactive"}, // Will be changed to "pending" by preprocessor
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_preproc",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_preproc.xlsx", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected success but got error: %s", body.Message)

	// Verify data was imported
	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(2), data["total"])
}

func (suite *ImportTestSuite) TestImportWithPostProcessor() {
	// Create test Excel file
	exporter := excel.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Post-processed User 1", Email: "postproc1@example.com", Age: 30, Status: "active"},
		{Name: "Post-processed User 2", Email: "postproc2@example.com", Age: 25, Status: "active"},
		{Name: "Post-processed User 3", Email: "postproc3@example.com", Age: 28, Status: "inactive"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_postproc",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_postproc.xlsx", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	suite.NotEmpty(resp.Header.Get("X-Import-Count"))

	body := suite.readBody(resp)
	suite.True(body.IsOk())

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(3), data["total"])
}

func (suite *ImportTestSuite) TestImportEmptyFile() {
	// Create empty Excel file (with headers but no data rows)
	exporter := excel.NewExporterFor[ImportUser]()

	var testUsers []ImportUser

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	resp := suite.makeMultipartAPIRequest(api.Request{
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
		suite.Equal(200, resp.StatusCode)
		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Empty file should return error")
	}
}

func (suite *ImportTestSuite) TestImportLargeFile() {
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
	suite.NoError(err)

	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_large.xlsx", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(100), data["total"])
}

func (suite *ImportTestSuite) TestImportNegativeCases() {
	suite.Run("MissingFile", func() {
		// Try to import without providing a file
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_import",
				Action:   "import",
				Version:  "v1",
			},
		})

		// Request should fail with status 500 or 200 with error
		body := suite.readBody(resp)
		suite.False(body.IsOk())
	})

	suite.Run("InvalidFileFormat", func() {
		// Try to import a non-Excel file
		resp := suite.makeMultipartAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_import",
				Action:   "import",
				Version:  "v1",
			},
		}, "test.txt", []byte("This is not an Excel file"))

		// Should return error (either 500 or 200 with error body)
		if resp.StatusCode == 200 {
			body := suite.readBody(resp)
			suite.False(body.IsOk())
		} else {
			// 500 error is also acceptable for invalid file format
			suite.NotEqual(200, resp.StatusCode)
		}
	})

	suite.Run("JSONRequest", func() {
		// Import requires multipart/form-data, not JSON
		resp := suite.makeAPIRequest(api.Request{
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
			suite.False(body.IsOk())
		} else {
			// Error status is also acceptable
			suite.NotEqual(200, resp.StatusCode)
		}
	})

	suite.Run("CorruptedExcelFile", func() {
		// Try to import corrupted Excel file
		corruptedData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10} // Invalid Excel data

		resp := suite.makeMultipartAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "test/user_import",
				Action:   "import",
				Version:  "v1",
			},
		}, "corrupted.xlsx", corruptedData)

		// Should return error (either 500 or 200 with error body)
		if resp.StatusCode == 200 {
			body := suite.readBody(resp)
			suite.False(body.IsOk())
		} else {
			// 500 error is also acceptable for corrupted file
			suite.NotEqual(200, resp.StatusCode)
		}
	})
}

// CSV Import Tests

func (suite *ImportTestSuite) TestImportCSVBasic() {
	// Create test CSV file
	exporter := csv.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "CSV User 1", Email: "csv1@example.com", Age: 30, Status: "active"},
		{Name: "CSV User 2", Email: "csv2@example.com", Age: 25, Status: "active"},
		{Name: "CSV User 3", Email: "csv3@example.com", Age: 28, Status: "inactive"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	// Create multipart request
	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_csv",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import.csv", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())
	suite.Equal(i18n.T(result.OkMessage), body.Message)

	// Verify response data
	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(3), data["total"])
}

func (suite *ImportTestSuite) TestImportCSVWithValidationErrors() {
	// Create test CSV file with invalid data
	exporter := csv.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Valid User", Email: "valid@example.com", Age: 30, Status: "active"},
		{Name: "Invalid Email", Email: "invalid-email", Age: 25, Status: "active"},     // Invalid email
		{Name: "Invalid Age", Email: "test@example.com", Age: 200, Status: "active"},   // Invalid age > 150
		{Name: "Invalid Status", Email: "test2@example.com", Age: 25, Status: "wrong"}, // Invalid status
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	// Import should detect validation errors
	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_csv",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_invalid.csv", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.False(body.IsOk())

	// Verify error data contains validation errors
	data := suite.readDataAsMap(body.Data)
	suite.NotNil(data["errors"])
	errors := suite.readDataAsSlice(data["errors"])
	suite.NotEmpty(errors)
}

func (suite *ImportTestSuite) TestImportCSVWithOptions() {
	// Create test CSV file with semicolon delimiter
	exporter := csv.NewExporterFor[ImportUser](csv.WithExportDelimiter(';'))
	testUsers := []ImportUser{
		{Name: "CSV Options User 1", Email: "csvopts1@example.com", Age: 30, Status: "active"},
		{Name: "CSV Options User 2", Email: "csvopts2@example.com", Age: 25, Status: "active"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import_csv_opts",
			Action:   "import",
			Version:  "v1",
		},
	}, "test_import_opts.csv", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(2), data["total"])
}

func (suite *ImportTestSuite) TestImportFormatOverride() {
	// Test format parameter override
	exporter := csv.NewExporterFor[ImportUser]()
	testUsers := []ImportUser{
		{Name: "Format Override User", Email: "override@example.com", Age: 30, Status: "active"},
	}

	buf, err := exporter.Export(testUsers)
	suite.NoError(err)

	// Use Excel endpoint but override format to CSV via parameter
	resp := suite.makeMultipartAPIRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "test/user_import",
			Action:   "import",
			Version:  "v1",
		},
		Params: map[string]any{
			"format": "csv",
		},
	}, "test_import_override.csv", buf.Bytes())

	suite.Equal(200, resp.StatusCode)
	body := suite.readBody(resp)
	suite.True(body.IsOk())

	data := suite.readDataAsMap(body.Data)
	suite.Equal(float64(1), data["total"])
}

// Helper method for multipart requests.
func (suite *ImportTestSuite) makeMultipartAPIRequest(req api.Request, filename string, fileContent []byte) *http.Response {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	// Add API request fields
	_ = writer.WriteField("resource", req.Resource)
	_ = writer.WriteField("action", req.Action)
	_ = writer.WriteField("version", req.Version)

	// Add params as JSON string if present
	if req.Params != nil {
		paramsJSON, err := encoding.ToJSON(req.Params)
		suite.NoError(err)

		_ = writer.WriteField("params", paramsJSON)
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
