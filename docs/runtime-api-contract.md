# Runtime API Contract

## Transport
- Implemented transport: local HTTP fallback daemon (`127.0.0.1:7777` by default)
- Bind safety: non-loopback binds are rejected unless `AGENTD_ALLOW_NON_LOOPBACK=true`
- Compatibility targets (planned): local IPC and stdio JSONL profile

## Mandatory Handshake
All mutating runtime endpoints currently require `X-Agent-Token` auth and a 2-step handshake:

1) `POST /runtime/initialize`
- Required JSON body:
  - `protocolVersion` (string, must be `"v1"`)
- Optional JSON body:
  - `clientName` (string)
- Success response includes:
  - `status: "ok"`
  - `phase: "initialize"`
  - `sessionId` (handshake session ID; required for step 2)
  - `protocolVersion: "v1"`

2) `POST /runtime/initialized`
- Required JSON body:
  - `sessionId` (string, must match `sessionId` returned by initialize)
- Success response includes:
  - `status: "ok"`
  - `phase: "initialized"`

Mutating operations before handshake completion return an error.

## Auth Requirements
- Header: `X-Agent-Token`
- Applied to current POST routes:
  - `POST /runtime/initialize`
  - `POST /runtime/initialized`
  - `POST /session/create`
- `GET /health` and `GET /ready` are unauthenticated.

## Core API Groups
- Runtime: `/runtime/initialize`, `/runtime/initialized`, `/ready`, `/health`
- Session: `/session/create` (create path implemented)
- Other groups in blueprint remain planned for later phases

## Current Error Contract (Implemented)
- `400 Bad Request`
  - invalid JSON body
  - missing/unsupported `protocolVersion`
  - missing `sessionId` on `/runtime/initialized`
- `401 Unauthorized`
  - missing or invalid `X-Agent-Token` on protected routes
- `409 Conflict`
  - `/runtime/initialized` called before initialize
  - `sessionId` mismatch during handshake
- `428 Precondition Required`
  - `/session/create` called before handshake completion

## Event Stream
Implemented baseline:
- `session.created` event published on successful session creation (best-effort)

Planned stream expansion:
- orchestrator, approvals, tool/workflow/team lifecycle, provider usage, safeguard rejections

## API Stability Rules
- Version all breaking changes
- Keep IDs and field names stable across CLI/TUI and web clients
- Emit explainable errors with remediation hints when possible
