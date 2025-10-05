package sort

import "github.com/ilxqx/vef-framework-go/constants"

// OrderDirection represents the direction of ordering in SQL queries.
// It defines whether results should be sorted in ascending or descending order.
type OrderDirection int

const (
	// OrderAsc represents ascending order (A-Z, 0-9, oldest to newest).
	OrderAsc OrderDirection = iota
	// OrderDesc represents descending order (Z-A, 9-0, newest to oldest).
	OrderDesc
)

// String returns the string representation of OrderDirection.
func (od OrderDirection) String() string {
	switch od {
	case OrderDesc:
		return "DESC"
	default:
		return "ASC"
	}
}

// NullsOrder represents how to handle NULL values in ordering.
// Different databases have different default behaviors for NULL values,
// so explicit specification ensures consistent behavior across databases.
type NullsOrder int

const (
	// NullsDefault uses database default behavior for NULLs.
	// WARNING: This may cause inconsistent behavior across different databases.
	// PostgreSQL: ASC=NULLS LAST, DESC=NULLS FIRST
	// MySQL/SQLite: ASC=NULLS FIRST, DESC=NULLS LAST
	// Oracle: ASC=NULLS LAST, DESC=NULLS FIRST.
	NullsDefault NullsOrder = iota
	// NullsFirst places NULL values before non-NULL values in the result set.
	NullsFirst
	// NullsLast places NULL values after non-NULL values in the result set.
	NullsLast
)

// String returns the string representation of NullsOrder.
func (no NullsOrder) String() string {
	switch no {
	case NullsFirst:
		return "NULLS FIRST"
	case NullsLast:
		return "NULLS LAST"
	default:
		return constants.Empty
	}
}

// OrderSpec represents a ordering specification for a single column.
// It encapsulates all the information needed to generate the ORDER BY clause for one item.
type OrderSpec struct {
	// Column is the column name to order by.
	// Should be empty if Expression is used instead.
	Column string
	// Direction specifies the ordering direction (ASC or DESC).
	Direction OrderDirection
	// NullsOrder specifies how to handle NULL values in the ordering.
	// Use NullsDefault to rely on database defaults (not recommended for cross-database compatibility).
	NullsOrder NullsOrder
}

// IsValid checks if the OrderSpec is valid.
// An OrderSpec is valid if it has either a Column or an Expression, but not both.
func (os OrderSpec) IsValid() bool {
	return os.Column != constants.Empty
}
