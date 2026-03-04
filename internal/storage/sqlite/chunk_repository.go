package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrChunkIDRequired = errors.New("sqlite: chunk id is required")

type Chunk struct {
	ID          string
	SourcePath  string
	StartByte   int
	EndByte     int
	Language    string
	Content     string
	ContentHash string
	UpdatedAt   time.Time
}

type SearchHit struct {
	Chunk Chunk
	Score float64
}

type ChunkRepository struct {
	db *sql.DB
}

func NewChunkRepository(db *sql.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) UpsertChunks(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: start chunk upsert tx: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO chunks(
			chunk_id,
			source_path,
			start_byte,
			end_byte,
			language,
			content,
			content_hash,
			updated_at
		)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(chunk_id) DO UPDATE SET
			source_path = excluded.source_path,
			start_byte = excluded.start_byte,
			end_byte = excluded.end_byte,
			language = excluded.language,
			content = excluded.content,
			content_hash = excluded.content_hash,
			updated_at = excluded.updated_at
	`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("sqlite: prepare chunk upsert statement: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		if chunk.ID == "" {
			_ = tx.Rollback()
			return ErrChunkIDRequired
		}

		updatedAt := chunk.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = time.Now().UTC()
		}

		if _, err := stmt.ExecContext(
			ctx,
			chunk.ID,
			chunk.SourcePath,
			chunk.StartByte,
			chunk.EndByte,
			chunk.Language,
			chunk.Content,
			chunk.ContentHash,
			updatedAt.Format(time.RFC3339Nano),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("sqlite: upsert chunk %q: %w", chunk.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite: commit chunk upsert tx: %w", err)
	}

	return nil
}

func (r *ChunkRepository) RebuildLexicalIndex(ctx context.Context) error {
	queries := []string{
		"DELETE FROM chunks_fts;",
		"INSERT INTO chunks_fts(rowid, chunk_id, content) SELECT rowid, chunk_id, content FROM chunks;",
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("sqlite: rebuild lexical index query %q: %w", query, err)
		}
	}

	return nil
}

func (r *ChunkRepository) SearchLexical(ctx context.Context, query string, limit int) ([]SearchHit, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			c.chunk_id,
			c.source_path,
			c.start_byte,
			c.end_byte,
			c.language,
			c.content,
			c.content_hash,
			c.updated_at,
			bm25(chunks_fts) AS score
		FROM chunks_fts
		JOIN chunks c ON c.rowid = chunks_fts.rowid
		WHERE chunks_fts MATCH ?
		ORDER BY score
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("sqlite: lexical search: %w", err)
	}
	defer rows.Close()

	hits := make([]SearchHit, 0, limit)
	for rows.Next() {
		var (
			hit       SearchHit
			updatedAt string
		)

		if err := rows.Scan(
			&hit.Chunk.ID,
			&hit.Chunk.SourcePath,
			&hit.Chunk.StartByte,
			&hit.Chunk.EndByte,
			&hit.Chunk.Language,
			&hit.Chunk.Content,
			&hit.Chunk.ContentHash,
			&updatedAt,
			&hit.Score,
		); err != nil {
			return nil, fmt.Errorf("sqlite: scan lexical search row: %w", err)
		}

		parsed, err := time.Parse(time.RFC3339Nano, updatedAt)
		if err == nil {
			hit.Chunk.UpdatedAt = parsed
		}

		hits = append(hits, hit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: lexical search rows: %w", err)
	}

	return hits, nil
}
