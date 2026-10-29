# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Volt pulls daily electricity consumption from the EVNHCMC customer portal, upserts it into PostgreSQL, computes tiered billing cost (with VAT), and reports it through a Telegram bot and a small HTTP API. Module path: `github.com/gamee1910/volt`.

## Commands

```bash
docker compose up -d db            # local Postgres 16 (admin/password, db "volt", port 5432)
make migrate-up                    # apply migrations (golang-migrate CLI; DB_URL set in Makefile)
make migrate-create name=foo       # new sequential migration in ./migrations
make migrate-down-one | migrate-version | migrate-force version=N
make gen-docs                      # swag docs from cmd/api/main.go into ./docs

air                                # run API with live reload (builds ./cmd/api to ./bin/main)
go run ./cmd/api                   # API server + Telegram bot + in-process scheduler
go run ./cmd/cli/sync              # one-shot sync: first of month -> today
go run ./cmd/cli/report            # one-shot: send yesterday's report to the bot's first chat

go build ./...
go vet ./...
go test ./...                      # there are currently no _test.go files
go test ./internal/common/utils -run TestName   # single test
```

Config comes only from environment variables (`config/config.go`), normally loaded via direnv from `.envrc` (template: `.envrc.example`). Note the customer code variable is `EVN_CUSTOMER` (the README's `EVN_CUSTOMER_CODE` is outdated). The GitHub workflows in `.github/workflows/` run the CLI commands on a self-hosted runner and are `workflow_dispatch` only (their cron triggers are commented out).

## Architecture

Clean/hexagonal layout with CQRS in the application layer:

- `internal/domain` — entities (`entity.ElectricityConsumption`) and the `repository.ElectricityRepository` interface. No outward dependencies.
- `internal/application` — CQRS handlers. `application.Application` groups `Commands` (e.g. `SyncElectricity`) and `Queries` (`AllElectricity`, `YesterdayUsage`). Every handler is built as a private struct and returned wrapped by `decorator.ApplyCommandDecorators` / `ApplyQueryDecorators` (logging → metrics → handler). Follow this pattern for new handlers and register them in `application.NewApplication`. `application/types` holds the DTOs that queries return.
- `internal/common` — decorators (generic `CommandHandler[C]` / `QueryHandler[Q,R]`), a log-only metrics client, and `utils` (Vietnam timezone, EVN date format/parse, `CalculateElectricityBill` tier pricing, VND formatting).
- `internal/infrastructure` — `persistences/postgres` implements the repository with `database/sql` + `lib/pq` (`redis/` is a placeholder). `di/container.go` is the single composition root: it builds the EVN and Telegram clients, repository, `Application`, HTTP handler, bot command handler, scheduler, and attaches the bot router as the Telegram message handler.
- `internal/interfaces` — the three entry adapters, all of which call into `Application` only:
  - `api` — chi router (`/health`, `/api/v1/{sync,"",yesterday}`) and handlers with request/response DTOs.
  - `bot` — Telegram command router (`/yesterday`, `/sync [from] [to]`, `/get`, `/help`); messages are in Vietnamese.
  - `scheduler` — `robfig/cron` job (`0 */12 * * *`, Asia/Ho_Chi_Minh) that runs SyncElectricity for the current month, plus once immediately on `Start()`. After each cron run (not the startup run) it reports to `TELEGRAM_CHAT_ID` via `ElectricityCommand.NotifySyncResult`; unset/0 disables this.
- `pkg/` — reusable clients with no dependency on `internal`: `evnhcmc` (portal login + daily usage API), `telegram` (wrapper around `go-telegram/bot`), `logger` (zap).
- `cmd/` — `api` (long-running server) and `cli/{sync,report}` (one-shot jobs). Each entrypoint builds config → logger → DB → `di.NewContainer`, so all three share the same wiring.

Domain details worth knowing:
- All date logic uses the Vietnam timezone via `utils.LoadVietnamTimezone()` (tzdata is embedded via `time/tzdata`, so it works on the alpine runtime image); sync ranges are always "first day of month → now" unless given explicitly.
- Sync upserts one row per `measurement_date` (unique constraint). If EVN reports `TotalOutput` of 0, the handler falls back to `TotalIndex`.
- Monthly cost = `CalculateElectricityBill(total kWh from first of month to yesterday)`.
- `docs/mock_evnhcmc.json` is a sample EVN API response.

## Docs in `.claude/`

- `.claude/project-summary.md` — project overview and known gaps (Vietnamese).
- `.claude/deploy-checklist.md` — step-by-step Botkeep deploy checklist (Vietnamese).
