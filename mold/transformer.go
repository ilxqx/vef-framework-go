package mold

import (
	"context"
	"reflect"
)

// Func defines a transform function for use.
type Func func(ctx context.Context, fl FieldLevel) error

// StructLevelFunc accepts all values needed for struct level manipulation.
//
// Why does this exist? For structs for which you may not have access or rights to add tags too,
// from other packages your using.
type StructLevelFunc func(ctx context.Context, sl StructLevel) error

// InterceptorFunc is a way to intercept custom types to redirect the functions to be applied to an inner typ/value.
// Eg. Sql.NullString, the manipulation should be done on the inner string.
type InterceptorFunc func(current reflect.Value) (inner reflect.Value)
