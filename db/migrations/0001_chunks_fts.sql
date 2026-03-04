BEGIN;

CREATE TABLE IF NOT EXISTS chunks (
    rowid INTEGER PRIMARY KEY,
    chunk_id TEXT NOT NULL UNIQUE,
    source_path TEXT NOT NULL,
    start_byte INTEGER NOT NULL,
    end_byte INTEGER NOT NULL,
    language TEXT NOT NULL,
    content TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chunks_source_path ON chunks(source_path);
CREATE INDEX IF NOT EXISTS idx_chunks_updated_at ON chunks(updated_at);

CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
    chunk_id UNINDEXED,
    content,
    content='chunks',
    content_rowid='rowid'
);

CREATE TRIGGER IF NOT EXISTS chunks_ai AFTER INSERT ON chunks BEGIN
    INSERT INTO chunks_fts(rowid, chunk_id, content)
    VALUES (new.rowid, new.chunk_id, new.content);
END;

CREATE TRIGGER IF NOT EXISTS chunks_ad AFTER DELETE ON chunks BEGIN
    INSERT INTO chunks_fts(chunks_fts, rowid, chunk_id, content)
    VALUES('delete', old.rowid, old.chunk_id, old.content);
END;

CREATE TRIGGER IF NOT EXISTS chunks_au AFTER UPDATE ON chunks BEGIN
    INSERT INTO chunks_fts(chunks_fts, rowid, chunk_id, content)
    VALUES('delete', old.rowid, old.chunk_id, old.content);
    INSERT INTO chunks_fts(rowid, chunk_id, content)
    VALUES (new.rowid, new.chunk_id, new.content);
END;

COMMIT;
