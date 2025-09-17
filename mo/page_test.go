package mo

import (
	"encoding/json"
	"testing"
)

func TestPageableNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    Pageable
		expected Pageable
	}{
		{
			"Normal values",
			Pageable{Page: 2, Size: 10, Sort: "name"},
			Pageable{Page: 2, Size: 10, Sort: "name"},
		},
		{
			"Page less than 1",
			Pageable{Page: 0, Size: 10, Sort: "name"},
			Pageable{Page: DefaultPageNumber, Size: 10, Sort: "name"},
		},
		{
			"Negative page",
			Pageable{Page: -1, Size: 10, Sort: "name"},
			Pageable{Page: DefaultPageNumber, Size: 10, Sort: "name"},
		},
		{
			"Size less than 1",
			Pageable{Page: 1, Size: 0, Sort: "name"},
			Pageable{Page: 1, Size: DefaultPageSize, Sort: "name"},
		},
		{
			"Negative size",
			Pageable{Page: 1, Size: -5, Sort: "name"},
			Pageable{Page: 1, Size: DefaultPageSize, Sort: "name"},
		},
		{
			"Size exceeds maximum",
			Pageable{Page: 1, Size: 2000, Sort: "name"},
			Pageable{Page: 1, Size: MaxPageSize, Sort: "name"},
		},
		{
			"All invalid values",
			Pageable{Page: -1, Size: -1, Sort: "name"},
			Pageable{Page: DefaultPageNumber, Size: DefaultPageSize, Sort: "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			if tt.input != tt.expected {
				t.Errorf("Expected %+v, got %+v", tt.expected, tt.input)
			}
		})
	}
}

func TestPageableOffset(t *testing.T) {
	tests := []struct {
		name     string
		pageable Pageable
		expected int
	}{
		{"Page 1, Size 10", Pageable{Page: 1, Size: 10}, 0},
		{"Page 2, Size 10", Pageable{Page: 2, Size: 10}, 10},
		{"Page 3, Size 15", Pageable{Page: 3, Size: 15}, 30},
		{"Page 1, Size 1", Pageable{Page: 1, Size: 1}, 0},
		{"Page 5, Size 20", Pageable{Page: 5, Size: 20}, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := tt.pageable.Offset()
			if offset != tt.expected {
				t.Errorf("Expected offset %d, got %d", tt.expected, offset)
			}
		})
	}
}

func TestNewPage(t *testing.T) {
	pageable := Pageable{Page: 2, Size: 10, Sort: "name"}
	items := []string{"item1", "item2", "item3"}
	total := int64(25)

	page := NewPage(pageable, total, items)

	if page.Page != 2 {
		t.Errorf("Expected Page to be 2, got %d", page.Page)
	}
	if page.Size != 10 {
		t.Errorf("Expected Size to be 10, got %d", page.Size)
	}
	if page.Total != 25 {
		t.Errorf("Expected Total to be 25, got %d", page.Total)
	}
	if len(page.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(page.Items))
	}
	if page.Items[0] != "item1" {
		t.Errorf("Expected first item to be 'item1', got %s", page.Items[0])
	}
}

func TestNewPageWithNilItems(t *testing.T) {
	pageable := Pageable{Page: 1, Size: 10}
	total := int64(0)

	page := NewPage[string](pageable, total, nil)

	if page.Items == nil {
		t.Error("Expected Items to be empty slice, not nil")
	}
	if len(page.Items) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(page.Items))
	}
}

func TestPageTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		page     Page[string]
		expected int
	}{
		{
			"Normal case",
			Page[string]{Size: 10, Total: 25},
			3,
		},
		{
			"Exact division",
			Page[string]{Size: 10, Total: 20},
			2,
		},
		{
			"Single item",
			Page[string]{Size: 10, Total: 1},
			1,
		},
		{
			"Zero items",
			Page[string]{Size: 10, Total: 0},
			0,
		},
		{
			"Zero size",
			Page[string]{Size: 0, Total: 10},
			0,
		},
		{
			"Large numbers",
			Page[string]{Size: 15, Total: 100},
			7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalPages := tt.page.TotalPages()
			if totalPages != tt.expected {
				t.Errorf("Expected %d total pages, got %d", tt.expected, totalPages)
			}
		})
	}
}

func TestPageHasNext(t *testing.T) {
	tests := []struct {
		name     string
		page     Page[string]
		expected bool
	}{
		{
			"Has next page",
			Page[string]{Page: 1, Size: 10, Total: 25},
			true,
		},
		{
			"Last page",
			Page[string]{Page: 3, Size: 10, Total: 25},
			false,
		},
		{
			"Single page",
			Page[string]{Page: 1, Size: 10, Total: 5},
			false,
		},
		{
			"Empty result",
			Page[string]{Page: 1, Size: 10, Total: 0},
			false,
		},
		{
			"Middle page",
			Page[string]{Page: 2, Size: 10, Total: 30},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasNext := tt.page.HasNext()
			if hasNext != tt.expected {
				t.Errorf("Expected HasNext to be %t, got %t", tt.expected, hasNext)
			}
		})
	}
}

