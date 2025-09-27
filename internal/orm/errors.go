package orm

import "errors"

var (
	ErrSubQuery                     = errors.New("cannot execute a subquery directly; use it as part of a parent query")
	ErrAggregateMissingArgs         = errors.New("aggregate function requires at least one argument")
	ErrDialectUnsupportedOperation  = errors.New("operation not supported by current database dialect")
	ErrAggregateUnsupportedFunction = errors.New("aggregate function not supported by current database dialect")
	ErrDialectHandlerMissing        = errors.New("no dialect handler available for requested operation")
	ErrMissingColumnOrExpression    = errors.New("order clause requires at least one column or expression")
	ErrModelMustBePointerToStruct   = errors.New("model must be a pointer to struct")
	ErrPrimaryKeyUnsupportedType    = errors.New("unsupported primary key type")
)
