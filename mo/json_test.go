package mo

import (
	"encoding/json"
	"testing"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestNewJSON(t *testing.T) {
	value := TestStruct{Name: "Alice", Age: 25}
	j := NewJSON(value)

	if j.Unwrap().Name != "Alice" {
		t.Errorf("Expected Name to be Alice, got %s", j.Unwrap().Name)
	}
	if j.Unwrap().Age != 25 {
		t.Errorf("Expected Age to be 25, got %d", j.Unwrap().Age)
	}
}

func TestJSONUnwrap(t *testing.T) {
	value := 42
	j := NewJSON(value)

	unwrapped := j.Unwrap()
	if unwrapped != 42 {
		t.Errorf("Expected 42, got %d", unwrapped)
	}
}

func TestJSONMarshalJSON(t *testing.T) {
	value := TestStruct{Name: "Bob", Age: 30}
	j := NewJSON(value)

	data, err := j.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `{"name":"Bob","age":30}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestJSONUnmarshalJSON(t *testing.T) {
	data := []byte(`{"name":"Charlie","age":35}`)

	var j JSON[TestStruct]
	err := j.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	value := j.Unwrap()
	if value.Name != "Charlie" {
		t.Errorf("Expected Name to be Charlie, got %s", value.Name)
	}
	if value.Age != 35 {
		t.Errorf("Expected Age to be 35, got %d", value.Age)
	}
}

func TestJSONUnmarshalJSONInvalid(t *testing.T) {
	data := []byte(`invalid json`)

	var j JSON[TestStruct]
	err := j.UnmarshalJSON(data)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestJSONValue(t *testing.T) {
	value := TestStruct{Name: "David", Age: 40}
	j := NewJSON(value)

	driverValue, err := j.Value()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `{"name":"David","age":40}`
	if str, ok := driverValue.(string); !ok || str != expected {
		t.Errorf("Expected %s, got %v", expected, driverValue)
	}
}

func TestJSONScan(t *testing.T) {
	// Test scanning from []byte
	data := []byte(`{"name":"Eve","age":45}`)
	var j JSON[TestStruct]
	err := j.Scan(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	value := j.Unwrap()
	if value.Name != "Eve" {
		t.Errorf("Expected Name to be Eve, got %s", value.Name)
	}
	if value.Age != 45 {
		t.Errorf("Expected Age to be 45, got %d", value.Age)
	}
}

func TestJSONScanFromString(t *testing.T) {
	// Test scanning from string
	data := `{"name":"Frank","age":50}`
	var j JSON[TestStruct]
	err := j.Scan(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	value := j.Unwrap()
	if value.Name != "Frank" {
		t.Errorf("Expected Name to be Frank, got %s", value.Name)
	}
	if value.Age != 50 {
		t.Errorf("Expected Age to be 50, got %d", value.Age)
	}
}

func TestJSONScanFromBytePointer(t *testing.T) {
	// Test scanning from *[]byte
	data := []byte(`{"name":"Grace","age":55}`)
	var j JSON[TestStruct]
	err := j.Scan(&data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	value := j.Unwrap()
	if value.Name != "Grace" {
		t.Errorf("Expected Name to be Grace, got %s", value.Name)
	}
	if value.Age != 55 {
		t.Errorf("Expected Age to be 55, got %d", value.Age)
	}
}

func TestJSONScanFromStringPointer(t *testing.T) {
	// Test scanning from *string
	data := `{"name":"Henry","age":60}`
	var j JSON[TestStruct]
	err := j.Scan(&data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	value := j.Unwrap()
	if value.Name != "Henry" {
		t.Errorf("Expected Name to be Henry, got %s", value.Name)
	}
	if value.Age != 60 {
		t.Errorf("Expected Age to be 60, got %d", value.Age)
	}
}

func TestJSONScanNil(t *testing.T) {
	// Test scanning nil
	var j JSON[TestStruct]
	err := j.Scan(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Value should remain zero value
	value := j.Unwrap()
	if value.Name != "" || value.Age != 0 {
		t.Errorf("Expected zero value after scanning nil")
	}
}

func TestJSONScanNilPointers(t *testing.T) {
	var j JSON[TestStruct]

	// Test nil *[]byte
	err := j.Scan((*[]byte)(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test nil *string
	err = j.Scan((*string)(nil))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestJSONScanWithCast(t *testing.T) {
	// Test scanning from int (should be cast to string then parsed)
	var j JSON[int]
	err := j.Scan(123)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	value := j.Unwrap()
	if value != 123 {
		t.Errorf("Expected 123, got %d", value)
	}
}

func TestJSONScanInvalidJSON(t *testing.T) {
	// Test scanning invalid JSON
	var j JSON[TestStruct]
	err := j.Scan("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestJSONScanUnsupportedType(t *testing.T) {
	// Test scanning from unsupported type
	var j JSON[TestStruct]
	err := j.Scan(complex(1, 2)) // complex numbers can't be cast to string
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original JSON[TestStruct]
	}{
		{
			"Simple struct",
			NewJSON(TestStruct{Name: "Test", Age: 25}),
		},
		{
			"Empty struct",
			NewJSON(TestStruct{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.original)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			// Unmarshal
			var result JSON[TestStruct]
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			// Compare
			original := tt.original.Unwrap()
			resultValue := result.Unwrap()
			if original.Name != resultValue.Name || original.Age != resultValue.Age {
				t.Errorf("Round trip failed: %+v != %+v", original, resultValue)
			}
		})
	}
}

func TestJSONWithPrimitiveTypes(t *testing.T) {
	// Test with int
	intJSON := NewJSON(42)
	data, err := intJSON.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != "42" {
		t.Errorf("Expected 42, got %s", string(data))
	}

	var intResult JSON[int]
	err = intResult.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if intResult.Unwrap() != 42 {
		t.Errorf("Expected 42, got %d", intResult.Unwrap())
	}

	// Test with string
	strJSON := NewJSON("hello")
	data, err = strJSON.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != `"hello"` {
		t.Errorf("Expected \"hello\", got %s", string(data))
	}

	var strResult JSON[string]
	err = strResult.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if strResult.Unwrap() != "hello" {
		t.Errorf("Expected hello, got %s", strResult.Unwrap())
	}

	// Test with bool
	boolJSON := NewJSON(true)
	data, err = boolJSON.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(data) != "true" {
		t.Errorf("Expected true, got %s", string(data))
	}

	var boolResult JSON[bool]
	err = boolResult.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if boolResult.Unwrap() != true {
		t.Errorf("Expected true, got %t", boolResult.Unwrap())
	}
}