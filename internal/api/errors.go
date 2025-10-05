package api

import "errors"

var (
	// ErrResolveParamType indicates failing to resolve handler parameter type.
	ErrResolveParamType = errors.New("failed to resolve api handler parameter type")
	// ErrUnmarshalParamsMustPointerStruct indicates unmarshal target must be pointer to struct.
	ErrUnmarshalParamsMustPointerStruct = errors.New("the parameter of UnmarshalParams function must be a pointer to a struct")
	// ErrProvidedHandlerNil indicates provided handler is nil.
	ErrProvidedHandlerNil = errors.New("provided handler cannot be nil")
	// ErrProvidedHandlerMustFunc indicates provided handler must be a function.
	ErrProvidedHandlerMustFunc = errors.New("provided handler must be a function")
	// ErrProvidedHandlerFuncNil indicates provided handler function is nil.
	ErrProvidedHandlerFuncNil = errors.New("provided handler function cannot be nil")
	// ErrHandlerFactoryRequireDB indicates handler factory requires db.
	ErrHandlerFactoryRequireDB = errors.New("handler factory function requires database connection but none provided")
	// ErrHandlerFactoryMethodRequireDB indicates handler factory method requires db.
	ErrHandlerFactoryMethodRequireDB = errors.New("handler factory method requires database connection but none provided")
	// ErrAPIMethodNotFound indicates api action method not found.
	ErrAPIMethodNotFound = errors.New("api action method not found in resource")
	// ErrHandlerMethodInvalidReturn indicates handler method invalid return type.
	ErrHandlerMethodInvalidReturn = errors.New("handler method has invalid return type, must be 'error'")
	// ErrHandlerMethodTooManyReturns indicates handler method has too many returns.
	ErrHandlerMethodTooManyReturns = errors.New("handler method has too many return values, must have at most 1 (error) or none")
	// ErrHandlerFactoryInvalidReturn indicates handler factory invalid return count.
	ErrHandlerFactoryInvalidReturn = errors.New("handler factory method should return 1 or 2 values")
)
