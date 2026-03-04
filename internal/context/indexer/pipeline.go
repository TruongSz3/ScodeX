package indexer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sz3/scodex/internal/storage/sqlite"
)

type ChunkSink interface {
	UpsertChunks(ctx context.Context, chunks []sqlite.Chunk) error
	RebuildLexicalIndex(ctx context.Context) error
}

type Options struct {
	ChunkBytes   int
	OverlapBytes int
	BatchSize    int
}

type Stats struct {
	FilesIndexed  int
	ChunksIndexed int
}

type Pipeline struct {
	chunker   Chunker
	repo      ChunkSink
	batchSize int
}

func NewPipeline(repo ChunkSink, opts Options) (*Pipeline, error) {
	if repo == nil {
		return nil, fmt.Errorf("indexer: repository is required")
	}

	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 64
	}

	return &Pipeline{
		chunker:   NewChunker(opts.ChunkBytes, opts.OverlapBytes),
		repo:      repo,
		batchSize: batchSize,
	}, nil
}

func (p *Pipeline) IndexPath(ctx context.Context, root string) (Stats, error) {
	stats := Stats{}
	pending := make([]sqlite.Chunk, 0, p.batchSize)

	flush := func() error {
		if len(pending) == 0 {
			return nil
		}
		if err := p.repo.UpsertChunks(ctx, pending); err != nil {
			return err
		}
		pending = pending[:0]
		return nil
	}

	err := WalkTextFiles(root, func(path string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("indexer: read file %q: %w", path, err)
		}

		chunks := p.chunker.Chunk(path, data)
		if len(chunks) == 0 {
			return nil
		}

		now := time.Now().UTC()
		for _, chunk := range chunks {
			pending = append(pending, sqlite.Chunk{
				ID:          chunk.ID,
				SourcePath:  chunk.SourcePath,
				StartByte:   chunk.StartByte,
				EndByte:     chunk.EndByte,
				Language:    chunk.Language,
				Content:     chunk.Content,
				ContentHash: chunk.ContentHash,
				UpdatedAt:   now,
			})
			if len(pending) >= p.batchSize {
				if err := flush(); err != nil {
					return fmt.Errorf("indexer: flush chunk batch: %w", err)
				}
			}
		}

		stats.FilesIndexed++
		stats.ChunksIndexed += len(chunks)
		return nil
	})
	if err != nil {
		return stats, err
	}

	if err := flush(); err != nil {
		return stats, fmt.Errorf("indexer: flush final chunk batch: %w", err)
	}

	if err := p.repo.RebuildLexicalIndex(ctx); err != nil {
		return stats, fmt.Errorf("indexer: rebuild lexical index: %w", err)
	}

	return stats, nil
}
