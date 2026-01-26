package treebuilder

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/constants"
)

type TestNode struct {
	ID       string     `json:"id"`
	ParentID string     `json:"parentID"`
	Name     string     `json:"name"`
	Children []TestNode `json:"children"`
}

type TestCategory struct {
	CategoryID    string         `json:"categoryID"`
	ParentCatID   string         `json:"parentCatId"`
	CategoryName  string         `json:"categoryName"`
	SubCategories []TestCategory `json:"subCategories"`
	Level         int            `json:"level"`
}

func createTestNodeAdapter() Adapter[TestNode] {
	return Adapter[TestNode]{
		GetID:       func(node TestNode) string { return node.ID },
		GetParentID: func(node TestNode) string { return node.ParentID },
		GetChildren: func(node TestNode) []TestNode { return node.Children },
		SetChildren: func(node *TestNode, children []TestNode) { node.Children = children },
	}
}

func createTestCategoryAdapter() Adapter[TestCategory] {
	return Adapter[TestCategory]{
		GetID:       func(cat TestCategory) string { return cat.CategoryID },
		GetParentID: func(cat TestCategory) string { return cat.ParentCatID },
		GetChildren: func(cat TestCategory) []TestCategory { return cat.SubCategories },
		SetChildren: func(cat *TestCategory, children []TestCategory) { cat.SubCategories = children },
	}
}

func createTestNodes() []TestNode {
	return []TestNode{
		{ID: "1", ParentID: constants.Empty, Name: "Root 1"},
		{ID: "2", ParentID: "1", Name: "Child 1-1"},
		{ID: "3", ParentID: "1", Name: "Child 1-2"},
		{ID: "4", ParentID: "2", Name: "Child 1-1-1"},
		{ID: "5", ParentID: "2", Name: "Child 1-1-2"},
		{ID: "6", ParentID: constants.Empty, Name: "Root 2"},
		{ID: "7", ParentID: "6", Name: "Child 2-1"},
		{ID: "8", ParentID: "nonexistent", Name: "Orphan"},
	}
}

func createComplexTestNodes() []TestNode {
	return []TestNode{
		{ID: "root1", ParentID: constants.Empty, Name: "Root 1"},
		{ID: "root2", ParentID: constants.Empty, Name: "Root 2"},
		{ID: "a", ParentID: "root1", Name: "A"},
		{ID: "b", ParentID: "root1", Name: "B"},
		{ID: "c", ParentID: "a", Name: "C"},
		{ID: "d", ParentID: "a", Name: "D"},
		{ID: "e", ParentID: "b", Name: "E"},
		{ID: "f", ParentID: "c", Name: "F"},
		{ID: "g", ParentID: "c", Name: "G"},
		{ID: "h", ParentID: "root2", Name: "H"},
		{ID: "i", ParentID: "h", Name: "I"},
	}
}

func findNodeByID(nodes []TestNode, id string) *TestNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}

func findCategoryByID(categories []TestCategory, id string) *TestCategory {
	for i := range categories {
		if categories[i].CategoryID == id {
			return &categories[i]
		}
	}
	return nil
}

