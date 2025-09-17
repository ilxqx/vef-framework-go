// Package mo provides model objects and utilities for common data types.
// It includes optional values, JSON handling, pagination, ranges, and time utilities.
package mo

const (
	JSONNull       = "null" // JSONNull is the string representation of the null value in JSON
	JSONQuote byte = '"'    // JSONQuote is the quote character used in JSON strings

	DefaultPageNumber int = 1    // DefaultPageNumber is the default page number for pagination (starts from 1)
	DefaultPageSize   int = 15   // DefaultPageSize is the default page size for pagination
	MaxPageSize       int = 1000 // MaxPageSize is the maximum allowed page size to prevent excessive data loading
)

var (
	JSONNullBytes = []byte(JSONNull) // JSONNullBytes is the byte representation of JSON null for efficient comparisons
)
