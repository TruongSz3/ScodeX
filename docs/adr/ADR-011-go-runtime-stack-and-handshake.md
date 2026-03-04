# ADR-011: Go Runtime Stack and Handshake Contract

- Status: accepted
- Date: 2026-03-05

## Context
Need one practical stack for cross-platform local runtime with clear protocol lifecycle.

## Decision
- Implement runtime in Go
- Use typed local RPC and mandatory initialization handshake:
  - `runtime.initialize`
  - `runtime.initialized`

## Consequences
- Pros: strong portability, performance, and operational simplicity
- Cons: additional upfront protocol/schema discipline

## Alternatives
- JavaScript/TypeScript runtime first
- Protocol without explicit handshake
