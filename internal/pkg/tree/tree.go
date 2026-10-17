// Package tree 把扁平列表组装成树。
package tree

// Node 是树节点。
type Node[T any] struct {
	Data     T
	Children []*Node[T]
	Depth    int
}

// Build 根据 ID/ParentID 组装树，保持输入顺序；父节点不存在或自引用时视为根。
func Build[T any](items []T, id func(T) uint64, parent func(T) uint64) []*Node[T] {
	nodes := make(map[uint64]*Node[T], len(items))
	order := make([]uint64, 0, len(items))
	for _, item := range items {
		nodes[id(item)] = &Node[T]{Data: item}
		order = append(order, id(item))
	}

	roots := make([]*Node[T], 0)
	for _, item := range items {
		node := nodes[id(item)]
		pid := parent(item)
		if pid == 0 || pid == id(item) {
			roots = append(roots, node)
			continue
		}
		if p, ok := nodes[pid]; ok {
			p.Children = append(p.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	var setDepth func(*Node[T], int)
	setDepth = func(n *Node[T], depth int) {
		n.Depth = depth
		for _, child := range n.Children {
			setDepth(child, depth+1)
		}
	}
	for _, root := range roots {
		setDepth(root, 0)
	}
	_ = order
	return roots
}

// Flatten 按前序遍历展开树。
func Flatten[T any](roots []*Node[T]) []*Node[T] {
	var out []*Node[T]
	var walk func(*Node[T])
	walk = func(n *Node[T]) {
		out = append(out, n)
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, root := range roots {
		walk(root)
	}
	return out
}
