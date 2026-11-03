package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/o98k-ok/aggregation/doc"
)

func TestPrivateCacheAndLogout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := cachePath()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(path) {
		t.Fatal("cache must not depend on workflow directory")
	}
	if err := saveCache(path, []doc.Entity{{Title: "private title"}}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("cache is not private")
	}
	if err := logout(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("logout retained cache")
	}
}
