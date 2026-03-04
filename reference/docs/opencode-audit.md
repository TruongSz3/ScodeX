# OpenCode Repository Audit

Ngày audit: 2026-03-03
Repo: `reference/opencode/`
Phương pháp: đọc `reference/repomix-opencode.xml`, README/docs/workflows/manifests/infra files, đối chiếu git metadata.

## Executive snapshot

- OpenCode là **open-source AI coding agent** cho developer, local-first, nhưng có thêm cloud/web/desktop surfaces.
- Repo là **Bun + TypeScript monorepo** với nhiều package: core runtime, app web, desktop, docs, console, enterprise, sdk.
- Core surface: local server, CLI/TUI, built-in agents (`build`, `plan`, `general`), provider-agnostic model access, LSP support.
- Kiến trúc là **modular monolith local + edge/cloud services** trên Cloudflare/SST, không phải microservices/Kubernetes truyền thống.
- Product breadth lớn hơn Codex: desktop beta, docs site, console, enterprise, sharing API, cloud infra.
- Security posture minh bạch nhưng thực dụng: **không sandbox**; permission system là UX guardrail chứ không phải isolation boundary.
- Server mode có thể bật HTTP Basic Auth bằng `OPENCODE_SERVER_PASSWORD`; nếu không đặt, server chạy unauthenticated kèm warning.
- CI có test/unit/e2e/typecheck tốt; PR governance automation rất chặt.
- Deployment/DevOps mature: SST deploy theo `dev`/`production`, Cloudflare Workers/KV/R2/Durable Objects, PlanetScale, Stripe, containers, signing.
- Code structure nhìn chung rõ, nhưng có vài file lớn trong core server/file logic.
- Documentation root tốt, đa ngôn ngữ, nhưng package-level docs không đồng đều; có template debt.
- Repo đang **active mạnh**: last commit ngày 2026-03-03; contributor count theo `git shortlog` là 799.
- License là **MIT**, rất dễ reuse.
- Dependencies được pin khá chặt (`bunfig.toml exact = true`, lockfile, patches), nhưng không thấy evidence mạnh cho automated vuln scanning như Dependabot/CodeQL trong phạm vi đã đọc.

## 1. Features + Scope

### Findings

- Product chính: open-source AI coding agent.
- Core features:
  - Local CLI/TUI.
  - Built-in agents: `build`, `plan`, `general`.
  - Provider-agnostic model support.
  - LSP support.
  - Client/server architecture.
  - Project/session/tool/permission flows.
- Features phụ / adjacent product surfaces:
  - Desktop app beta.
  - Static/docs/web app.
  - Console/cloud features.
  - Enterprise package.
  - GitHub action / SDK surfaces.
- Problem repo giải quyết:
  - Cho developer dùng AI coding agent local-first, có thể dùng từ terminal hoặc client khác, và mở rộng lên cloud surfaces.
- In-scope:
  - Core coding agent.
  - Local server/API.
  - TUI/app/desktop clients.
  - Cloud features cho docs/share/console/enterprise.
- Out-of-scope / explicit boundaries:
  - Security doc nói rõ một số category ngoài scope: sandbox escapes, provider data handling, MCP server behavior, server access when user opted-in.
- Product type:
  - **Developer tool platform OSS**.
  - Không phải library thuần.
  - Không phải internal tool.
- Roadmap/backlog:
  - Không thấy `ROADMAP.md`/changelog chuẩn ở root.
  - Backlog flow hiện diện qua issue-first policy, issue templates, specs như `specs/project.md`.

### Evidence

- `reference/opencode/README.md`
- `reference/opencode/CONTRIBUTING.md`
- `reference/opencode/SECURITY.md`
- `reference/opencode/specs/project.md`
- `reference/opencode/packages/`

### Assessment

- Product breadth lớn.
- Roadmap visibility thấp hơn mức lý tưởng.

## 2. Target Users

### Findings

- Primary users:
  - Developers terminal-first.
  - Neovim/TUI-oriented power users.
  - OSS contributors.
- Secondary users:
  - Desktop app users.
  - Teams/enterprise users của cloud/console surfaces.
- Expected scale:
  - Repo không nêu active-user/SLA/concurrency public.
  - `STATS.md` cho thấy tổng downloads đạt **10,190,453** vào `2026-01-29`.

### Evidence

- `reference/opencode/README.md`
- `reference/opencode/STATS.md`

### Assessment

