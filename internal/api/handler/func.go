package handler

import "reflect"

type Func interface {
	IsFactory() bool
	H() reflect.Value
}
