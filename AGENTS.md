# AGENTS.md - md-spec-tool (monorepo)

## Structure
- **`backend/`** — Go 1.24 API server (Gin, port 8080). OpenAI-powered AI suggestions, Excel→Markdown conversion, diff, sharing, quotas. See `backend/AGENTS.md`.
- **`frontend/`** — Next.js 16 + React 19 + TypeScript app (port 3000). Tailwind CSS 4, Zustand 5, TanStack React Query. See `frontend/AGENTS.md`.
- **`use-cases/`** — Sample/test data for evaluation runners.
- **`.env.example`** — Required env vars (OpenAI key, ports, rate limits, cookie secret).

## Build & Test
- **Backend**: `cd backend && go test ./...` | single test: `go test -run TestName ./internal/pkg` | run: `go run ./cmd/server`
- **Frontend**: `cd frontend && npm test` (Vitest) | single test: `npx vitest run path/to/file.test.ts` | dev: `npm run dev` | e2e: `npm run test:e2e`

## Code Style
- **Go**: absolute imports from `github.com/yourorg/md-spec-tool`; wrap errors with `fmt.Errorf("ctx: %w", err)`; use `log/slog`; Gin handlers return `c.JSON()`
- **TypeScript**: strict mode; `@/` path alias; Tailwind (no inline styles); `clsx`/`tailwind-merge`; Radix UI primitives; `lucide-react` icons
- **Naming**: `PascalCase` exports/components/types, `camelCase` unexported/functions, `UPPER_SNAKE_CASE` constants

## Key APIs
- Backend serves REST endpoints consumed by frontend via `lib/httpClient.ts` → `lib/mdflowApi.ts`, `shareApi.ts`, `quotaApi.ts`
- AI endpoints use OpenAI (`internal/ai/`) with streaming support (`internal/http/handlers/stream_handler.go`)
