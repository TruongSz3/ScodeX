package indexer

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestWalkTextFilesSkipsHiddenAndNodeModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "visible.go"), "package main")
	writeFile(t, filepath.Join(root, ".git", "ignored.txt"), "nope")
	writeFile(t, filepath.Join(root, "node_modules", "lib.js"), "nope")
	writeFile(t, filepath.Join(root, "nested", "readme.md"), "ok")
	writeFile(t, filepath.Join(root, "nested", "image.bin"), string([]byte{0, 1, 2, 3}))

	got := make([]string, 0, 2)
	err := WalkTextFiles(root, func(path string) error {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		got = append(got, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	sort.Strings(got)
	want := []string{"nested/readme.md", "visible.go"}
	if len(got) != len(want) {
		t.Fatalf("unexpected count: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected path at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