- Adoption public có tín hiệu lớn.
- Scale runtime thực tế vẫn chưa có public SLO.

## 3. Architecture

### Findings

- Architecture style: **modular monolith local + edge/cloud services**.
- Local core:
  - `packages/opencode` chứa CLI/server/runtime logic.
- Cloud/edge:
  - `infra/` dùng SST.
  - `packages/function/src/api.ts` chạy Cloudflare Worker + Durable Object + R2.
  - `packages/console/*` và `packages/enterprise` cho cloud surfaces.
- Repo không phải microservices/Kubernetes kiểu truyền thống.
- README xác nhận client/server architecture.
- Không thấy diagram chính thức trong docs đã đọc.

### Inferred diagram

```text
User
 |-- CLI/TUI ----------------------> packages/opencode
 |                                   |- local server (Hono)
 |                                   |- project/session/tool/provider
 |                                   |- local SQLite
 |
 |-- Web app ----------------------> packages/app
 |-- Desktop app ------------------> packages/desktop (Tauri)
 |
 +-- Cloud surfaces (SST/Cloudflare)
      |- Worker API
      |- Durable Object SyncServer
      |- KV / R2
      |- Console/Auth workers
      |- PlanetScale MySQL
      |- Stripe billing
```

### Evidence

- `reference/opencode/README.md`
- `reference/opencode/packages/opencode/src/server/server.ts`
- `reference/opencode/infra/app.ts`
- `reference/opencode/infra/console.ts`
- `reference/opencode/packages/function/src/api.ts`

### Assessment

- Kiến trúc phù hợp với product breadth.
- Complexity cao hơn local-only tool vì có cả cloud stack.

## 4. Tech Stack

### Findings

- Backend/runtime:
  - Bun + TypeScript.
  - Hono.
  - Zod.
  - `hono-openapi`.
- Frontend:
  - SolidJS + Vite (`packages/app`).
  - Astro/Starlight (`packages/web`).
  - SolidStart (`packages/console/app`, `packages/enterprise`).
- Desktop:
  - Tauri.
- Database:
  - Local SQLite qua Drizzle (`packages/opencode/drizzle.config.ts`).
  - Cloud MySQL/PlanetScale (`packages/console/core/drizzle.config.ts`, `infra/console.ts`).
- Cache / state infra:
  - Cloudflare KV.
  - R2 buckets.
  - Durable Objects.
- Queue:
  - Không thấy dedicated queue như Redis/Kafka/Rabbit trong phạm vi đã đọc.
- Search:
  - `ripgrep` + `fuzzysort` signals trong core package.
- Infra:
  - SST.
  - Cloudflare.
  - Stripe.
  - GitHub Actions.
  - Nix.

### Evidence

- `reference/opencode/package.json`
- `reference/opencode/packages/opencode/drizzle.config.ts`
- `reference/opencode/packages/console/core/drizzle.config.ts`
- `reference/opencode/infra/app.ts`
- `reference/opencode/infra/console.ts`
- `reference/opencode/infra/enterprise.ts`
- `reference/opencode/packages/function/src/api.ts`

### Assessment

- Stack hiện đại, tốc độ cao, nhưng vận hành phức tạp hơn vì nhiều runtime/platform.

## 5. Code Structure

### Findings

- Root workspace chia package rõ:
  - `packages/opencode` core runtime.
  - `packages/app` web app.
  - `packages/desktop` desktop app.
  - `packages/console/*` cloud console.
  - `packages/web` docs/marketing.
  - `packages/sdk/js`, `sdks/`, `github/`, `infra/`.
- Separation of concerns tốt ở package level.
- Naming nhìn chung rõ, nhưng style nội bộ project ưu tiên single-word identifiers ở code TS.
- Một số file lớn tồn tại, ví dụ `packages/opencode/src/server/server.ts`.

### Evidence

- `reference/opencode/package.json`
- `reference/opencode/packages/`
- `reference/opencode/packages/opencode/src/server/server.ts`
- `reference/opencode/AGENTS.md`

### Assessment

- Folder structure đủ rõ để làm việc.
- Technical debt cục bộ ở vài file lớn.

## 6. Code Quality

### Findings

- Tests:
  - Unit tests qua `bun turbo test`.
  - E2E Playwright cho app.
  - Typecheck riêng qua workflow.
- Lint/format:
  - Prettier config có ở root.
  - Một số package có lint/check scripts.
  - Không thấy repo-wide lint workflow mạnh tương đương test/typecheck trong phạm vi đã đọc.
