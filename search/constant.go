package search

type Operator string

const (
	// Comparison operators
	Equals             Operator = "eq"
	NotEquals          Operator = "neq"
	GreaterThan        Operator = "gt"
	GreaterThanOrEqual Operator = "gte"
	LessThan           Operator = "lt"
	LessThanOrEqual    Operator = "lte"

	// Range operators
	Between    Operator = "between"
	NotBetween Operator = "notBetween"

	// Collection operators
	In    Operator = "in"
	NotIn Operator = "notIn"

	// Null check operators
	IsNull    Operator = "isNull"
	IsNotNull Operator = "isNotNull"

	// String matching operators (case sensitive)
	Contains      Operator = "contains"
	NotContains   Operator = "notContains"
	StartsWith    Operator = "startsWith"
	NotStartsWith Operator = "notStartsWith"
	EndsWith      Operator = "endsWith"
	NotEndsWith   Operator = "notEndsWith"

	// String matching operators (case insensitive)
	ContainsIgnoreCase      Operator = "iContains"
	NotContainsIgnoreCase   Operator = "iNotContains"
	StartsWithIgnoreCase    Operator = "iStartsWith"
	NotStartsWithIgnoreCase Operator = "iNotStartsWith"
	EndsWithIgnoreCase      Operator = "iEndsWith"
	NotEndsWithIgnoreCase   Operator = "iNotEndsWith"

	// Struct tag configuration
	TagSearch = "search"

	// Tag attributes for field configuration
	AttrDive     = "dive"
	AttrAlias    = "alias"
	AttrColumn   = "column"
	AttrOperator = "operator"
	AttrParams   = "params"

	// Parameters for value processing
	ParamDelimiter = "delimiter"
	ParamType      = "type"

	// Special tag value
	IgnoreField = "-" // Ignore this field
)
