package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	defaultChunkBytes   = 1400
	defaultOverlapBytes = 200
)

type Chunk struct {
	ID          string
	SourcePath  string
	StartByte   int
	EndByte     int
	Language    string
	Content     string
	ContentHash string
}

type Chunker struct {
	ChunkBytes   int
	OverlapBytes int
}

func NewChunker(chunkBytes, overlapBytes int) Chunker {
	if chunkBytes <= 0 {
		chunkBytes = defaultChunkBytes
	}
	if overlapBytes < 0 || overlapBytes >= chunkBytes {
		overlapBytes = defaultOverlapBytes
	}

	return Chunker{ChunkBytes: chunkBytes, OverlapBytes: overlapBytes}
}

func (c Chunker) Chunk(path string, payload []byte) []Chunk {
	if len(payload) == 0 {
		return nil
	}

	chunks := make([]Chunk, 0, (len(payload)/c.ChunkBytes)+1)
	language := inferLanguage(path)

	for start := 0; start < len(payload); {
		end := start + c.ChunkBytes
		if end > len(payload) {
			end = len(payload)
		} else {
			end = bestSplit(payload, start, end)
		}

		segment := strings.TrimSpace(string(payload[start:end]))
		if segment != "" {
			contentHash := hashHex(segment)
			idSeed := fmt.Sprintf("%s:%d:%d:%s", filepath.Clean(path), start, end, contentHash)
			chunks = append(chunks, Chunk{
				ID:          hashHex(idSeed),
				SourcePath:  path,
				StartByte:   start,
				EndByte:     end,
				Language:    language,
				Content:     segment,
				ContentHash: contentHash,
			})
		}

		if end >= len(payload) {
			break
		}

		next := end - c.OverlapBytes
		if next <= start {
			next = end
		}
		start = next
	}

	return chunks
}

func bestSplit(payload []byte, start, targetEnd int) int {
	searchStart := start + (targetEnd-start)/2
	for i := targetEnd; i > searchStart; i-- {
		if payload[i-1] == '\n' {
			return i
		}
	}
	return targetEnd
}

func inferLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".ts", ".tsx", ".js", ".jsx", ".mjs":
		return "javascript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".sql":
		return "sql"
	case ".html":
		return "html"
	case ".css":
		return "css"
	default:
		return "text"
	}
}

func hashHex(data string) string {
	digest := sha256.Sum256([]byte(data))
	return hex.EncodeToString(digest[:])
}
