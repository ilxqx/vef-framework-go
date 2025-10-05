package app

import "github.com/ilxqx/vef-framework-go/validator"

// structValidator implements Fiber's struct validator interface.
type structValidator struct{}

func (*structValidator) Validate(out any) error {
	// Validate delegates to the framework's validator
	return validator.Validate(out)
}

func newStructValidator() *structValidator {
	return new(structValidator)
}
