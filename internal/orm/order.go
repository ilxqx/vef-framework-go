package orm

import (
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/sort"
)

// OrderBuilder provides a fluent interface for building ORDER BY clauses.
type OrderBuilder interface {
	// Column specifies the column name to order by
	Column(column string) OrderBuilder
	// Expr allows ordering by a SQL expression
	Expr(expr any) OrderBuilder
	// Asc sets ascending order (default)
	Asc() OrderBuilder
	// Desc sets descending order
	Desc() OrderBuilder
	// NullsFirst sets NULL values to appear first
	NullsFirst() OrderBuilder
	// NullsLast sets NULL values to appear last
	NullsLast() OrderBuilder
}

// orderExpr implements OrderBuilder interface.
type orderExpr struct {
	builders   ExprBuilder
	column     string
	direction  sort.OrderDirection
	nullsOrder sort.NullsOrder
	expr       any
}

func (o *orderExpr) Column(column string) OrderBuilder {
	o.column = column
	o.expr = nil

	return o
}

func (o *orderExpr) Expr(expr any) OrderBuilder {
	o.expr = expr
	o.column = constants.Empty

	return o
}

func (o *orderExpr) Asc() OrderBuilder {
	o.direction = sort.OrderAsc

	return o
}

func (o *orderExpr) Desc() OrderBuilder {
	o.direction = sort.OrderDesc

	return o
}

func (o *orderExpr) NullsFirst() OrderBuilder {
	o.nullsOrder = sort.NullsFirst

	return o
}

func (o *orderExpr) NullsLast() OrderBuilder {
	o.nullsOrder = sort.NullsLast

	return o
}

func (o *orderExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if o.column == constants.Empty && o.expr == nil {
		return nil, ErrMissingColumnOrExpression
	}

	if o.column != constants.Empty {
		b, err = o.builders.Column(o.column).AppendQuery(fmter, b)
	} else {
		b, err = o.builders.Expr("?", o.expr).AppendQuery(fmter, b)
	}

	if err != nil {
		return
	}

	b = append(b, constants.ByteSpace)
	b = append(b, o.direction.String()...)

	if o.nullsOrder != sort.NullsDefault {
		b = append(b, constants.ByteSpace)
		b = append(b, o.nullsOrder.String()...)
	}

	return b, nil
}

type orderByClause struct {
	exprs []orderExpr
}

// newOrderExpr creates a new OrderBuilder with default ascending order.
func (o *orderByClause) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	b = append(b, "ORDER BY "...)

	for i, expr := range o.exprs {
		if i > 0 {
			b = append(b, constants.CommaSpace...)
		}

		if b, err = expr.AppendQuery(fmter, b); err != nil {
			return
		}
	}

	return b, nil
}

func newOrderExpr(builders ExprBuilder) *orderExpr {
	return &orderExpr{
		builders:   builders,
		direction:  sort.OrderAsc,
		nullsOrder: sort.NullsDefault,
	}
}

func newOrderByClause(exprs ...orderExpr) *orderByClause {
	return &orderByClause{
		exprs: exprs,
	}
}
