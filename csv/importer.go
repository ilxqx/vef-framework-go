package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/tabular"
	"github.com/ilxqx/vef-framework-go/validator"
)

var logger = log.Named("csv")

// importer is the csv implementation of Importer.
type importer struct {
	schema  *tabular.Schema
	parsers map[string]tabular.ValueParser
	options importOptions
	typ     reflect.Type
}

// newImporter creates a new importer with the specified type.
func newImporter(typ reflect.Type, opts ...ImportOption) *importer {
	options := importOptions{
		delimiter: ',',
		hasHeader: true,
		skipRows:  0,
		trimSpace: true,
		comment:   0,
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

// ImportFromFile imports data from a CSV file.
func (i *importer) ImportFromFile(filename string) (any, []tabular.ImportError, error) {
	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("open CSV file %s: %w", filename, err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Errorf("Failed to close CSV file %s: %v", filename, closeErr)
		}
	}()

	return i.Import(f)
}

// Import imports data from an io.Reader.
func (i *importer) Import(reader io.Reader) (any, []tabular.ImportError, error) {
	// Create CSV reader
	csvReader := csv.NewReader(reader)
	csvReader.Comma = i.options.delimiter
	csvReader.TrimLeadingSpace = i.options.trimSpace
	csvReader.Comment = i.options.comment
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all rows
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("read CSV: %w", err)
	}

	// Check if file has data
	minRows := i.options.skipRows
	if i.options.hasHeader {
		minRows++
	}

	if len(rows) <= minRows {
		return nil, nil, fmt.Errorf("%w (total rows: %d, skip rows: %d, has header: %v)",
			ErrNoDataRowsFound, len(rows), i.options.skipRows, i.options.hasHeader)
	}

	// Build column mapping
	var columnMapping map[int]int

	dataStartIdx := i.options.skipRows

	if i.options.hasHeader {
		// Skip rows and get header
		headerRow := rows[i.options.skipRows]

		var err error

		columnMapping, err = i.buildColumnMapping(headerRow)
		if err != nil {
			return nil, nil, fmt.Errorf("build column mapping: %w", err)
		}

		dataStartIdx++
	} else {
		// No header: map columns by index
		columnMapping = i.buildDefaultMapping()
	}

	// Parse data rows
	dataRows := rows[dataStartIdx:]
	resultSlice := reflect.MakeSlice(reflect.SliceOf(i.typ), 0, len(dataRows))

	var importErrors []tabular.ImportError

	for rowIdx, row := range dataRows {
		csvRow := dataStartIdx + rowIdx + 1 // CSV row number (1-based)

		// Skip empty rows
		if i.isEmptyRow(row) {
			continue
		}

		// Parse row to struct
		item, rowErrors := i.parseRow(row, columnMapping, csvRow)
		if len(rowErrors) > 0 {
			importErrors = append(importErrors, rowErrors...)

			continue
		}

		// Validate item
		if err := validator.Validate(item); err != nil {
			importErrors = append(importErrors, tabular.ImportError{
				Row: csvRow,
				Err: fmt.Errorf("validation failed: %w", err),
			})

			continue
		}

		resultSlice = reflect.Append(resultSlice, reflect.ValueOf(item))
	}

	return resultSlice.Interface(), importErrors, nil
}

// buildColumnMapping builds a mapping from CSV column index to Schema column index.
func (i *importer) buildColumnMapping(headerRow []string) (map[int]int, error) {
	columns := i.schema.Columns()
	mapping := make(map[int]int)

	// Create a map from column name to schema column index
	nameToSchemaIdx := make(map[string]int)
	for schemaIdx, col := range columns {
		nameToSchemaIdx[col.Name] = schemaIdx
	}

	// Map CSV columns to schema columns and detect duplicates
	seen := make(map[string]bool)
	for csvIdx, headerName := range headerRow {
		// Trim header name if needed
		if i.options.trimSpace {
			headerName = strings.TrimSpace(headerName)
		}

		if headerName == constants.Empty {
			continue
		}

		if seen[headerName] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateColumnName, headerName)
		}

		seen[headerName] = true

		if schemaIdx, ok := nameToSchemaIdx[headerName]; ok {
			mapping[csvIdx] = schemaIdx
		}
	}

	return mapping, nil
}

// buildDefaultMapping builds a default mapping when there's no header (column index -> column index).
func (i *importer) buildDefaultMapping() map[int]int {
	columns := i.schema.Columns()

	mapping := make(map[int]int)
	for idx := range columns {
		mapping[idx] = idx
	}

	return mapping
}

// parseRow parses a CSV row to a struct instance.
func (i *importer) parseRow(row []string, columnMapping map[int]int, csvRow int) (any, []tabular.ImportError) {
	result := reflect.New(i.typ).Elem()

	var errors []tabular.ImportError

	columns := i.schema.Columns()

	// Parse each cell
	for csvIdx, schemaIdx := range columnMapping {
		col := columns[schemaIdx]

		// Get cell value
		var cellValue string
		if csvIdx < len(row) {
			cellValue = row[csvIdx]
			if i.options.trimSpace {
				cellValue = strings.TrimSpace(cellValue)
			}
		}

		// Use default value if cell is empty
		if cellValue == constants.Empty && col.Default != constants.Empty {
			cellValue = col.Default
		}

		// Get field
		field := result.FieldByIndex(col.Index)
		if !field.CanSet() {
			errors = append(errors, tabular.ImportError{
				Row:    csvRow,
				Column: col.Name,
				Field:  field.Type().Name(),
				Err:    fmt.Errorf("%w", ErrFieldNotSettable),
			})

			continue
		}

		// Parse value
		value, err := i.parseValue(cellValue, field.Type(), col)
		if err != nil {
			errors = append(errors, tabular.ImportError{
				Row:    csvRow,
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
		trimmed := cell
		if i.options.trimSpace {
			trimmed = strings.TrimSpace(cell)
		}

		if trimmed != constants.Empty {
			return false
		}
	}

	return true
}
