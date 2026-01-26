package resolver

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hbollon/go-edlib"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/api/handler"
	"github.com/ilxqx/vef-framework-go/reflectx"
)

var errorType = reflect.TypeFor[error]()

type funcHandler struct {
	isFactory bool
	h         reflect.Value
}

func (f *funcHandler) IsFactory() bool {
	return f.isFactory
}

func (f *funcHandler) H() reflect.Value {
	return f.h
}

func newFuncHandler(isFactory bool, h reflect.Value) handler.Func {
	return &funcHandler{
		isFactory: isFactory,
		h:         h,
	}
}

// findHandlerMethod locates a method on the target resource.
func findHandlerMethod(target reflect.Value, name string) (reflect.Value, error) {
	method := reflectx.FindMethod(target, name)
	if method.IsValid() {
		return method, nil
	}

	allMethods := reflectx.CollectMethods(target)
	lowerName := strings.ToLower(name)

	var matches []string

	for actualName := range allMethods {
		if strings.ToLower(actualName) == lowerName {
			matches = append(matches, actualName)
		}
	}

	switch len(matches) {
	case 0:
		return reflect.Value{}, fmt.Errorf("api action method %q not found in resource %q", name, target.Type().String())
	case 1:
		return allMethods[matches[0]], nil
	default:
		best := selectClosestMatch(name, matches)
		if best != constants.Empty {
			return allMethods[best], nil
		}

		return reflect.Value{}, fmt.Errorf("api action method %q matches multiple methods %v in resource %q",
			name, matches, target.Type().String())
	}
}

// selectClosestMatch finds the closest match from candidates using Levenshtein distance.
func selectClosestMatch(target string, candidates []string) string {
	if len(candidates) == 0 {
		return constants.Empty
	}

	var (
		bestMatch   string
		minDistance = -1
		ambiguous   bool
	)

	for _, candidate := range candidates {
		distance := edlib.LevenshteinDistance(target, candidate)
		if minDistance < 0 || distance < minDistance {
			minDistance = distance
			bestMatch = candidate
			ambiguous = false
		} else if distance == minDistance {
			ambiguous = true
		}
	}

	if ambiguous {
		return constants.Empty
	}

	return bestMatch
}

func validateHandlerSignature(method reflect.Type) error {
	numOut := method.NumOut()

	if numOut == 0 {
		return nil
	}

	if numOut == 1 {
		if method.Out(0) == errorType {
			return nil
		}

		return fmt.Errorf("handler method has invalid return type, must be 'error': %q -> %q",
			method.String(), method.Out(0).String())
	}

	return fmt.Errorf("handler method has too many return values, must have at most 1 (error) or none: %q has %d returns",
		method.String(), numOut)
}

// isHandlerFactory checks for factory signatures that return handler closures.
func isHandlerFactory(method reflect.Type) bool {
	numOut := method.NumOut()
	if numOut < 1 || numOut > 2 {
		return false
	}

	handlerType := method.Out(0)
	if handlerType.Kind() != reflect.Func {
		return false
	}

	if validateHandlerSignature(handlerType) != nil {
		return false
	}

	return numOut == 1 || method.Out(1) == errorType
}

func validateHandler(handler reflect.Value) error {
	if handler.Kind() != reflect.Func {
		return fmt.Errorf("provided handler must be a function, got %s", handler.Kind())
	}

	if handler.IsNil() {
		return fmt.Errorf("provided handler function cannot be nil")
	}

	if isHandlerFactory(handler.Type()) {
		return nil
	}

	return validateHandlerSignature(handler.Type())
}

func resolveHandlerFromSpec(spec api.OperationSpec, resource api.Resource) (any, error) {
	var h reflect.Value

	if methodName, ok := spec.Handler.(string); ok {
		method, err := findHandlerMethod(reflect.ValueOf(resource), methodName)
		if err != nil {
			return nil, err
		}

		h = method
	} else {
		h = reflect.ValueOf(spec.Handler)
	}

	if err := validateHandler(h); err != nil {
		return nil, err
	}

	return newFuncHandler(isHandlerFactory(h.Type()), h), nil
}
