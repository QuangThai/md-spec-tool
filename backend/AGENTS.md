# AGENTS.md - backend (Go 1.24+, Gin, OpenAI v3)

## Build & Test
- **Test all**: `go test ./...`
- **Test single package**: `go test ./internal/ai`
- **Test single test**: `go test -run TestMyFunc ./internal/ai`
- **Run server**: `go run ./cmd/server` (port 8080)
- **Build CLI**: `go build -o ../bin/mdflow ./cmd/cli`

## Architecture
- `cmd/server/` — HTTP server entry point (Gin, port 8080)
- `cmd/cli/` — CLI tool entry point
- `cmd/evalrunner/`, `cmd/usecases/`, `cmd/usecases_diff/` — evaluation & use-case runners
- `internal/http/` — routes, handlers, middleware (Gin)
- `internal/ai/` — OpenAI integration (mapping, validation, refinement, caching)
- `internal/converter/` — Excel→Markdown conversion (Excelize)
- `internal/diff/` — diff computation (go-difflib)
- `internal/config/` — env-var config (godotenv), `LoadConfig()`
- `internal/share/`, `internal/suggest/`, `internal/feedback/`, `internal/quota/`, `internal/gsheetutils/` — domain services

## Code Style
- Module: `github.com/yourorg/md-spec-tool`; use absolute imports
- Errors: always check and wrap with `fmt.Errorf("context: %w", err)`
- Logging: `log/slog` (`slog.Info`, `slog.Error`); no `fmt.Println` or `log.Fatal` in library code
- HTTP handlers: `func(c *gin.Context)`, return `c.JSON(status, body)`
- Naming: `PascalCase` exports, `camelCase` unexported; `UPPER_SNAKE` constants
