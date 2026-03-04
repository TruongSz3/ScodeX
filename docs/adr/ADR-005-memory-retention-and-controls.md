# ADR-005: Memory Retention and User Controls

- Status: accepted
- Date: 2026-03-05

## Context
Need useful memory without uncontrolled drift or hidden state.

## Decision
Keep layered memory:
- session memory (short-term)
- project memory (long-term)

Expose inspect/edit/delete controls and provenance metadata.

## Consequences
- Pros: transparency and user control
- Cons: additional UI and schema complexity

## Alternatives
- Opaque automatic memory only
- No long-term memory
