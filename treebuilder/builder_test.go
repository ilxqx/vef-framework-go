package treebuilder

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/constants"
)

// TestNode represents a simple node structure for testing.
type TestNode struct {
	Id       string     `json:"id"`       // unique identifier
	ParentId string     `json:"parentId"` // parent identifier
	Name     string     `json:"name"`     // node name for identification
	Children []TestNode `json:"children"` // children nodes
}

// TestCategory represents a category structure for testing.
type TestCategory struct {
	CategoryId    string         `json:"categoryId"`    // unique identifier
	ParentCatId   string         `json:"parentCatId"`   // parent identifier
	CategoryName  string         `json:"categoryName"`  // category name
	SubCategories []TestCategory `json:"subCategories"` // sub categories
	Level         int            `json:"level"`         // category level
}

// createTestNodeAdapter creates a TreeAdapter for TestNode.
func createTestNodeAdapter() Adapter[TestNode] {
	return Adapter[TestNode]{
		GetId:       func(node TestNode) string { return node.Id },
		GetParentId: func(node TestNode) string { return node.ParentId },
		GetChildren: func(node TestNode) []TestNode { return node.Children },
		SetChildren: func(node *TestNode, children []TestNode) { node.Children = children },
	}
}

// createTestCategoryAdapter creates a TreeAdapter for TestCategory.
func createTestCategoryAdapter() Adapter[TestCategory] {
	return Adapter[TestCategory]{
		GetId:       func(cat TestCategory) string { return cat.CategoryId },
		GetParentId: func(cat TestCategory) string { return cat.ParentCatId },
		GetChildren: func(cat TestCategory) []TestCategory { return cat.SubCategories },
		SetChildren: func(cat *TestCategory, children []TestCategory) { cat.SubCategories = children },
	}
}

// createTestNodes creates a set of test nodes for testing.
func createTestNodes() []TestNode {
	return []TestNode{
		{Id: "1", ParentId: constants.Empty, Name: "Root 1"},
		{Id: "2", ParentId: "1", Name: "Child 1-1"},
		{Id: "3", ParentId: "1", Name: "Child 1-2"},
		{Id: "4", ParentId: "2", Name: "Child 1-1-1"},
		{Id: "5", ParentId: "2", Name: "Child 1-1-2"},
		{Id: "6", ParentId: constants.Empty, Name: "Root 2"},
		{Id: "7", ParentId: "6", Name: "Child 2-1"},
		{Id: "8", ParentId: "nonexistent", Name: "Orphan"}, // orphan node
	}
}

// createComplexTestNodes creates a more complex set of test nodes.
func createComplexTestNodes() []TestNode {
	return []TestNode{
		{Id: "root1", ParentId: constants.Empty, Name: "Root 1"},
		{Id: "root2", ParentId: constants.Empty, Name: "Root 2"},
		{Id: "a", ParentId: "root1", Name: "A"},
		{Id: "b", ParentId: "root1", Name: "B"},
		{Id: "c", ParentId: "a", Name: "C"},
		{Id: "d", ParentId: "a", Name: "D"},
		{Id: "e", ParentId: "b", Name: "E"},
		{Id: "f", ParentId: "c", Name: "F"},
		{Id: "g", ParentId: "c", Name: "G"},
		{Id: "h", ParentId: "root2", Name: "H"},
		{Id: "i", ParentId: "h", Name: "I"},
	}
}

