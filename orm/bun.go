package orm

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type (
	BaseModel         = bun.BaseModel
	BeforeScanRowHook = bun.BeforeScanRowHook
	AfterScanRowHook  = bun.AfterScanRowHook
	BeforeSelectHook  = bun.BeforeSelectHook
	AfterSelectHook   = bun.AfterSelectHook
	BeforeInsertHook  = bun.BeforeInsertHook
	AfterInsertHook   = bun.AfterInsertHook
	BeforeUpdateHook  = bun.BeforeUpdateHook
	AfterUpdateHook   = bun.AfterUpdateHook
	BeforeDeleteHook  = bun.BeforeDeleteHook
	AfterDeleteHook   = bun.AfterDeleteHook
	SelectQuery       = bun.SelectQuery
	InsertQuery       = bun.InsertQuery
	UpdateQuery       = bun.UpdateQuery
	DeleteQuery       = bun.DeleteQuery
	Ident             = bun.Ident
	Name              = bun.Name
	Safe              = bun.Safe
	Table             = schema.Table
	Field             = schema.Field
	Relation          = schema.Relation
	Dialect           = schema.Dialect
)
