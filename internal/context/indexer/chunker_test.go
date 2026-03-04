package indexer

import (
	"testing"
)

func TestChunkerDeterministicChunkIDs(t *testing.T) {
	t.Parallel()

	chunker := NewChunker(64, 8)
	payload := []byte("line one\nline two\nline three\nline four\nline five\n")

	first := chunker.Chunk("a/b/file.go", payload)
	second := chunker.Chunk("a/b/file.go", payload)

	if len(first) != len(second) {
		t.Fatalf("chunk count mismatch: got %d and %d", len(first), len(second))
	}

	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("chunk ID mismatch at %d", i)
		}
	}
}

func TestChunkerUsesOverlap(t *testing.T) {
	t.Parallel()

	chunker := NewChunker(12, 4)
	chunks := chunker.Chunk("/tmp/demo.txt", []byte("abcdefghijklmnopqrstuvwx"))

	if len(chunks) < 2 {
		t.Fatalf("expected at least two chunks, got %d", len(chunks))
	}

	if chunks[1].StartByte >= chunks[0].EndByte {
		t.Fatalf("expected overlapping chunks, got start=%d end=%d", chunks[1].StartByte, chunks[0].EndByte)
	}
}