func TestBuild(t *testing.T) {
	adapter := createTestNodeAdapter()

	t.Run("Builds simple tree structure", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "1", ParentID: constants.Empty, Name: "Root"},
			{ID: "2", ParentID: "1", Name: "Child 1"},
			{ID: "3", ParentID: "1", Name: "Child 2"},
		}

		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		root := result[0]
		assert.Equal(t, "1", root.ID)
		assert.Equal(t, "Root", root.Name)
		require.Len(t, root.Children, 2)

		assert.Equal(t, "2", root.Children[0].ID)
		assert.Equal(t, "3", root.Children[1].ID)
		assert.Empty(t, root.Children[0].Children)
		assert.Empty(t, root.Children[1].Children)
	})

	t.Run("Builds tree with multiple roots", func(t *testing.T) {
		nodes := createTestNodes()

		result := Build(nodes, adapter)

		require.Len(t, result, 3)

		root1 := findNodeByID(result, "1")
		root2 := findNodeByID(result, "6")
		orphan := findNodeByID(result, "8")

		require.NotNil(t, root1)
		require.NotNil(t, root2)
		require.NotNil(t, orphan)

		assert.Equal(t, "Root 1", root1.Name)
		assert.Len(t, root1.Children, 2)

		assert.Equal(t, "Root 2", root2.Name)
		assert.Len(t, root2.Children, 1)

		assert.Equal(t, "Orphan", orphan.Name)
		assert.Empty(t, orphan.Children)
	})

	t.Run("Builds deep nested tree", func(t *testing.T) {
		nodes := createComplexTestNodes()

		result := Build(nodes, adapter)

		require.Len(t, result, 2)

		root1 := findNodeByID(result, "root1")
		require.NotNil(t, root1)
		require.Len(t, root1.Children, 2)

		childA := findNodeByID(root1.Children, "a")
		require.NotNil(t, childA)
		require.Len(t, childA.Children, 2)

		childC := findNodeByID(childA.Children, "c")
		require.NotNil(t, childC)
		assert.Len(t, childC.Children, 2)
	})

	t.Run("Handles empty slice", func(t *testing.T) {
		var nodes []TestNode

		result := Build(nodes, adapter)

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("Handles single node", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "1", ParentID: constants.Empty, Name: "Single"},
		}

		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		assert.Equal(t, "1", result[0].ID)
		assert.Equal(t, "Single", result[0].Name)
		assert.Empty(t, result[0].Children)
	})

	t.Run("Handles nodes with empty IDs", func(t *testing.T) {
		nodes := []TestNode{
			{ID: constants.Empty, ParentID: constants.Empty, Name: "Empty ID"},
			{ID: "1", ParentID: constants.Empty, Name: "Valid"},
		}

		result := Build(nodes, adapter)

		require.Len(t, result, 2)

		validNode := findNodeByID(result, "1")
		require.NotNil(t, validNode)
		assert.Equal(t, "Valid", validNode.Name)
	})

	t.Run("Handles circular references gracefully", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "1", ParentID: "2", Name: "Node 1"},
			{ID: "2", ParentID: "1", Name: "Node 2"},
		}

		result := Build(nodes, adapter)

		require.Empty(t, result)
	})

	t.Run("Handles partial circular references", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "root", ParentID: constants.Empty, Name: "Root"},
			{ID: "1", ParentID: "2", Name: "Node 1"},
			{ID: "2", ParentID: "1", Name: "Node 2"},
			{ID: "3", ParentID: "root", Name: "Node 3"},
		}

		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		root := result[0]
		assert.Equal(t, "root", root.ID)
		require.Len(t, root.Children, 1)
		assert.Equal(t, "3", root.Children[0].ID)
	})

	t.Run("Works with different data types", func(t *testing.T) {
		categories := []TestCategory{
			{CategoryID: "tech", ParentCatID: constants.Empty, CategoryName: "Technology", Level: 1},
			{CategoryID: "software", ParentCatID: "tech", CategoryName: "Software", Level: 2},
			{CategoryID: "hardware", ParentCatID: "tech", CategoryName: "Hardware", Level: 2},
			{CategoryID: "ai", ParentCatID: "software", CategoryName: "AI", Level: 3},
		}

		categoryAdapter := createTestCategoryAdapter()
		result := Build(categories, categoryAdapter)

		require.Len(t, result, 1)
		tech := result[0]
		assert.Equal(t, "tech", tech.CategoryID)
		assert.Equal(t, "Technology", tech.CategoryName)
		require.Len(t, tech.SubCategories, 2)

		software := findCategoryByID(tech.SubCategories, "software")
		require.NotNil(t, software)
		require.Len(t, software.SubCategories, 1)
		assert.Equal(t, "ai", software.SubCategories[0].CategoryID)
	})
}

