# Product State — md-spec-tool

> Last updated: 2026-03-08
> Agent: content

## Product
- **Name**: md-spec-tool (Markdown Spec Tool — converts Excel/TSV to Markdown with AI)
- **Backend**: Go 1.24 + Gin + OpenAI v3 + Excelize + SQLite (modernc)
- **Frontend**: Next.js 16 + React 19 + TypeScript + Tailwind CSS 4 + Zustand 5 + TanStack React Query
- **Deploy**: Docker Compose (backend on port 8080, frontend on port 3000)

## Current Focus
<!-- Updated by brainstorm/dev agents -->
- [x] **SPEC-001**: AI Pipeline Logic Improvement (P1) — parse-first unification, AI+fallback merge, stream options — released ✅
- [ ] Document & clarify transcript upload validation rules (P1)
- [ ] Upload page UI/UX improvements (P1) — ErrorAlert label, touch targets, layout

## Recent Decisions
<!-- Auto-appended by kd-handoff-spec -->
- **2026-03-08**: SPEC-001 approved. detect_only: when parse succeeds (hasTable) return "table"; else DetectInputType.
- **2026-03-08**: SPEC-001 released. Parse-first, AI merge, prompt injection defense, stream options shipped. (content generated)

## Active Specs
<!-- Auto-appended by kd-handoff-spec, removed by kd-release -->
- ~~SPEC-001: AI Pipeline Logic Improvement (P1, approved)~~ released ✅

## Quality Metrics
<!-- Updated by kd-qa -->
- Backend vet: `go vet ./...`
- Backend tests: `go test ./...`
- Backend build: `go build ./...`
- Frontend build: `npm run build`
- Frontend tests: `npm test`
