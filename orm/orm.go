package orm

import "github.com/ilxqx/vef-framework-go/internal/orm"

type (
	DB                         = orm.DB
	SelectQuery                = orm.SelectQuery
	InsertQuery                = orm.InsertQuery
	UpdateQuery                = orm.UpdateQuery
	DeleteQuery                = orm.DeleteQuery
	MergeQuery                 = orm.MergeQuery
	RawQuery                   = orm.RawQuery
	QueryBuilder               = orm.QueryBuilder
	ConditionBuilder           = orm.ConditionBuilder
	Applier[T any]             = orm.Applier[T]
	ApplyFunc[T any]           = orm.ApplyFunc[T]
	RelationSpec               = orm.RelationSpec
	JoinType                   = orm.JoinType
	FuzzyKind                  = orm.FuzzyKind
	NullsMode                  = orm.NullsMode
	FromDirection              = orm.FromDirection
	FrameType                  = orm.FrameType
	FrameBoundKind             = orm.FrameBoundKind
	StatisticalMode            = orm.StatisticalMode
	ConflictAction             = orm.ConflictAction
	DateTimeUnit               = orm.DateTimeUnit
	ColumnInfo                 = orm.ColumnInfo
	Model                      = orm.Model
	IDModel                    = orm.IDModel
	CreatedModel               = orm.CreatedModel
	AuditedModel               = orm.AuditedModel
	PKField                    = orm.PKField
	ExprBuilder                = orm.ExprBuilder
	OrderBuilder               = orm.OrderBuilder
	CaseBuilder                = orm.CaseBuilder
	CaseWhenBuilder            = orm.CaseWhenBuilder
	ConflictBuilder            = orm.ConflictBuilder
	ConflictUpdateBuilder      = orm.ConflictUpdateBuilder
	MergeWhenBuilder           = orm.MergeWhenBuilder
	MergeUpdateBuilder         = orm.MergeUpdateBuilder
	MergeInsertBuilder         = orm.MergeInsertBuilder
	CountBuilder               = orm.CountBuilder
	SumBuilder                 = orm.SumBuilder
	AvgBuilder                 = orm.AvgBuilder
	MinBuilder                 = orm.MinBuilder
	MaxBuilder                 = orm.MaxBuilder
	StringAggBuilder           = orm.StringAggBuilder
	ArrayAggBuilder            = orm.ArrayAggBuilder
	StdDevBuilder              = orm.StdDevBuilder
	VarianceBuilder            = orm.VarianceBuilder
	JSONObjectAggBuilder       = orm.JSONObjectAggBuilder
	JSONArrayAggBuilder        = orm.JSONArrayAggBuilder
	BitOrBuilder               = orm.BitOrBuilder
	BitAndBuilder              = orm.BitAndBuilder
	BoolOrBuilder              = orm.BoolOrBuilder
	BoolAndBuilder             = orm.BoolAndBuilder
	WindowCountBuilder         = orm.WindowCountBuilder
	WindowSumBuilder           = orm.WindowSumBuilder
	WindowAvgBuilder           = orm.WindowAvgBuilder
	WindowMinBuilder           = orm.WindowMinBuilder
	WindowMaxBuilder           = orm.WindowMaxBuilder
	WindowStringAggBuilder     = orm.WindowStringAggBuilder
	WindowArrayAggBuilder      = orm.WindowArrayAggBuilder
	WindowStdDevBuilder        = orm.WindowStdDevBuilder
	WindowVarianceBuilder      = orm.WindowVarianceBuilder
	WindowJSONObjectAggBuilder = orm.WindowJSONObjectAggBuilder
	WindowJSONArrayAggBuilder  = orm.WindowJSONArrayAggBuilder
	WindowBitOrBuilder         = orm.WindowBitOrBuilder
	WindowBitAndBuilder        = orm.WindowBitAndBuilder
	WindowBoolOrBuilder        = orm.WindowBoolOrBuilder
	WindowBoolAndBuilder       = orm.WindowBoolAndBuilder
	RowNumberBuilder           = orm.RowNumberBuilder
	RankBuilder                = orm.RankBuilder
	DenseRankBuilder           = orm.DenseRankBuilder
	PercentRankBuilder         = orm.PercentRankBuilder
	CumeDistBuilder            = orm.CumeDistBuilder
	NTileBuilder               = orm.NTileBuilder
	LagBuilder                 = orm.LagBuilder
	LeadBuilder                = orm.LeadBuilder
	FirstValueBuilder          = orm.FirstValueBuilder
	LastValueBuilder           = orm.LastValueBuilder
	NthValueBuilder            = orm.NthValueBuilder
)

const (
	// JoinType constants.
	JoinDefault = orm.JoinDefault
	JoinInner   = orm.JoinInner
	JoinLeft    = orm.JoinLeft
	JoinRight   = orm.JoinRight
	JoinFull    = orm.JoinFull
	JoinCross   = orm.JoinCross

	// FuzzyKind constants.
	FuzzyStarts   = orm.FuzzyStarts
	FuzzyEnds     = orm.FuzzyEnds
	FuzzyContains = orm.FuzzyContains

	// NullsMode constants.
	NullsDefault = orm.NullsDefault
	NullsRespect = orm.NullsRespect
	NullsIgnore  = orm.NullsIgnore

	// FromDirection constants.
	FromDefault = orm.FromDefault
	FromFirst   = orm.FromFirst
	FromLast    = orm.FromLast

	// FrameType constants.
	FrameDefault = orm.FrameDefault
	FrameRows    = orm.FrameRows
	FrameRange   = orm.FrameRange
	FrameGroups  = orm.FrameGroups

	// FrameBoundKind constants.
	FrameBoundNone               = orm.FrameBoundNone
	FrameBoundUnboundedPreceding = orm.FrameBoundUnboundedPreceding
	FrameBoundUnboundedFollowing = orm.FrameBoundUnboundedFollowing
	FrameBoundCurrentRow         = orm.FrameBoundCurrentRow
	FrameBoundPreceding          = orm.FrameBoundPreceding
	FrameBoundFollowing          = orm.FrameBoundFollowing

	// StatisticalMode constants.
	StatisticalDefault    = orm.StatisticalDefault
	StatisticalPopulation = orm.StatisticalPopulation
	StatisticalSample     = orm.StatisticalSample

	// ConflictAction constants.
	ConflictDoNothing = orm.ConflictDoNothing
	ConflictDoUpdate  = orm.ConflictDoUpdate

	// DateTimeUnit constants.
	UnitYear   = orm.UnitYear
	UnitMonth  = orm.UnitMonth
	UnitDay    = orm.UnitDay
	UnitHour   = orm.UnitHour
	UnitMinute = orm.UnitMinute
	UnitSecond = orm.UnitSecond
)

var ApplySort = orm.ApplySort
