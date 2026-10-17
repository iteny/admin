package str

import "testing"

func TestSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		enabled bool
		want    string
	}{
		{"english", "Hello Gin CMS", false, "hello-gin-cms"},
		{"symbols", "Hello,  World!", false, "hello-world"},
		{"chinese", "你好世界", true, "ni-hao-shi-jie"},
		{"chinese-disabled", "你好世界", false, "item"},
		{"empty", "", true, "item"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slug(tt.input, tt.enabled); got != tt.want {
				t.Fatalf("Slug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeSlug(t *testing.T) {
	if got := NormalizeSlug("  Hello -- World / Go  "); got != "hello-world-go" {
		t.Fatalf("unexpected slug: %q", got)
	}
}

func TestRandom(t *testing.T) {
	a := Random(12)
	b := Random(12)
	if len(a) != 12 || len(b) != 12 {
		t.Fatalf("unexpected lengths: %d, %d", len(a), len(b))
	}
	if a == b {
		t.Fatal("random values should differ")
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("你好世界", 2); got != "你好…" {
		t.Fatalf("unexpected truncate: %q", got)
	}
	if got := Truncate("abc", 5); got != "abc" {
		t.Fatalf("unexpected truncate: %q", got)
	}
}
