package csv

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/tabular"
)

type TestUser struct {
	Id       string      `tabular:"用户ID"                 validate:"required"`
	Name     string      `tabular:"姓名"                   validate:"required"`
	Email    string      `tabular:"邮箱"                   validate:"email"`
	Age      int         `tabular:"年龄"                   validate:"gte=0,lte=150"`
	Salary   float64     `tabular:"薪资,format=%.2f"`
	Birthday time.Time   `tabular:"生日,format=2006-01-02"`
	Active   bool        `tabular:"激活状态"`
	Remark   null.String `tabular:"备注"`
	Password string      `tabular:"-"` // Ignored field
}

func TestCSVExportImport(t *testing.T) {
	// Test data
	users := []TestUser{
		{
			Id:       "1",
			Name:     "张三",
			Email:    "zhangsan@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Date(1994, 1, 15, 0, 0, 0, 0, time.UTC),
			Active:   true,
			Remark:   null.StringFrom("测试用户1"),
		},
		{
			Id:       "2",
			Name:     "李四",
			Email:    "lisi@example.com",
			Age:      25,
			Salary:   8000.75,
			Birthday: time.Date(1999, 5, 20, 0, 0, 0, 0, time.UTC),
			Active:   false,
			Remark:   null.String{},
		},
	}

	// Export to CSV
	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(users)
	require.NoError(t, err)
	require.NotNil(t, buf)

	csvContent := buf.String()
	t.Logf("Exported CSV:\n%s", csvContent)

	// Verify header
	assert.Contains(t, csvContent, "用户ID")
	assert.Contains(t, csvContent, "姓名")
	assert.Contains(t, csvContent, "邮箱")

	// Import from CSV
	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	// Verify imported data
	importedUsers, ok := result.([]TestUser)
	require.True(t, ok)
	require.Len(t, importedUsers, 2)

	// Check first user
	assert.Equal(t, "1", importedUsers[0].Id)
	assert.Equal(t, "张三", importedUsers[0].Name)
	assert.Equal(t, "zhangsan@example.com", importedUsers[0].Email)
	assert.Equal(t, 30, importedUsers[0].Age)
	assert.InDelta(t, 10000.50, importedUsers[0].Salary, 0.01)
	assert.Equal(t, "1994-01-15", importedUsers[0].Birthday.Format("2006-01-02"))
	assert.True(t, importedUsers[0].Active)
	assert.True(t, importedUsers[0].Remark.Valid)
	assert.Equal(t, "测试用户1", importedUsers[0].Remark.ValueOrZero())

	// Check second user
	assert.Equal(t, "2", importedUsers[1].Id)
	assert.Equal(t, "李四", importedUsers[1].Name)
	assert.Equal(t, "lisi@example.com", importedUsers[1].Email)
	assert.Equal(t, 25, importedUsers[1].Age)
	assert.InDelta(t, 8000.75, importedUsers[1].Salary, 0.01)
	assert.Equal(t, "1999-05-20", importedUsers[1].Birthday.Format("2006-01-02"))
	assert.False(t, importedUsers[1].Active)
	assert.False(t, importedUsers[1].Remark.Valid)
}

