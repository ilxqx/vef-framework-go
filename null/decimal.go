package null

import (
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type Decimal = decimal.NullDecimal

func NewDecimal(d decimal.Decimal, valid bool) Decimal {
	return decimal.NullDecimal{
		Decimal: d,
		Valid:   valid,
	}
}

func DecimalFrom(d decimal.Decimal) Decimal {
	return decimal.NewNullDecimal(d)
}

func DecimalFromPtr(d *decimal.Decimal) Decimal {
	if d == nil {
		return NewDecimal(lo.Empty[decimal.Decimal](), false)
	}

	return NewDecimal(*d, true)
}
