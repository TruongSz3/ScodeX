# Codex Repository Audit

Ngày audit: 2026-03-03
Repo: `reference/codex/`
Phương pháp: đọc `reference/repomix-codex.xml`, README/docs/workflows/manifests, đối chiếu git metadata.

## Executive snapshot

- Codex là **local coding agent platform** của OpenAI, trọng tâm là CLI chạy local, không phải SaaS thuần.
- Repo là **monorepo Rust-first**; TS CLI cũ tồn tại nhưng Rust CLI là implementation được maintain chính.
- Core surface: CLI, TUI, `codex exec`, app-server JSON-RPC, MCP client/server, TypeScript SDK.
- Kiến trúc là **modular monolith trong monorepo**, tách theo crates và process boundaries; không có dấu microservices/Kubernetes.
- Security posture mạnh: sandbox theo OS, network policy proxy, cargo-deny/cargo-audit, security reporting qua Bugcrowd.
- CI/CD và release engineering rất mature: multi-OS/multi-arch builds, signed artifacts, notarization, trusted publish.
- Code structure tốt, naming nhất quán kiểu `codex-*`, separation of concerns tốt ở crate level.
- Test/lint coverage mạnh: unit/integration/snapshot, `cargo fmt`, `clippy`, `cargo shear`, codespell, Bazel/CI.
- Documentation sâu ở vài domain lớn như `app-server`, `network-proxy`, nhưng một phần docs chỉ trỏ ra docs external.
- Contribution flow chặt: external code contribution **by invitation only**, có PR template + CLA.
- Repo đang **active mạnh**: last commit ngày 2026-03-03; contributor count theo `git shortlog` là 467.
- License là **Apache-2.0**, phù hợp reuse thương mại/OSS.
- Dependency governance tốt hơn nhiều repo OSS thông thường, nhưng có **RUSTSEC ignores có chủ đích** nên vẫn có security debt được ghi nhận.

## 1. Features + Scope

### Findings

- Product chính: `Codex CLI` là coding agent chạy local trên máy người dùng.
- Core features:
  - Local CLI/TUI runtime.
  - `codex exec` cho automation/non-interactive use.
  - App-server dùng JSON-RPC để embed vào client khác.
  - MCP client support.
  - MCP server experimental.
  - IDE/desktop/web surfaces được repo root nhắc tới như các mặt trải nghiệm liên quan.
- Features phụ/experimental:
  - WebSocket transport của app-server là experimental/unsupported.
  - Realtime, dynamic tools, background terminal cleanup, một số app-server methods gated bởi `experimentalApi`.
  - `shell-tool-mcp` là experimental.
- Repo giải quyết problem gì:
  - Cho developer chạy coding agent local, có thể tích hợp vào terminal, IDE, app UI, hoặc client khác qua protocol.
- In-scope:
  - Local coding workflow.
  - Agent execution/sandbox/policy.
  - Embeddable app-server protocol.
  - MCP interoperability.
- Out-of-scope / not-ready:
  - Một số surface experimental chưa production-ready.
  - Repo không định vị là SaaS backend platform hay library thuần.
- Product type:
  - **Developer tool platform**, local-first, embeddable.
- Roadmap/backlog:
  - Không thấy file `ROADMAP` hay backlog public rõ ràng trong repo.
  - `docs/contributing.md` nói ưu tiên feature dựa trên community feedback, alignment với roadmap, consistency across Codex surfaces.

### Evidence

- `reference/codex/README.md`
- `reference/codex/codex-rs/README.md`
- `reference/codex/codex-rs/app-server/README.md`
- `reference/codex/shell-tool-mcp/README.md`
- `reference/codex/docs/contributing.md`
- `reference/codex/CHANGELOG.md`

### Assessment

- Scope rõ: local coding agent platform.
- Feature set core mạnh và có modular extension path.
- Roadmap/backlog visibility thấp.

## 2. Target Users

### Findings

- Primary users:
  - Developers dùng terminal.
  - Dev teams dùng ChatGPT Plus/Pro/Team/Edu/Enterprise plans.
  - Integrators muốn embed Codex qua app-server hoặc SDK.
- Secondary users:
  - IDE extension users.
  - Experimental MCP ecosystem users.
- Expected scale:
  - Không thấy số liệu active users, throughput target, concurrency target, SLO/SLA trong repo.
  - Có enterprise signal nhưng không có public sizing guidance.

### Evidence