- CI/CD:
  - `test.yml` chạy Linux + Windows unit tests.
  - `test.yml` chạy Linux + Windows e2e app tests.
  - `typecheck.yml` chạy trên `dev` + PR.
  - Có nhiều workflows phụ cho review, publish, deploy, triage.

### Evidence

- `reference/opencode/.github/workflows/test.yml`
- `reference/opencode/.github/workflows/typecheck.yml`
- `reference/opencode/package.json`
- `reference/opencode/.husky/pre-push`
- `reference/opencode/packages/app/e2e/`
- `reference/opencode/packages/opencode/test/`

### Assessment

- Test/typecheck discipline tốt.
- Lint/security automation chưa rõ ràng bằng Codex.

## 7. Deployment & DevOps

### Findings

- Docker/containers:
  - Có workflow `containers.yml` và package containers.
- Kubernetes:
  - Không thấy K8s/Helm manifests.
- Environments:
  - `deploy.yml` deploy theo branch `dev` và `production`.
  - `sst.config.ts` dùng `stage`; `production` được `protect` và `retain` resources.
  - `infra/stage.ts` map domain theo `production`, `dev`, và các stage khác.
- Cloud infra:
  - Cloudflare Workers/KV/R2/Durable Objects.
  - PlanetScale MySQL.
  - Stripe.
- Config management:
  - Dùng SST secrets/linkables/resources.
  - Nhiều secrets được define rõ trong infra.
- Distribution:
  - Installer script đa OS.
  - Publish workflows cho CLI/desktop/vscode/github-action.

### Evidence

- `reference/opencode/.github/workflows/deploy.yml`
- `reference/opencode/sst.config.ts`
- `reference/opencode/infra/stage.ts`
- `reference/opencode/infra/app.ts`
- `reference/opencode/infra/console.ts`
- `reference/opencode/infra/enterprise.ts`
- `reference/opencode/install`

### Assessment

- DevOps mature và đa surface.
- Không phải repo Kubernetes-native.

## 8. Security

### Findings

- Auth mechanism:
  - Server mode có thể dùng HTTP Basic Auth qua `OPENCODE_SERVER_PASSWORD`.
  - OAuth/provider auth flows tồn tại trong server routes.
- Validation:
  - `zod` + `hono-openapi` + validators được dùng trong server.
- CORS / boundary controls:
  - Server có CORS whitelist logic.
  - Có error handling chặt ở server.
- Permission model:
  - Repo nói rõ permission system là UX feature, **không phải sandbox**.
- Rate limit:
  - Có domain-specific rate limit logic ở Zen/subscription parts.
  - Có retry handling cho `429` ở vài luồng.
  - Không thấy generic API rate limiting middleware toàn repo cho local server.
- Secret management:
  - SST secrets.
  - Env vars cho deploy/runtime.
  - Infra files define nhiều secrets cho Cloudflare/Stripe/GitHub/R2/auth.

### Evidence

- `reference/opencode/SECURITY.md`
- `reference/opencode/packages/opencode/src/cli/cmd/serve.ts`
- `reference/opencode/packages/opencode/src/server/server.ts`
- `reference/opencode/sst.config.ts`
- `reference/opencode/infra/app.ts`
- `reference/opencode/infra/console.ts`
- `reference/opencode/packages/console/app/src/routes/zen/util/handler.ts`

### Assessment

- Security communication rất honest.
- Trade-off rõ: tốc độ/phạm vi product > isolation.

## 9. Performance

### Findings

- Caching:
  - Cloudflare KV được dùng ở console/auth/gateway flows.
  - Durable Objects + R2 giúp state/share distribution.
- DB/indexing:
  - Có local SQLite và cloud MySQL, nhưng không thấy benchmark/indexing strategy doc tổng thể.
- Async processing:
  - Bun async runtime.
  - Worker/Durable Object flows.
  - SSE streaming trong server.
- Build/runtime optimization:
  - `turbo` cho workspace tasks.
  - Containers để tối ưu CI.
- Benchmark:
  - Không thấy benchmark suite public hoặc perf budget/SLO công khai.

### Evidence

- `reference/opencode/infra/console.ts`
- `reference/opencode/packages/function/src/api.ts`
- `reference/opencode/packages/opencode/src/server/server.ts`
- `reference/opencode/turbo.json`
- `reference/opencode/packages/containers/README.md`

### Assessment

