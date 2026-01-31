# AGENTS.md - md-spec-tool

## Build, Test & Lint Commands

### Backend (Go)
- **Test single package**: `cd backend && go test ./path/to/package`
- **Test all**: `cd backend && go test ./...`
- **Run dev server**: `make dev-backend` (runs `go run ./cmd/server`)
- **Build**: `docker-compose build`

### Frontend (Next.js)
- **Dev server**: `make dev-frontend` (runs `npm run dev`)
- **Build**: `cd frontend && npm run build`
- **Test**: `cd frontend && npm test` (currently not configured)

### Docker
- **Start services**: `make up`
- **Stop services**: `make down`
- **View logs**: `make logs`

## Architecture & Codebase

**Tech Stack**:
- **Backend**: Go 1.20+ + Gin web framework (stateless conversion API)
- **Frontend**: Next.js 14 + React 18 + TypeScript + Tailwind CSS + Zustand

**Key Directories**:
- `backend/cmd/server/`: Entry point (main.go)
- `backend/internal/`: Business logic
  - `http/`: Routes & handlers (HTTP transport layer)
  - `converter/`: Excel→Markdown + Markdown diff conversion logic
  - `config/`: Configuration management
- `frontend/app/`: Next.js app router
- `frontend/components/`: React components
- `frontend/lib/`: Utilities

**Ports**: Frontend 3000, Backend 8080

## Code Style & Guidelines

**Go**:
- Standard Go conventions (CamelCase, error handling with explicit checks)
- Error handling: always check and propagate errors
- Use Gin middleware for HTTP handling
- Package structure: `internal/` for private packages

**TypeScript/React**:
- TypeScript strict mode
- Functional components with hooks
- Zustand for state management
- Tailwind CSS for styling
- App router pattern (Next.js 14)

**Imports**: Use relative imports within packages; absolute imports for cross-package

**Naming**: 
- Go: PascalCase for exported, camelCase for unexported
- TS: camelCase functions/vars, PascalCase components/types
