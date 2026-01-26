package encoding

import (
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// SimpleStruct represents a simple test structure with basic fields.
type SimpleStruct struct {
	Name   string `json:"name" xml:"name"`
	Age    int    `json:"age" xml:"age"`
	Active bool   `json:"active" xml:"active"`
}

// MediumStruct represents a medium complexity test structure.
type MediumStruct struct {
	ID       string            `json:"id" xml:"id"`
	Items    []SimpleStruct    `json:"items" xml:"Items>Item"`
	Tags     []string          `json:"tags" xml:"Tags>Tag"`
	Metadata map[string]string `json:"metadata" xml:"Metadata>Entry"`
}

// ComplexStruct represents a complex test structure with nested data.
type ComplexStruct struct {
	ID       string         `json:"id" xml:"id"`
	Data     map[string]any `json:"data" xml:"data"`
	Items    []MediumStruct `json:"items" xml:"Items>Item"`
	Nested   *ComplexStruct `json:"nested,omitempty" xml:"nested,omitempty"`
	Created  time.Time      `json:"created" xml:"created"`
	Score    float64        `json:"score" xml:"score"`
	Count    int64          `json:"count" xml:"count"`
	Enabled  bool           `json:"enabled" xml:"enabled"`
	Optional *string        `json:"optional,omitempty" xml:"optional,omitempty"`
}

// generateLargeData generates random binary data of specified size.
func generateLargeData(size int) []byte {
	data := make([]byte, size)
	_, _ = rand.Read(data)

	return data
}

// generateRandomString generates a random string of specified length.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)

	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

// generateSimpleStruct generates a random SimpleStruct.
func generateSimpleStruct() SimpleStruct {
	return SimpleStruct{
		Name:   generateRandomString(10),
		Age:    int(time.Now().Unix() % 100),
		Active: time.Now().Unix()%2 == 0,
	}
}

// generateMediumStruct generates a random MediumStruct.
func generateMediumStruct(itemCount int) MediumStruct {
	items := make([]SimpleStruct, itemCount)
	for i := range itemCount {
		items[i] = generateSimpleStruct()
	}

	tags := make([]string, itemCount)
	for i := range itemCount {
		tags[i] = generateRandomString(5)
	}

	metadata := make(map[string]string)
	for range itemCount {
		metadata[generateRandomString(5)] = generateRandomString(10)
	}

	return MediumStruct{
		ID:       generateRandomString(16),
		Items:    items,
		Tags:     tags,
		Metadata: metadata,
	}
}

// generateComplexStruct generates a random ComplexStruct.
func generateComplexStruct(depth int) ComplexStruct {
	data := map[string]any{
		"key1": generateRandomString(10),
		"key2": time.Now().Unix(),
		"key3": true,
	}

	items := make([]MediumStruct, 2)
	for i := range 2 {
		items[i] = generateMediumStruct(3)
	}

	optional := generateRandomString(10)

	result := ComplexStruct{
		ID:       generateRandomString(16),
		Data:     data,
		Items:    items,
		Created:  time.Now().UTC(),
		Score:    float64(time.Now().Unix()%100) + 0.5,
		Count:    time.Now().Unix(),
		Enabled:  true,
		Optional: &optional,
	}

	// Add nested structure if depth > 0
	if depth > 0 {
		nested := generateComplexStruct(depth - 1)
		result.Nested = &nested
	}

	return result
}

// generateNestedStruct generates a nested structure with specified depth.
func generateNestedStruct(depth int) any {
	if depth <= 0 {
		return generateSimpleStruct()
	}

	return generateComplexStruct(depth - 1)
}

// generateLargeStruct generates a large structure with many items.
func generateLargeStruct(targetSize int) ComplexStruct {
	// Estimate: each MediumStruct is roughly 1KB when encoded
	itemCount := max(targetSize/1024, 1)

	items := make([]MediumStruct, itemCount)
	for i := range itemCount {
		items[i] = generateMediumStruct(10)
	}

	return ComplexStruct{
		ID:      generateRandomString(16),
		Data:    map[string]any{"large": true},
		Items:   items,
		Created: time.Now().UTC(),
		Score:   99.9,
		Count:   int64(itemCount),
		Enabled: true,
	}
}

// generateUnicodeString generates a string with Unicode characters.
func generateUnicodeString() string {
	return "Hello世界🌍Привет مرحبا"
}

// generateSpecialCharString generates a string with special characters.
func generateSpecialCharString() string {
	return `Special chars: \n\t\r"'<>&`
}

// generateControlCharString generates a string with control characters.
func generateControlCharString() string {
	var sb strings.Builder
	// Add some printable chars
	sb.WriteString("Start")
	// Add control characters (ASCII 0-31)
	for i := range 32 {
		sb.WriteByte(byte(i))
	}

	sb.WriteString("End")

	return sb.String()
}

// assertStructEqual compares two structs and provides detailed error messages.
func assertStructEqual(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()

	return assert.Equal(t, expected, actual, msgAndArgs...)
}

// assertErrorContains checks if error contains expected substring.
func assertErrorContains(t *testing.T, err error, substring string, msgAndArgs ...any) bool {
	t.Helper()

	if !assert.Error(t, err, msgAndArgs...) {
		return false
	}

	return assert.Contains(t, err.Error(), substring, msgAndArgs...)
}

// assertNoErrorWithContext asserts no error with context message.
func assertNoErrorWithContext(t *testing.T, err error, context string) bool {
	t.Helper()

	return assert.NoError(t, err, fmt.Sprintf("should not return error when %s", context))
}

// assertErrorWithContext asserts error exists with context message.
func assertErrorWithContext(t *testing.T, err error, context string) bool {
	t.Helper()

	return assert.Error(t, err, fmt.Sprintf("should return error when %s", context))
}

// assertNotEmptyWithContext asserts value is not empty with context message.
func assertNotEmptyWithContext(t *testing.T, value any, context string) bool {
	t.Helper()

	return assert.NotEmpty(t, value, fmt.Sprintf("%s should not be empty", context))
}

// assertContainsWithContext asserts string contains substring with context message.
func assertContainsWithContext(t *testing.T, s, substring, context string) bool {
	t.Helper()

	return assert.Contains(t, s, substring, fmt.Sprintf("%s should contain '%s'", context, substring))
}

// assertEqualWithContext asserts equality with context message.
func assertEqualWithContext(t *testing.T, expected, actual any, fieldName string) bool {
	t.Helper()

	return assert.Equal(t, expected, actual, fmt.Sprintf("field %s should equal %v, but got %v", fieldName, expected, actual))
}

// Test functions for fixtures

func TestGenerateLargeData(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"Small_100B", 100},
		{"Medium_1KB", 1024},
		{"Large_10KB", 10240},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := generateLargeData(tt.size)
			assert.Len(t, data, tt.size, "generated data size should equal %d bytes", tt.size)
			assert.NotNil(t, data, "generated data should not be nil")
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"Short_5", 5},
		{"Medium_20", 20},
		{"Long_100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := generateRandomString(tt.length)
			assert.Len(t, str, tt.length, "generated string length should equal %d", tt.length)
			assert.NotEmpty(t, str, "generated string should not be empty")
			// Verify all characters are alphanumeric
			for _, ch := range str {
				assert.True(t, (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9'),
					"string should only contain alphanumeric characters, but contains '%c'", ch)
			}
		})
	}
}

