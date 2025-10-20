package orm

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

var (
	// TxOptions defines transaction options for read-write transactions.
	txOptions = &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	// ReadOnlyTxOptions defines transaction options for read-only transactions.
	readOnlyTxOptions = &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}
)

// BunDb is a wrapper around the bun.DB type.
type BunDb struct {
	db bun.IDB
}

func (d *BunDb) NewSelect() SelectQuery {
	return NewSelectQuery(d)
}

func (d *BunDb) NewInsert() InsertQuery {
	return NewInsertQuery(d)
}

func (d *BunDb) NewUpdate() UpdateQuery {
	return NewUpdateQuery(d)
}

func (d *BunDb) NewDelete() DeleteQuery {
	return NewDeleteQuery(d)
}

func (d *BunDb) NewMerge() MergeQuery {
	return NewMergeQuery(d)
}

func (d *BunDb) NewRaw(query string, args ...any) RawQuery {
	return newRawQuery(d, query, args...)
}

func (d *BunDb) RunInTx(ctx context.Context, fn func(context.Context, Db) error) error {
	return d.db.RunInTx(
		ctx,
		txOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, New(tx))
		},
	)
}

func (d *BunDb) RunInReadOnlyTx(ctx context.Context, fn func(context.Context, Db) error) error {
	return d.db.RunInTx(
		ctx,
		readOnlyTxOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, New(tx))
		},
	)
}

func (d *BunDb) WithNamedArg(name string, value any) Db {
	if db, ok := d.db.(*bun.DB); ok {
		return New(db.WithNamedArg(name, value))
	}

	logger.Panic("'WithNamedArg' is not supported within a transaction context")

	return d
}

func (d *BunDb) ModelPks(model any) (map[string]any, error) {
	pks := d.ModelPkFields(model)
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

func (d *BunDb) ModelPkFields(model any) []*PkField {
	var db *bun.DB
	if bd, ok := d.db.(*bun.DB); ok {
		db = bd
	} else {
		db = d.db.NewDropTable().DB()
	}

	table := getTableSchema(model, db)
	pks := make([]*PkField, len(table.PKs))

	for i, pk := range table.PKs {
		pks[i] = NewPkField(pk)
	}

	return pks
}

func (d *BunDb) TableOf(model any) *schema.Table {
	var db *bun.DB
	if bd, ok := d.db.(*bun.DB); ok {
		db = bd
	} else {
		db = d.db.NewDropTable().DB()
	}

	return getTableSchema(model, db)
}

func (d *BunDb) Unwrap() bun.IDB {
	return d.db
}
