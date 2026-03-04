# ADR-003: Context Retrieval Architecture

- Status: accepted
- Date: 2026-03-05

## Context
Need fast local retrieval with optional semantic quality boost.

## Decision
Default to lexical retrieval (SQLite FTS5), with optional semantic rerank behind feature flag.

## Consequences
- Pros: predictable latency, lower external dependency
- Cons: semantic quality lower when rerank disabled

## Alternatives
- Semantic-only retrieval
- Full AST-only retrieval
