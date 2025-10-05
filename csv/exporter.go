package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"reflect"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/tabular"
)

// exporter is the csv implementation of Exporter.
type exporter struct {
	schema     *tabular.Schema
	formatters map[string]tabular.Formatter
	options    exportOptions
	typ        reflect.Type
}

// newExporter creates a new exporter with the specified type.
func newExporter(typ reflect.Type, opts ...ExportOption) *exporter {
	options := exportOptions{
		delimiter:   constants.ByteComma,
		writeHeader: true,
		useCRLF:     false,
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

// ExportToFile exports data to a CSV file.
func (e *exporter) ExportToFile(data any, filename string) error {
	// Create CSV file
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create CSV file %s: %w", filename, err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Errorf("Failed to close CSV file %s: %v", filename, closeErr)
		}
	}()

	return e.writeToWriter(csv.NewWriter(f), data)
}

// Export exports data to a bytes.Buffer.
func (e *exporter) Export(data any) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if err := e.writeToWriter(csv.NewWriter(buf), data); err != nil {
		return nil, err
	}

	return buf, nil
}

// writeToWriter configures a CSV writer, writes data, and flushes.
func (e *exporter) writeToWriter(csvWriter *csv.Writer, data any) error {
	// Configure CSV writer
	csvWriter.Comma = e.options.delimiter
	csvWriter.UseCRLF = e.options.useCRLF

	// Write data
	if err := e.doExport(csvWriter, data); err != nil {
		return err
	}

	// Flush and check for errors
	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("flush CSV writer: %w", err)
	}

	return nil
}

// doExport exports data to a CSV writer.
func (e *exporter) doExport(csvWriter *csv.Writer, data any) error {
	// Write header row
	if e.options.writeHeader {
		if err := e.writeHeader(csvWriter); err != nil {
			return fmt.Errorf("write header: %w", err)
		}
	}

	// Write data rows
	if err := e.writeData(csvWriter, data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

// writeHeader writes the header row to the CSV writer.
func (e *exporter) writeHeader(csvWriter *csv.Writer) error {
	columns := e.schema.Columns()
	headerRow := make([]string, len(columns))

	for colIdx, col := range columns {
		headerRow[colIdx] = col.Name
	}

	if err := csvWriter.Write(headerRow); err != nil {
		return fmt.Errorf("write header row: %w", err)
	}

	return nil
}

// writeData writes data rows to the CSV writer.
func (e *exporter) writeData(csvWriter *csv.Writer, data any) error {
	columns := e.schema.Columns()

	// Convert data to slice using reflection
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Slice {
		return fmt.Errorf("%w, got %s", ErrDataMustBeSlice, dataValue.Kind())
	}

	for rowIdx := 0; rowIdx < dataValue.Len(); rowIdx++ {
		item := dataValue.Index(rowIdx)
		row := make([]string, len(columns))

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

			row[colIdx] = cellValue
		}

		// Write row
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("write row %d: %w", rowIdx, err)
		}
	}

	return nil
}

// formatValue formats a field value to CSV cell string.
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