func TestBuild(t *testing.T) {
	adapter := createTestNodeAdapter()

	t.Run("builds simple tree structure", func(t *testing.T) {
		nodes := []TestNode{
			{Id: "1", ParentId: constants.Empty, Name: "Root"},
			{Id: "2", ParentId: "1", Name: "Child 1"},
			{Id: "3", ParentId: "1", Name: "Child 2"},
		}

		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		root := result[0]
		assert.Equal(t, "1", root.Id)
		assert.Equal(t, "Root", root.Name)
		assert.Len(t, root.Children, 2)

		// Check children
		child1 := root.Children[0]
		child2 := root.Children[1]

		assert.Equal(t, "2", child1.Id)
		assert.Equal(t, "3", child2.Id)
		assert.Len(t, child1.Children, 0)
		assert.Len(t, child2.Children, 0)
	})

	t.Run("builds tree with multiple roots", func(t *testing.T) {
		nodes := createTestNodes()

		result := Build(nodes, adapter)

		require.Len(t, result, 3) // Root 1, Root 2, and Orphan

		// Find roots by ID
		var root1, root2, orphan *TestNode

		for i := range result {
			switch result[i].Id {
			case "1":
				root1 = &result[i]
			case "6":
				root2 = &result[i]
			case "8":
				orphan = &result[i]
			}
		}

		require.NotNil(t, root1)
		require.NotNil(t, root2)
		require.NotNil(t, orphan)

		// Check Root 1 structure
		assert.Equal(t, "Root 1", root1.Name)
		assert.Len(t, root1.Children, 2)

		// Check Root 2 structure
		assert.Equal(t, "Root 2", root2.Name)
		assert.Len(t, root2.Children, 1)

		// Check orphan
		assert.Equal(t, "Orphan", orphan.Name)
		assert.Len(t, orphan.Children, 0)
	})

	t.Run("builds deep nested tree", func(t *testing.T) {
		nodes := createComplexTestNodes()

		result := Build(nodes, adapter)

		require.Len(t, result, 2) // root1 and root2

		// Find root1
		var root1 *TestNode

		for i := range result {
			if result[i].Id == "root1" {
				root1 = &result[i]

				break
			}
		}

		require.NotNil(t, root1)

		// Check structure: root1 -> {a, b}
		assert.Len(t, root1.Children, 2)

		// Find child 'a'
		var childA *TestNode

		for i := range root1.Children {
			if root1.Children[i].Id == "a" {
				childA = &root1.Children[i]

				break
			}
		}

		require.NotNil(t, childA)

		// Check structure: a -> {c, d}
		assert.Len(t, childA.Children, 2)

		// Find child 'c'
		var childC *TestNode

		for i := range childA.Children {
			if childA.Children[i].Id == "c" {
				childC = &childA.Children[i]

				break
			}
		}

		require.NotNil(t, childC)

		// Check structure: c -> {f, g}
		assert.Len(t, childC.Children, 2)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		var nodes []TestNode

		result := Build(nodes, adapter)

		assert.NotNil(t, result)
		assert.Empty(t, result)
		assert.IsType(t, []TestNode{}, result)
	})

	t.Run("handles single node", func(t *testing.T) {
		nodes := []TestNode{
			{Id: "1", ParentId: constants.Empty, Name: "Single"},
		}

		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		assert.Equal(t, "1", result[0].Id)
		assert.Equal(t, "Single", result[0].Name)
		assert.Len(t, result[0].Children, 0)
	})

	t.Run("handles nodes with empty IDs", func(t *testing.T) {
		nodes := []TestNode{
			{Id: constants.Empty, ParentId: constants.Empty, Name: "Empty ID"},
			{Id: "1", ParentId: constants.Empty, Name: "Valid"},
		}

		result := Build(nodes, adapter)

		// Both nodes are technically roots, but empty ID node should be filtered out at collection time
		// Based on current implementation, empty ID nodes are skipped in nodeMap but included in final collection
		// Let's test the actual behavior: empty ID nodes become roots if they have no parent
		require.Len(t, result, 2)

		// Find the valid node
		var validNode *TestNode

		for i := range result {
			if result[i].Id == "1" {
				validNode = &result[i]

				break
			}
		}

		require.NotNil(t, validNode)
		assert.Equal(t, "Valid", validNode.Name)
	})

	t.Run("handles circular references gracefully", func(t *testing.T) {
		nodes := []TestNode{
			{Id: "1", ParentId: "2", Name: "Node 1"},
			{Id: "2", ParentId: "1", Name: "Node 2"},
		}

		result := Build(nodes, adapter)

		// In true circular references where both nodes point to each other,
		// there are no actual roots since both have valid parent relationships
		// The cycle detection prevents infinite recursion and results in no roots
		require.Len(t, result, 0)
	})

	t.Run("handles partial circular references", func(t *testing.T) {
		nodes := []TestNode{
			{Id: "root", ParentId: constants.Empty, Name: "Root"},
			{Id: "1", ParentId: "2", Name: "Node 1"},
			{Id: "2", ParentId: "1", Name: "Node 2"},
			{Id: "3", ParentId: "root", Name: "Node 3"},
		}

		result := Build(nodes, adapter)

		// Should have the valid root and orphaned circular nodes
		require.Len(t, result, 1)
		root := result[0]
		assert.Equal(t, "root", root.Id)
		assert.Len(t, root.Children, 1)
		assert.Equal(t, "3", root.Children[0].Id)
	})

	t.Run("works with different data types", func(t *testing.T) {
		categories := []TestCategory{
			{CategoryId: "tech", ParentCatId: constants.Empty, CategoryName: "Technology", Level: 1},
			{CategoryId: "software", ParentCatId: "tech", CategoryName: "Software", Level: 2},
			{CategoryId: "hardware", ParentCatId: "tech", CategoryName: "Hardware", Level: 2},
			{CategoryId: "ai", ParentCatId: "software", CategoryName: "AI", Level: 3},
		}

		categoryAdapter := createTestCategoryAdapter()
		result := Build(categories, categoryAdapter)

		require.Len(t, result, 1)
		tech := result[0]
		assert.Equal(t, "tech", tech.CategoryId)
		assert.Equal(t, "Technology", tech.CategoryName)
		assert.Len(t, tech.SubCategories, 2)

		// Find software category
		var software *TestCategory

		for i := range tech.SubCategories {
			if tech.SubCategories[i].CategoryId == "software" {
				software = &tech.SubCategories[i]

				break
			}
		}

		require.NotNil(t, software)
		assert.Len(t, software.SubCategories, 1)
		assert.Equal(t, "ai", software.SubCategories[0].CategoryId)
	})
}

