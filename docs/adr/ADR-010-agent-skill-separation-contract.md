# ADR-010: Agent-Skill Separation Contract

- Status: accepted
- Date: 2026-03-05

## Context
Mixed role/procedure definitions create ambiguity and runtime drift.

## Decision
Enforce strict separation:
- Agent = WHO
- Skill = HOW

Workflow steps must bind both role and skill plus tool scope.

## Consequences
- Pros: clear responsibility boundaries and better validation
- Cons: stricter authoring requirements

## Alternatives
- Combined role+procedure documents without enforcement
