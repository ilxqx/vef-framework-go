package constants

const (
	// JsonNull is the string representation of the null value in JSON.
	JsonNull = "null"
	// JsonQuote is the quote character used in JSON strings.
	JsonQuote byte = '"'
)

// JsonNullBytes is the byte representation of JSON null.
var JsonNullBytes = []byte(JsonNull)