func TestGenerateSimpleStruct(t *testing.T) {
	s := generateSimpleStruct()
	assert.NotEmpty(t, s.Name, "Name field should not be empty")
	assert.GreaterOrEqual(t, s.Age, 0, "Age field should be greater than or equal to 0")
	assert.LessOrEqual(t, s.Age, 100, "Age field should be less than or equal to 100")
}

func TestGenerateMediumStruct(t *testing.T) {
	itemCount := 5
	s := generateMediumStruct(itemCount)

	assert.NotEmpty(t, s.ID, "ID field should not be empty")
	assert.Len(t, s.Items, itemCount, "Items count should equal %d", itemCount)
	assert.Len(t, s.Tags, itemCount, "Tags count should equal %d", itemCount)
	assert.Len(t, s.Metadata, itemCount, "Metadata count should equal %d", itemCount)

	// Verify all items are valid
	for i, item := range s.Items {
		assert.NotEmpty(t, item.Name, "Items[%d].Name should not be empty", i)
	}
}

func TestGenerateComplexStruct(t *testing.T) {
	tests := []struct {
		name  string
		depth int
	}{
		{"NoNesting", 0},
		{"OneLevel", 1},
		{"TwoLevels", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := generateComplexStruct(tt.depth)

			assert.NotEmpty(t, s.ID, "ID field should not be empty")
			assert.NotNil(t, s.Data, "Data field should not be nil")
			assert.NotEmpty(t, s.Items, "Items field should not be empty")
			assert.NotZero(t, s.Created, "Created field should not be zero")
			assert.Greater(t, s.Score, 0.0, "Score field should be greater than 0")
			assert.Greater(t, s.Count, int64(0), "Count field should be greater than 0")
			assert.True(t, s.Enabled, "Enabled field should be true")
			assert.NotNil(t, s.Optional, "Optional field should not be nil")

			if tt.depth > 0 {
				assert.NotNil(t, s.Nested, "Nested field should not be nil at depth %d", tt.depth)
			} else {
				assert.Nil(t, s.Nested, "Nested field should be nil at depth 0")
			}
		})
	}
}

