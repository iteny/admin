package web

import (
	"io/fs"
	"testing"
)

func TestEmbeddedTemplates(t *testing.T) {
	count := 0
	err := fs.WalkDir(FS, "template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("no embedded templates")
	}
	sub, err := fs.Sub(FS, "template")
	if err != nil {
		t.Fatal(err)
	}
	adminCount := 0
	if err := fs.WalkDir(sub, "admin", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			adminCount++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("embedded templates: %d, admin: %d", count, adminCount)
}