func TestCSVImportWithCustomDelimiter(t *testing.T) {
	// CSV content with semicolon delimiter
	csvContent := `用户ID;姓名;邮箱
1;张三;zhangsan@example.com
2;李四;lisi@example.com`

	type SimpleUser struct {
		Id    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	// Import with custom delimiter
	importer := NewImporterFor[SimpleUser](WithImportDelimiter(';'))
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	users, ok := result.([]SimpleUser)
	require.True(t, ok)
	require.Len(t, users, 2)

	assert.Equal(t, 1, users[0].Id)
	assert.Equal(t, "张三", users[0].Name)
}

func TestCSVImportWithoutHeader(t *testing.T) {
	// CSV content without header
	csvContent := `1,张三,zhangsan@example.com
2,李四,lisi@example.com`

	type SimpleUser struct {
		Id    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	// Import without header
	importer := NewImporterFor[SimpleUser](WithoutHeader())
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	users, ok := result.([]SimpleUser)
	require.True(t, ok)
	require.Len(t, users, 2)

	assert.Equal(t, 1, users[0].Id)
	assert.Equal(t, "张三", users[0].Name)
}

func TestCSVExportWithoutHeader(t *testing.T) {
	type SimpleUser struct {
		Id    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	users := []SimpleUser{
		{Id: 1, Name: "张三", Email: "zhangsan@example.com"},
	}

	// Export without header
	exporter := NewExporterFor[SimpleUser](WithoutWriteHeader())
	buf, err := exporter.Export(users)
	require.NoError(t, err)

	csvContent := buf.String()
	assert.NotContains(t, csvContent, "用户ID")
	assert.Contains(t, csvContent, "1,张三,zhangsan@example.com")
}

func TestCSVImportWithSkipRows(t *testing.T) {
	// CSV content with title row to skip (with proper field count)
	csvContent := `用户数据表,,,
用户ID,姓名,邮箱
1,张三,zhangsan@example.com`

	type SimpleUser struct {
		Id    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	// Import with skip rows
	importer := NewImporterFor[SimpleUser](WithSkipRows(1))
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	users, ok := result.([]SimpleUser)
	require.True(t, ok)
	require.Len(t, users, 1)

	assert.Equal(t, 1, users[0].Id)
	assert.Equal(t, "张三", users[0].Name)
}

// TestSchema_ParseTags tests schema tag parsing.
func TestSchema_ParseTags(t *testing.T) {
	schema := tabular.NewSchemaFor[TestUser]()

	columns := schema.Columns()
	assert.NotEmpty(t, columns)

	// Find specific columns
	var idCol, nameCol, passwordCol *tabular.Column

	for i := range columns {
		col := columns[i]
		switch col.Name {
		case "用户ID":
			idCol = col
		case "姓名":
			nameCol = col
		case "Password":
			passwordCol = col
		}
	}

	// Verify ID column
	require.NotNil(t, idCol)
	assert.Equal(t, "用户ID", idCol.Name)

	// Verify Name column
	require.NotNil(t, nameCol)
	assert.Equal(t, "姓名", nameCol.Name)

	// Password field should be ignored
	assert.Nil(t, passwordCol)
}

// TestNoTagStruct tests struct fields without tags.
type TestNoTagStruct struct {
	ID   string
	Name string
	Age  int
}

func TestSchema_NoTags(t *testing.T) {
	schema := tabular.NewSchemaFor[TestNoTagStruct]()

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

	// Export to CSV
	exporter := NewExporterFor[TestNoTagStruct]()
	buf, err := exporter.Export(data)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestNoTagStruct]()
	result, importErrors, err := importer.Import(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestNoTagStruct)
	require.True(t, ok)
	assert.Len(t, imported, 2)

	// Verify data
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "Alice", imported[0].Name)
	assert.Equal(t, 30, imported[0].Age)
}

// TestImporter_ValidationErrors tests validation error handling.
func TestImporter_ValidationErrors(t *testing.T) {
	// CSV content with invalid data
	csvContent := `用户ID,姓名,邮箱,年龄
1,张三,invalid-email,200`

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Empty(t, imported) // No valid data
	assert.NotEmpty(t, importErrors)
}

// prefixFormatter is a custom formatter that adds prefix.
type prefixFormatter struct {
	prefix string
}

func (f *prefixFormatter) Format(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	return f.prefix + " " + fmt.Sprint(value), nil
}

// TestExport_CustomFormatter tests export with custom formatter.
func TestExport_CustomFormatter(t *testing.T) {
	users := []TestUser{
		{
			Id:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Active:   true,
		},
	}

	// Register custom formatter
	exporter := NewExporterFor[TestUser]()
	exporter.RegisterFormatter("prefix", &prefixFormatter{prefix: "ID:"})

	buf, err := exporter.Export(users)
	require.NoError(t, err)
	assert.NotNil(t, buf)
}

// prefixParser is a custom parser that removes prefix.
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

// TestImport_CustomParser tests import with custom parser.
func TestImport_CustomParser(t *testing.T) {
	// CSV content
	csvContent := `用户ID,姓名,邮箱
ID: 1,张三,zhang@example.com`

	importer := NewImporterFor[TestUser]()
	importer.RegisterParser("prefix_parser", &prefixParser{})

	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 1)
}

// TestExport_EmptyData tests exporting empty data.
func TestExport_EmptyData(t *testing.T) {
	var emptyUsers []TestUser

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(emptyUsers)
	require.NoError(t, err)

	csvContent := buf.String()
	// Should have header
	assert.Contains(t, csvContent, "用户ID")
	assert.Contains(t, csvContent, "姓名")
}

// TestExport_ToFile tests exporting to file.
func TestExport_ToFile(t *testing.T) {
	users := []TestUser{
		{
			Id:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Now(),
			Active:   true,
		},
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_csv_export_*.csv")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

// TestImport_FromFile tests importing from file.
func TestImport_FromFile(t *testing.T) {
	// First create a test CSV file
	users := []TestUser{
		{
			Id:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Now(),
			Active:   true,
		},
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_csv_import_*.csv")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Import from file
	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.ImportFromFile(filename)
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 1)
	assert.Equal(t, "1", imported[0].Id)
	assert.Equal(t, "张三", imported[0].Name)
}

// TestImport_EmptyRows tests importing CSV with empty rows.
func TestImport_EmptyRows(t *testing.T) {
	csvContent := `用户ID,姓名,邮箱
1,张三,zhang@example.com

2,李四,li@example.com`

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 2) // Empty row should be skipped
}

// TestImport_MissingColumns tests importing with missing columns.
func TestImport_MissingColumns(t *testing.T) {
	// CSV with only required fields
	csvContent := `用户ID,姓名,邮箱,年龄
1,张三,zhang@example.com,30`

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 1)

	// Verify missing fields have zero values
	assert.Equal(t, "1", imported[0].Id)
	assert.Equal(t, "张三", imported[0].Name)
	assert.Equal(t, 0.0, imported[0].Salary)
	assert.False(t, imported[0].Remark.Valid)
}

// TestImport_InvalidData tests importing invalid data.
func TestImport_InvalidData(t *testing.T) {
	csvContent := `用户ID,姓名,邮箱,年龄
1,张三,invalid-email,not-a-number`

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Empty(t, imported)
	assert.NotEmpty(t, importErrors)
}

// TestImport_LargeFile tests importing large CSV file.
func TestImport_LargeFile(t *testing.T) {
	// Create large test data
	count := 1000
	users := make([]TestUser, count)

	for i := range count {
		users[i] = TestUser{
			Id:       fmt.Sprintf("%d", i+1),
			Name:     fmt.Sprintf("用户%d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Age:      20 + (i % 50),
			Salary:   5000.0 + float64(i*100),
			Birthday: time.Now(),
			Active:   i%2 == 0,
		}
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_csv_large_*.csv")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.ImportFromFile(filename)
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, count)

	// Spot check
	assert.Equal(t, "1", imported[0].Id)
	assert.Equal(t, "用户1", imported[0].Name)
	assert.Equal(t, fmt.Sprintf("%d", count), imported[count-1].Id)
}

// TestExport_NullValues tests exporting null values.
func TestExport_NullValues(t *testing.T) {
	users := []TestUser{
		{
			Id:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Now(),
			Active:   true,
			Remark:   null.String{}, // Null value
		},
		{
			Id:       "2",
			Name:     "李四",
			Email:    "li@example.com",
			Age:      25,
			Salary:   8000.00,
			Birthday: time.Now(),
			Active:   false,
			Remark:   null.StringFrom("有备注"), // Valid value
		},
	}

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(users)
	require.NoError(t, err)

	// Import and verify
	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 2)

	// Verify null handling
	assert.False(t, imported[0].Remark.Valid)
	assert.True(t, imported[1].Remark.Valid)
	assert.Equal(t, "有备注", imported[1].Remark.ValueOrZero())
}

// TestRoundTrip tests full export-import cycle.
func TestRoundTrip(t *testing.T) {
	original := []TestUser{
		{
			Id:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Active:   true,
			Remark:   null.StringFrom("测试用户1"),
		},
		{
			Id:       "2",
			Name:     "李四",
			Email:    "li@example.com",
			Age:      25,
			Salary:   8000.75,
			Birthday: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			Active:   false,
			Remark:   null.String{},
		},
	}

	// Export
	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(original)
	require.NoError(t, err)

	// Import
	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, len(original))

	// Verify all fields match
	for i := range original {
		assert.Equal(t, original[i].Id, imported[i].Id)
		assert.Equal(t, original[i].Name, imported[i].Name)
		assert.Equal(t, original[i].Email, imported[i].Email)
		assert.Equal(t, original[i].Age, imported[i].Age)
		assert.InDelta(t, original[i].Salary, imported[i].Salary, 0.01)
		assert.Equal(t, original[i].Active, imported[i].Active)
		assert.Equal(t, original[i].Remark.Valid, imported[i].Remark.Valid)

		if original[i].Remark.Valid {
			assert.Equal(t, original[i].Remark.ValueOrZero(), imported[i].Remark.ValueOrZero())
		}
	}
}
