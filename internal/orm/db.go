package orm

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

var (
	// txOptions defines transaction options for read-write transactions
	txOptions = &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	// readOnlyTxOptions defines transaction options for read-only transactions
	readOnlyTxOptions = &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}
)

// bunDb is a wrapper around the bun.DB type.
type bunDb struct {
	db bun.IDB
}

func (d *bunDb) NewSelect() SelectQuery {
	return NewSelectQuery(d.db)
}

func (d *bunDb) NewInsert() InsertQuery {
	return NewInsertQuery(d.db)
}

func (d *bunDb) NewUpdate() UpdateQuery {
	return NewUpdateQuery(d.db)
}

func (d *bunDb) NewDelete() DeleteQuery {
	return NewDeleteQuery(d.db)
}

func (d *bunDb) NewMerge() MergeQuery {
	return NewMergeQuery(d.db)
}

func (d *bunDb) NewRaw(query string, args ...any) RawQuery {
	return newRawQuery(d.db, query, args...)
}

func (d *bunDb) RunInTx(ctx context.Context, fn func(context.Context, Db) error) error {
	return d.db.RunInTx(
		ctx,
		txOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, New(tx))
		},
	)
}

func (d *bunDb) RunInReadOnlyTx(ctx context.Context, fn func(context.Context, Db) error) error {
	return d.db.RunInTx(
		ctx,
		readOnlyTxOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, New(tx))
		},
	)
}

func (d *bunDb) WithNamedArg(name string, value any) Db {
	if db, ok := d.db.(*bun.DB); ok {
		return New(db.WithNamedArg(name, value))
	}

	panic("method WithNamedArg can not be called in transaction")
}

func (d *bunDb) ModelPKs(model any) (map[string]any, error) {
	pks := d.ModelPKFields(model)
	pkValues := make(map[string]any, len(pks))

	for _, pk := range pks {
		value, err := pk.Value(model)
		if err != nil {
			return nil, err
		}
		pkValues[pk.Name] = value
	}

	return pkValues, nil
}

func (d *bunDb) ModelPKFields(model any) []*PKField {
	var db *bun.DB
	if bd, ok := d.db.(*bun.DB); ok {
		db = bd
	} else {
		db = d.db.NewDropTable().DB()
	}

	table := getTableSchema(model, db)
	pks := make([]*PKField, len(table.PKs))

	for i, pk := range table.PKs {
		pks[i] = NewPKField(pk)
	}

	return pks
}

func (d *bunDb) Schema(model any) *schema.Table {
	var db *bun.DB
	if bd, ok := d.db.(*bun.DB); ok {
		db = bd
	} else {
		db = d.db.NewDropTable().DB()
	}

	return getTableSchema(model, db)
}
