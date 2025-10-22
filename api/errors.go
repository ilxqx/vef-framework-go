package api

import "errors"

var (
	// ErrResourceNameEmpty is returned when the resource name is empty.
	ErrResourceNameEmpty = errors.New("resource name cannot be empty")
	// ErrResourceNameInvalidFormat is returned when the resource name is not in snake_case format.
	ErrResourceNameInvalidFormat = errors.New("resource name must be in snake_case format with optional slashes")
	// ErrResourceNameInvalidSlash is returned when the resource name starts or ends with a slash.
	ErrResourceNameInvalidSlash = errors.New("resource name cannot start or end with a slash")
	// ErrResourceNameConsecutiveSlashes is returned when the resource name contains consecutive slashes.
	ErrResourceNameConsecutiveSlashes = errors.New("resource name cannot contain consecutive slashes")
	// ErrActionNameEmpty is returned when the action name is empty.
	ErrActionNameEmpty = errors.New("action name cannot be empty")
	// ErrActionNameInvalidFormat is returned when the action name is not in snake_case format.
	ErrActionNameInvalidFormat = errors.New("action name must be in snake_case format")
	// ErrParamsDecodeTypeMismatch is returned when Params.Decode receives a non-pointer or non-struct type.
	ErrParamsDecodeTypeMismatch = errors.New("Params.Decode requires a pointer to struct")
	// ErrMetaDecodeTypeMismatch is returned when Meta.Decode receives a non-pointer or non-struct type.
	ErrMetaDecodeTypeMismatch = errors.New("Meta.Decode requires a pointer to struct")
)
