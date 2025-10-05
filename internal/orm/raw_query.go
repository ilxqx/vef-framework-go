package orm

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type bunRawQuery struct {
	query *bun.RawQuery
}

func newRawQuery(db bun.IDB, query string, args ...any) *bunRawQuery {
	return &bunRawQuery{query: db.NewRaw(query, args...)}
}

func (b *bunRawQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	return b.query.Exec(ctx, dest...)
}

func (b *bunRawQuery) Scan(ctx context.Context, dest ...any) error {
	return b.query.Scan(ctx, dest...)
}
