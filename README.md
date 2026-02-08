# MD-Spec-Tool

MD-Spec-Tool converts spreadsheet and pasted tabular data into consistent Markdown specifications (MDFlow), with optional AI-assisted mapping, preview, diffing, Google Sheets import, and sharing workflows.

## Features

- Multi-input conversion: paste text, `.tsv`, `.xlsx`, and Google Sheets URLs.
- Canonical output formats: `spec` (structured requirements/spec) and `table` (clean markdown table).
- Smart parsing pipeline: input detection, header detection, column mapping, and warning metadata.
- AI support with safe fallback: optional OpenAI mapping/suggestions, plus rule-based degraded mode.
- BYOK (Bring Your Own Key): send `X-OpenAI-API-Key` per request without server-side key storage.
- Collaboration features: share links, public listing, and comment threads.
- Studio UX: live preview, batch conversion, template preview, diff viewer, and history.

## Tech Stack

### Backend

- Go `1.24` + Gin
- OpenAI integration via `openai-go/v3`
- Google Sheets integration (service account + OAuth bearer token)
- Converter pipeline in `backend/internal/converter`
- API handlers in `backend/internal/http/handlers`

### Frontend

- Next.js `16` + React `19` + TypeScript
- Tailwind CSS `4`
- Zustand `5` + TanStack Query `5`
- Next API routes for Google OAuth/session-aware gsheet proxying

## Project Structure

```text
md-spec-tool/
├── backend/
│   ├── cmd/
│   │   ├── server/              # HTTP API server
│   │   ├── cli/                 # mdflow CLI
│   │   ├── usecases/            # use-case conversion checker
│   │   └── usecases_diff/       # use-case diff/coverage checker
│   └── internal/
│       ├── ai/                  # AI client/service, mapping, validation
│       ├── converter/           # parsing, mapping, rendering pipeline
│       ├── http/                # router, middleware, handlers
│       ├── share/               # share store + comments
│       └── suggest/             # AI suggestion logic
├── frontend/
│   ├── app/                     # Next.js app router (studio, docs, batch, share)
│   ├── components/
│   ├── hooks/
│   └── lib/
├── use-cases/                   # sample inputs and generated output checks
├── Makefile
├── AGENTS.md
└── README.md
```

## Quick Start (Local)

### Prerequisites

- Go `1.24+`
- Node.js `20+`
- npm

### 1) Run backend

```bash
cd backend
go mod download
go run ./cmd/server
```

Backend runs at `http://localhost:8080`.

### 2) Run frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs at `http://localhost:3000`.

## API Overview

Terminology note: requests accept both `template` and `format` as aliases for output mode (`spec` or `table`).

### Health

- `GET /health`

### Conversion & Preview

- `POST /api/mdflow/paste` (JSON: `paste_text`, `template?`, `format?`, `?detect_only=true`)
- `POST /api/mdflow/preview` (JSON: `paste_text`, `template?`, `format?`, `?skip_ai=false`)
- `POST /api/mdflow/tsv` (multipart: `file`, `template?`, `format?`)
- `POST /api/mdflow/tsv/preview` (multipart: `file`, `template?`, `format?`, `?skip_ai=false`)
- `POST /api/mdflow/xlsx` (multipart: `file`, `sheet_name?`, `template?`, `format?`)
- `POST /api/mdflow/xlsx/preview` (multipart: `file`, `sheet_name?`, `template?`, `format?`, `?skip_ai=false`)
- `POST /api/mdflow/xlsx/sheets` (multipart: `file`)

### Templates & Validation

- `GET /api/mdflow/templates`
- `GET /api/mdflow/templates/info`
- `GET /api/mdflow/templates/:name`
- `POST /api/mdflow/templates/preview` (JSON: `template_content`, `sample_data?`)
- `POST /api/mdflow/validate` (JSON: `paste_text`, `validation_rules?`, `template?`)

### Diff & AI

- `POST /api/mdflow/diff` (JSON: `before`, `after`)
- `POST /api/mdflow/ai/suggest` (JSON: `paste_text`, `template?`)

### Google Sheets

- `POST /api/mdflow/gsheet` (JSON: `url`, `gid?`)
- `POST /api/mdflow/gsheet/sheets` (JSON: `url`)
- `POST /api/mdflow/gsheet/preview` (JSON: `url`, `template?`, `gid?`)
- `POST /api/mdflow/gsheet/convert` (JSON: `url`, `template?`, `format?`, `gid?`)

### Share API

- `POST /api/share`
- `GET /api/share/public`
- `GET /api/share/:key`
- `PATCH /api/share/:key`
- `GET /api/share/:key/comments`
- `POST /api/share/:key/comments`
- `PATCH /api/share/:key/comments/:commentId`

## CLI

Build CLI:

```bash
make cli
```

Examples:

```bash
./bin/mdflow convert --input spec.tsv --output spec.mdflow.md --template spec
./bin/mdflow convert --input data.xlsx --sheet "Sheet1" --template table
./bin/mdflow diff before.md after.md --json
./bin/mdflow templates
```

## Useful Commands

```bash
make test           # backend tests (go test ./...)
make dev-backend    # run backend server
make dev-frontend   # run frontend dev server
make cli            # build mdflow CLI
make install-cli    # install CLI to /usr/local/bin/mdflow
```

## Environment Variables

Copy from `.env.example` and adjust values.

Core:

- `HOST`, `PORT`, `CORS_ORIGINS`
- `MAX_UPLOAD_BYTES`, `MAX_PASTE_BYTES`
- `HTTP_CLIENT_TIMEOUT`

AI:

- `OPENAI_API_KEY` (optional)
- `OPENAI_MODEL`
- `AI_REQUEST_TIMEOUT`, `AI_MAX_RETRIES`, `AI_CACHE_TTL`, `AI_MAX_CACHE_SIZE`, `AI_RETRY_BASE_DELAY`
- `AI_PREVIEW_TIMEOUT`, `AI_PREVIEW_MAX_RETRIES`

Google Sheets / OAuth:

- `GOOGLE_APPLICATION_CREDENTIALS` (backend service account path; optional)
- `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET` (frontend OAuth)
- `COOKIE_SECRET` (required for encrypted OAuth session cookie)

Frontend:

- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_APP_URL`

Share store:

- `SHARE_STORE_PATH` (optional persisted storage path)

## Notes

- Supported output modes are strictly `spec` and `table`.
- Prefer `template` for consistency; `format` is kept as a backward-compatible alias.
- Preview defaults to rule-based mapping for speed (`skip_ai=true`) and can opt-in AI mapping per request.
- AI endpoints remain safe when unconfigured; they return structured non-fatal responses.
