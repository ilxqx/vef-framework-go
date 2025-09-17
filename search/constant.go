package search

type Operator string

const (
	Equals                  Operator = "eq"             // Equals is the equal operator
	NotEquals               Operator = "neq"            // NotEquals is the not equal operator
	GreaterThan             Operator = "gt"             // GreaterThan is the greater than operator
	GreaterThanOrEqual      Operator = "gte"            // GreaterThanOrEqual is the greater than or equal operator
	LessThan                Operator = "lt"             // LessThan is the less than operator
	LessThanOrEqual         Operator = "lte"            // LessThanOrEqual is the less than or equal operator
	Between                 Operator = "between"        // Between is the between operator
	NotBetween              Operator = "notBetween"     // NotBetween is the not between operator
	In                      Operator = "in"             // In is the in operator
	NotIn                   Operator = "notIn"          // NotIn is the not in operator
	IsNull                  Operator = "isNull"         // IsNull is the is null operator
	IsNotNull               Operator = "isNotNull"      // IsNotNull is the is not null operator
	Contains                Operator = "contains"       // Contains is the contains operator
	NotContains             Operator = "notContains"    // NotContains is the not contains operator
	StartsWith              Operator = "startsWith"     // StartsWith is the starts with operator
	NotStartsWith           Operator = "notStartsWith"  // NotStartsWith is the not starts with operator
	EndsWith                Operator = "endsWith"       // EndsWith is the ends with operator
	NotEndsWith             Operator = "notEndsWith"    // NotEndsWith is the not ends with operator
	ContainsIgnoreCase      Operator = "iContains"      // ContainsIgnoreCase is the contains ignore case operator
	NotContainsIgnoreCase   Operator = "iNotContains"   // NotContainsIgnoreCase is the not contains ignore case operator
	StartsWithIgnoreCase    Operator = "iStartsWith"    // StartsWithIgnoreCase is the starts with ignore case operator
	NotStartsWithIgnoreCase Operator = "iNotStartsWith" // NotStartsWithIgnoreCase is the not starts with ignore case operator
	EndsWithIgnoreCase      Operator = "iEndsWith"      // EndsWithIgnoreCase is the ends with ignore case operator
	NotEndsWithIgnoreCase   Operator = "iNotEndsWith"   // NotEndsWithIgnoreCase is the not ends with ignore case operator

	TagSearch = "search" // TagSearch is the tag name for search

	AttrDive     = "dive"     // AttrDive is the dive attribute
	AttrAlias    = "alias"    // AttrAlias is the alias attribute
	AttrColumn   = "column"   // AttrColumn is the column attribute
	AttrOperator = "operator" // AttrOperator is the operator attribute
	AttrArgs     = "args"     // AttrArgs is the args attribute
	AttrDefault  = "default"  // AttrDefault is the default attribute

	ArgDelimiter = "delimiter" // ArgDelimiter is the delimiter for the range
	ArgType      = "type"      // ArgType is the type for the value
	ArgDefault   = "default"   // ArgDefault is the default argument
)
