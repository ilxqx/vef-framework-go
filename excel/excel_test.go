package excel

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ilxqx/vef-framework-go/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// TestUser is a test struct for Excel operations.
type TestUser struct {
	ID        string      `excel:"width=15" validate:"required"`
	Name      string      `excel:"姓名,width=20" validate:"required"`
	Email     string      `excel:"邮箱,width=25" validate:"email"`
	Age       int         `excel:"name=年龄,width=10" validate:"gte=0,lte=150"`
	Salary    float64     `excel:"name=薪资,width=15,format=%.2f"`
	CreatedAt time.Time   `excel:"name=创建时间,width=20,format=2006-01-02 15:04:05"`
	Status    int         `excel:"name=状态,width=10"`
	Remark    null.String `excel:"name=备注,width=30"`
	Password  string      `excel:"-"` // Ignored field
}

func TestExporter_ExportToFile(t *testing.T) {
	// Prepare test data
	now := time.Now()
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: now,
			Status:    1,
			Remark:    null.StringFrom("测试用户1"),
			Password:  "secret123", // Should be ignored
		},
		{
			ID:        "2",
			Name:      "李四",
			Email:     "li@example.com",
			Age:       25,
			Salary:    8000.75,
			CreatedAt: now,
			Status:    2,
			Remark:    null.String{}, // Null value
			Password:  "secret456",
		},
	}

	// Create exporter
	exporter := NewExporterFor[TestUser]()

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_users_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

func TestImporter_ImportFromFile(t *testing.T) {
	// First, create a test Excel file
	now := time.Now()
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: now,
			Status:    1,
			Remark:    null.StringFrom("测试用户1"),
		},
		{
			ID:        "2",
			Name:      "李四",
			Email:     "li@example.com",
			Age:       25,
			Salary:    8000.75,
			CreatedAt: now,
			Status:    2,
			Remark:    null.String{},
		},
	}

	// Export to temp file
	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_import_users_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2)

	// Verify data
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "张三", imported[0].Name)
	assert.Equal(t, "zhang@example.com", imported[0].Email)
	assert.Equal(t, 30, imported[0].Age)
	assert.InDelta(t, 10000.50, imported[0].Salary, 0.01)
	assert.Equal(t, 1, imported[0].Status)
	assert.True(t, imported[0].Remark.Valid)
	assert.Equal(t, "测试用户1", imported[0].Remark.ValueOrZero())

	assert.Equal(t, "2", imported[1].ID)
	assert.Equal(t, "李四", imported[1].Name)
	assert.False(t, imported[1].Remark.Valid)
}

func TestSchema_ParseTags(t *testing.T) {
	schema := NewSchemaFor[TestUser]()

	columns := schema.Columns()
	assert.NotEmpty(t, columns)

	// Find specific columns
	var idCol, nameCol, passwordCol *Column
	for i := range columns {
		col := columns[i]
		switch col.Name {
		case "ID":
			idCol = col
		case "姓名":
			nameCol = col
		case "Password":
			passwordCol = col
		}
	}

	// Verify ID column (no name specified, uses field name)
	require.NotNil(t, idCol)
	assert.Equal(t, "ID", idCol.Name)
	assert.Equal(t, 15.0, idCol.Width)

	// Verify Name column (using shorthand syntax: "姓名")
	require.NotNil(t, nameCol)
	assert.Equal(t, "姓名", nameCol.Name)
	assert.Equal(t, 20.0, nameCol.Width)

	// Password field should be ignored
	assert.Nil(t, passwordCol)
}

func TestImporter_ValidationErrors(t *testing.T) {
	// Create a test file with invalid data
	invalidUsers := []TestUser{
		{
			ID:     "1",
			Name:   "张三",
			Email:  "invalid-email", // Invalid email
			Age:    200,             // Age > 150
			Salary: 10000,
		},
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_invalid_users_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(invalidUsers, filename)
	require.NoError(t, err)

	// Import should catch validation errors
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, imported) // No valid data
	assert.NotEmpty(t, importErrors)
}

// TestNoTagStruct tests struct fields without excel tags (should use field name as column name)
type TestNoTagStruct struct {
	ID   string
	Name string
	Age  int
}

func TestSchema_NoTags(t *testing.T) {
	schema := NewSchemaFor[TestNoTagStruct]()

	columns := schema.Columns()
	assert.Len(t, columns, 3)

	// Verify columns use field names
	assert.Equal(t, "ID", columns[0].Name)
	assert.Equal(t, "Name", columns[1].Name)
	assert.Equal(t, "Age", columns[2].Name)
}

