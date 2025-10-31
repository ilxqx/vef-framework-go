package treebuilder

import (
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
)

// Adapter defines the adapter for building trees from arbitrary data.
type Adapter[T any] struct {
	GetId       func(T) string // extracts the unique identifier
	GetParentId func(T) string // extracts the parent identifier
	GetChildren func(T) []T    // extracts the children slice
	SetChildren func(*T, []T)  // sets the children slice (requires pointer)
}

// Build converts a flat slice of nodes into a tree structure using the provided adapter.
// Time complexity: O(n), Space complexity: O(n).
func Build[T any](nodes []T, adapter Adapter[T]) []T {
	if len(nodes) == 0 {
		return make([]T, 0)
	}

	// Build lookup maps - use pointers for modification
	nodeMap := make(map[string]*T, len(nodes))
	childrenMap := make(map[string][]*T)

	// First pass: build node map with pointers
	for i := range nodes {
		node := &nodes[i]
		if id := adapter.GetId(*node); id != constants.Empty {
			nodeMap[id] = node
		}
	}

	// Second pass: build parent-child relationships with pointers
	for i := range nodes {
		node := &nodes[i]
		if parentId := adapter.GetParentId(*node); parentId != constants.Empty {
			childrenMap[parentId] = append(childrenMap[parentId], node)
		}
	}

	// Third pass: recursively build tree structure with cycle detection
	visited := make(map[string]bool)

	var setChildrenRecursively func(*T)

	setChildrenRecursively = func(nodePtr *T) {
		id := adapter.GetId(*nodePtr)
		if id == constants.Empty {
			return
		}

		// Prevent infinite recursion with cycle detection
		if visited[id] {
			return
		}

		visited[id] = true

		if childrenPtrs, exists := childrenMap[id]; exists {
			// First, recursively set children for all child nodes
			for _, childPtr := range childrenPtrs {
				setChildrenRecursively(childPtr)
			}

			// Then convert to values and set children for this node
			children := make([]T, len(childrenPtrs))
			for j, childPtr := range childrenPtrs {
				children[j] = *childPtr
			}

			adapter.SetChildren(nodePtr, children)
		}

		// Reset visited for this node after processing its subtree
		visited[id] = false
	}

	// Apply recursive children setting to all nodes
	for i := range nodes {
		setChildrenRecursively(&nodes[i])
	}

	// Fourth pass: collect roots
	var roots []T

	for i := range nodes {
		node := &nodes[i]

		// Check if this is a root node
		parentId := adapter.GetParentId(*node)
		if parentId == constants.Empty {
			roots = append(roots, *node)
		} else {
			if _, exists := nodeMap[parentId]; !exists {
				// Parent doesn't exist, treat as root (orphan)
				roots = append(roots, *node)
			}
		}
	}

	return roots
}

// FindNode finds a node by Id in the tree structure built with Adapter.
func FindNode[T any](roots []T, targetId string, adapter Adapter[T]) (T, bool) {
	if targetId == constants.Empty {
		return lo.Empty[T](), false
	}

	return findNodeRecursive(roots, targetId, adapter)
}

func FindNodePath[T any](roots []T, targetId string, adapter Adapter[T]) ([]T, bool) {
	if targetId == constants.Empty {
		return nil, false
	}

	for _, root := range roots {
		if path, ok := findNodePathRecursive(root, targetId, nil, adapter); ok {
			return path, true
		}
	}

	return nil, false
}

// findNodeRecursive recursively searches for a node by ID using adapter.
func findNodeRecursive[T any](nodes []T, targetKey string, adapter Adapter[T]) (T, bool) {
	for _, node := range nodes {
		if id := adapter.GetId(node); id == targetKey {
			return node, true
		}

		if found, ok := findNodeRecursive(adapter.GetChildren(node), targetKey, adapter); ok {
			return found, true
		}
	}

	return lo.Empty[T](), false
}

// findNodePathRecursive recursively builds the path to a target node using adapter.
func findNodePathRecursive[T any](node T, targetKey string, currentPath []T, adapter Adapter[T]) ([]T, bool) {
	path := append(currentPath, node)

	if id := adapter.GetId(node); id == targetKey {
		return path, true
	}

	for _, child := range adapter.GetChildren(node) {
		if result, ok := findNodePathRecursive(child, targetKey, path, adapter); ok {
			return result, true
		}
	}

	return nil, false
}