func TestFindNode(t *testing.T) {
	adapter := createTestNodeAdapter()

	t.Run("Finds root node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "1", adapter)

		assert.True(t, found)
		assert.Equal(t, "1", result.ID)
		assert.Equal(t, "Root 1", result.Name)
	})

	t.Run("Finds deep nested node", func(t *testing.T) {
		nodes := createComplexTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "f", adapter)

		assert.True(t, found)
		assert.Equal(t, "f", result.ID)
		assert.Equal(t, "F", result.Name)
	})

	t.Run("Finds leaf node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "4", adapter)

		assert.True(t, found)
		assert.Equal(t, "4", result.ID)
		assert.Equal(t, "Child 1-1-1", result.Name)
	})

	t.Run("Finds intermediate node with children", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "2", adapter)

		assert.True(t, found)
		assert.Equal(t, "2", result.ID)
		assert.Equal(t, "Child 1-1", result.Name)
		assert.Len(t, result.Children, 2)
	})

	t.Run("Returns false for non-existent node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result, found := FindNode(tree, "nonexistent", adapter)

		assert.False(t, found)
		assert.Equal(t, constants.Empty, result.ID)
	})

	t.Run("Returns false for empty target ID", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		_, found := FindNode(tree, constants.Empty, adapter)

		assert.False(t, found)
	})

	t.Run("Handles empty tree", func(t *testing.T) {
		var tree []TestNode

		_, found := FindNode(tree, "1", adapter)

		assert.False(t, found)
	})

	t.Run("Finds nodes in different branches", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		result1, found1 := FindNode(tree, "2", adapter)
		assert.True(t, found1)
		assert.Equal(t, "2", result1.ID)

		result2, found2 := FindNode(tree, "7", adapter)
		assert.True(t, found2)
		assert.Equal(t, "7", result2.ID)

		result3, found3 := FindNode(tree, "8", adapter)
		assert.True(t, found3)
		assert.Equal(t, "8", result3.ID)
	})

	t.Run("Works with different data types", func(t *testing.T) {
		categories := []TestCategory{
			{CategoryID: "tech", ParentCatID: constants.Empty, CategoryName: "Technology"},
			{CategoryID: "software", ParentCatID: "tech", CategoryName: "Software"},
			{CategoryID: "ai", ParentCatID: "software", CategoryName: "AI"},
		}

		categoryAdapter := createTestCategoryAdapter()
		tree := Build(categories, categoryAdapter)

		result, found := FindNode(tree, "ai", categoryAdapter)

		assert.True(t, found)
		assert.Equal(t, "ai", result.CategoryID)
		assert.Equal(t, "AI", result.CategoryName)
	})

	t.Run("Finds first occurrence with duplicate IDs", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "1", ParentID: constants.Empty, Name: "Root"},
			{ID: "2", ParentID: "1", Name: "Child 1"},
			{ID: "2", ParentID: "1", Name: "Child 2"},
		}

		tree := Build(nodes, adapter)
		result, found := FindNode(tree, "2", adapter)

		assert.True(t, found)
		assert.Equal(t, "2", result.ID)
		assert.Contains(t, []string{"Child 1", "Child 2"}, result.Name)
	})
}

