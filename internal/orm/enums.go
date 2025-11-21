package orm

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

// JoinType specifies the type of JOIN operation.
type JoinType int

const (
	JoinDefault JoinType = iota
	JoinInner
	JoinLeft
	JoinRight
	JoinFull // Note: Not supported by SQLite
	JoinCross
)

func (j JoinType) String() string {
	switch j {
	case JoinLeft, JoinDefault:
		return "LEFT JOIN"
	case JoinRight:
		return "RIGHT JOIN"
	case JoinFull:
		return "FULL JOIN"
	case JoinCross:
		return "CROSS JOIN"
	case JoinInner:
		fallthrough
	default:
		return "JOIN"
	}
}

// FuzzyKind represents the wildcard placement for LIKE patterns.
type FuzzyKind uint8

const (
	FuzzyStarts   FuzzyKind = 0 // value%
	FuzzyEnds     FuzzyKind = 1 // %value
	FuzzyContains FuzzyKind = 2 // %value%
)

// BuildPattern constructs a LIKE pattern string based on the FuzzyKind.
// It efficiently builds the pattern using strings.Builder with pre-allocated capacity.
func (k FuzzyKind) BuildPattern(value string) string {
	var sb strings.Builder
	if k == FuzzyStarts {
		sb.Grow(len(value) + 1)
	} else {
		sb.Grow(len(value) + int(k))
	}

	switch k {
	case FuzzyEnds, FuzzyContains:
		_ = sb.WriteByte(constants.BytePercent)
	}

	_, _ = sb.WriteString(value)

	switch k {
	case FuzzyStarts, FuzzyContains:
		_ = sb.WriteByte(constants.BytePercent)
	}

	return sb.String()
}

// NullsMode controls how NULLs are treated in window functions.
type NullsMode int

const (
	NullsDefault NullsMode = iota
	NullsRespect
	NullsIgnore
)

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
	FromDefault FromDirection = iota
	FromFirst
	FromLast
)

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

// FrameType specifies the window frame unit.
type FrameType int

const (
	FrameDefault FrameType = iota
	FrameRows
	FrameRange
	FrameGroups
)

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
	FrameBoundNone FrameBoundKind = iota
	FrameBoundUnboundedPreceding
	FrameBoundUnboundedFollowing
	FrameBoundCurrentRow
	FrameBoundPreceding
	FrameBoundFollowing
)

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

// StatisticalMode selects the statistical variant for aggregates.
type StatisticalMode int

const (
	StatisticalDefault    StatisticalMode = iota
	StatisticalPopulation                 // POP
	StatisticalSample                     // SAMP
)

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
	ConflictDoNothing ConflictAction = iota
	ConflictDoUpdate
)

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

// DateTimeUnit represents date and time interval units for date arithmetic operations.
type DateTimeUnit int

const (
	UnitYear DateTimeUnit = iota
	UnitMonth
	UnitDay
	UnitHour
	UnitMinute
	UnitSecond
)

func (u DateTimeUnit) String() string {
	switch u {
	case UnitYear:
		return "YEAR"
	case UnitMonth:
		return "MONTH"
	case UnitDay:
		return "DAY"
	case UnitHour:
		return "HOUR"
	case UnitMinute:
		return "MINUTE"
	case UnitSecond:
		return "SECOND"
	default:
		return "DAY"
	}
}

// ForPostgres returns the PostgreSQL interval unit string (YEAR, MONTH, DAY, etc.).
func (u DateTimeUnit) ForPostgres() string {
	return u.String()
}

// ForMySQL returns the MySQL interval unit string (YEAR, MONTH, DAY, etc.).
func (u DateTimeUnit) ForMySQL() string {
	return u.String()
}

// ForSQLite returns the SQLite datetime modifier string (years, months, days, etc.).
func (u DateTimeUnit) ForSQLite() string {
	switch u {
	case UnitYear:
		return "years"
	case UnitMonth:
		return "months"
	case UnitDay:
		return "days"
	case UnitHour:
		return "hours"
	case UnitMinute:
		return "minutes"
	case UnitSecond:
		return "seconds"
	default:
		return "days"
	}
}

// ForDateTrunc returns the lowercase string for DateTrunc precision parameter.
func (u DateTimeUnit) ForDateTrunc() string {
	switch u {
	case UnitYear:
		return "year"
	case UnitMonth:
		return "month"
	case UnitDay:
		return "day"
	case UnitHour:
		return "hour"
	case UnitMinute:
		return "minute"
	case UnitSecond:
		return "second"
	default:
		return "day"
	}
}
