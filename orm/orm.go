package orm

import "github.com/ilxqx/vef-framework-go/internal/orm"

type (
	Db                         = orm.Db
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
	ColumnInfo                 = orm.ColumnInfo
	Model                      = orm.Model
	ModelPK                    = orm.ModelPK
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
	NtileBuilder               = orm.NtileBuilder
	LagBuilder                 = orm.LagBuilder
	LeadBuilder                = orm.LeadBuilder
	FirstValueBuilder          = orm.FirstValueBuilder
	LastValueBuilder           = orm.LastValueBuilder
	NthValueBuilder            = orm.NthValueBuilder
)

const (
	// InnerJoin performs an INNER JOIN.
	InnerJoin = orm.JoinInner
	// LeftJoin performs a LEFT JOIN (default).
	LeftJoin = orm.JoinLeft
	// RightJoin performs a RIGHT JOIN.
	RightJoin = orm.JoinRight
)