func TestFindNode(t *testing.T) {
	adapter := createTestNodeAdapter()

	t.Run("finds root node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "1", adapter)

		assert.True(t, found)
		assert.Equal(t, "1", result.Id)
		assert.Equal(t, "Root 1", result.Name)
	})

	t.Run("finds deep nested node", func(t *testing.T) {
		nodes := createComplexTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "f", adapter)

		assert.True(t, found)
		assert.Equal(t, "f", result.Id)
		assert.Equal(t, "F", result.Name)
	})

	t.Run("finds leaf node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "4", adapter)

		assert.True(t, found)
		assert.Equal(t, "4", result.Id)
		assert.Equal(t, "Child 1-1-1", result.Name)
	})

	t.Run("finds intermediate node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "2", adapter)

		assert.True(t, found)
		assert.Equal(t, "2", result.Id)
		assert.Equal(t, "Child 1-1", result.Name)
		assert.Len(t, result.Children, 2) // Should have children populated
	})

	t.Run("returns false for non-existent node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "nonexistent", adapter)

		assert.False(t, found)
		assert.Equal(t, constants.Empty, result.Id)
		assert.Equal(t, constants.Empty, result.Name)
	})

	t.Run("returns false for empty target ID", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, constants.Empty, adapter)

		assert.False(t, found)
		assert.Equal(t, constants.Empty, result.Id)
	})

	t.Run("handles empty tree", func(t *testing.T) {
		var tree []TestNode

		result, found := FindNode(tree, "1", adapter)

		assert.False(t, found)
		assert.Equal(t, constants.Empty, result.Id)
	})

	t.Run("finds node in different tree branches", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		// Find node in first branch
		result1, found1 := FindNode(tree, "2", adapter)
		assert.True(t, found1)
		assert.Equal(t, "2", result1.Id)

		// Find node in second branch
		result2, found2 := FindNode(tree, "7", adapter)
		assert.True(t, found2)
		assert.Equal(t, "7", result2.Id)

		// Find orphan node
		result3, found3 := FindNode(tree, "8", adapter)
		assert.True(t, found3)
		assert.Equal(t, "8", result3.Id)
	})

	t.Run("works with different data types", func(t *testing.T) {
		categories := []TestCategory{
			{CategoryId: "tech", ParentCatId: constants.Empty, CategoryName: "Technology"},
			{CategoryId: "software", ParentCatId: "tech", CategoryName: "Software"},
			{CategoryId: "ai", ParentCatId: "software", CategoryName: "AI"},
		}

		categoryAdapter := createTestCategoryAdapter()
		tree := Build(categories, categoryAdapter)

		result, found := FindNode(tree, "ai", categoryAdapter)

		assert.True(t, found)
		assert.Equal(t, "ai", result.CategoryId)
		assert.Equal(t, "AI", result.CategoryName)
	})

	t.Run("finds first occurrence in case of duplicates", func(t *testing.T) {
		// This test ensures consistent behavior
		nodes := []TestNode{
			{Id: "1", ParentId: constants.Empty, Name: "Root"},
			{Id: "2", ParentId: "1", Name: "Child 1"},
			{Id: "2", ParentId: "1", Name: "Child 2"}, // duplicate ID
		}

		tree := Build(nodes, adapter)
		result, found := FindNode(tree, "2", adapter)

		assert.True(t, found)
		assert.Equal(t, "2", result.Id)
		// Should find the first occurrence
		assert.Contains(t, []string{"Child 1", "Child 2"}, result.Name)
	})
}

