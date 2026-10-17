package tree

import "testing"

func TestBuild(t *testing.T) {
	type item struct {
		ID       uint64
		ParentID uint64
		Name     string
	}
	items := []item{{1, 0, "root"}, {2, 1, "child"}, {3, 2, "grand"}, {4, 99, "orphan"}}
	roots := Build(items, func(i item) uint64 { return i.ID }, func(i item) uint64 { return i.ParentID })
	if len(roots) != 2 {
		t.Fatalf("want 2 roots, got %d", len(roots))
	}
	if len(roots[0].Children) != 1 || roots[0].Children[0].Depth != 1 {
		t.Fatalf("unexpected tree: %#v", roots[0])
	}
	if roots[0].Children[0].Children[0].Depth != 2 {
		t.Fatalf("unexpected depth: %#v", roots[0].Children[0])
	}
}
