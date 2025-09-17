package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"
)

// Names returns a query appender that appends a list of names to the query.
func Names(ns ...string) schema.QueryAppender {
	return &names{
		ns: ns,
	}
}

type names struct {
	ns []string // ns contains the list of names to append to the query
}

var _ schema.QueryAppender = (*names)(nil)

func (n *names) AppendQuery(formatter schema.Formatter, b []byte) ([]byte, error) {
	nsLen := len(n.ns)

	if nsLen == 0 {
		return dialect.AppendNull(b), nil
	}

	for i := range nsLen {
		if i > 0 {
			b = append(b, constants.CommaSpace...)
		}

		b = formatter.AppendName(b, n.ns[i])
	}

	return b, nil
}
