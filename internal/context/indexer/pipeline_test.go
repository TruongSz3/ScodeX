package indexer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sz3/scodex/internal/storage/sqlite"
)

type fakeRepo struct {
	batches     int
	chunkCount  int
	rebuildCall int
	err         error
}

func (f *fakeRepo) UpsertChunks(_ context.Context, chunks []sqlite.Chunk) error {
	if f.err != nil {
		return f.err
	}
	f.batches++
	f.chunkCount += len(chunks)
	return nil
}

func (f *fakeRepo) RebuildLexicalIndex(_ context.Context) error {
	f.rebuildCall++
	return f.err
}

func TestPipelineIndexesAndRebuilds(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("alpha beta gamma delta epsilon"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	repo := &fakeRepo{}
	pipeline, err := NewPipeline(repo, Options{ChunkBytes: 8, OverlapBytes: 2, BatchSize: 2})
	if err != nil {
		t.Fatalf("new pipeline failed: %v", err)
	}

	stats, err := pipeline.IndexPath(context.Background(), root)
	if err != nil {
		t.Fatalf("index failed: %v", err)
	}

	if stats.FilesIndexed != 1 {
		t.Fatalf("unexpected indexed file count: %d", stats.FilesIndexed)
	}
	if stats.ChunksIndexed == 0 {
		t.Fatal("expected indexed chunks")
	}
	if repo.chunkCount != stats.ChunksIndexed {
		t.Fatalf("repo chunk count mismatch: got=%d want=%d", repo.chunkCount, stats.ChunksIndexed)
	}
	if repo.rebuildCall != 1 {
		t.Fatalf("expected rebuild call once, got %d", repo.rebuildCall)
	}
}

func TestPipelinePropagatesRepoError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("abc def ghi"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	repoErr := errors.New("boom")
	pipeline, err := NewPipeline(&fakeRepo{err: repoErr}, Options{})
	if err != nil {
		t.Fatalf("new pipeline failed: %v", err)
	}

	_, err = pipeline.IndexPath(context.Background(), root)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
