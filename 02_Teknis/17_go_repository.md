# 17 — Struktur repositori Go dan engineering standards

## Tech choices

- **Go 1.24+** (`os.OpenRoot` feature) dengan versi minor terbaru yang disetujui; pin toolchain pada CI/Docker.
- **Gin** sebagai HTTP router; middleware auth/security.
- **pgx/v5** PostgreSQL driver & pools.
- **sqlc** generated query types; SQL eksplisit, tidak mengandalkan ORM dinamis untuk inti tenant.
- **goose** migrations.
- **OpenAPI 3.1** kontrak API, generated clients bila perlu.
- **PostgreSQL jobs/outbox** untuk P0, tanpa memaksa Redis di fase awal.
- React + Tailwind untuk frontend dashboard (repo frontend terpisah atau `web/`).

## Repository tree

```text
ponpes-saas/
├── cmd/
│   ├── api/main.go
│   └── worker/main.go
├── internal/
│   ├── platform/        # config, auth, middleware, errors, clock, transaction
│   ├── tenant/          # lifecycle, admin
│   ├── identity/        # users, roles, session/MFA
│   ├── hermes/          # provisioner adapter, profiles, gateway ops
│   ├── telegram/        # verify bot, token lifecycle
│   ├── storage/         # safe DataPonpes file operations
│   ├── knowledge/       # extraction, retrieval, source citations
│   ├── crawler/         # P1: sources, scheduler tick, safe fetcher (SSRF guard), extraction
│   ├── telemetry/       # signed receiver, usage, analytics
│   ├── admission/       # P1 leads/status/handoff
│   ├── jobs/            # queue, retry, worker execution
│   └── audit/           # append-only audit
├── db/
│   ├── migrations/
│   ├── queries/
│   └── sqlc.yaml
├── internal/db/generated/
├── configs/
│   └── hermes/         # versioned baseline templates, no credentials
├── docs/
│   └── openapi.yaml
├── deploy/
│   ├── docker-compose.yml
│   ├── Dockerfile.api
│   ├── Dockerfile.worker
│   └── scripts/
├── tests/
│   ├── integration/
│   ├── security/
│   └── fixtures/
├── web/                  # React UI optional monorepo
├── Makefile
├── go.mod
└── README.md
```

## Per module conventions

`handler.go` HTTP validation and transport; `service.go` business orchestration; `repository.go` storage interface where needed; `model.go` domain structs; `policy.go` authorization policy; tests near source. Hindari over-abstraction: interface di boundary eksternal (Hermes, Telegram, filesystem, clock) atau saat test butuh, bukan setiap service.

## Engineering quality gates

```bash
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
# sqlc generate + goose migration checks + static checks as pinned CI
```

Add `govulncheck`, linter versi pin, Testcontainers PostgreSQL untuk integration, per-tenant negative tests setiap merge, `go test` schema fixtures. Jangan commit API keys, token Telegram, `.env`, logs private atau dokumen asli ke git.

## Error handling

Use typed codes (`TENANT_NOT_FOUND`, `PROVISION_FAILED`, `BOT_TOKEN_INVALID`, `DOCUMENT_NOT_READY`), map error to HTTP/status retryable, wrap with `%w`, log internal detail redacted, request correlation ID; jangan error string langsung diekspos ke UI.

## Deployment discipline

Small reversible migrations, image pinned by digest, read-only root filesystem containers jika mungkin, separate workload credentials, graceful shutdown, bounded goroutines/worker pools, connection pool limits and runtime metrics.