- `reference/codex/README.md`
- `reference/codex/codex-rs/app-server/README.md`
- `reference/codex/sdk/typescript/README.md`

### Assessment

- Persona khá rõ.
- Scale định lượng còn mờ.

## 3. Architecture

### Findings

- Architecture style: **modular monolith** trong monorepo.
- Không có evidence cho microservices + Kubernetes deployment model.
- Ranh giới chính là crates/processes:
  - `core` cho business logic.
  - `cli` cho multi-tool CLI.
  - `tui` cho terminal UI.
  - `exec` cho headless automation.
  - `app-server` cho JSON-RPC embedding.
  - `network-proxy`, `linux-sandbox`, `windows-sandbox-rs`, `mcp-server`, `state`.
- App-server dùng JSON-RPC 2.0 semantics qua stdio và websocket; websocket bị đánh dấu experimental/unsupported.
- Không thấy architecture diagram chính thức trong repo docs đã đọc.

### Inferred diagram

```text
User / IDE / SDK
      |
      v
  codex CLI
   |   |   \
   |   |    +--> exec (headless)
   |   +-------> tui
   +-----------> app-server (JSON-RPC)
                    |
                    v
                 codex-core
                    |
   +----------------+----------------+
   | sandbox        | network proxy  |
   | state          | MCP            |
   | auth/config    | file search    |
```

### Evidence

- `reference/codex/codex-rs/README.md`
- `reference/codex/codex-rs/Cargo.toml`
- `reference/codex/codex-rs/app-server/README.md`

### Assessment

- Tổ chức kiến trúc hợp lý cho local runtime phức tạp.
- Thiếu diagram chính thức làm onboarding khó hơn mức cần thiết.

## 4. Tech Stack

### Findings

- Backend/runtime:
  - Rust 2024 edition.
  - Tokio, Axum, Reqwest, RMCP, OpenTelemetry, Sentry.
- Frontend/UI:
  - Terminal UI với Ratatui.
  - TypeScript SDK cho embedding.
  - Không thấy web frontend app lớn là phần chính của repo này.
- Database/state:
  - SQLite qua `sqlx` trong workspace/state stack.
- Cache:
  - Có local utility/cache crate và CI build caching; không thấy external cache kiểu Redis.
- Queue:
  - App-server dùng bounded queues giữa transport ingress / processing / outbound writes.
- Search:
  - Fuzzy file search, `ignore`, `nucleo`/matcher stack.
- Infra/build:
  - GitHub Actions, Bazel, Nix, devcontainer, Docker.
- Secrets/auth stack:
  - Keyring support và secrets-related crates tồn tại.

### Evidence

- `reference/codex/codex-rs/Cargo.toml`
- `reference/codex/codex-rs/README.md`
- `reference/codex/codex-rs/file-search/README.md`
- `reference/codex/codex-rs/app-server/README.md`
- `reference/codex/.devcontainer/`
- `reference/codex/flake.nix`
- `reference/codex/codex-cli/Dockerfile`

### Assessment

- Stack thiên native performance và cross-platform binaries.
- Không có dấu hiệu của Redis/Kafka/Elasticsearch/Postgres là core runtime dependency.

## 5. Code Structure

### Findings

- Root là monorepo; phần core code nằm ở `codex-rs/`.
- `codex-rs` là Cargo workspace rất lớn, chia theo domain rõ.
- Naming convention nhất quán: crates prefixed bằng `codex-`.
- Separation of concerns tốt:
  - `core/` business logic.
  - `cli/` entrypoint/tooling.
  - `tui/` UI.
  - `exec/` automation.
  - `app-server/` protocol surface.
  - `network-proxy/`, `state/`, `mcp-server/`, sandbox crates chuyên biệt.

### Evidence

- `reference/codex/codex-rs/README.md`
- `reference/codex/codex-rs/Cargo.toml`
- `reference/codex/AGENTS.md`

### Assessment

- Folder structure rõ, mature, dễ map trách nhiệm.
- Nhược điểm chính là số crate lớn => learning curve cao.

## 6. Code Quality

### Findings

- Tests:
  - Có unit tests.
  - Có integration tests.
  - Có snapshot tests, đặc biệt ở TUI.
  - Có test suites riêng cho app-server, core, API, SDK, shell-tool-mcp.
- Lint/format:
  - `cargo fmt`.
  - `clippy`.
  - `cargo shear`.
  - `codespell`.
