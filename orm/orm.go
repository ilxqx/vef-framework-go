package orm

import "github.com/ilxqx/vef-framework-go/internal/orm"

type (
	Db                 = orm.Db
	SelectQuery        = orm.SelectQuery
	InsertQuery        = orm.InsertQuery
	UpdateQuery        = orm.UpdateQuery
	DeleteQuery        = orm.DeleteQuery
	RawQuery           = orm.RawQuery
	ConditionBuilder   = orm.ConditionBuilder
	ApplyFunc[T any]   = orm.ApplyFunc[T]
	ModelRelation      = orm.ModelRelation
	Model              = orm.Model
	ExpressionBuilders = orm.ExprBuilder
	OrderBuilder       = orm.OrderBuilder
	CaseBuilder        = orm.CaseBuilder
	CaseWhenBuilder    = orm.CaseWhenBuilder
)
