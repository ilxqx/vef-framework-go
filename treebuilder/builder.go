package treebuilder

import (
	"github.com/ilxqx/go-streams"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
)

type Adapter[T any] struct {
	GetId       func(T) string
	GetParentId func(T) string
	GetChildren func(T) []T
	SetChildren func(*T, []T)
}

func Build[T any](nodes []T, adapter Adapter[T]) []T {
	if len(nodes) == 0 {
		return make([]T, 0)
	}

	nodeMap := make(map[string]*T, len(nodes))
	childrenMap := make(map[string][]*T)

	for i := range nodes {
		node := &nodes[i]
		if id := adapter.GetId(*node); id != constants.Empty {
			nodeMap[id] = node
		}
	}

	for i := range nodes {
		node := &nodes[i]
		if parentId := adapter.GetParentId(*node); parentId != constants.Empty {
			childrenMap[parentId] = append(childrenMap[parentId], node)
		}
	}

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
			for _, childPtr := range childrenPtrs {
				setChildrenRecursively(childPtr)
			}

			children := streams.MapTo(
				streams.FromSlice(childrenPtrs),
				func(childPtr *T) T { return *childPtr },
			).Collect()

			adapter.SetChildren(nodePtr, children)
		}

		visited[id] = false
	}

	for i := range nodes {
		setChildrenRecursively(&nodes[i])
	}

	// Use streams.Filter to find root nodes (nodes without parent or with non-existent parent)
	roots := streams.FromSlice(nodes).Filter(func(node T) bool {
		parentId := adapter.GetParentId(node)
		if parentId == constants.Empty {
			return true
		}

		_, exists := nodeMap[parentId]

		return !exists
	}).Collect()

	return roots
}

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
