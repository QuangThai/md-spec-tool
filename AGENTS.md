# AGENTS.md - Development Guidelines

## Commands

### Build & Test
- **Backend tests (all)**: `make test` or `cd backend && go test ./...`
- **Backend single test**: `cd backend && go test -run TestName ./...`
- **Backend build**: `go build -o main ./cmd/server`
- **Frontend build**: `cd frontend && npm run build`
- **Frontend tests**: `cd frontend && npm test` (not yet configured)

### Development
- **Dev backend**: `make dev-backend` → runs `cd backend && go run ./cmd/server`
- **Dev frontend**: `make dev-frontend` → runs `cd frontend && npm run dev`
- **All services**: `make up` → Docker Compose (db, backend, frontend)
- **Logs**: `make logs` → `docker-compose logs -f`

## Architecture

**Full-Stack**: Go 1.20 + Gin backend, Next.js 14 + React frontend, PostgreSQL 15

**Backend** (`/backend`): Entry point `cmd/server/main.go`, layered: `internal/{http,services,repositories,models,converters,config}`
- HTTP handlers, JWT auth, Excel parsing (excelize), Markdown generation (text/template)
- PostgreSQL with migrations in `migrations/`

**Frontend** (`/frontend`): Next.js App Router, Tailwind CSS, Zustand state, Zod validation
- Path alias: `@/*` maps to project root

**Database**: PostgreSQL 15, auto-migrated on startup

**Key APIs**: Auth (JWT login), Import (Excel), Convert (Markdown), Spec (CRUD), Sharing/Comments/Activity

## Style

**Go**: CamelCase, `type StructName struct{}`, receiver functions `(r *Type) Method()`, error returns, organized by layer
**TypeScript**: Strict mode, ES2020+, `@/` imports, React 18, component pattern, Zustand stores
**Frontend**: Tailwind for styling, `ErrorBoundary` component error handling, custom fonts (Sora, Fraunces, Space Grotesk)