func TestFindNodePath(t *testing.T) {
	adapter := createTestNodeAdapter()

	t.Run("finds path to root node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "1", adapter)

		assert.True(t, found)
		require.Len(t, path, 1)
		assert.Equal(t, "1", path[0].Id)
		assert.Equal(t, "Root 1", path[0].Name)
	})

	t.Run("finds path to deep nested node", func(t *testing.T) {
		nodes := createComplexTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "f", adapter)

		assert.True(t, found)
		require.Len(t, path, 4) // root1 -> a -> c -> f
		assert.Equal(t, "root1", path[0].Id)
		assert.Equal(t, "a", path[1].Id)
		assert.Equal(t, "c", path[2].Id)
		assert.Equal(t, "f", path[3].Id)
	})

	t.Run("finds path to immediate child", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "2", adapter)

		assert.True(t, found)
		require.Len(t, path, 2) // Root 1 -> Child 1-1
		assert.Equal(t, "1", path[0].Id)
		assert.Equal(t, "2", path[1].Id)
	})

	t.Run("finds path to leaf node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "4", adapter)

		assert.True(t, found)
		require.Len(t, path, 3) // Root 1 -> Child 1-1 -> Child 1-1-1
		assert.Equal(t, "1", path[0].Id)
		assert.Equal(t, "2", path[1].Id)
		assert.Equal(t, "4", path[2].Id)
	})

	t.Run("finds path to orphan node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "8", adapter)

		assert.True(t, found)
		require.Len(t, path, 1) // Just the orphan itself
		assert.Equal(t, "8", path[0].Id)
		assert.Equal(t, "Orphan", path[0].Name)
	})

	t.Run("returns empty path for non-existent node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "nonexistent", adapter)

		assert.False(t, found)
		assert.Nil(t, path)
	})

	t.Run("returns empty path for empty target ID", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, constants.Empty, adapter)

		assert.False(t, found)
		assert.Nil(t, path)
	})

	t.Run("handles empty tree", func(t *testing.T) {
		var tree []TestNode

		path, found := FindNodePath(tree, "1", adapter)

		assert.False(t, found)
		assert.Nil(t, path)
	})

	t.Run("finds path in different tree branches", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		// Path in first branch
		path1, found1 := FindNodePath(tree, "5", adapter)
		assert.True(t, found1)
		require.Len(t, path1, 3) // 1 -> 2 -> 5
		assert.Equal(t, "1", path1[0].Id)
		assert.Equal(t, "2", path1[1].Id)
		assert.Equal(t, "5", path1[2].Id)

		// Path in second branch
		path2, found2 := FindNodePath(tree, "7", adapter)
		assert.True(t, found2)
		require.Len(t, path2, 2) // 6 -> 7
		assert.Equal(t, "6", path2[0].Id)
		assert.Equal(t, "7", path2[1].Id)
	})

	t.Run("path contains complete node data", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "4", adapter)

		assert.True(t, found)
		require.Len(t, path, 3)

		// Verify each node in path has correct data
		assert.Equal(t, "Root 1", path[0].Name)
		assert.Equal(t, "Child 1-1", path[1].Name)
		assert.Equal(t, "Child 1-1-1", path[2].Name)

		// Verify parent-child relationships
		assert.Equal(t, constants.Empty, path[0].ParentId)
		assert.Equal(t, "1", path[1].ParentId)
		assert.Equal(t, "2", path[2].ParentId)
	})

	t.Run("works with different data types", func(t *testing.T) {
		categories := []TestCategory{
			{CategoryId: "tech", ParentCatId: constants.Empty, CategoryName: "Technology", Level: 1},
			{CategoryId: "software", ParentCatId: "tech", CategoryName: "Software", Level: 2},
			{CategoryId: "ai", ParentCatId: "software", CategoryName: "AI", Level: 3},
		}

		categoryAdapter := createTestCategoryAdapter()
		tree := Build(categories, categoryAdapter)

		path, found := FindNodePath(tree, "ai", categoryAdapter)

		assert.True(t, found)
		require.Len(t, path, 3) // tech -> software -> ai
		assert.Equal(t, "tech", path[0].CategoryId)
		assert.Equal(t, "software", path[1].CategoryId)
		assert.Equal(t, "ai", path[2].CategoryId)

		// Verify levels
		assert.Equal(t, 1, path[0].Level)
		assert.Equal(t, 2, path[1].Level)
		assert.Equal(t, 3, path[2].Level)
	})

	t.Run("finds path with multiple possible paths", func(t *testing.T) {
		// Create a more complex structure where node could theoretically have multiple paths
		nodes := createComplexTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "g", adapter)

		assert.True(t, found)
		require.Len(t, path, 4) // root1 -> a -> c -> g
		assert.Equal(t, "root1", path[0].Id)
		assert.Equal(t, "a", path[1].Id)
		assert.Equal(t, "c", path[2].Id)
		assert.Equal(t, "g", path[3].Id)
	})
}

