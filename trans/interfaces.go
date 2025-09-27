package trans

import (
	"context"
	"reflect"

	"github.com/ilxqx/vef-framework-go/null"
)

// Transformer defines the main interface for struct transformers that provide tag-based data transformation
type Transformer interface {
	// Struct applies transformations to the entire struct based on field tags
	Struct(ctx context.Context, value any) error
	// Field applies specified transformation tags to a single field
	Field(ctx context.Context, value any, tags string) error
}

// FieldTransformer defines the field-level transformer interface for extending custom field transformation logic
type FieldTransformer interface {
	// Tag returns the tag name corresponding to this transformer, used for referencing in struct tags
	Tag() string
	// Transform executes field transformation logic, fl provides field-level context information
	Transform(ctx context.Context, fl FieldLevel) error
}

// StructTransformer defines the struct-level transformer interface for custom processing of entire structs
type StructTransformer interface {
	// Transform executes struct-level transformation logic, sl provides struct-level context information
	Transform(ctx context.Context, sl StructLevel) error
}

// Interceptor defines the interceptor interface for redirecting transformation operations of certain types to inner values
// For example: sql.NullString transformations should operate on its inner string value
type Interceptor interface {
	// Intercept intercepts the current value and returns the inner value that should be actually operated on
	Intercept(current reflect.Value) (inner reflect.Value)
}

// DataDictNameResolver defines the data dictionary resolver interface for converting codes to readable names
// Supports multi-level data dictionaries, using key to distinguish different dictionary types
type DataDictNameResolver interface {
	// Resolve resolves the corresponding name based on dictionary key and code value
	// Returns Option type, None indicates no corresponding name was found
	Resolve(ctx context.Context, key, code string) null.String
}
