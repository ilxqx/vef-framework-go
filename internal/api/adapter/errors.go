package adapter

import (
	"errors"
)

// ErrHandlerFactoryInvalidReturn indicates handler factory invalid return count.
var ErrHandlerFactoryInvalidReturn = errors.New("handler factory method should return 1 or 2 values")