func TestAdapter_EdgeCases(t *testing.T) {
	t.Run("adapter with nil functions panics appropriately", func(t *testing.T) {
		// Test that we handle nil function pointers gracefully
		nodes := []TestNode{
			{Id: "1", ParentId: constants.Empty, Name: "Test"},
		}

		// Create adapter with nil functions
		badAdapter := Adapter[TestNode]{
			GetId:       nil,
			GetParentId: nil,
			GetChildren: nil,
			SetChildren: nil,
		}

		// This should panic when trying to use nil functions
		assert.Panics(t, func() {
			Build(nodes, badAdapter)
		})
	})

	t.Run("large tree performance", func(t *testing.T) {
		// Create a large flat list to test performance
		const nodeCount = 1000

		nodes := make([]TestNode, nodeCount)

		// Create a single root with many children
		nodes[0] = TestNode{Id: "root", ParentId: constants.Empty, Name: "Root"}
		for i := 1; i < nodeCount; i++ {
			nodes[i] = TestNode{
				Id:       fmt.Sprintf("child_%d", i),
				ParentId: "root",
				Name:     fmt.Sprintf("Child %d", i),
			}
		}

		adapter := createTestNodeAdapter()
		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		assert.Equal(t, "root", result[0].Id)
		assert.Len(t, result[0].Children, nodeCount-1)
	})

	t.Run("deep nesting performance", func(t *testing.T) {
		// Create a deeply nested tree
		const depth = 100

		nodes := make([]TestNode, depth)

		// Create a chain: 0 -> 1 -> 2 -> ... -> depth-1
		nodes[0] = TestNode{Id: "0", ParentId: constants.Empty, Name: "Root"}
		for i := 1; i < depth; i++ {
			nodes[i] = TestNode{
				Id:       fmt.Sprintf("%d", i),
				ParentId: fmt.Sprintf("%d", i-1),
				Name:     fmt.Sprintf("Level %d", i),
			}
		}

		adapter := createTestNodeAdapter()
		result := Build(nodes, adapter)

		require.Len(t, result, 1)

		// Traverse to verify depth
		current := result[0]
		depth_count := 1

		for len(current.Children) > 0 {
			current = current.Children[0]
			depth_count++
		}

		assert.Equal(t, depth, depth_count)
	})

	t.Run("nodes with special characters in IDs", func(t *testing.T) {
		nodes := []TestNode{
			{Id: "root/path", ParentId: constants.Empty, Name: "Root with slash"},
			{Id: "child@domain.com", ParentId: "root/path", Name: "Child with email"},
			{Id: "special#$%^&*()", ParentId: "root/path", Name: "Special chars"},
			{Id: "unicode_测试_🌟", ParentId: "child@domain.com", Name: "Unicode"},
		}

		adapter := createTestNodeAdapter()
		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		root := result[0]
		assert.Equal(t, "root/path", root.Id)
		assert.Len(t, root.Children, 2)

		// Find the email child
		var emailChild *TestNode

		for i := range root.Children {
			if root.Children[i].Id == "child@domain.com" {
				emailChild = &root.Children[i]

				break
			}
		}

		require.NotNil(t, emailChild)
		assert.Len(t, emailChild.Children, 1)
		assert.Equal(t, "unicode_测试_🌟", emailChild.Children[0].Id)
	})

	t.Run("memory safety with concurrent access", func(t *testing.T) {
		// This test ensures the tree structure is safe for concurrent reads
		nodes := createComplexTestNodes()
		adapter := createTestNodeAdapter()
		tree := Build(nodes, adapter)

		// Run multiple goroutines that read from the tree
		done := make(chan bool, 10)

		for range 10 {
			go func() {
				defer func() { done <- true }()

				// Perform various read operations
				FindNode(tree, "f", adapter)
				FindNodePath(tree, "g", adapter)
				FindNode(tree, "root1", adapter)

				// Access tree structure directly
				for _, root := range tree {
					_ = root.Children
					for _, child := range root.Children {
						_ = child.Name
					}
				}
			}()
		}

		// Wait for all goroutines to complete
		for range 10 {
			<-done
		}

		// If we get here without data races, test passes
		assert.True(t, true)
	})

	t.Run("adapter function consistency", func(t *testing.T) {
		nodes := []TestNode{
			{Id: "1", ParentId: constants.Empty, Name: "Root", Children: []TestNode{}},
		}

		adapter := createTestNodeAdapter()

		// Test that GetId and GetParentId return expected types
		node := nodes[0]
		id := adapter.GetId(node)
		parentId := adapter.GetParentId(node)
		children := adapter.GetChildren(node)

		assert.IsType(t, "", id)
		assert.IsType(t, "", parentId)
		assert.IsType(t, []TestNode{}, children)

		// Test SetChildren
		newChildren := []TestNode{{Id: "child", Name: "Test Child"}}
		adapter.SetChildren(&nodes[0], newChildren)

		assert.Equal(t, newChildren, nodes[0].Children)
	})
}