func TestExportImport_NoTags(t *testing.T) {
	// Test data
	data := []TestNoTagStruct{
		{ID: "1", Name: "Alice", Age: 30},
		{ID: "2", Name: "Bob", Age: 25},
	}

	// Export to temp file
	exporter := NewExporterFor[TestNoTagStruct]()
	tmpFile, err := os.CreateTemp("", "test_no_tags_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(data, filename)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestNoTagStruct]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestNoTagStruct)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2)

	// Verify data
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "Alice", imported[0].Name)
	assert.Equal(t, 30, imported[0].Age)
}

// prefixFormatter is a custom formatter that adds prefix
type prefixFormatter struct {
	prefix string
}

func (f *prefixFormatter) Format(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	return f.prefix + " " + fmt.Sprint(value), nil
}

// TestExport_CustomFormatter tests export with custom formatter
func TestExport_CustomFormatter(t *testing.T) {
	// Test data
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.Local),
			Status:    1,
			Remark:    null.StringFrom("测试用户"),
		},
	}

	// Register custom formatter for ID field
	exporter := NewExporterFor[TestUser]()
	exporter.RegisterFormatter("prefix", &prefixFormatter{prefix: "ID:"})

	tmpFile, err := os.CreateTemp("", "test_custom_formatter_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

// TestExport_ToBuffer tests exporting to buffer
func TestExport_ToBuffer(t *testing.T) {
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: time.Now(),
			Status:    1,
			Remark:    null.StringFrom("测试"),
		},
	}

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(users)

	require.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

// TestExport_EmptyData tests exporting empty data
func TestExport_EmptyData(t *testing.T) {
	var emptyUsers []TestUser

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_empty_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(emptyUsers, filename)
	require.NoError(t, err)

	// Verify file exists and has header
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

// TestExport_WithOptions tests export with custom options
func TestExport_WithOptions(t *testing.T) {
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: time.Now(),
			Status:    1,
		},
	}

	exporter := NewExporterFor[TestUser](WithSheetName("用户数据"))
	tmpFile, err := os.CreateTemp("", "test_options_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

// prefixParser is a custom parser that removes prefix
type prefixParser struct{}

func (p *prefixParser) Parse(cellValue string, targetType reflect.Type) (any, error) {
	if cellValue == "" {
		return "", nil
	}
	// Remove "ID: " prefix
	if len(cellValue) > 4 {
		return cellValue[4:], nil
	}
	return cellValue, nil
}

// TestImport_CustomParser tests import with custom parser
func TestImport_CustomParser(t *testing.T) {
	// Create test file manually
	now := time.Now()
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: now,
			Status:    1,
			Remark:    null.StringFrom("测试"),
		},
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_custom_parser_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Import with custom parser
	importer := NewImporterFor[TestUser]()
	importer.RegisterParser("prefix_parser", &prefixParser{})

	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 1)
}

// TestImport_FromReader tests importing from io.Reader
func TestImport_FromReader(t *testing.T) {
	// Create test file
	now := time.Now()
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: now,
			Status:    1,
		},
	}

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(users)
	require.NoError(t, err)

	// Import from buffer
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.Import(buf)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 1)
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "张三", imported[0].Name)
}

// TestImport_WithOptions tests import with custom options
func TestImport_WithOptions(t *testing.T) {
	// Create test file with custom sheet name
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: time.Now(),
			Status:    1,
		},
	}

	exporter := NewExporterFor[TestUser](WithSheetName("用户数据"))
	tmpFile, err := os.CreateTemp("", "test_import_options_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Import with sheet name option
	importer := NewImporterFor[TestUser](WithImportSheetName("用户数据"))
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 1)
}

// TestImport_EmptyRows tests importing file with empty rows
func TestImport_EmptyRows(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_empty_rows_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	// Create file with empty rows manually using excelize
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Write header
	_ = f.SetCellValue(sheetName, "A1", "ID")
	_ = f.SetCellValue(sheetName, "B1", "姓名")
	_ = f.SetCellValue(sheetName, "C1", "邮箱")

	// Write data with empty row in middle
	_ = f.SetCellValue(sheetName, "A2", "1")
	_ = f.SetCellValue(sheetName, "B2", "张三")
	_ = f.SetCellValue(sheetName, "C2", "zhang@example.com")

	// Row 3 is empty

	_ = f.SetCellValue(sheetName, "A4", "2")
	_ = f.SetCellValue(sheetName, "B4", "李四")
	_ = f.SetCellValue(sheetName, "C4", "li@example.com")

	err = f.SaveAs(filename)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2) // Empty row should be skipped
}

