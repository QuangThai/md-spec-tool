# AGENTS.md - md-spec-tool

## Build, Test & Lint Commands

### Backend (Go 1.24+)
- **Test all**: `cd backend && go test ./...`
- **Test single package**: `cd backend && go test ./internal/ai` (replace with package path)
- **Test single file**: `cd backend && go test -run TestNamePrefix ./internal/ai`
- **Run dev server**: `make dev-backend` (runs `go run ./cmd/server`)
- **Build CLI**: `make cli` (builds to `bin/mdflow`)

### Frontend (Next.js 16)
- **Dev server**: `make dev-frontend` (runs `npm run dev`, port 3000)
- **Build**: `cd frontend && npm run build`
- **Test**: `cd frontend && npm test` (Vitest)

### Docker
- **Build**: `make build` (docker-compose build)
- **Start**: `make up` (starts all services)
- **Stop**: `make down`
- **Logs**: `make logs`

## Architecture & Codebase

**Tech Stack**:
- **Backend**: Go 1.24+ (Gin web framework) + OpenAI API v3 + Excelize (Excel parsing)
- **Frontend**: Next.js 16 + React 19 + TypeScript + Tailwind CSS + Zustand state
- **Services**: Docker Compose, Google APIs (OAuth2)

**Key Backend Packages**:
- `backend/internal/http/`: Routes & handlers (HTTP transport)
- `backend/internal/ai/`: OpenAI integration (column mapping, semantic validation, refinement, caching)
- `backend/internal/converter/`: Excel→Markdown conversion + diff logic
- `backend/internal/config/`: Configuration via env vars (.env, .env.example)
- `backend/internal/share/`: Shared content storage & retrieval
- `backend/internal/suggest/`: AI suggestions for specs
- `backend/cmd/server/`: Entry point (main.go, starts HTTP server on 8080)
- `backend/cmd/cli/`: CLI tool entry point

**Frontend Structure**:
- `frontend/app/`: Next.js App Router pages
- `frontend/components/`: React UI components
- `frontend/lib/`: Utilities (API calls, helpers)
- `frontend/hooks/`: Custom React hooks

**Ports**: Frontend 3000, Backend 8080. Config via env vars: `HOST`, `PORT`, `CORS_ORIGINS`, `OPENAI_API_KEY`.

## Code Style & Guidelines

**Go**:
- CamelCase exports, camelCase unexported fields/functions
- Explicit error handling: check and propagate with `fmt.Errorf()` for context
- Gin HTTP: return `c.JSON(status, data)` or `c.Error(err)`
- Middleware: use `func(c *gin.Context)`, call `c.Next()` to continue
- Package structure: `internal/` for private, `cmd/` for binaries, absolute imports (e.g., `github.com/yourorg/md-spec-tool/internal/ai`)
- Logging: use stdlib `log/slog` with `slog.Info()`, `slog.Warn()`, `slog.Error()`
- Config: load env vars in `LoadConfig()`, provide defaults, validate required fields

**TypeScript/React**:
- TypeScript strict mode enabled
- Functional components with hooks only
- Zustand state: `create((set, get) => ({ ... }))`
- Tailwind CSS for styling (no inline styles)
- Next.js 14+ App Router (no Pages Router)
- Import paths: absolute from project root

**Naming**:
- Go: `PascalCase` exports, `camelCase` private
- TS: `camelCase` functions/variables, `PascalCase` components/types/interfaces
- Constants: `UPPERCASE_SNAKE_CASE`
- Interfaces: `Readable`, `Valid`, `Service` (suffix with domain concept)
