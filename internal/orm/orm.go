package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
)

var (
	logger = log.Named("orm")

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

func (d *bunDb) NewQuery() orm.Query {
	return NewQuery(d.db)
}

func (d *bunDb) NewCreate() orm.Create {
	return NewCreate(d.db)
}

func (d *bunDb) NewUpdate() orm.Update {
	return NewUpdate(d.db)
}

func (d *bunDb) NewDelete() orm.Delete {
	return NewDelete(d.db)
}

func (d *bunDb) NewRawQuery(query string, args ...any) orm.RawQuery {
	return newBunRawQuery(d.db, query, args...)
}

func (d *bunDb) RunInTx(ctx context.Context, fn func(context.Context, orm.Db) error) error {
	return d.db.RunInTx(
		ctx,
		txOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, newDb(tx))
		},
	)
}

func (d *bunDb) RunInReadOnlyTx(ctx context.Context, fn func(context.Context, orm.Db) error) error {
	return d.db.RunInTx(
		ctx,
		readOnlyTxOptions,
		func(ctx context.Context, tx bun.Tx) error {
			return fn(ctx, newDb(tx))
		},
	)
}

func (d *bunDb) WithNamedArg(name string, value any) orm.Db {
	if db, ok := d.db.(*bun.DB); ok {
		return newDb(db.WithNamedArg(name, value))
	}

	panic("method WithNamedArg can not be called in transaction")
}

func (d *bunDb) ModelPKs(model any) map[string]any {
	var db *bun.DB
	if bd, ok := d.db.(*bun.DB); ok {
		db = bd
	} else {
		db = d.db.NewDropTable().DB()
	}

	table := getTableSchema(model, db)
	modelValue := reflect.ValueOf(model).Elem()
	pks := make(map[string]any, len(table.PKs))

	for _, pk := range table.PKs {
		pks[pk.Name] = pk.Value(modelValue).Interface()
	}

	return pks
}

func (d *bunDb) Schema(model any) *orm.Table {
	var db *bun.DB
	if bd, ok := d.db.(*bun.DB); ok {
		db = bd
	} else {
		db = d.db.NewDropTable().DB()
	}

	return getTableSchema(model, db)
}

// newDb creates a new Db instance.
func newDb(db bun.IDB) orm.Db {
	return &bunDb{db: db}
}