func TestPageHasPrevious(t *testing.T) {
	tests := []struct {
		name     string
		page     Page[string]
		expected bool
	}{
		{
			"First page",
			Page[string]{Page: 1, Size: 10, Total: 25},
			false,
		},
		{
			"Second page",
			Page[string]{Page: 2, Size: 10, Total: 25},
			true,
		},
		{
			"Last page",
			Page[string]{Page: 3, Size: 10, Total: 25},
			true,
		},
		{
			"Middle page",
			Page[string]{Page: 5, Size: 10, Total: 100},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPrevious := tt.page.HasPrevious()
			if hasPrevious != tt.expected {
				t.Errorf("Expected HasPrevious to be %t, got %t", tt.expected, hasPrevious)
			}
		})
	}
}

func TestPageableJSONMarshaling(t *testing.T) {
	pageable := Pageable{
		Page: 2,
		Size: 15,
		Sort: "name",
	}

	// Marshal
	data, err := json.Marshal(pageable)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var result Pageable
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Compare
	if result != pageable {
		t.Errorf("Expected %+v, got %+v", pageable, result)
	}
}

func TestPageJSONMarshaling(t *testing.T) {
	items := []string{"item1", "item2", "item3"}
	page := Page[string]{
		Page:  2,
		Size:  10,
		Total: 25,
		Items: items,
	}

	// Marshal
	data, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var result Page[string]
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Compare
	if result.Page != page.Page {
		t.Errorf("Expected Page %d, got %d", page.Page, result.Page)
	}
	if result.Size != page.Size {
		t.Errorf("Expected Size %d, got %d", page.Size, result.Size)
	}
	if result.Total != page.Total {
		t.Errorf("Expected Total %d, got %d", page.Total, result.Total)
	}
	if len(result.Items) != len(page.Items) {
		t.Errorf("Expected %d items, got %d", len(page.Items), len(result.Items))
	}
	for i, item := range page.Items {
		if result.Items[i] != item {
			t.Errorf("Expected item[%d] to be %s, got %s", i, item, result.Items[i])
		}
	}
}

func TestPageWithDifferentTypes(t *testing.T) {
	// Test with struct
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}

	pageable := Pageable{Page: 1, Size: 10}
	userPage := NewPage(pageable, 2, users)

	if userPage.Items[0].Name != "Alice" {
		t.Errorf("Expected first user name to be Alice, got %s", userPage.Items[0].Name)
	}

	// Test with int
	numbers := []int{1, 2, 3, 4, 5}
	numberPage := NewPage(pageable, 5, numbers)

	if len(numberPage.Items) != 5 {
		t.Errorf("Expected 5 numbers, got %d", len(numberPage.Items))
	}
	if numberPage.Items[0] != 1 {
		t.Errorf("Expected first number to be 1, got %d", numberPage.Items[0])
	}
}

func TestPaginationScenarios(t *testing.T) {
	// Scenario: API pagination
	t.Run("API pagination workflow", func(t *testing.T) {
		// Client sends request for page 2, size 10
		pageable := Pageable{Page: 2, Size: 10, Sort: "created_at"}
		pageable.Normalize()

		// Database query would use offset
		offset := pageable.Offset()
		if offset != 10 {
			t.Errorf("Expected offset 10, got %d", offset)
		}

		// Mock data from database
		items := []string{"item11", "item12", "item13", "item14", "item15"}
		total := int64(45)

		// Create response page
		page := NewPage(pageable, total, items)

		// Verify page metadata
		if !page.HasPrevious() {
			t.Error("Expected page to have previous")
		}
		if !page.HasNext() {
			t.Error("Expected page to have next")
		}
		if page.TotalPages() != 5 {
			t.Errorf("Expected 5 total pages, got %d", page.TotalPages())
		}
	})

	// Scenario: Edge cases
	t.Run("Edge cases", func(t *testing.T) {
		// Empty result set
		emptyPageable := Pageable{Page: 1, Size: 10}
		emptyPage := NewPage[string](emptyPageable, 0, nil)

		if emptyPage.HasNext() {
			t.Error("Empty page should not have next")
		}
		if emptyPage.HasPrevious() {
			t.Error("Empty page should not have previous")
		}
		if emptyPage.TotalPages() != 0 {
			t.Errorf("Empty page should have 0 total pages, got %d", emptyPage.TotalPages())
		}

		// Single item result
		singlePageable := Pageable{Page: 1, Size: 10}
		singlePage := NewPage(singlePageable, 1, []string{"only item"})

		if singlePage.HasNext() {
			t.Error("Single item page should not have next")
		}
		if singlePage.TotalPages() != 1 {
			t.Errorf("Single item page should have 1 total page, got %d", singlePage.TotalPages())
		}
	})
}
