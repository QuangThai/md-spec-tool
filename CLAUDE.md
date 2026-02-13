# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Does

MD-Spec-Tool converts spreadsheet and pasted tabular data into Markdown specifications ("MDFlow"). It supports paste text, TSV, XLSX, and Google Sheets inputs with two output modes: `spec` (structured requirements) and `table` (clean markdown table). AI-assisted column mapping, suggestions, and validation are optional (gracefully degrades without an API key). BYOK (Bring Your Own Key) allows per-request AI via `X-OpenAI-API-Key` header.

## Build & Dev Commands

### Backend (Go 1.24+ / Gin)
```bash
cd backend && go run ./cmd/server          # Dev server on :8080
cd backend && go test ./...                 # All tests
cd backend && go test ./internal/converter  # Single package
cd backend && go test -run TestFoo ./internal/ai  # Single test by name
make cli                                    # Build CLI to bin/mdflow
```

### Frontend (Next.js 16 / React 19)
```bash
cd frontend && npm install    # Install deps
cd frontend && npm run dev    # Dev server on :3000
cd frontend && npm run build  # Production build
cd frontend && npm test       # Vitest
```

### Docker
```bash
make build    # docker-compose build
make up       # Start services
make down     # Stop services
```

## Architecture

Monorepo with Go backend and Next.js frontend. Backend serves API on port 8080, frontend on port 3000.

### Backend (`backend/`)

**Entry points** in `cmd/`: `server` (HTTP API), `cli` (mdflow CLI), `usecases` / `usecases_diff` (test fixtures against use-case data).

**Core packages** in `internal/`:

- **converter/** — Central domain logic. Pipeline: detect input type → parse to matrix → detect header row → column mapping (rule-based + optional AI) → build SpecDoc → render output. Has a `UseNewConverterPipeline` feature flag for a newer Table-based pipeline. Key files: `converter.go` (orchestrator), `paste_parser.go`, `header_detect.go`, `header_resolver.go`, `model.go` (SpecDoc/SpecRow types), `template_registry.go`, `renderer_factory.go`, `ai_mapping.go` (AI confidence thresholds and fallback).

- **ai/** — OpenAI integration layer. Structured JSON schema output with retry/backoff. Used for column mapping, paste analysis, suggestions, diff summarization, and semantic validation. `client.go` wraps the OpenAI API, `service.go` defines the service interface, `prompts.go` has all system prompts, `schemas.go` defines JSON schemas for structured output.

- **http/** — Gin router setup (`router.go`), middleware (CORS, rate limiting), and handlers. Handlers are split by concern: `mdflow_convert.go`, `mdflow_preview.go`, `mdflow_gsheet.go`, `mdflow_ai_suggest.go`, `mdflow_validation.go`. BYOK support creates per-request AI service instances.

- **config/** — Env-based configuration with defaults. AI is auto-enabled when `OPENAI_API_KEY` is set.

- **share/** — File-backed JSON store for shareable MDFlow documents and comment threads. Atomic save via temp file + rename.

- **suggest/** — Thin adapter wrapping AI suggestions into API response structures.

- **diff/** — Line-based unified diff, optionally summarized via AI.

### Frontend (`frontend/`)

**Next.js App Router** with routes: `/` (landing), `/studio` (main workbench), `/batch`, `/docs`, `/gallery`, `/share`, `/s/[key]` (share slug).

**API proxy routes** (`app/api/`): `gsheet/*` and `oauth/google/*` proxy to the Go backend, injecting OAuth tokens for Google Sheets access.

**State management**: React Query (TanStack Query 5) for server state, Zustand 5 for client UI state with persistence (history, BYOK key). Query hooks in `lib/mdflowQueries.ts`, API functions in `lib/mdflowApi.ts`, HTTP client in `lib/httpClient.ts`.

**Styling**: Tailwind CSS v4 with `@theme` tokens in `styles/globals.css`. Utility composition via `cn()` from `lib/utils.ts` (clsx + tailwind-merge). UI components in `components/ui/` are custom wrappers around Radix UI primitives.

**Key UI component**: `MDFlowWorkbench.tsx` is the main studio controller that composes conversion flow hooks, preview, and AI suggestions.

## Code Conventions

**Go**: `log/slog` for logging. Errors propagated with `fmt.Errorf()` context. Gin handlers return `c.JSON(status, data)`. Config loaded from env vars with defaults. Package imports use full module path `github.com/yourorg/md-spec-tool/internal/...`.

**TypeScript**: Strict mode. Functional components only. Zustand stores split into state and actions. Tailwind classes only (no inline styles). Path alias `@/*` maps to project root.

**Naming**: Go exports `PascalCase`, unexported `camelCase`. TS functions/variables `camelCase`, components/types `PascalCase`, constants `UPPERCASE_SNAKE_CASE`.

## Key Design Decisions

- Preview endpoints default to `skip_ai=true` for speed; AI mapping is opt-in per request.
- AI endpoints return structured non-fatal responses when unconfigured — the app works without any API key.
- Output modes are strictly `spec` and `table`. The API accepts both `template` and `format` as aliases for backward compatibility; prefer `template`.
- The converter pipeline uses confidence thresholds to decide between AI and rule-based mapping, falling back gracefully.
- Google Sheets requests go through Next.js API proxy routes that inject OAuth tokens before forwarding to the Go backend.

## Environment

Copy `.env.example` to `.env`. Core variables: `OPENAI_API_KEY` (optional), `OPENAI_MODEL`, `NEXT_PUBLIC_API_URL` (default `http://localhost:8080`), `NEXT_PUBLIC_APP_URL`, `COOKIE_SECRET` (required for Google OAuth sessions).
