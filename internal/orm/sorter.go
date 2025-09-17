package orm

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"

	"github.com/samber/lo"
)

// Sorter is a struct that contains the orders of the query.
type Sorter struct {
	orders []Order // orders contains the list of order specifications
}

// Order is a struct that contains the column and the direction of the order.
type Order struct {
	Alias  string // Alias is the alias of the table
	Column string // Column is the column to order by
	Desc   bool   // Desc is true if the order is descending
}

// Apply applies the orders to the database.
func (s Sorter) Apply(query orm.Query, defaultAlias ...string) {
	var defaultAliasToUse string
	if len(defaultAlias) > 0 {
		defaultAliasToUse = defaultAlias[0]
	}

	for _, order := range s.orders {
		alias := lo.Ternary(order.Alias == constants.Empty, defaultAliasToUse, order.Alias)
		if alias != constants.Empty {
			if order.Desc {
				query.OrderByExpr("?.? DESC", bun.Name(alias), bun.Name(order.Column))
			} else {
				query.OrderByExpr("?.?", bun.Name(alias), bun.Name(order.Column))
			}
		} else {
			if order.Desc {
				query.OrderByExpr("? DESC", bun.Name(order.Column))
			} else {
				query.OrderByExpr("?", bun.Name(order.Column))
			}
		}
	}
}

// NewSorter creates a new Sorter from a sort string.
func NewSorter(sortString string) Sorter {
	sortString = strings.TrimSpace(sortString)
	if sortString == constants.Empty {
		return Sorter{}
	}

	sorts := strings.Split(sortString, constants.Comma)
	orders := make([]Order, 0, len(sorts))
	for _, sort := range sorts {
		sort = strings.TrimSpace(sort)
		if sort == constants.Empty {
			continue
		}

		pair := strings.SplitN(sort, constants.Colon, 2)
		alias, column := parseColumn(strings.TrimSpace(pair[0]))
		if len(pair) == 2 {
			direction := strings.ToLower(strings.TrimSpace(pair[1]))
			if direction == orm.OrderDesc {
				orders = append(
					orders,
					Order{
						Alias:  alias,
						Column: column,
						Desc:   true,
					},
				)
				continue
			}
		}

		orders = append(
			orders,
			Order{
				Alias:  alias,
				Column: column,
				Desc:   false,
			},
		)
	}

	return Sorter{orders: orders}
}

// parseColumn parses the column and returns the alias and the column name.
func parseColumn(column string) (string, string) {
	parts := strings.SplitN(column, constants.Dot, 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	return constants.Empty, column
}
