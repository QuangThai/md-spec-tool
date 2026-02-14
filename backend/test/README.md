# Backend Test Layout

This folder centralizes backend Go tests that were previously spread across `internal/*` packages.

## Structure

- `test/ai/` tests for `internal/ai`
- `test/converter/` tests for `internal/converter`
- `test/handlers/` tests for `internal/http/handlers`
- `test/suggest/` tests for `internal/suggest`

Each subfolder maps to one internal package area.

## Conventions

- Use external test packages (`<pkg>_test`) in this folder.
- Import tested package APIs from `internal/...`.
- Keep tests black-box by default (test exported behavior first).
- If a test must verify internal behavior, use explicit test export wrappers in the corresponding package:
  - `internal/ai/test_exports.go`
  - `internal/converter/test_exports.go`
  - `internal/http/handlers/test_exports.go`
  - `internal/suggest/test_exports.go`
- Do not expose production-only internals without clear test need.

## Running Tests

From `backend/`:

```bash
go test ./...
```

Run a specific test package:

```bash
go test ./test/converter
```

## Adding New Tests

1. Add test files under `backend/test/<domain>/`.
2. Use `<domain>_test` package naming.
3. Reuse existing wrappers in `test_exports.go` when needed.
4. Run `go test ./...` before opening PR.
