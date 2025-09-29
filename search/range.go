// Package search provides range value parsing functionality for search conditions.
// This file handles the extraction and parsing of range values from various data types
// including structs, strings, and slices, with support for multiple data formats
// like integers, decimals, dates, times, and datetimes.
package search

import (
	"reflect"
	"strings"
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/monad"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

var (
	rangeStartFieldIndex []int
	rangeEndFieldIndex   []int
)

func init() {
	field, ok := reflect.TypeFor[monad.Range[int]]().FieldByName("Start")
	if !ok {
		panic("mo.Range[int] struct must have a 'Start' field for range operations to work properly")
	}

	rangeStartFieldIndex = field.Index

	field, ok = reflect.TypeFor[monad.Range[int]]().FieldByName("End")
	if !ok {
		panic("mo.Range[int] struct must have an 'End' field for range operations to work properly")
	}

	rangeEndFieldIndex = field.Index
}

// getRangeValue gets the start and end values of the value.
func getRangeValue(fieldValue any, conditionParams map[string]string) (any, any, bool) {
	value := reflect.Indirect(reflect.ValueOf(fieldValue))
	valueType := value.Type()
	kind := valueType.Kind()

	if kind == reflect.Struct && reflectx.IsSimilarType(valueType, rangeType) {
		return value.FieldByIndex(rangeStartFieldIndex).Interface(), value.FieldByIndex(rangeEndFieldIndex).Interface(), true
	} else if kind == reflect.String {
		return parseStringRange(value.String(), conditionParams)
	} else if kind == reflect.Slice {
		return parseSliceRange(value)
	}

	return nil, nil, false
}

// parseStringRange parses the string range.
func parseStringRange(value string, conditionParams map[string]string) (any, any, bool) {
	delimiter := lo.CoalesceOrEmpty(conditionParams[ParamDelimiter], constants.Comma)
	values := strings.SplitN(value, delimiter, 2)
	if len(values) != 2 {
		logger.Warnf("Invalid range value, expected value delimited by %s, got %v", delimiter, value)
		return nil, nil, false
	}

	// Map type to corresponding parser function
	parserMap := map[string]func([]string) (any, any, bool){
		constants.TypeInt:      parseIntRange,
		constants.TypeDecimal:  parseDecimalRange,
		constants.TypeDate:     parseDateRange,
		constants.TypeTime:     parseTimeRange,
		constants.TypeDateTime: parseDateTimeRange,
	}

	if parser, exists := parserMap[conditionParams[ParamType]]; exists {
		return parser(values)
	}

	return nil, nil, false
}

// parseSliceRange parses slice range values.
func parseSliceRange(value reflect.Value) (any, any, bool) {
	if value.Len() != 2 {
		logger.Warnf("Invalid range value, expected slice of length 2, got %v", value.Interface())
		return nil, nil, false
	}

	return value.Index(0).Interface(), value.Index(1).Interface(), true
}

// parseIntRange parses integer range values.
func parseIntRange(values []string) (any, any, bool) {
	start, err := cast.ToIntE(values[0])
	if err != nil {
		logger.Warnf("Invalid range value, expected int, got %v", values[0])
		return nil, nil, false
	}

	end, err := cast.ToIntE(values[1])
	if err != nil {
		logger.Warnf("Invalid range value, expected int, got %v", values[1])
		return nil, nil, false
	}

	return start, end, true
}

// parseDecimalRange parses decimal range values.
func parseDecimalRange(values []string) (any, any, bool) {
	start, err := decimal.NewFromString(values[0])
	if err != nil {
		logger.Warnf("Invalid range value, expected decimal, got %v", values[0])
		return nil, nil, false
	}

	end, err := decimal.NewFromString(values[1])
	if err != nil {
		logger.Warnf("Invalid range value, expected decimal, got %v", values[1])
		return nil, nil, false
	}

	return start, end, true
}

// parseDateRange parses date range values.
func parseDateRange(values []string) (any, any, bool) {
	start, err := time.ParseInLocation(time.DateOnly, values[0], time.Local)
	if err != nil {
		logger.Warnf("Invalid range value, expected date, got %v", values[0])
		return nil, nil, false
	}

	end, err := time.ParseInLocation(time.DateOnly, values[1], time.Local)
	if err != nil {
		logger.Warnf("Invalid range value, expected date, got %v", values[1])
		return nil, nil, false
	}

	return start, end, true
}

// parseTimeRange parses time range values.
func parseTimeRange(values []string) (any, any, bool) {
	start, err := time.ParseInLocation(time.TimeOnly, values[0], time.Local)
	if err != nil {
		logger.Warnf("Invalid range value, expected time, got %v", values[0])
		return nil, nil, false
	}

	end, err := time.ParseInLocation(time.TimeOnly, values[1], time.Local)
	if err != nil {
		logger.Warnf("Invalid range value, expected time, got %v", values[1])
		return nil, nil, false
	}

	return start, end, true
}

// parseDateTimeRange parses datetime range values.
func parseDateTimeRange(values []string) (any, any, bool) {
	start, err := time.ParseInLocation(time.DateTime, values[0], time.Local)
	if err != nil {
		logger.Warnf("Invalid range value, expected datetime, got %v", values[0])
		return nil, nil, false
	}

	end, err := time.ParseInLocation(time.DateTime, values[1], time.Local)
	if err != nil {
		logger.Warnf("Invalid range value, expected datetime, got %v", values[1])
		return nil, nil, false
	}

	return start, end, true
}
