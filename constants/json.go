package constants

const (
	// JSONNull is the string representation of the null value in JSON.
	JSONNull = "null"
	// JSONQuote is the quote character used in JSON strings.
	JSONQuote byte = '"'
)

// JSONNullBytes is the byte representation of JSON null.
var JSONNullBytes = []byte(JSONNull)