- CI/CD:
  - `rust-ci.yml` chạy trên `main` + PR.
  - Matrix build nhiều OS/arch.
  - Có Bazel CI, SDK CI, shell-tool-mcp CI.

### Evidence

- `reference/codex/.github/workflows/rust-ci.yml`
- `reference/codex/.github/workflows/codespell.yml`
- `reference/codex/.github/workflows/bazel.yml`
- `reference/codex/.github/workflows/sdk.yml`
- `reference/codex/codex-rs/tui/src/**/snapshots/`
- `reference/codex/codex-rs/core/tests/`

### Assessment

- Formal quality gates rất mạnh.
- Đây là một trong các điểm mạnh nhất của repo.

## 7. Deployment & DevOps

### Findings

- Docker:
  - Có Dockerfile cho `codex-cli` và CI build helpers.
- Kubernetes:
  - Không thấy evidence K8s/Helm manifests.
- Release/distribution:
  - GitHub tag-based release workflow.
  - Multi-arch binaries cho macOS/Linux/Windows.
  - Linux artifact signing, macOS notarization, Windows signing.
  - npm publish/trusted publishing flows tồn tại.
- Environments:
  - Repo chủ yếu ship client/runtime artifacts; staging/production env app-hosted không phải concern chính.
  - Có alpha/beta/release tagging patterns, nhưng không thấy staging environment docs theo kiểu web service.
- Config management:
  - `config.toml` là cấu hình chính.
  - Có docs config và schema generation.
  - Dev environments qua Nix/devcontainer.

### Evidence

- `reference/codex/.github/workflows/rust-release.yml`
- `reference/codex/.github/actions/linux-code-sign/action.yml`
- `reference/codex/.github/actions/macos-code-sign/action.yml`
- `reference/codex/.github/actions/windows-code-sign/action.yml`
- `reference/codex/.devcontainer/`
- `reference/codex/flake.nix`
- `reference/codex/docs/config.md`

### Assessment

- DevOps/release discipline rất cao.
- Không phải repo Kubernetes-native.

## 8. Security

### Findings

- Auth mechanism:
  - ChatGPT sign-in hoặc API key.
  - App-server yêu cầu initialize handshake; auth/account methods có surface rõ.
- Isolation:
  - Sandbox theo OS: macOS, Linux, Windows.
  - `codex-network-proxy` enforce allow/deny policy, loopback clamps, unix socket restrictions.
- Rate limit:
  - Repo surfacing ChatGPT/OpenAI rate limits cho client (`account/rateLimits/read`, headers, snapshots).
  - Không thấy internal application-level rate limiter kiểu API gateway quota.
  - Overload control có: app-server dùng bounded queues và reject với `-32001` khi saturated.
- Validation:
  - App-server protocol docs rõ.
  - Experimental API gated qua `experimentalApi` capability.
  - JSON-RPC initialization rules chặt.
- Secret management:
  - Keyring store crate.
  - Sandbox secrets trên Windows dùng DPAPI-encrypted blobs.
  - Security disclosure qua Bugcrowd.

### Evidence

- `reference/codex/README.md`
- `reference/codex/SECURITY.md`
- `reference/codex/codex-rs/core/README.md`
- `reference/codex/codex-rs/network-proxy/README.md`
- `reference/codex/codex-rs/app-server/README.md`
- `reference/codex/codex-rs/windows-sandbox-rs/src/setup_orchestrator.rs`
- `reference/codex/codex-rs/keyring-store/`

### Assessment

- Security posture mạnh hơn mặt bằng OSS devtool.
- Gap chính: rate limiting là mostly upstream/account oriented, không phải generic gateway rate limiter.

## 9. Performance

### Findings

- Caching:
  - CI dùng sccache/build caching.
  - Có cache utilities nội bộ.
- DB/indexing:
  - Có SQLite state layer, nhưng không thấy benchmark/indexing overview public trong docs đã đọc.
- Async processing:
  - Tokio async runtime.
  - Bounded queues trong app-server.
  - Background jobs/state pipeline có trong `state/`.
- Performance tuning:
  - Release profile tối ưu.
  - TUI stream chunking có docs tuning/review/validation.
- Benchmark:
  - Không thấy benchmark dashboard/SLO public rõ ràng.

### Evidence

