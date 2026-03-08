# AGENTS.md - md-spec-tool/frontend

## Build & Test
- **Dev**: `npm run dev` (Next.js on port 3000, backend on 8080)
- **Build**: `npm run build`
- **Test all**: `npm test` (Vitest, excludes e2e)
- **Test single**: `npx vitest run path/to/file.test.ts`
- **E2E**: `npm run test:e2e` (Playwright)

## Architecture
- **Stack**: Next.js 16 + React 19 + TypeScript (strict) + Tailwind CSS 4 + Zustand 5 + TanStack React Query
- **`app/`**: App Router pages — routes include `/batch`, `/dashboard`, `/docs`, `/gallery`, `/share`, `/studio`, `/transcribe`, `/oauth`
- **`components/`**: React components; subdirs `ui/` (primitives), `workbench/`, `remotion/`, `templateEditor/`
- **`lib/`**: API clients (`httpClient.ts`, `mdflowApi.ts`, `shareApi.ts`, `quotaApi.ts`), Zustand stores (`mdflowStore.ts`, `toastStore.ts`), types, utilities
- **`hooks/`**: Custom hooks — `useWorkbench*` family for workbench state, `useConversionFlow`, `useFileHandling`, `useGoogleAuth`, etc.
- **`constants/`**: App constants; `@/*` path alias maps to project root

## Code Style
- TypeScript strict mode; functional components with hooks only
- Tailwind CSS for styling (no inline styles); use `clsx`/`tailwind-merge` for conditional classes
- Zustand stores: `create((set, get) => ({ ... }))`; React Query for server state (`lib/mdflowQueries.ts`, `lib/shareQueries.ts`)
- Imports: use `@/` absolute paths (e.g., `@/lib/types`, `@/components/ui/Button`)
- Naming: `PascalCase` components/types, `camelCase` functions/variables, `UPPER_SNAKE_CASE` constants
- UI primitives from Radix UI (`@radix-ui/*`); icons from `lucide-react`