- Có nền tảng scale-out ở cloud pieces.
- Thiếu benchmark story chính thức.

## 10. Documentation

### Findings

- README root tốt, đa ngôn ngữ, install guide rõ.
- Contributing và security docs đầy đủ.
- Docs chính được trỏ ra `opencode.ai/docs`.
- Package-level docs không đồng đều; một số README mang tính template/starter hơn là product docs hoàn chỉnh.
- API docs local có thể generate từ server/OpenAPI, nhưng docs công khai trong repo chưa thống nhất thành một internal source of truth.

### Evidence

- `reference/opencode/README.md`
- `reference/opencode/CONTRIBUTING.md`
- `reference/opencode/SECURITY.md`
- `reference/opencode/packages/docs/`
- `reference/opencode/packages/web/README.md`
- `reference/opencode/packages/opencode/src/server/server.ts`

### Assessment

- Root-level docs tốt cho onboarding.
- Package/docs consistency còn yếu.

## 11. Contribution Flow

### Findings

- PR template: có, bắt buộc điền đúng sections.
- Commit convention:
  - PR titles phải theo conventional commit format (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`).
- Branch strategy:
  - Default branch hiện tại là `dev`.
  - Deploy flows chạy trên `dev` và `production`.
- Issue-first policy:
  - Feature request nên bắt đầu bằng issue/design discussion.
- Automation:
  - Workflow check PR title.
  - Workflow check linked issue.
  - Workflow check PR template compliance.
  - Auto-close issue/PR không compliant sau 2 giờ.

### Evidence

- `reference/opencode/CONTRIBUTING.md`
- `reference/opencode/.github/pull_request_template.md`
- `reference/opencode/.github/workflows/pr-standards.yml`
- `reference/opencode/.github/workflows/compliance-close.yml`
- `reference/opencode/.github/ISSUE_TEMPLATE/`

### Assessment

- Contribution process rất chặt nhưng vẫn mở cho contributor ngoài.
- Governance automation là điểm mạnh rõ rệt.

## 12. Activity

### Findings

- Repo active: **Có**.
- Current branch / remote default branch: `dev` / `origin/dev`.
- Last commit:
  - Hash: `cbf0570489b30a366d5e93de5640034086f84281`
  - Date: `2026-03-03 15:27:54 +0300`
  - Author: `İbrahim Hakkı Ergin`
  - Subject: `fix: update Turkish translations (#15835)`
- Contributor count (`git shortlog -sn --all | wc -l`): **799**.

### Evidence

- `git log -1 --date=iso --pretty=format:'hash=%H%ndate=%ad%nauthor=%an%nsubject=%s%n'`
- `git shortlog -sn --all | wc -l`

### Assessment

- Repo active mạnh, cộng đồng contributor lớn.

## 13. License

### Findings

- License: **MIT**.
- Rất permissive, reuse-friendly cho commercial và OSS use cases.

### Evidence

- `reference/opencode/LICENSE`
- `reference/opencode/package.json`

### Assessment

- Phù hợp reuse rất tốt.

## 14. Dependencies

### Findings

- Outdated management signals:
  - `bunfig.toml` đặt `exact = true`.
  - Có `bun.lock`.
  - Root `package.json` dùng nhiều version pin cụ thể.
  - Có `patchedDependencies`.
- Risk signals:
  - Có dùng một số pre-release/custom URLs, ví dụ `@solidjs/start` từ `pkg.pr.new`, drizzle beta, TS native preview.
  - Không thấy `dependabot`, `codeql`, hay workflow audit bảo mật dependency rõ ràng trong `.github/workflows` đã đọc.
- Security vulnerability status:
  - Audit này **không chạy live `bun audit`/OSV/outdated scan**.
  - Vì vậy chưa thể kết luận chắc có bao nhiêu outdated packages hay CVE active.

### Evidence

- `reference/opencode/package.json`
- `reference/opencode/bunfig.toml`
- `reference/opencode/bun.lock`
- `reference/opencode/.github/workflows/`

### Assessment

- Dependency pinning tốt.
- Visibility về vuln/outdated automation yếu hơn Codex.

## Unresolved questions

1. Có roadmap/changelog public chính thức ở ngoài repo không?
2. Có dependency/security scanning ở private pipeline không?
3. Có benchmark/perf budget/SLO public cho local server hoặc cloud surfaces không?
4. Package docs nào là source of truth: repo hay docs site ngoài repo?