func TestGenerateNestedStruct(t *testing.T) {
	tests := []struct {
		name  string
		depth int
	}{
		{"Depth0", 0},
		{"Depth1", 1},
		{"Depth2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateNestedStruct(tt.depth)
			assert.NotNil(t, result, "generated struct should not be nil")

			if tt.depth == 0 {
				_, ok := result.(SimpleStruct)
				assert.True(t, ok, "should return SimpleStruct at depth 0")
			} else {
				_, ok := result.(ComplexStruct)
				assert.True(t, ok, "should return ComplexStruct at depth > 0")
			}
		})
	}
}

func TestGenerateLargeStruct(t *testing.T) {
	tests := []struct {
		name       string
		targetSize int
		minItems   int
	}{
		{"Small_1KB", 1024, 1},
		{"Medium_10KB", 10240, 10},
		{"Large_100KB", 102400, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := generateLargeStruct(tt.targetSize)

			assert.NotEmpty(t, s.ID, "ID field should not be empty")
			assert.GreaterOrEqual(t, len(s.Items), tt.minItems,
				"Items count should be at least %d", tt.minItems)
		})
	}
}

func TestGenerateUnicodeString(t *testing.T) {
	str := generateUnicodeString()
	assert.NotEmpty(t, str, "Unicode string should not be empty")
	assert.Contains(t, str, "世界", "should contain Chinese characters")
	assert.Contains(t, str, "🌍", "should contain emoji")
	assert.Contains(t, str, "Привет", "should contain Russian characters")
	assert.Contains(t, str, "مرحبا", "should contain Arabic characters")
}

func TestGenerateSpecialCharString(t *testing.T) {
	str := generateSpecialCharString()
	assert.NotEmpty(t, str, "special char string should not be empty")
	assert.Contains(t, str, `\n`, "should contain newline escape")
	assert.Contains(t, str, `"`, "should contain double quote")
	assert.Contains(t, str, `'`, "should contain single quote")
	assert.Contains(t, str, `<`, "should contain less-than sign")
	assert.Contains(t, str, `>`, "should contain greater-than sign")
	assert.Contains(t, str, `&`, "should contain ampersand")
}

func TestGenerateControlCharString(t *testing.T) {
	str := generateControlCharString()
	assert.NotEmpty(t, str, "control char string should not be empty")
	assert.Contains(t, str, "Start", "should contain 'Start'")
	assert.Contains(t, str, "End", "should contain 'End'")
	// Verify it contains control characters
	hasControlChar := false
	for _, ch := range str {
		if ch < 32 {
			hasControlChar = true

			break
		}
	}

	assert.True(t, hasControlChar, "should contain control characters")
}

func TestAssertStructEqual(t *testing.T) {
	s1 := SimpleStruct{Name: "Test", Age: 30, Active: true}
	s2 := SimpleStruct{Name: "Test", Age: 30, Active: true}

	// Just verify the function works without panicking
	_ = assertStructEqual(t, s1, s2, "equal structs should be equal")
}

func TestAssertErrorContains(t *testing.T) {
	err := fmt.Errorf("this is a test error with keyword")

	// Just verify the function works without panicking
	_ = assertErrorContains(t, err, "keyword", "error should contain keyword")
}

func TestAssertHelpers(t *testing.T) {
	t.Run("assertNoErrorWithContext", func(t *testing.T) {
		// Just verify the function works without panicking
		_ = assertNoErrorWithContext(t, nil, "test context")
	})

	t.Run("assertErrorWithContext", func(t *testing.T) {
		// Just verify the function works without panicking
		_ = assertErrorWithContext(t, fmt.Errorf("error"), "test context")
	})

	t.Run("assertNotEmptyWithContext", func(t *testing.T) {
		// Just verify the function works without panicking
		_ = assertNotEmptyWithContext(t, "value", "test context")
	})

	t.Run("assertContainsWithContext", func(t *testing.T) {
		// Just verify the function works without panicking
		_ = assertContainsWithContext(t, "hello world", "world", "test context")
	})

	t.Run("assertEqualWithContext", func(t *testing.T) {
		// Just verify the function works without panicking
		_ = assertEqualWithContext(t, 42, 42, "number field")
	})
}
