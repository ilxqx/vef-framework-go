package orm

import "github.com/ilxqx/vef-framework-go/constants"

// FuzzyKind represents the wildcard placement for LIKE patterns.
// 0: startsWith (value%), 1: endsWith (%value), 2: contains (%value%)
type FuzzyKind uint8

const (
	// FuzzyStarts builds a pattern for starts-with (value%).
	FuzzyStarts FuzzyKind = 0
	// FuzzyEnds builds a pattern for ends-with (%value).
	FuzzyEnds FuzzyKind = 1
	// FuzzyContains builds a pattern for contains (%value%).
	FuzzyContains FuzzyKind = 2
)

// NullsMode controls how NULLs are treated in window functions.
type NullsMode int

const (
	// NullsDefault leaves NULL handling to the database default behavior.
	NullsDefault NullsMode = iota
	// NullsRespect uses RESPECT NULLS.
	NullsRespect
	// NullsIgnore uses IGNORE NULLS.
	NullsIgnore
)

// String returns the SQL snippet for the given NullsMode.
func (n NullsMode) String() string {
	switch n {
	case NullsRespect:
		return "RESPECT NULLS"
	case NullsIgnore:
		return "IGNORE NULLS"
	default:
		return constants.Empty
	}
}

// FromDirection specifies the direction for window frame FROM clause.
type FromDirection int

const (
	// FromDefault leaves direction unspecified.
	FromDefault FromDirection = iota
	// FromFirst emits FROM FIRST.
	FromFirst
	// FromLast emits FROM LAST.
	FromLast
)

// String returns the SQL snippet for the given FromDirection.
func (f FromDirection) String() string {
	switch f {
	case FromFirst:
		return "FROM FIRST"
	case FromLast:
		return "FROM LAST"
	default:
		return constants.Empty
	}
}

// FrameType specifies the window frame unit (ROWS, RANGE, GROUPS).
type FrameType int

const (
	// FrameDefault uses database default frame type.
	FrameDefault FrameType = iota
	// FrameRows emits ROWS.
	FrameRows
	// FrameRange emits RANGE.
	FrameRange
	// FrameGroups emits GROUPS.
	FrameGroups
)

// String returns the SQL snippet for the given FrameType.
func (f FrameType) String() string {
	switch f {
	case FrameRows:
		return "ROWS"
	case FrameRange:
		return "RANGE"
	case FrameGroups:
		return "GROUPS"
	default:
		return constants.Empty
	}
}

// FrameBoundKind specifies the bound type in a window frame.
type FrameBoundKind int

const (
	// FrameBoundNone indicates no bound.
	FrameBoundNone FrameBoundKind = iota
	// FrameBoundUnboundedPreceding emits UNBOUNDED PRECEDING.
	FrameBoundUnboundedPreceding
	// FrameBoundUnboundedFollowing emits UNBOUNDED FOLLOWING.
	FrameBoundUnboundedFollowing
	// FrameBoundCurrentRow emits CURRENT ROW.
	FrameBoundCurrentRow
	// FrameBoundPreceding emits PRECEDING.
	FrameBoundPreceding
	// FrameBoundFollowing emits FOLLOWING.
	FrameBoundFollowing
)

// String returns the SQL snippet for the given FrameBoundKind.
func (f FrameBoundKind) String() string {
	switch f {
	case FrameBoundUnboundedPreceding:
		return "UNBOUNDED PRECEDING"
	case FrameBoundUnboundedFollowing:
		return "UNBOUNDED FOLLOWING"
	case FrameBoundCurrentRow:
		return "CURRENT ROW"
	case FrameBoundPreceding:
		return "PRECEDING"
	case FrameBoundFollowing:
		return "FOLLOWING"
	default:
		return constants.Empty
	}
}

// StatisticalMode selects the statistical variant for aggregates (POP vs SAMP).
type StatisticalMode int

const (
	// StatisticalDefault leaves mode unspecified.
	StatisticalDefault StatisticalMode = iota
	// StatisticalPopulation emits POP (population).
	StatisticalPopulation
	// StatisticalSample emits SAMP (sample).
	StatisticalSample
)

// String returns the SQL snippet for the given StatisticalMode.
func (s StatisticalMode) String() string {
	switch s {
	case StatisticalPopulation:
		return "POP"
	case StatisticalSample:
		return "SAMP"
	default:
		return constants.Empty
	}
}

// ConflictAction represents the action strategy for INSERT ... ON CONFLICT.
type ConflictAction int

const (
	// ConflictDoNothing emits DO NOTHING.
	ConflictDoNothing ConflictAction = iota
	// ConflictDoUpdate emits DO UPDATE.
	ConflictDoUpdate
)

// String returns the SQL snippet for the given ConflictAction.
func (c ConflictAction) String() string {
	switch c {
	case ConflictDoNothing:
		return "DO NOTHING"
	case ConflictDoUpdate:
		return "DO UPDATE"
	default:
		return constants.Empty
	}
}
