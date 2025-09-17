package utils

import (
	"encoding/xml"
	"errors"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-json"
	"github.com/ilxqx/vef-framework-go/constants"
)

var (
	mapDecoderTagName = "json"                              // mapDecoderTagName is the struct tag name for map decoding
	MapDecoderHook    = mapstructure.ComposeDecodeHookFunc( // mapDecoderHook composes multiple decode hooks for type conversion
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

// ToMap converts a struct value to a map[string]any.
func ToMap(value any) (map[string]any, error) {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() != reflect.Struct {
		return nil, errors.New("the value of ToMap function must be a struct")
	}

	var result map[string]any
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:              mapDecoderTagName,
		IgnoreUntaggedFields: false,
		DecodeHook:           MapDecoderHook,
		Result:               &result,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(value); err != nil {
		return nil, err
	}
	return result, nil
}

// FromMap converts a map[string]any to a struct value.
func FromMap[T any](value map[string]any) (*T, error) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		return nil, errors.New("the type parameter of FromMap function must be a struct")
	}

	var result T
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:              mapDecoderTagName,
		IgnoreUntaggedFields: false,
		DecodeHook:           MapDecoderHook,
		Result:               &result,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(value); err != nil {
		return nil, err
	}
	return &result, nil
}

// NewMapDecoder creates a mapstructure decoder with the given result.
func NewMapDecoder(result any) (*mapstructure.Decoder, error) {
	return mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:              mapDecoderTagName,
		IgnoreUntaggedFields: false,
		DecodeHook:           MapDecoderHook,
		Result:               result,
	})
}

// ToJSON converts a struct value to a JSON string.
func ToJSON(value any) (string, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return constants.Empty, err
	}

	return string(jsonBytes), nil
}

// FromJSON converts a JSON string to a struct value.
func FromJSON[T any](value string) (*T, error) {
	var result T
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ToXML converts a struct value to an XML string.
func ToXML(value any) (string, error) {
	xmlBytes, err := xml.Marshal(value)
	if err != nil {
		return constants.Empty, err
	}

	return string(xmlBytes), nil
}

// FromXML converts an XML string to a struct value.
func FromXML[T any](value string) (*T, error) {
	var result T
	if err := xml.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