- `reference/codex/.github/workflows/rust-ci.yml`
- `reference/codex/codex-rs/app-server/README.md`
- `reference/codex/docs/tui-stream-chunking-review.md`
- `reference/codex/docs/tui-stream-chunking-tuning.md`
- `reference/codex/docs/tui-stream-chunking-validation.md`
- `reference/codex/codex-rs/Cargo.toml`

### Assessment

- Perf mindset rõ.
- Public benchmark story chưa đủ mạnh.

## 10. Documentation

### Findings

- README root tốt cho install/quickstart.
- Setup/build guide có.
- Config docs có.
- API docs mạnh nhất nằm ở `app-server/README.md`.
- Một số docs sâu và chất lượng cao.
- Nhưng nhiều docs chỉ là pointer sang developers.openai.com, nên self-contained quality không đồng đều.

### Evidence

- `reference/codex/README.md`
- `reference/codex/docs/install.md`
- `reference/codex/docs/config.md`
- `reference/codex/codex-rs/app-server/README.md`
- `reference/codex/docs/getting-started.md`

### Assessment

- Technical depth tốt.
- Offline/local-only audit vẫn bị hụt context ở vài khu vực do external docs dependency.

## 11. Contribution Flow

### Findings

- PR template: có.
- Commit convention:
  - Không thấy commitlint hard-enforced từ các file đã đọc.
  - Git history và changelog tooling nghiêng về conventional style.
- Branch strategy:
  - Topic branch từ `main`.
  - Maintainer squash-and-merge.
- CLA:
  - Có workflow CLA assistant.
- Governance:
  - External code contribution **by invitation only**.
  - PR không được invite sẽ bị đóng.

### Evidence

- `reference/codex/docs/contributing.md`
- `reference/codex/.github/pull_request_template.md`
- `reference/codex/.github/workflows/cla.yml`
- `reference/codex/cliff.toml`

### Assessment

- Contribution process rất chặt.
- Open discussion được khuyến khích hơn open code contribution.

## 12. Activity

### Findings

- Repo active: **Có**.
- Current branch / remote default branch: `main` / `origin/main`.
- Last commit:
  - Hash: `8c5e50ef3962614180e3fb84393cdc669764d6a1`
  - Date: `2026-03-03 12:25:40 +0000`
  - Author: `jif-oai`
  - Subject: `feat: spreadsheet artifact (#13345)`
- Contributor count (`git shortlog -sn --all | wc -l`): **467**.

### Evidence

- `git log -1 --date=iso --pretty=format:'hash=%H%ndate=%ad%nauthor=%an%nsubject=%s%n'`
- `git shortlog -sn --all | wc -l`

### Assessment

- Repo sống, nhịp commit cao.

## 13. License

### Findings

- License: **Apache-2.0**.
- Có `NOTICE` và third-party licenses.
- Phù hợp reuse thương mại, OSS, và có explicit patent grant.

### Evidence

- `reference/codex/LICENSE`
- `reference/codex/NOTICE`
- `reference/codex/docs/license.md`

### Assessment

- Reuse-friendly.
- Compliance hygiene tốt.

## 14. Dependencies

### Findings

- Outdated management signals:
  - Dependabot weekly cho Cargo, GitHub Actions, Docker, devcontainers, bun, rust-toolchain.
- Security vulnerability signals:
  - `cargo-deny` workflow có.
  - `cargo-audit` workflow có.
  - `deny.toml` và `.cargo/audit.toml` ignore một số `RUSTSEC-*` advisories có ghi lý do cụ thể.
- Patched/forked deps:
  - Có git/fork dependencies trong Rust workspace.
- Điều chưa xác minh:
  - Audit này **không chạy live `cargo audit`/`cargo deny`/outdated scan** trong phiên.

### Evidence

- `reference/codex/.github/dependabot.yaml`
- `reference/codex/.github/workflows/cargo-deny.yml`
- `reference/codex/codex-rs/.github/workflows/cargo-audit.yml`
- `reference/codex/codex-rs/deny.toml`
- `reference/codex/codex-rs/.cargo/audit.toml`
- `reference/codex/codex-rs/Cargo.toml`

### Assessment

- Dependency governance mature.
- Vẫn có accepted security debt do upstream/unmaintained deps chưa có safe upgrade path.

## Unresolved questions

1. Có roadmap/backlog public nào ngoài repo không?
2. Có benchmark/SLO public cho app-server hoặc TUI không?
3. Có commit convention enforcement cụ thể ngoài PR/review culture không?
4. ETA remediation cho các `RUSTSEC` đang ignore là gì?
