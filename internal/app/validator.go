package app

import "github.com/ilxqx/vef-framework-go/validator"

type structValidator struct {
	// structValidator implements Fiber's struct validator interface
}

func (*structValidator) Validate(out any) error {
	return validator.Validate(out) // Validate delegates to the framework's validator
}

func newStructValidator() *structValidator {
	return new(structValidator) // newStructValidator creates a new struct validator instance
}
