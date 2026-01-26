package excel

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/tabular"
)

// TestUser is a test struct for Excel operations.
type TestUser struct {
	ID        string      `tabular:"width=15"                                      validate:"required"`
	Name      string      `tabular:"姓名,width=20"                                   validate:"required"`
	Email     string      `tabular:"邮箱,width=25"                                   validate:"email"`
	Age       int         `tabular:"name=年龄,width=10"                              validate:"gte=0,lte=150"`
	Salary    float64     `tabular:"name=薪资,width=15,format=%.2f"`
	CreatedAt time.Time   `tabular:"name=创建时间,width=20,format=2006-01-02 15:04:05"`
	Status    int         `tabular:"name=状态,width=10"`
	Remark    null.String `tabular:"name=备注,width=30"`
	Password  string      `tabular:"-"` // Ignored field
}

func TestExporterExportToFile(t *testing.T) {
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
			Password:  "secret123",
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
			Password:  "secret456",
		},
	}

	exporter := NewExporterFor[TestUser]()

	tmpFile, err := os.CreateTemp("", "test_users_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

func TestImporterImportFromFile(t *testing.T) {
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

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_import_users_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2)

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

func TestSchemaParseTags(t *testing.T) {
	schema := tabular.NewSchemaFor[TestUser]()

	columns := schema.Columns()
	assert.NotEmpty(t, columns)

	var idCol, nameCol, passwordCol *tabular.Column

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

	require.NotNil(t, idCol)
	assert.Equal(t, "ID", idCol.Name)
	assert.Equal(t, 15.0, idCol.Width)

	require.NotNil(t, nameCol)
	assert.Equal(t, "姓名", nameCol.Name)
	assert.Equal(t, 20.0, nameCol.Width)

	assert.Nil(t, passwordCol)
}

func TestImporterValidationErrors(t *testing.T) {
	invalidUsers := []TestUser{
		{
			ID:     "1",
			Name:   "张三",
			Email:  "invalid-email",
			Age:    200,
			Salary: 10000,
		},
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_invalid_users_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(invalidUsers, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, imported)
	assert.NotEmpty(t, importErrors)
}

type TestNoTagStruct struct {
	ID   string
	Name string
	Age  int
}

func TestSchemaNoTags(t *testing.T) {
	schema := tabular.NewSchemaFor[TestNoTagStruct]()

	columns := schema.Columns()
	assert.Len(t, columns, 3)

	assert.Equal(t, "ID", columns[0].Name)
	assert.Equal(t, "Name", columns[1].Name)
	assert.Equal(t, "Age", columns[2].Name)
}

func TestExportImportNoTags(t *testing.T) {
	data := []TestNoTagStruct{
		{ID: "1", Name: "Alice", Age: 30},
		{ID: "2", Name: "Bob", Age: 25},
	}

	exporter := NewExporterFor[TestNoTagStruct]()
	tmpFile, err := os.CreateTemp("", "test_no_tags_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(data, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestNoTagStruct]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestNoTagStruct)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2)

	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "Alice", imported[0].Name)
	assert.Equal(t, 30, imported[0].Age)
}

type prefixFormatter struct {
	prefix string
}

func (f *prefixFormatter) Format(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	return f.prefix + " " + fmt.Sprint(value), nil
}

func TestExportCustomFormatter(t *testing.T) {
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

	exporter := NewExporterFor[TestUser]()
	exporter.RegisterFormatter("prefix", &prefixFormatter{prefix: "ID:"})

	tmpFile, err := os.CreateTemp("", "test_custom_formatter_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

func TestExportToBuffer(t *testing.T) {
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

func TestExportEmptyData(t *testing.T) {
	var emptyUsers []TestUser

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_empty_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(emptyUsers, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

func TestExportWithOptions(t *testing.T) {
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
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

type prefixParser struct{}

func (p *prefixParser) Parse(cellValue string, targetType reflect.Type) (any, error) {
	if cellValue == "" {
		return "", nil
	}

	if len(cellValue) > 4 {
		return cellValue[4:], nil
	}

	return cellValue, nil
}

func TestImportCustomParser(t *testing.T) {
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
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importer.RegisterParser("prefix_parser", &prefixParser{})

	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 1)
}

func TestImportFromReader(t *testing.T) {
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

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.Import(buf)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 1)
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "张三", imported[0].Name)
}

func TestImportWithOptions(t *testing.T) {
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
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser](WithImportSheetName("用户数据"))
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 1)
}

func TestImportEmptyRows(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_empty_rows_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	f := excelize.NewFile()
	sheetName := "Sheet1"

	_ = f.SetCellValue(sheetName, "A1", "ID")
	_ = f.SetCellValue(sheetName, "B1", "姓名")
	_ = f.SetCellValue(sheetName, "C1", "邮箱")

	_ = f.SetCellValue(sheetName, "A2", "1")
	_ = f.SetCellValue(sheetName, "B2", "张三")
	_ = f.SetCellValue(sheetName, "C2", "zhang@example.com")

	_ = f.SetCellValue(sheetName, "A4", "2")
	_ = f.SetCellValue(sheetName, "B4", "李四")
	_ = f.SetCellValue(sheetName, "C4", "li@example.com")

	err = f.SaveAs(filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2)
}

func TestImportMissingColumns(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_missing_columns_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	f := excelize.NewFile()
	sheetName := "Sheet1"

	_ = f.SetCellValue(sheetName, "A1", "ID")
	_ = f.SetCellValue(sheetName, "B1", "姓名")
	_ = f.SetCellValue(sheetName, "C1", "邮箱")
	_ = f.SetCellValue(sheetName, "D1", "年龄")

	_ = f.SetCellValue(sheetName, "A2", "1")
	_ = f.SetCellValue(sheetName, "B2", "张三")
	_ = f.SetCellValue(sheetName, "C2", "zhang@example.com")
	_ = f.SetCellValue(sheetName, "D2", "30")

	err = f.SaveAs(filename)
	require.NoError(t, err)

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
	assert.Equal(t, 0.0, imported[0].Salary)
	assert.Equal(t, 0, imported[0].Status)
	assert.False(t, imported[0].Remark.Valid)
}

func TestImportInvalidData(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_invalid_data_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	f := excelize.NewFile()
	sheetName := "Sheet1"

	_ = f.SetCellValue(sheetName, "A1", "ID")
	_ = f.SetCellValue(sheetName, "B1", "姓名")
	_ = f.SetCellValue(sheetName, "C1", "邮箱")
	_ = f.SetCellValue(sheetName, "D1", "年龄")

	_ = f.SetCellValue(sheetName, "A2", "1")
	_ = f.SetCellValue(sheetName, "B2", "张三")
	_ = f.SetCellValue(sheetName, "C2", "invalid-email")
	_ = f.SetCellValue(sheetName, "D2", "not-a-number")

	err = f.SaveAs(filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, imported)
	assert.NotEmpty(t, importErrors)
}

func TestImportLargeFile(t *testing.T) {
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
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, count)

	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "用户1", imported[0].Name)
	assert.Equal(t, fmt.Sprintf("%d", count), imported[count-1].ID)
}

func TestExportNullValues(t *testing.T) {
	users := []TestUser{
		{
			ID:        "1",
			Name:      "张三",
			Email:     "zhang@example.com",
			Age:       30,
			Salary:    10000.50,
			CreatedAt: time.Now(),
			Status:    1,
			Remark:    null.String{},
		},
		{
			ID:        "2",
			Name:      "李四",
			Email:     "li@example.com",
			Age:       25,
			Salary:    8000.00,
			CreatedAt: time.Now(),
			Status:    2,
			Remark:    null.StringFrom("有备注"),
		},
	}

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_null_values_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(users, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, 2)

	assert.False(t, imported[0].Remark.Valid)
	assert.True(t, imported[1].Remark.Valid)
	assert.Equal(t, "有备注", imported[1].Remark.ValueOrZero())
}

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

	exporter := NewExporterFor[TestUser]()
	tmpFile, err := os.CreateTemp("", "test_roundtrip_*.xlsx")
	require.NoError(t, err)

	filename := tmpFile.Name()
	_ = tmpFile.Close()

	defer os.Remove(filename)

	err = exporter.ExportToFile(original, filename)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	importedAny, importErrors, err := importer.ImportFromFile(filename)
	imported := importedAny.([]TestUser)

	require.NoError(t, err)
	assert.Empty(t, importErrors)
	assert.Len(t, imported, len(original))

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
