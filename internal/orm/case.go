package orm

import (
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// CaseBuilder is an interface for building CASE expressions.
// Supports both searched CASE (WHEN condition) and simple CASE (WHEN value).
type CaseBuilder interface {
	// Case adds a CASE expression.
	Case(expr any) CaseBuilder
	// CaseColumn adds a CASE expression with a column.
	CaseColumn(column string) CaseBuilder
	// CaseSubQuery adds a CASE expression with a subquery.
	CaseSubQuery(func(query SelectQuery)) CaseBuilder
	// When adds a WHEN condition for searched CASE.
	When(func(cb ConditionBuilder)) CaseWhenBuilder
	// WhenExpr adds a WHEN expression for searched CASE.
	WhenExpr(expr any) CaseWhenBuilder
	// WhenSubQuery adds a WHEN subquery for searched CASE.
	WhenSubQuery(func(query SelectQuery)) CaseWhenBuilder
	// ElseExpr sets the ELSE expression for the CASE expression.
	Else(expr any)
	// ElseSubQuery adds a ELSE subquery for the CASE expression.
	ElseSubQuery(func(query SelectQuery))
}

// CaseWhenBuilder is an interface for building the THEN part of WHEN clauses.
type CaseWhenBuilder interface {
	// ThenExpr sets the result expression when the WHEN condition is true.
	Then(expr any) CaseBuilder
	// ThenSubQuery sets the result as a subquery when the WHEN condition is true.
	ThenSubQuery(func(query SelectQuery)) CaseBuilder
}

// newCaseExpr creates a new CASE expression builder.
func newCaseExpr(qb QueryBuilder) *caseExpr {
	return &caseExpr{
		qb:      qb,
		eb:      qb.ExprBuilder(),
		clauses: make([]caseClause, 0),
	}
}

// caseExpr implements CaseBuilder interface.
type caseExpr struct {
	qb       QueryBuilder
	eb       ExprBuilder
	caseExpr schema.QueryAppender
	clauses  []caseClause
	elseExpr schema.QueryAppender
	hasElse  bool
}

// caseClause represents a WHEN...THEN clause.
type caseClause struct {
	whenExpr schema.QueryAppender
	thenExpr schema.QueryAppender
}

// caseWhenExpr implements CaseWhenBuilder interface.
type caseWhenExpr struct {
	parent   *caseExpr
	whenExpr schema.QueryAppender
}

func (c *caseExpr) Case(expr any) CaseBuilder {
	c.caseExpr = c.eb.Expr("?", expr)

	return c
}

func (c *caseExpr) CaseColumn(column string) CaseBuilder {
	c.caseExpr = c.eb.Column(column)

	return c
}

func (c *caseExpr) CaseSubQuery(builder func(query SelectQuery)) CaseBuilder {
	c.caseExpr = c.eb.Expr("(?)", c.qb.BuildSubQuery(builder))

	return c
}

func (c *caseExpr) When(builder func(cb ConditionBuilder)) CaseWhenBuilder {
	return &caseWhenExpr{
		parent:   c,
		whenExpr: c.qb.BuildCondition(builder),
	}
}

func (c *caseExpr) WhenExpr(expr any) CaseWhenBuilder {
	return &caseWhenExpr{
		parent:   c,
		whenExpr: c.eb.Expr("?", expr),
	}
}

func (c *caseExpr) WhenSubQuery(builder func(query SelectQuery)) CaseWhenBuilder {
	return &caseWhenExpr{
		parent:   c,
		whenExpr: c.eb.Expr("(?)", c.qb.BuildSubQuery(builder)),
	}
}

func (c *caseExpr) Else(expr any) {
	c.elseExpr = c.eb.Expr("?", expr)
	c.hasElse = true
}

func (c *caseExpr) ElseSubQuery(builder func(query SelectQuery)) {
	c.elseExpr = c.eb.Expr("(?)", c.qb.BuildSubQuery(builder))
	c.hasElse = true
}

func (cw *caseWhenExpr) Then(expr any) CaseBuilder {
	cw.parent.clauses = append(cw.parent.clauses, caseClause{
		whenExpr: cw.whenExpr,
		thenExpr: cw.parent.eb.Expr("?", expr),
	})

	return cw.parent
}

func (cw *caseWhenExpr) ThenSubQuery(builder func(query SelectQuery)) CaseBuilder {
	cw.parent.clauses = append(cw.parent.clauses, caseClause{
		whenExpr: cw.whenExpr,
		thenExpr: cw.parent.eb.Expr("(?)", cw.parent.qb.BuildSubQuery(builder)),
	})

	return cw.parent
}

// AppendQuery implements schema.QueryAppender interface.
func (c *caseExpr) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	b = append(b, "CASE"...)

	// Add the CASE expression if it exists (for simple CASE)
	if c.caseExpr != nil {
		b = append(b, constants.ByteSpace)
		if b, err = c.caseExpr.AppendQuery(fmter, b); err != nil {
			return
		}
	}

	// Add WHEN...THEN clauses
	for _, clause := range c.clauses {
		b = append(b, " WHEN "...)
		if b, err = clause.whenExpr.AppendQuery(fmter, b); err != nil {
			return
		}

		b = append(b, " THEN "...)
		if b, err = clause.thenExpr.AppendQuery(fmter, b); err != nil {
			return
		}
	}

	// Add ELSE clause if exists
	if c.hasElse {
		b = append(b, " ELSE "...)
		if b, err = c.elseExpr.AppendQuery(fmter, b); err != nil {
			return
		}
	}

	b = append(b, " END"...)

	return b, nil
}
