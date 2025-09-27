package orm

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type (
	BaseModel         = bun.BaseModel
	BeforeScanRowHook = bun.BeforeScanRowHook
	AfterScanRowHook  = bun.AfterScanRowHook
	BunSelectQuery    = bun.SelectQuery
	BunInsertQuery    = bun.InsertQuery
	BunUpdateQuery    = bun.UpdateQuery
	BunDeleteQuery    = bun.DeleteQuery
	Table             = schema.Table
	Field             = schema.Field
	Relation          = schema.Relation
	Dialect           = schema.Dialect
)

type BeforeSelectHook interface {
	bun.BeforeSelectHook
	BeforeSelect(ctx context.Context, query *BunSelectQuery) error
}

type AfterSelectHook interface {
	bun.AfterSelectHook
	AfterSelect(ctx context.Context, query *BunSelectQuery) error
}

type BeforeInsertHook interface {
	bun.BeforeInsertHook
	BeforeInsert(ctx context.Context, query *BunInsertQuery) error
}

type AfterInsertHook interface {
	bun.AfterInsertHook
	AfterInsert(ctx context.Context, query *BunInsertQuery) error
}

type BeforeUpdateHook interface {
	bun.BeforeUpdateHook
	BeforeUpdate(ctx context.Context, query *BunUpdateQuery) error
}

type AfterUpdateHook interface {
	bun.AfterUpdateHook
	AfterUpdate(ctx context.Context, query *BunUpdateQuery) error
}

type BeforeDeleteHook interface {
	bun.BeforeDeleteHook
	BeforeDelete(ctx context.Context, query *BunDeleteQuery) error
}

type AfterDeleteHook interface {
	bun.AfterDeleteHook
	AfterDelete(ctx context.Context, query *BunDeleteQuery) error
}
