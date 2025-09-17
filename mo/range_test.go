package mo

import (
	"encoding/json"
	"testing"
)

func TestNewRange(t *testing.T) {
	// Test with int
	intRange := NewRange(1, 10)
	if intRange.Start != 1 {
		t.Errorf("Expected Start to be 1, got %d", intRange.Start)
	}
	if intRange.End != 10 {
		t.Errorf("Expected End to be 10, got %d", intRange.End)
	}

	// Test with string
	strRange := NewRange("a", "z")
	if strRange.Start != "a" {
		t.Errorf("Expected Start to be 'a', got %s", strRange.Start)
	}
	if strRange.End != "z" {
		t.Errorf("Expected End to be 'z', got %s", strRange.End)
	}

	// Test with float
	floatRange := NewRange(1.5, 9.5)
	if floatRange.Start != 1.5 {
		t.Errorf("Expected Start to be 1.5, got %f", floatRange.Start)
	}
	if floatRange.End != 9.5 {
		t.Errorf("Expected End to be 9.5, got %f", floatRange.End)
	}
}

func TestRangeContains(t *testing.T) {
	tests := []struct {
		name     string
		r        Range[int]
		value    int
		expected bool
	}{
		{"Value in range", NewRange(1, 10), 5, true},
		{"Value at start", NewRange(1, 10), 1, true},
		{"Value at end", NewRange(1, 10), 10, true},
		{"Value below range", NewRange(1, 10), 0, false},
		{"Value above range", NewRange(1, 10), 11, false},
		{"Single value range - contains", NewRange(5, 5), 5, true},
		{"Single value range - not contains", NewRange(5, 5), 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r.Contains(tt.value)
			if result != tt.expected {
				t.Errorf("Expected Contains(%d) to be %t, got %t", tt.value, tt.expected, result)
			}
		})
	}
}

func TestRangeContainsString(t *testing.T) {
	strRange := NewRange("b", "y")

	tests := []struct {
		value    string
		expected bool
	}{
		{"a", false},
		{"b", true},
		{"m", true},
		{"y", true},
		{"z", false},
	}

	for _, tt := range tests {
		t.Run("String "+tt.value, func(t *testing.T) {
			result := strRange.Contains(tt.value)
			if result != tt.expected {
				t.Errorf("Expected Contains(%s) to be %t, got %t", tt.value, tt.expected, result)
			}
		})
	}
}

func TestRangeContainsFloat(t *testing.T) {
	floatRange := NewRange(1.5, 9.5)

	tests := []struct {
		value    float64
		expected bool
	}{
		{1.0, false},
		{1.5, true},
		{5.25, true},
		{9.5, true},
		{10.0, false},
	}

	for _, tt := range tests {
		t.Run("Float", func(t *testing.T) {
			result := floatRange.Contains(tt.value)
			if result != tt.expected {
				t.Errorf("Expected Contains(%f) to be %t, got %t", tt.value, tt.expected, result)
			}
		})
	}
}

func TestRangeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		r        Range[int]
		expected bool
	}{
		{"Valid range", NewRange(1, 10), true},
		{"Single value range", NewRange(5, 5), true},
		{"Invalid range", NewRange(10, 1), false},
		{"Zero range", NewRange(0, 0), true},
		{"Negative range", NewRange(-10, -1), true},
		{"Invalid negative range", NewRange(-1, -10), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r.IsValid()
			if result != tt.expected {
				t.Errorf("Expected IsValid() to be %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestRangeIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		r        Range[int]
		expected bool
	}{
		{"Valid range", NewRange(1, 10), false},
		{"Single value range", NewRange(5, 5), false},
		{"Empty range", NewRange(10, 1), true},
		{"Zero range", NewRange(0, 0), false},
		{"Empty negative range", NewRange(-1, -10), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r.IsEmpty()
			if result != tt.expected {
				t.Errorf("Expected IsEmpty() to be %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestRangeOverlaps(t *testing.T) {
	tests := []struct {
		name     string
		r1       Range[int]
		r2       Range[int]
		expected bool
	}{
		{"Complete overlap", NewRange(1, 10), NewRange(3, 7), true},
		{"Partial overlap", NewRange(1, 5), NewRange(3, 8), true},
		{"Adjacent ranges", NewRange(1, 5), NewRange(5, 10), true},
		{"No overlap", NewRange(1, 3), NewRange(5, 10), false},
		{"Reverse overlap", NewRange(5, 10), NewRange(1, 7), true},
		{"Same range", NewRange(1, 10), NewRange(1, 10), true},
		{"Single point overlap", NewRange(1, 5), NewRange(5, 5), true},
		{"No overlap with gap", NewRange(1, 3), NewRange(6, 10), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r1.Overlaps(tt.r2)
			if result != tt.expected {
				t.Errorf("Expected %v.Overlaps(%v) to be %t, got %t", tt.r1, tt.r2, tt.expected, result)
			}

			// Test symmetry
			reverseResult := tt.r2.Overlaps(tt.r1)
			if reverseResult != tt.expected {
				t.Errorf("Expected symmetry: %v.Overlaps(%v) should also be %t, got %t", tt.r2, tt.r1, tt.expected, reverseResult)
			}
		})
	}
}

func TestRangeIntersection(t *testing.T) {
	tests := []struct {
		name     string
		r1       Range[int]
		r2       Range[int]
		expected Range[int]
		isEmpty  bool
	}{
		{
			"Complete overlap",
			NewRange(1, 10),
			NewRange(3, 7),
			NewRange(3, 7),
			false,
		},
		{
			"Partial overlap",
			NewRange(1, 5),
			NewRange(3, 8),
			NewRange(3, 5),
			false,
		},
		{
			"Adjacent ranges",
			NewRange(1, 5),
			NewRange(5, 10),
			NewRange(5, 5),
			false,
		},
		{
			"No overlap",
			NewRange(1, 3),
			NewRange(5, 10),
			NewRange(3, 1), // Empty range
			true,
		},
		{
			"Same range",
			NewRange(1, 10),
			NewRange(1, 10),
			NewRange(1, 10),
			false,
		},
		{
			"Reverse overlap",
			NewRange(5, 10),
			NewRange(1, 7),
			NewRange(5, 7),
			false,
		},
		{
			"Single point",
			NewRange(1, 5),
			NewRange(5, 5),
			NewRange(5, 5),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r1.Intersection(tt.r2)

			if tt.isEmpty {
				if !result.IsEmpty() {
					t.Errorf("Expected empty intersection, got %v", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("Expected intersection %v, got %v", tt.expected, result)
				}
			}

			// Test symmetry
			reverseResult := tt.r2.Intersection(tt.r1)
			if tt.isEmpty {
				if !reverseResult.IsEmpty() {
					t.Errorf("Expected symmetric empty intersection, got %v", reverseResult)
				}
			} else {
				if reverseResult != tt.expected {
					t.Errorf("Expected symmetric intersection %v, got %v", tt.expected, reverseResult)
				}
			}
		})
	}
}

func TestRangeIntersectionString(t *testing.T) {
	r1 := NewRange("c", "m")
	r2 := NewRange("f", "z")

	intersection := r1.Intersection(r2)
	expected := NewRange("f", "m")

	if intersection != expected {
		t.Errorf("Expected string intersection %v, got %v", expected, intersection)
	}

	// Test no overlap
	r3 := NewRange("a", "b")
	r4 := NewRange("x", "z")

	noOverlap := r3.Intersection(r4)
	if !noOverlap.IsEmpty() {
		t.Errorf("Expected empty intersection for non-overlapping string ranges, got %v", noOverlap)
	}
}

func TestRangeJSONMarshaling(t *testing.T) {
	r := NewRange(5, 15)

	// Marshal
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var result Range[int]
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Compare
	if result != r {
		t.Errorf("Expected %v, got %v", r, result)
	}
}

func TestRangeWithDifferentTypes(t *testing.T) {
	// Test with different numeric types
	int8Range := NewRange[int8](1, 10)
	if !int8Range.Contains(5) {
		t.Error("int8 range should contain 5")
	}

	int16Range := NewRange[int16](100, 200)
	if !int16Range.Contains(150) {
		t.Error("int16 range should contain 150")
	}

	int32Range := NewRange[int32](1000, 2000)
	if !int32Range.Contains(1500) {
		t.Error("int32 range should contain 1500")
	}

	int64Range := NewRange[int64](10000, 20000)
	if !int64Range.Contains(15000) {
		t.Error("int64 range should contain 15000")
	}

	uint8Range := NewRange[uint8](10, 250)
	if !uint8Range.Contains(100) {
		t.Error("uint8 range should contain 100")
	}

	uint16Range := NewRange[uint16](1000, 50000)
	if !uint16Range.Contains(25000) {
		t.Error("uint16 range should contain 25000")
	}

	uint32Range := NewRange[uint32](100000, 500000)
	if !uint32Range.Contains(300000) {
		t.Error("uint32 range should contain 300000")
	}

	uint64Range := NewRange[uint64](1000000, 5000000)
	if !uint64Range.Contains(3000000) {
		t.Error("uint64 range should contain 3000000")
	}

	float32Range := NewRange[float32](1.1, 9.9)
	if !float32Range.Contains(5.5) {
		t.Error("float32 range should contain 5.5")
	}

	float64Range := NewRange[float64](1.1, 9.9)
	if !float64Range.Contains(5.5) {
		t.Error("float64 range should contain 5.5")
	}
}

func TestRangeEdgeCases(t *testing.T) {
	// Test with maximum values
	maxRange := NewRange[uint8](0, 255)
	if !maxRange.IsValid() {
		t.Error("Max uint8 range should be valid")
	}
	if !maxRange.Contains(255) {
		t.Error("Max uint8 range should contain 255")
	}

	// Test with negative values
	negRange := NewRange(-100, -10)
	if !negRange.IsValid() {
		t.Error("Negative range should be valid")
	}
	if !negRange.Contains(-50) {
		t.Error("Negative range should contain -50")
	}
	if negRange.Contains(-5) {
		t.Error("Negative range should not contain -5")
	}

	// Test range operations with negative numbers
	r1 := NewRange(-10, 0)
	r2 := NewRange(-5, 5)

	if !r1.Overlaps(r2) {
		t.Error("Negative ranges should overlap")
	}

	intersection := r1.Intersection(r2)
	expected := NewRange(-5, 0)
	if intersection != expected {
		t.Errorf("Expected negative intersection %v, got %v", expected, intersection)
	}
}

func TestRangeStringOperations(t *testing.T) {
	// Test comprehensive string range operations
	r1 := NewRange("apple", "orange")
	r2 := NewRange("banana", "zebra")

	// Test contains
	if !r1.Contains("mango") {
		t.Error("Range [apple, orange] should contain 'mango'")
	}
	if r1.Contains("pear") {
		t.Error("Range [apple, orange] should not contain 'pear'")
	}

	// Test overlap
	if !r1.Overlaps(r2) {
		t.Error("String ranges should overlap")
	}

	// Test intersection
	intersection := r1.Intersection(r2)
	expectedStart := "banana"
	expectedEnd := "orange"
	if intersection.Start != expectedStart || intersection.End != expectedEnd {
		t.Errorf("Expected string intersection [%s, %s], got [%s, %s]",
			expectedStart, expectedEnd, intersection.Start, intersection.End)
	}

	// Test non-overlapping string ranges
	r3 := NewRange("aaa", "bbb")
	r4 := NewRange("yyy", "zzz")

	if r3.Overlaps(r4) {
		t.Error("Non-overlapping string ranges should not overlap")
	}

	noOverlap := r3.Intersection(r4)
	if !noOverlap.IsEmpty() {
		t.Error("Non-overlapping string ranges should have empty intersection")
	}
}
