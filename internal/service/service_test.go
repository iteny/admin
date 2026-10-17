package service

import (
	"strings"
	"testing"

	"cms/internal/model"
)

func TestSanitizeHTML(t *testing.T) {
	input := `<h2>标题</h2><p onclick="alert(1)">正文 <strong>加粗</strong></p><script>alert(1)</script><img src="javascript:alert(1)" onerror="alert(1)">`
	got := SanitizeHTML(input)
	if strings.Contains(got, "<script") || strings.Contains(got, "onerror") || strings.Contains(got, "onclick") {
		t.Fatalf("unsafe html survived: %s", got)
	}
	if !strings.Contains(got, "<h2>") || !strings.Contains(got, "<strong>") {
		t.Fatalf("allowed html was removed: %s", got)
	}
	if strings.Contains(got, "javascript:") {
		t.Fatalf("javascript url survived: %s", got)
	}
}

func TestPermissions(t *testing.T) {
	auth := NewAuthService(nil)
	editor := &model.User{Role: model.Role{Slug: "editor", Permissions: []model.Permission{{Code: "article:view"}, {Code: "article:create"}}}}
	perms := auth.Permissions(editor)
	if !perms["article:view"] || !perms["article:create"] || perms["*"] {
		t.Fatalf("unexpected editor permissions: %#v", perms)
	}
	super := &model.User{Role: model.Role{Slug: "super-admin"}}
	perms = auth.Permissions(super)
	if !perms["*"] {
		t.Fatalf("super admin should have wildcard: %#v", perms)
	}
	if got := auth.Permissions(nil); len(got) != 0 {
		t.Fatalf("nil user should have no permissions: %#v", got)
	}
}
