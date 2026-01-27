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
	ID       string      `tabular:"用户ID"                 validate:"required"`
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
	users := []TestUser{
		{
			ID:       "1",
			Name:     "张三",
			Email:    "zhangsan@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Date(1994, 1, 15, 0, 0, 0, 0, time.UTC),
			Active:   true,
			Remark:   null.StringFrom("测试用户1"),
		},
		{
			ID:       "2",
			Name:     "李四",
			Email:    "lisi@example.com",
			Age:      25,
			Salary:   8000.75,
			Birthday: time.Date(1999, 5, 20, 0, 0, 0, 0, time.UTC),
			Active:   false,
			Remark:   null.String{},
		},
	}

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(users)
	require.NoError(t, err)
	require.NotNil(t, buf)

	csvContent := buf.String()
	t.Logf("Exported CSV:\n%s", csvContent)

	assert.Contains(t, csvContent, "用户ID")
	assert.Contains(t, csvContent, "姓名")
	assert.Contains(t, csvContent, "邮箱")

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	importedUsers, ok := result.([]TestUser)
	require.True(t, ok)
	require.Len(t, importedUsers, 2)

	assert.Equal(t, "1", importedUsers[0].ID)
	assert.Equal(t, "张三", importedUsers[0].Name)
	assert.Equal(t, "zhangsan@example.com", importedUsers[0].Email)
	assert.Equal(t, 30, importedUsers[0].Age)
	assert.InDelta(t, 10000.50, importedUsers[0].Salary, 0.01)
	assert.Equal(t, "1994-01-15", importedUsers[0].Birthday.Format("2006-01-02"))
	assert.True(t, importedUsers[0].Active)
	assert.True(t, importedUsers[0].Remark.Valid)
	assert.Equal(t, "测试用户1", importedUsers[0].Remark.ValueOrZero())

	assert.Equal(t, "2", importedUsers[1].ID)
	assert.Equal(t, "李四", importedUsers[1].Name)
	assert.Equal(t, "lisi@example.com", importedUsers[1].Email)
	assert.Equal(t, 25, importedUsers[1].Age)
	assert.InDelta(t, 8000.75, importedUsers[1].Salary, 0.01)
	assert.Equal(t, "1999-05-20", importedUsers[1].Birthday.Format("2006-01-02"))
	assert.False(t, importedUsers[1].Active)
	assert.False(t, importedUsers[1].Remark.Valid)
}

func TestCSVImportWithCustomDelimiter(t *testing.T) {
	csvContent := `用户ID;姓名;邮箱
1;张三;zhangsan@example.com
2;李四;lisi@example.com`

	type SimpleUser struct {
		ID    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	importer := NewImporterFor[SimpleUser](WithImportDelimiter(';'))
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	users, ok := result.([]SimpleUser)
	require.True(t, ok)
	require.Len(t, users, 2)

	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "张三", users[0].Name)
}

func TestCSVImportWithoutHeader(t *testing.T) {
	csvContent := `1,张三,zhangsan@example.com
2,李四,lisi@example.com`

	type SimpleUser struct {
		ID    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	importer := NewImporterFor[SimpleUser](WithoutHeader())
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	users, ok := result.([]SimpleUser)
	require.True(t, ok)
	require.Len(t, users, 2)

	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "张三", users[0].Name)
}

func TestCSVExportWithoutHeader(t *testing.T) {
	type SimpleUser struct {
		ID    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	users := []SimpleUser{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com"},
	}

	exporter := NewExporterFor[SimpleUser](WithoutWriteHeader())
	buf, err := exporter.Export(users)
	require.NoError(t, err)

	csvContent := buf.String()
	assert.NotContains(t, csvContent, "用户ID")
	assert.Contains(t, csvContent, "1,张三,zhangsan@example.com")
}

func TestCSVImportWithSkipRows(t *testing.T) {
	csvContent := `用户数据表,,,
用户ID,姓名,邮箱
1,张三,zhangsan@example.com`

	type SimpleUser struct {
		ID    int    `tabular:"用户ID"`
		Name  string `tabular:"姓名"`
		Email string `tabular:"邮箱"`
	}

	importer := NewImporterFor[SimpleUser](WithSkipRows(1))
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	users, ok := result.([]SimpleUser)
	require.True(t, ok)
	require.Len(t, users, 1)

	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "张三", users[0].Name)
}

