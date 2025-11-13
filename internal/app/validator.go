package app

import "github.com/ilxqx/vef-framework-go/validator"

// StructValidator implements Fiber's struct validator interface.
type StructValidator struct{}

func (*StructValidator) Validate(out any) error {
	// Validate delegates to the framework's validator
	return validator.Validate(out)
}

func newStructValidator() *StructValidator {
	return new(StructValidator)
}