func TestFindNodePath(t *testing.T) {
	adapter := createTestNodeAdapter()

	t.Run("Finds path to root node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "1", adapter)

		assert.True(t, found)
		require.Len(t, path, 1)
		assert.Equal(t, "1", path[0].ID)
		assert.Equal(t, "Root 1", path[0].Name)
	})

	t.Run("Finds path to deep nested node", func(t *testing.T) {
		nodes := createComplexTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "f", adapter)

		assert.True(t, found)
		require.Len(t, path, 4)
		assert.Equal(t, "root1", path[0].ID)
		assert.Equal(t, "a", path[1].ID)
		assert.Equal(t, "c", path[2].ID)
		assert.Equal(t, "f", path[3].ID)
	})

	t.Run("Finds path to immediate child", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "2", adapter)

		assert.True(t, found)
		require.Len(t, path, 2)
		assert.Equal(t, "1", path[0].ID)
		assert.Equal(t, "2", path[1].ID)
	})

	t.Run("Finds path to leaf node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "4", adapter)

		assert.True(t, found)
		require.Len(t, path, 3)
		assert.Equal(t, "1", path[0].ID)
		assert.Equal(t, "2", path[1].ID)
		assert.Equal(t, "4", path[2].ID)
	})

	t.Run("Finds path to orphan node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "8", adapter)

		assert.True(t, found)
		require.Len(t, path, 1)
		assert.Equal(t, "8", path[0].ID)
		assert.Equal(t, "Orphan", path[0].Name)
	})

	t.Run("Returns nil for non-existent node", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "nonexistent", adapter)

		assert.False(t, found)
		assert.Nil(t, path)
	})

	t.Run("Returns nil for empty target ID", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, constants.Empty, adapter)

		assert.False(t, found)
		assert.Nil(t, path)
	})

	t.Run("Handles empty tree", func(t *testing.T) {
		var tree []TestNode

		path, found := FindNodePath(tree, "1", adapter)

		assert.False(t, found)
		assert.Nil(t, path)
	})

	t.Run("Finds paths in different branches", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path1, found1 := FindNodePath(tree, "5", adapter)
		assert.True(t, found1)
		require.Len(t, path1, 3)
		assert.Equal(t, "1", path1[0].ID)
		assert.Equal(t, "2", path1[1].ID)
		assert.Equal(t, "5", path1[2].ID)

		path2, found2 := FindNodePath(tree, "7", adapter)
		assert.True(t, found2)
		require.Len(t, path2, 2)
		assert.Equal(t, "6", path2[0].ID)
		assert.Equal(t, "7", path2[1].ID)
	})

	t.Run("Path contains complete node data", func(t *testing.T) {
		nodes := createTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "4", adapter)

		assert.True(t, found)
		require.Len(t, path, 3)

		assert.Equal(t, "Root 1", path[0].Name)
		assert.Equal(t, "Child 1-1", path[1].Name)
		assert.Equal(t, "Child 1-1-1", path[2].Name)

		assert.Equal(t, constants.Empty, path[0].ParentID)
		assert.Equal(t, "1", path[1].ParentID)
		assert.Equal(t, "2", path[2].ParentID)
	})

	t.Run("Works with different data types", func(t *testing.T) {
		categories := []TestCategory{
			{CategoryID: "tech", ParentCatID: constants.Empty, CategoryName: "Technology", Level: 1},
			{CategoryID: "software", ParentCatID: "tech", CategoryName: "Software", Level: 2},
			{CategoryID: "ai", ParentCatID: "software", CategoryName: "AI", Level: 3},
		}

		categoryAdapter := createTestCategoryAdapter()
		tree := Build(categories, categoryAdapter)

		path, found := FindNodePath(tree, "ai", categoryAdapter)

		assert.True(t, found)
		require.Len(t, path, 3)
		assert.Equal(t, "tech", path[0].CategoryID)
		assert.Equal(t, "software", path[1].CategoryID)
		assert.Equal(t, "ai", path[2].CategoryID)

		assert.Equal(t, 1, path[0].Level)
		assert.Equal(t, 2, path[1].Level)
		assert.Equal(t, 3, path[2].Level)
	})

	t.Run("Finds correct path in complex tree", func(t *testing.T) {
		nodes := createComplexTestNodes()
		tree := Build(nodes, adapter)

		path, found := FindNodePath(tree, "g", adapter)

		assert.True(t, found)
		require.Len(t, path, 4)
		assert.Equal(t, "root1", path[0].ID)
		assert.Equal(t, "a", path[1].ID)
		assert.Equal(t, "c", path[2].ID)
		assert.Equal(t, "g", path[3].ID)
	})
}