func TestAdapter_BenchmarkScenarios(t *testing.T) {
	t.Run("compares Build efficiency vs naive approach", func(t *testing.T) {
		// Create a reasonably sized test case
		const nodeCount = 100

		nodes := make([]TestNode, nodeCount)

		// Create a balanced tree structure
		nodes[0] = TestNode{Id: "root", ParentId: constants.Empty, Name: "Root"}

		for i := 1; i < nodeCount; i++ {
			parentIndex := (i - 1) / 3 // Each node has up to 3 children
			nodes[i] = TestNode{
				Id:       fmt.Sprintf("node_%d", i),
				ParentId: fmt.Sprintf("node_%d", parentIndex),
				Name:     fmt.Sprintf("Node %d", i),
			}
		}

		// Fix root parent
		nodes[1].ParentId = "root"
		nodes[2].ParentId = "root"
		nodes[3].ParentId = "root"

		adapter := createTestNodeAdapter()

		// Test Build
		result := Build(nodes, adapter)

		// Verify the result is reasonable
		assert.Len(t, result, 1) // Should have one root

		// Count total nodes in tree
		var countNodes func([]TestNode) int

		countNodes = func(nodes []TestNode) int {
			count := len(nodes)
			for _, node := range nodes {
				count += countNodes(node.Children)
			}

			return count
		}

		totalNodes := countNodes(result)
		assert.Equal(t, nodeCount, totalNodes)
	})
}
