# Deployment Guide

## Scope
This guide covers local and cross-platform packaging for the Go runtime.

## Runtime Artifacts
- `agentd`: local daemon
- `agent`: CLI/TUI client
- Embedded local web UI assets served by `agentd`

## Local Prerequisites
- Go toolchain (target version aligned with blueprint)
- SQLite runtime support
- Provider credentials (BYOK API keys or OAuth setup)

## Configuration
- Provider config: base URL, auth mode, model map
- Runtime mode default: `ask`
- Provider switch mode default: `quota_failover`

## Startup Sequence
1. Start `agentd`
2. Start `agent` or open local web UI
3. Client performs `runtime.initialize` handshake
4. Create or resume session

## Packaging
- Cross-platform release (macOS/Linux/Windows)
- Signed artifacts and checksums
- Version compatibility matrix for client/daemon
- Rollback-capable updater

## Security Notes
- Bind web/API to localhost by default
- Store provider secrets in OS secure store when possible
- Keep telemetry opt-in by default