// TestImport_MissingColumns tests importing file with missing columns
func TestImport_MissingColumns(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_missing_columns_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	// Create file with only some columns (but enough to pass validation)
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Write header (include required fields to pass validation)
	_ = f.SetCellValue(sheetName, "A1", "ID")
	_ = f.SetCellValue(sheetName, "B1", "姓名")
	_ = f.SetCellValue(sheetName, "C1", "邮箱")
	_ = f.SetCellValue(sheetName, "D1", "年龄")

	// Write data - missing some optional columns
	_ = f.SetCellValue(sheetName, "A2", "1")
	_ = f.SetCellValue(sheetName, "B2", "张三")
	_ = f.SetCellValue(sheetName, "C2", "zhang@example.com")
	_ = f.SetCellValue(sheetName, "D2", "30")
	// Missing: Salary, CreatedAt, Status, Remark

	err = f.SaveAs(filename)
	require.NoError(t, err)

	// Import - should use zero values for missing fields
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 1)
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "张三", imported[0].Name)
	assert.Equal(t, "zhang@example.com", imported[0].Email)
	assert.Equal(t, 30, imported[0].Age)
	assert.Equal(t, 0.0, imported[0].Salary)  // Missing field should be zero value
	assert.Equal(t, 0, imported[0].Status)    // Missing field should be zero value
	assert.False(t, imported[0].Remark.Valid) // Missing field should be invalid null
}

// TestImport_InvalidData tests importing file with invalid data
func TestImport_InvalidData(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_invalid_data_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	// Create file with invalid data
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Write header
	_ = f.SetCellValue(sheetName, "A1", "ID")
	_ = f.SetCellValue(sheetName, "B1", "姓名")
	_ = f.SetCellValue(sheetName, "C1", "邮箱")
	_ = f.SetCellValue(sheetName, "D1", "年龄")

	// Write invalid data
	_ = f.SetCellValue(sheetName, "A2", "1")
	_ = f.SetCellValue(sheetName, "B2", "张三")
	_ = f.SetCellValue(sheetName, "C2", "invalid-email")
	_ = f.SetCellValue(sheetName, "D2", "not-a-number")

	err = f.SaveAs(filename)
	require.NoError(t, err)

	// Import - should have errors
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, imported)
	assert.NotEmpty(t, importErrors)
}

// TestImport_LargeFile tests importing file with many rows
func TestImport_LargeFile(t *testing.T) {
	// Create large test data
	count := 1000
	users := make([]TestUser, count)
	now := time.Now()

	for i := range count {
		users[i] = TestUser{
			ID:        fmt.Sprintf("%d", i+1),
			Name:      fmt.Sprintf("用户%d", i+1),
			Email:     fmt.Sprintf("user%d@example.com", i+1),
			Age:       20 + (i % 50),
			Salary:    5000.0 + float64(i*100),
			CreatedAt: now,
			Status:    i % 3,
		}
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_large_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, count)

	// Spot check some data
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "用户1", imported[0].Name)
	assert.Equal(t, fmt.Sprintf("%d", count), imported[count-1].ID)
}

// TestExport_NullValues tests exporting null values
func TestExport_NullValues(t *testing.T) {
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: time.Now(),
			Status:    1,
			Remark:    null.String{}, // Null value
		},
		{
			ID:        "2",
			Name:      "李四",
			Email:     "li@example.com",
			Age:       25,
			Salary:    8000.00,
			CreatedAt: time.Now(),
			Status:    2,
			Remark:    null.StringFrom("有备注"), // Valid value
		},
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_null_values_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Import and verify
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2)

	// Verify null handling
	assert.False(t, imported[0].Remark.Valid)
	assert.True(t, imported[1].Remark.Valid)
	assert.Equal(t, "有备注", imported[1].Remark.ValueOrZero())
}

// TestRoundTrip tests full export-import cycle
func TestRoundTrip(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.Local)
	original := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: now,
			Status:    1,
			Remark:    null.StringFrom("测试用户1"),
		},
		{
			ID:        "2",
			Name:      "李四",
			Email:     "li@example.com",
			Age:       25,
			Salary:    8000.75,
			CreatedAt: now,
			Status:    2,
			Remark:    null.String{},
		},
	}

	// Export to temp file
	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_roundtrip_*.xlsx")
	require.NoError(t, err)
	filename := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(filename)

	err = exporter.ExportToFile(original, filename)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, len(original))

	// Verify all fields match
	for i := range original {
		assert.Equal(t, original[i].ID, imported[i].ID)
		assert.Equal(t, original[i].Name, imported[i].Name)
		assert.Equal(t, original[i].Email, imported[i].Email)
		assert.Equal(t, original[i].Age, imported[i].Age)
		assert.InDelta(t, original[i].Salary, imported[i].Salary, 0.01)
		assert.Equal(t, original[i].Status, imported[i].Status)
		assert.Equal(t, original[i].Remark.Valid, imported[i].Remark.Valid)
		if original[i].Remark.Valid {
			assert.Equal(t, original[i].Remark.ValueOrZero(), imported[i].Remark.ValueOrZero())
		}
	}
}
