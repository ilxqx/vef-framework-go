package excel

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/xuri/excelize/v2"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/tabular"
)

// exporter is the excel implementation of Exporter.
type exporter struct {
	schema     *tabular.Schema
	formatters map[string]tabular.Formatter
	options    exportOptions
	typ        reflect.Type
}

// newExporter creates a new exporter with the specified type.
func newExporter(typ reflect.Type, opts ...ExportOption) *exporter {
	options := exportOptions{
		sheetName: "Sheet1",
	}
	for _, opt := range opts {
		opt(&options)
	}

	return &exporter{
		schema:     tabular.NewSchema(typ),
		formatters: make(map[string]tabular.Formatter),
		options:    options,
		typ:        typ,
	}
}

// RegisterFormatter registers a custom formatter with the given name.
func (e *exporter) RegisterFormatter(name string, formatter tabular.Formatter) {
	e.formatters[name] = formatter
}

// ExportToFile exports data to an Excel file.
func (e *exporter) ExportToFile(data any, filename string) error {
	// Create Excel file
	f, err := e.doExport(data)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Errorf("Failed to close Excel file: %v", closeErr)
		}
	}()

	// Save file
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("save file %s: %w", filename, err)
	}

	return nil
}

// Export exports data to a bytes.Buffer.
func (e *exporter) Export(data any) (*bytes.Buffer, error) {
	f, err := e.doExport(data)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Errorf("Failed to close Excel file after export: %v", closeErr)
		}
	}()

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write to buffer: %w", err)
	}

	return buf, nil
}

// doExport exports data to an excelize.File.
func (e *exporter) doExport(data any) (*excelize.File, error) {
	// Create Excel file
	f := excelize.NewFile()

	// Create or use existing sheet
	sheetIndex, err := f.GetSheetIndex(e.options.sheetName)
	if err != nil {
		return nil, fmt.Errorf("get sheet index: %w", err)
	}

	if sheetIndex == -1 {
		sheetIndex, err = f.NewSheet(e.options.sheetName)
		if err != nil {
			return nil, fmt.Errorf("create sheet: %w", err)
		}
	}

	// Write header row
	if err := e.writeHeader(f, e.options.sheetName); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}

	// Write data rows
	if err := e.writeData(f, e.options.sheetName, data); err != nil {
		return nil, fmt.Errorf("write data: %w", err)
	}

	// Set active sheet
	f.SetActiveSheet(sheetIndex)

	return f, nil
}

// writeHeader writes the header row to the Excel sheet.
func (e *exporter) writeHeader(f *excelize.File, sheetName string) error {
	columns := e.schema.Columns()

	for colIdx, col := range columns {
		// Column letter (A, B, C, ...)
		colLetter, err := excelize.ColumnNumberToName(colIdx + 1)
		if err != nil {
			return fmt.Errorf("convert column number to name: %w", err)
		}

		// Set header cell value
		cell := fmt.Sprintf("%s1", colLetter)
		if err := f.SetCellValue(sheetName, cell, col.Name); err != nil {
			return fmt.Errorf("set header cell %s: %w", cell, err)
		}

		// Set column width if specified
		if col.Width > 0 {
			if err := f.SetColWidth(sheetName, colLetter, colLetter, col.Width); err != nil {
				return fmt.Errorf("set column width for %s: %w", colLetter, err)
			}
		}
	}

	return nil
}

// writeData writes data rows to the Excel sheet.
func (e *exporter) writeData(f *excelize.File, sheetName string, data any) error {
	columns := e.schema.Columns()

	// Convert data to slice using reflection
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Slice {
		return fmt.Errorf("%w, got %s", ErrDataMustBeSlice, dataValue.Kind())
	}

	for rowIdx := 0; rowIdx < dataValue.Len(); rowIdx++ {
		item := dataValue.Index(rowIdx)
		excelRow := rowIdx + 2 // Excel rows start at 1, header is row 1

		for colIdx, col := range columns {
			// Get field value
			fieldValue := item.FieldByIndex(col.Index)

			// Format value
			cellValue, err := e.formatValue(fieldValue.Interface(), col)
			if err != nil {
				return tabular.ExportError{
					Row:    rowIdx,
					Column: col.Name,
					Field:  fieldValue.Type().Name(),
					Err:    fmt.Errorf("format value: %w", err),
				}
			}

			// Column letter
			colLetter, err := excelize.ColumnNumberToName(colIdx + 1)
			if err != nil {
				return fmt.Errorf("convert column number to name: %w", err)
			}

			// Set cell value
			cell := fmt.Sprintf("%s%d", colLetter, excelRow)
			if err := f.SetCellValue(sheetName, cell, cellValue); err != nil {
				return fmt.Errorf("set cell %s: %w", cell, err)
			}
		}
	}

	return nil
}

// formatValue formats a field value to Excel cell string.
func (e *exporter) formatValue(value any, col *tabular.Column) (string, error) {
	// Use custom formatter if specified
	if col.Formatter != constants.Empty {
		if formatter, ok := e.formatters[col.Formatter]; ok {
			return formatter.Format(value)
		}

		logger.Warnf("Formatter %s not found, using default formatter", col.Formatter)
	}

	// Use default formatter
	formatter := tabular.NewDefaultFormatter(col.Format)

	return formatter.Format(value)
}
