# ADR-006: Telemetry and Privacy Defaults

- Status: accepted
- Date: 2026-03-05

## Context
Need auditable runtime behavior while preserving local-first privacy expectations.

## Decision
- Local logs by default
- Optional telemetry opt-in
- Provider metadata limited to quota/rate-limit/token usage

## Consequences
- Pros: privacy-respecting defaults
- Cons: less out-of-box product analytics

## Alternatives
- Telemetry on by default
- Full provider payload logging
