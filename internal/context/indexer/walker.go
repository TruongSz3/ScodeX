package indexer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var allowedTextExt = map[string]struct{}{
	".c": {}, ".cc": {}, ".cpp": {}, ".css": {}, ".csv": {}, ".go": {}, ".h": {},
	".hpp": {}, ".html": {}, ".java": {}, ".js": {}, ".json": {}, ".jsx": {},
	".md": {}, ".mjs": {}, ".py": {}, ".rb": {}, ".rs": {}, ".sh": {},
	".sql": {}, ".toml": {}, ".ts": {}, ".tsx": {}, ".txt": {}, ".xml": {},
	".yaml": {}, ".yml": {},
}

func WalkTextFiles(root string, visitor func(path string) error) error {
	if root == "" {
		return fmt.Errorf("indexer: root path is required")
	}

	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		name := entry.Name()
		if entry.IsDir() {
			if shouldSkipDirectory(path, root, name) {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasPrefix(name, ".") {
			return nil
		}

		isText, err := isTextFile(path)
		if err != nil {
			return err
		}
		if !isText {
			return nil
		}

		return visitor(path)
	})
}

func shouldSkipDirectory(path, root, name string) bool {
	if path == root {
		return false
	}

	if strings.HasPrefix(name, ".") {
		return true
	}

	return name == "node_modules"
}

func isTextFile(path string) (bool, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if _, ok := allowedTextExt[ext]; ok {
		return true, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("indexer: open file %q: %w", path, err)
	}
	defer file.Close()

	buf := make([]byte, 4096)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("indexer: read file %q: %w", path, err)
	}

	if n == 0 {
		return true, nil
	}

	sample := buf[:n]
	if containsNull(sample) {
		return false, nil
	}

	return utf8.Valid(sample), nil
}

func containsNull(sample []byte) bool {
	for _, b := range sample {
		if b == 0 {
			return true
		}
	}
	return false
}