func TestSchemaParseTags(t *testing.T) {
	schema := tabular.NewSchemaFor[TestUser]()

	columns := schema.Columns()
	assert.NotEmpty(t, columns)

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

	require.NotNil(t, idCol)
	assert.Equal(t, "用户ID", idCol.Name)

	require.NotNil(t, nameCol)
	assert.Equal(t, "姓名", nameCol.Name)

	assert.Nil(t, passwordCol)
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
	buf, err := exporter.Export(data)
	require.NoError(t, err)

	importer := NewImporterFor[TestNoTagStruct]()
	result, importErrors, err := importer.Import(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestNoTagStruct)
	require.True(t, ok)
	assert.Len(t, imported, 2)

	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "Alice", imported[0].Name)
	assert.Equal(t, 30, imported[0].Age)
}

func TestImporterValidationErrors(t *testing.T) {
	csvContent := `用户ID,姓名,邮箱,年龄
1,张三,invalid-email,200`

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Empty(t, imported)
	assert.NotEmpty(t, importErrors)
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
			ID:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Active:   true,
		},
	}

	exporter := NewExporterFor[TestUser]()
	exporter.RegisterFormatter("prefix", &prefixFormatter{prefix: "ID:"})

	buf, err := exporter.Export(users)
	require.NoError(t, err)
	assert.NotNil(t, buf)
}

type prefixParser struct{}

func (*prefixParser) Parse(cellValue string, _ reflect.Type) (any, error) {
	if cellValue == "" {
		return "", nil
	}

	if len(cellValue) > 4 {
		return cellValue[4:], nil
	}

	return cellValue, nil
}

func TestImportCustomParser(t *testing.T) {
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

func TestExportEmptyData(t *testing.T) {
	var emptyUsers []TestUser

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(emptyUsers)
	require.NoError(t, err)

	csvContent := buf.String()
	assert.Contains(t, csvContent, "用户ID")
	assert.Contains(t, csvContent, "姓名")
}

func TestExportToFile(t *testing.T) {
	users := []TestUser{
		{
			ID:       "1",
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

	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

func TestImportFromFile(t *testing.T) {
	users := []TestUser{
		{
			ID:       "1",
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

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.ImportFromFile(filename)
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 1)
	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "张三", imported[0].Name)
}

func TestImportEmptyRows(t *testing.T) {
	csvContent := `用户ID,姓名,邮箱
1,张三,zhang@example.com

2,李四,li@example.com`

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 2)
}

func TestImportMissingColumns(t *testing.T) {
	csvContent := `用户ID,姓名,邮箱,年龄
1,张三,zhang@example.com,30`

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(csvContent))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 1)

	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "张三", imported[0].Name)
	assert.Equal(t, 0.0, imported[0].Salary)
	assert.False(t, imported[0].Remark.Valid)
}

func TestImportInvalidData(t *testing.T) {
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

func TestImportLargeFile(t *testing.T) {
	count := 1000
	users := make([]TestUser, count)

	for i := range count {
		users[i] = TestUser{
			ID:       fmt.Sprintf("%d", i+1),
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

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.ImportFromFile(filename)
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, count)

	assert.Equal(t, "1", imported[0].ID)
	assert.Equal(t, "用户1", imported[0].Name)
	assert.Equal(t, fmt.Sprintf("%d", count), imported[count-1].ID)
}

func TestExportNullValues(t *testing.T) {
	users := []TestUser{
		{
			ID:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Now(),
			Active:   true,
			Remark:   null.String{},
		},
		{
			ID:       "2",
			Name:     "李四",
			Email:    "li@example.com",
			Age:      25,
			Salary:   8000.00,
			Birthday: time.Now(),
			Active:   false,
			Remark:   null.StringFrom("有备注"),
		},
	}

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(users)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, 2)

	assert.False(t, imported[0].Remark.Valid)
	assert.True(t, imported[1].Remark.Valid)
	assert.Equal(t, "有备注", imported[1].Remark.ValueOrZero())
}

func TestRoundTrip(t *testing.T) {
	original := []TestUser{
		{
			ID:       "1",
			Name:     "张三",
			Email:    "zhang@example.com",
			Age:      30,
			Salary:   10000.50,
			Birthday: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Active:   true,
			Remark:   null.StringFrom("测试用户1"),
		},
		{
			ID:       "2",
			Name:     "李四",
			Email:    "li@example.com",
			Age:      25,
			Salary:   8000.75,
			Birthday: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			Active:   false,
			Remark:   null.String{},
		},
	}

	exporter := NewExporterFor[TestUser]()
	buf, err := exporter.Export(original)
	require.NoError(t, err)

	importer := NewImporterFor[TestUser]()
	result, importErrors, err := importer.Import(strings.NewReader(buf.String()))
	require.NoError(t, err)
	assert.Empty(t, importErrors)

	imported, ok := result.([]TestUser)
	require.True(t, ok)
	assert.Len(t, imported, len(original))

	for i := range original {
		assert.Equal(t, original[i].ID, imported[i].ID)
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
