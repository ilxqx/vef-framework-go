package excel

const (
	// TagExcel is the struct tag name for excel field configuration
	TagExcel = "excel"

	// Tag attributes for field configuration
	AttrDive      = "dive"      // Recursively parse embedded struct
	AttrName      = "name"      // Excel column name (header)
	AttrWidth     = "width"     // Column width
	AttrOrder     = "order"     // Column order
	AttrDefault   = "default"   // Default value for import
	AttrFormat    = "format"    // Format template (date/number format)
	AttrFormatter = "formatter" // Custom formatter name for export
	AttrParser    = "parser"    // Custom parser name for import

	// Special tag value
	IgnoreField = "-" // Ignore this field
)
