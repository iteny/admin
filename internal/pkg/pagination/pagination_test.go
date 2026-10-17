package pagination

import "testing"

func TestQueryNormalize(t *testing.T) {
	q := Query{}
	q.Normalize()
	if q.Page != 1 || q.PageSize != 15 {
		t.Fatalf("unexpected defaults: %#v", q)
	}
	q = Query{Page: 2, PageSize: 1000}
	q.Normalize()
	if q.PageSize != 100 {
		t.Fatalf("page size should be capped: %d", q.PageSize)
	}
}

func TestNewAndPages(t *testing.T) {
	r := New([]int{1, 2}, 101, 3, 10)
	if r.TotalPages != 11 || r.Page != 3 || !r.HasPrev() || !r.HasNext() {
		t.Fatalf("unexpected result: %#v", r)
	}
	pages := r.Pages()
	if len(pages) == 0 || pages[0] <= 0 || pages[len(pages)-1] > r.TotalPages {
		t.Fatalf("invalid pages: %#v", pages)
	}
}

func TestNewView(t *testing.T) {
	v := NewView(1, 1, func(page int) string { return "/?page=" + itoa(page) })
	if v.HasPrev || v.HasNext || len(v.Links) != 1 || !v.Links[0].Active {
		t.Fatalf("unexpected view: %#v", v)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	out := ""
	for n > 0 {
		out = string(rune('0'+n%10)) + out
		n /= 10
	}
	return out
}
