// Package mapx provides utilities for map and struct conversion using mapstructure.
// It offers flexible decoding options and type-safe conversions between maps and structs.
package mapx

import (
	"errors"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/samber/lo"
)

var (
	// defaultDecoderTagName is the default struct tag name used for map decoding.
	defaultDecoderTagName = "json"

	// DecoderHook composes multiple decode hooks for comprehensive type conversion.
	// It includes hooks for null types, text unmarshalling, time parsing, URL parsing, IP parsing, and basic type conversion.
	DecoderHook = mapstructure.ComposeDecodeHookFunc(
		decodeNullBool,
		decodeNullString,
		decodeNullInt,
		decodeNullInt16,
		decodeNullInt32,
		decodeNullFloat,
		decodeNullByte,
		decodeNullDateTime,
		decodeNullDate,
		decodeNullTime,
		decodeNullDecimal,
		decodeNullValue,
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToTimeHookFunc(time.DateTime),
		mapstructure.StringToTimeLocationHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToURLHookFunc(),
		mapstructure.StringToIPHookFunc(),
		mapstructure.StringToIPNetHookFunc(),
		mapstructure.StringToNetIPPrefixHookFunc(),
		mapstructure.StringToNetIPAddrHookFunc(),
		mapstructure.StringToNetIPAddrPortHookFunc(),
		mapstructure.StringToBasicTypeHookFunc(),
	)
)

type (
	// decoderOption is a function type for configuring decoder options.
	decoderOption func(c *mapstructure.DecoderConfig)

	// Metadata is an alias for mapstructure.Metadata that contains decoding metadata.
	Metadata = mapstructure.Metadata
)

// NewDecoder creates a mapstructure decoder with the given result and optional configuration.
// The decoder is configured with sensible defaults for struct-to-map and map-to-struct conversions.
func NewDecoder(result any, options ...decoderOption) (*mapstructure.Decoder, error) {
	config := &mapstructure.DecoderConfig{
		TagName:              defaultDecoderTagName,
		IgnoreUntaggedFields: false,
		DecodeHook:           DecoderHook,
		Squash:               true,
		SquashTagOption:      "inline",
		MatchName: func(mapKey, fieldName string) bool {
			return mapKey == lo.CamelCase(fieldName)
		},
		ErrorUnused:       false,
		ErrorUnset:        false,
		ZeroFields:        false,
		AllowUnsetPointer: false,
		Metadata:          nil,
		WeaklyTypedInput:  false,
		DecodeNil:         false,
		Result:            result,
	}

	for _, option := range options {
		option(config)
	}

	return mapstructure.NewDecoder(config)
}

// WithTagName sets the struct tag name used for field mapping.
// Default is "json". This specifies which struct tag to read for field names.
// Example: WithTagName("yaml") will use `yaml:"field_name"` tags.
func WithTagName(tagName string) decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.TagName = tagName
	}
}

// WithIgnoreUntaggedFields controls whether struct fields without explicit tags are ignored.
// When true, fields without tags are skipped during decoding (similar to `mapstructure:"-"`).
// When false (default), untagged fields use their Go field names for mapping.
func WithIgnoreUntaggedFields(ignoreUntaggedFields bool) decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.IgnoreUntaggedFields = ignoreUntaggedFields
	}
}

// WithDecodeHook sets a custom decode hook function for type conversion.
// The hook is called before decoding and allows modification of values before setting them.
// It's called for every map and value in the input. Returning an error will cause the entire decode to fail.
// This replaces the default DecoderHook which includes time, URL, IP, and basic type conversions.
func WithDecodeHook(decodeHook mapstructure.DecodeHookFunc) decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = decodeHook
	}
}

// WithMatchName sets a custom function to match map keys to struct field names.
// Default uses CamelCase matching. The function receives the map key and struct field name,
// and should return true if they match. This allows implementing case-sensitive matching,
// snake_case conversion, or other custom naming strategies.
func WithMatchName(matchName func(mapKey, fieldName string) bool) decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.MatchName = matchName
	}
}

// WithErrorUnused enables strict decoding that returns an error for unused keys in the input map.
// When enabled, any map keys that don't correspond to struct fields will cause decoding to fail.
// This is useful for validating that all input data is being processed.
func WithErrorUnused() decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.ErrorUnused = true
	}
}

// WithErrorUnset enables strict decoding that returns an error for unset fields in the result struct.
// When enabled, any struct fields that remain unset after decoding will cause an error.
// This applies to all nested structs and is useful for ensuring complete data population.
func WithErrorUnset() decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.ErrorUnset = true
	}
}

// WithZeroFields enables zeroing of fields before writing new values.
// When enabled, struct fields are reset to their zero values before decoding.
// For maps, this empties the map before adding decoded values.
// This is useful when reusing structs or ensuring clean state.
func WithZeroFields() decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.ZeroFields = true
	}
}

// WithAllowUnsetPointer prevents pointer-type fields from being reported as unset.
// When enabled, pointer fields that remain nil won't trigger errors with WithErrorUnset().
// This allows optional pointer fields without causing validation failures.
func WithAllowUnsetPointer() decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.AllowUnsetPointer = true
	}
}

// WithMetadata enables collection of decoding metadata.
// The provided Metadata struct will be populated with information about:
// - Keys: successfully decoded keys from the input
// - Unused: keys from input that weren't used
// - Unset: struct fields that weren't set during decoding
// This is useful for debugging and validation purposes.
func WithMetadata(metadata *Metadata) decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.Metadata = metadata
	}
}

// WithWeaklyTypedInput enables flexible type conversions during decoding.
// When enabled, the decoder will attempt to convert between types automatically:
// - Strings to numbers, booleans, and other types
// - Numbers to strings
// - Various other weak type transformations
// This is useful when working with loosely-typed data sources like JSON.
func WithWeaklyTypedInput() decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.WeaklyTypedInput = true
	}
}

// WithDecodeNil enables running DecodeHook even when input values are nil.
// When enabled, the decode hook function will be called for nil inputs,
// allowing the hook to provide default values or perform nil-specific processing.
// This is useful for implementing default value logic in decode hooks.
func WithDecodeNil() decoderOption {
	return func(c *mapstructure.DecoderConfig) {
		c.DecodeNil = true
	}
}

// ToMap converts a struct value to a map[string]any.
// The input value must be a struct or a pointer to a struct.
func ToMap(value any, options ...decoderOption) (result map[string]any, err error) {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() != reflect.Struct {
		return nil, errors.New("the value of ToMap function must be a struct")
	}

	var decoder *mapstructure.Decoder
	if decoder, err = NewDecoder(&result, options...); err != nil {
		return nil, err
	}

	if err = decoder.Decode(value); err != nil {
		return nil, err
	}

	return result, nil
}

// FromMap converts a map[string]any to a struct value.
// The type parameter T must be a struct type.
func FromMap[T any](value map[string]any, options ...decoderOption) (*T, error) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		return nil, errors.New("the type parameter of FromMap function must be a struct")
	}

	var (
		result  T
		decoder *mapstructure.Decoder
		err     error
	)

	if decoder, err = NewDecoder(&result, options...); err != nil {
		return nil, err
	}
	if err = decoder.Decode(value); err != nil {
		return nil, err
	}

	return &result, nil
}
