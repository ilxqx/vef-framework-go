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

// BunDB is a wrapper around the bun.DB type.
type BunDB struct {
	db bun.IDB
}

func (d *BunDB) NewSelect() SelectQuery {
	return NewSelectQuery(d)
}

func (d *BunDB) NewInsert() InsertQuery {
	return NewInsertQuery(d)
}

func (d *BunDB) NewUpdate() UpdateQuery {
	return NewUpdateQuery(d)
}

func (d *BunDB) NewDelete() DeleteQuery {
	return NewDeleteQuery(d)
}

func (d *BunDB) NewMerge() MergeQuery {
	return NewMergeQuery(d)
}

func (d *BunDB) NewRaw(query string, args ...any) RawQuery {
	return newRawQuery(d, query, args...)
}

func (d *BunDB) RunInTX(ctx context.Context, fn func(context.Context, DB) error) error {
	return d.db.RunInTx(
		ctx,
		txOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, &BunDB{db: tx})
		},
	)
}

func (d *BunDB) RunInReadOnlyTX(ctx context.Context, fn func(context.Context, DB) error) error {
	return d.db.RunInTx(
		ctx,
		readOnlyTxOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, &BunDB{db: tx})
		},
	)
}

func (d *BunDB) WithNamedArg(name string, value any) DB {
	if db, ok := d.db.(*bun.DB); ok {
		return &BunDB{db: db.WithNamedArg(name, value)}
	}

	logger.Panicf("%q is not supported within a transaction context", "WithNamedArg")

	return d
}

func (d *BunDB) ModelPKs(model any) (map[string]any, error) {
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

func (d *BunDB) ModelPKFields(model any) []*PKField {
	table := getTableSchema(model, d.getBunDB())
	pks := make([]*PKField, len(table.PKs))

	for i, pk := range table.PKs {
		pks[i] = NewPKField(pk)
	}

	return pks
}

func (d *BunDB) TableOf(model any) *schema.Table {
	return getTableSchema(model, d.getBunDB())
}

func (d *BunDB) Unwrap() bun.IDB {
	return d.db
}

// getBunDB extracts the underlying *bun.DB from the wrapper.
// If the wrapper contains a transaction, it retrieves the DB from the transaction.
func (d *BunDB) getBunDB() *bun.DB {
	if db, ok := d.db.(*bun.DB); ok {
		return db
	}

	return d.db.NewDropTable().DB()
}