func TestAdapter_EdgeCases(t *testing.T) {
	t.Run("Adapter with nil functions panics", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "1", ParentID: constants.Empty, Name: "Test"},
		}

		badAdapter := Adapter[TestNode]{}

		assert.Panics(t, func() {
			Build(nodes, badAdapter)
		})
	})

	t.Run("Large tree performance", func(t *testing.T) {
		const nodeCount = 1000

		nodes := make([]TestNode, nodeCount)
		nodes[0] = TestNode{ID: "root", ParentID: constants.Empty, Name: "Root"}
		for i := 1; i < nodeCount; i++ {
			nodes[i] = TestNode{
				ID:       fmt.Sprintf("child_%d", i),
				ParentID: "root",
				Name:     fmt.Sprintf("Child %d", i),
			}
		}

		adapter := createTestNodeAdapter()
		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		assert.Equal(t, "root", result[0].ID)
		assert.Len(t, result[0].Children, nodeCount-1)
	})

	t.Run("Deep nesting performance", func(t *testing.T) {
		const depth = 100

		nodes := make([]TestNode, depth)
		nodes[0] = TestNode{ID: "0", ParentID: constants.Empty, Name: "Root"}
		for i := 1; i < depth; i++ {
			nodes[i] = TestNode{
				ID:       fmt.Sprintf("%d", i),
				ParentID: fmt.Sprintf("%d", i-1),
				Name:     fmt.Sprintf("Level %d", i),
			}
		}

		adapter := createTestNodeAdapter()
		result := Build(nodes, adapter)

		require.Len(t, result, 1)

		current := result[0]
		depthCount := 1
		for len(current.Children) > 0 {
			current = current.Children[0]
			depthCount++
		}

		assert.Equal(t, depth, depthCount)
	})

	t.Run("Nodes with special characters in IDs", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "root/path", ParentID: constants.Empty, Name: "Root with slash"},
			{ID: "child@domain.com", ParentID: "root/path", Name: "Child with email"},
			{ID: "special#$%^&*()", ParentID: "root/path", Name: "Special chars"},
			{ID: "unicode_测试_🌟", ParentID: "child@domain.com", Name: "Unicode"},
		}

		adapter := createTestNodeAdapter()
		result := Build(nodes, adapter)

		require.Len(t, result, 1)
		root := result[0]
		assert.Equal(t, "root/path", root.ID)
		require.Len(t, root.Children, 2)

		emailChild := findNodeByID(root.Children, "child@domain.com")
		require.NotNil(t, emailChild)
		require.Len(t, emailChild.Children, 1)
		assert.Equal(t, "unicode_测试_🌟", emailChild.Children[0].ID)
	})

	t.Run("Concurrent read safety", func(t *testing.T) {
		nodes := createComplexTestNodes()
		adapter := createTestNodeAdapter()
		tree := Build(nodes, adapter)

		done := make(chan bool, 10)

		for range 10 {
			go func() {
				defer func() { done <- true }()

				FindNode(tree, "f", adapter)
				FindNodePath(tree, "g", adapter)
				FindNode(tree, "root1", adapter)

				for _, root := range tree {
					_ = root.Children
					for _, child := range root.Children {
						_ = child.Name
					}
				}
			}()
		}

		for range 10 {
			<-done
		}
	})

	t.Run("Adapter function consistency", func(t *testing.T) {
		nodes := []TestNode{
			{ID: "1", ParentID: constants.Empty, Name: "Root", Children: []TestNode{}},
		}

		adapter := createTestNodeAdapter()

		node := nodes[0]
		assert.Equal(t, "1", adapter.GetID(node))
		assert.Equal(t, constants.Empty, adapter.GetParentID(node))
		assert.Empty(t, adapter.GetChildren(node))

		newChildren := []TestNode{{ID: "child", Name: "Test Child"}}
		adapter.SetChildren(&nodes[0], newChildren)
		assert.Equal(t, newChildren, nodes[0].Children)
	})
}

func TestAdapter_BenchmarkScenarios(t *testing.T) {
	t.Run("Balanced tree structure", func(t *testing.T) {
		const nodeCount = 100

		nodes := make([]TestNode, nodeCount)
		nodes[0] = TestNode{ID: "root", ParentID: constants.Empty, Name: "Root"}

		for i := 1; i < nodeCount; i++ {
			parentIndex := (i - 1) / 3
			nodes[i] = TestNode{
				ID:       fmt.Sprintf("node_%d", i),
				ParentID: fmt.Sprintf("node_%d", parentIndex),
				Name:     fmt.Sprintf("Node %d", i),
			}
		}

		nodes[1].ParentID = "root"
		nodes[2].ParentID = "root"
		nodes[3].ParentID = "root"

		adapter := createTestNodeAdapter()
		result := Build(nodes, adapter)

		require.Len(t, result, 1)

		var countNodes func([]TestNode) int
		countNodes = func(nodes []TestNode) int {
			count := len(nodes)
			for _, node := range nodes {
				count += countNodes(node.Children)
			}
			return count
		}

		assert.Equal(t, nodeCount, countNodes(result))
	})
}
