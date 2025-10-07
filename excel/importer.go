package excel

import (
	"fmt"
	"io"
	"reflect"

	"github.com/xuri/excelize/v2"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/tabular"
	"github.com/ilxqx/vef-framework-go/validator"
)

var logger = log.Named("excel")

// importer is the default implementation of Importer.
type importer struct {
	schema  *tabular.Schema
	parsers map[string]tabular.ValueParser
	options importOptions
	typ     reflect.Type
}

// newImporter creates a new importer with the specified type.
func newImporter(typ reflect.Type, opts ...ImportOption) *importer {
	options := importOptions{
		sheetIndex: 0,
	}
	for _, opt := range opts {
		opt(&options)
	}

	return &importer{
		schema:  tabular.NewSchema(typ),
		parsers: make(map[string]tabular.ValueParser),
		options: options,
		typ:     typ,
	}
}

// RegisterParser registers a custom parser with the given name.
func (i *importer) RegisterParser(name string, parser tabular.ValueParser) {
	i.parsers[name] = parser
}

// ImportFromFile imports data from an Excel file.
func (i *importer) ImportFromFile(filename string) (any, []tabular.ImportError, error) {
	// Open Excel file
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("open Excel file %s: %w", filename, err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Errorf("Failed to close Excel file %s: %v", filename, closeErr)
		}
	}()

	return i.doImport(f)
}

// Import imports data from an io.Reader.
func (i *importer) Import(reader io.Reader) (any, []tabular.ImportError, error) {
	// Open reader
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("open Excel from reader: %w", err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Errorf("Failed to close Excel file from reader: %v", closeErr)
		}
	}()

	return i.doImport(f)
}

// doImport imports data from an excelize.File.
func (i *importer) doImport(f *excelize.File) (any, []tabular.ImportError, error) {
	// Get sheet name
	sheetName := i.options.sheetName
	if sheetName == constants.Empty {
		// Use sheet at specified index
		sheets := f.GetSheetList()
		if i.options.sheetIndex >= len(sheets) {
			return nil, nil, fmt.Errorf("%w: %d (total sheets: %d)", ErrSheetIndexOutOfRange, i.options.sheetIndex, len(sheets))
		}

		sheetName = sheets[i.options.sheetIndex]
	}

	// Read all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("get rows: %w", err)
	}

	// Check if file has data
	if len(rows) <= i.options.skipRows+1 {
		return nil, nil, fmt.Errorf("%w (total rows: %d, skip rows: %d)", ErrNoDataRowsFound, len(rows), i.options.skipRows)
	}

	// Skip rows and get header
	headerRowIdx := i.options.skipRows
	headerRow := rows[headerRowIdx]

	// Build column mapping (Excel column index -> Schema column index)
	columnMapping, err := i.buildColumnMapping(headerRow)
	if err != nil {
		return nil, nil, fmt.Errorf("build column mapping: %w", err)
	}

	// Parse data rows
	dataRows := rows[headerRowIdx+1:]
	resultSlice := reflect.MakeSlice(reflect.SliceOf(i.typ), 0, len(dataRows))

	var importErrors []tabular.ImportError

	for rowIdx, row := range dataRows {
		excelRow := headerRowIdx + rowIdx + 2 // Excel row number (1-based, includes header)

		// Skip empty rows
		if i.isEmptyRow(row) {
			continue
		}

		// Parse row to struct
		item, rowErrors := i.parseRow(row, columnMapping, excelRow)
		if len(rowErrors) > 0 {
			importErrors = append(importErrors, rowErrors...)

			continue
		}

		// Validate item
		if err := validator.Validate(item); err != nil {
			importErrors = append(importErrors, tabular.ImportError{
				Row: excelRow,
				Err: fmt.Errorf("validation failed: %w", err),
			})

			continue
		}

		resultSlice = reflect.Append(resultSlice, reflect.ValueOf(item))
	}

	return resultSlice.Interface(), importErrors, nil
}

// buildColumnMapping builds a mapping from Excel column index to Schema column index.
func (i *importer) buildColumnMapping(headerRow []string) (map[int]int, error) {
	columns := i.schema.Columns()
	mapping := make(map[int]int)

	// Create a map from column name to schema column index
	nameToSchemaIdx := make(map[string]int)
	for schemaIdx, col := range columns {
		nameToSchemaIdx[col.Name] = schemaIdx
	}

	// Map Excel columns to schema columns and detect duplicates
	seen := make(map[string]bool)
	for excelIdx, headerName := range headerRow {
		if headerName == constants.Empty {
			continue
		}

		if seen[headerName] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateColumnName, headerName)
		}

		seen[headerName] = true

		if schemaIdx, ok := nameToSchemaIdx[headerName]; ok {
			mapping[excelIdx] = schemaIdx
		}
	}

	return mapping, nil
}

// parseRow parses an Excel row to a struct instance.
func (i *importer) parseRow(row []string, columnMapping map[int]int, excelRow int) (any, []tabular.ImportError) {
	result := reflect.New(i.typ).Elem()

	var errors []tabular.ImportError

	columns := i.schema.Columns()

	// Parse each cell
	for excelIdx, schemaIdx := range columnMapping {
		col := columns[schemaIdx]

		// Get cell value
		var cellValue string
		if excelIdx < len(row) {
			cellValue = row[excelIdx]
		}

		// Use default value if cell is empty
		if cellValue == constants.Empty && col.Default != constants.Empty {
			cellValue = col.Default
		}

		// Get field
		field := result.FieldByIndex(col.Index)
		if !field.CanSet() {
			errors = append(errors, tabular.ImportError{
				Row:    excelRow,
				Column: col.Name,
				Field:  field.Type().Name(),
				Err:    ErrFieldNotSettable,
			})

			continue
		}

		// Parse value
		value, err := i.parseValue(cellValue, field.Type(), col)
		if err != nil {
			errors = append(errors, tabular.ImportError{
				Row:    excelRow,
				Column: col.Name,
				Field:  field.Type().Name(),
				Err:    fmt.Errorf("parse value: %w", err),
			})

			continue
		}

		// Set field value
		field.Set(reflect.ValueOf(value))
	}

	return result.Interface(), errors
}

// parseValue parses a cell value to the target type.
func (i *importer) parseValue(cellValue string, targetType reflect.Type, col *tabular.Column) (any, error) {
	// Use custom parser if specified
	if col.Parser != constants.Empty {
		if parser, ok := i.parsers[col.Parser]; ok {
			return parser.Parse(cellValue, targetType)
		}

		logger.Warnf("Parser %s not found, using default parser", col.Parser)
	}

	// Use default parser
	parser := tabular.NewDefaultParser(col.Format)

	return parser.Parse(cellValue, targetType)
}

// isEmptyRow checks if a row is empty (all cells are empty).
func (i *importer) isEmptyRow(row []string) bool {
	for _, cell := range row {
		if cell != constants.Empty {
			return false
		}
	}

	return true
}
