package trans

import (
	"github.com/go-playground/mold/v4"
)

// Type aliases that directly reference mold library type definitions to maintain API consistency
type (
	FieldLevel      = mold.FieldLevel
	StructLevel     = mold.StructLevel
	Func            = mold.Func
	StructLevelFunc = mold.StructLevelFunc
	InterceptorFunc = mold.InterceptorFunc
)
