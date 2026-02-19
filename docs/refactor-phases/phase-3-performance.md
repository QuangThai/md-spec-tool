# Phase 3: Performance Optimization

> **Prerequisite**: Phase 2 complete (UI components extracted)  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## Goal

Apply Vercel React Best Practices now that structure is clean.

---

## Optimizations

### 3a. `React.memo()` wrappers

Wrap these components (they receive stable props from orchestrator):

| Component | Why |
|-----------|-----|
| `SourcePanel` | Large subtree, only re-renders when source-related props change |
| `OutputPanel` | Large subtree, only re-renders when output-related props change |
| `OutputToolbar` | Many buttons, only re-renders when output/snapshot state changes |
| `DiffModal` | Only visible when `showDiff` is true |

### 3b. `next/dynamic` for toggled panels

These components are only rendered when a boolean toggle is true:

```tsx
const ApiKeyPanel = dynamic(() => import("./workbench/ApiKeyPanel"), { ssr: false });
const ReviewGateBanner = dynamic(() => import("./workbench/ReviewGateBanner"), { ssr: false });
const DiffModal = dynamic(() => import("./workbench/DiffModal"), { ssr: false });
```

### 3c. `rerender-defer-reads` (already done in Phase 1c)

`handleCopy` and `handleDownload` already read from `useMDFlowStore.getState()` instead of subscribing.

### 3d. `rerender-functional-setstate`

Audit all `setState` calls — use functional form where dependent on previous value:

```tsx
// ✅ Already correct in useReviewGate:
setReviewedColumns((prev) => ({ ...prev, [column]: true }));

// Audit remaining setState calls for similar patterns
```

### 3e. `rendering-conditional-render`

Convert `&&` patterns to ternary in JSX:

```tsx
// Before:
{error && <ErrorBanner ... />}

// After:
{error ? <ErrorBanner ... /> : null}
```

### 3f. `rerender-memo-with-default-value`

Hoist non-primitive default props to module scope:

```tsx
// Module-level constants
const EMPTY_ARRAY: string[] = [];
const EMPTY_OBJECT: Record<string, string> = {};

// Use in props instead of inline `[]` or `{}`
```

---

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] React DevTools Profiler: no unnecessary re-renders when typing in paste area
- [ ] React DevTools Profiler: OutputPanel doesn't re-render when SourcePanel state changes
- [ ] Bundle size: `npm run build` → compare JS bundle sizes (no regression)

## Commit

```
git add frontend/components/
git commit -m "refactor(workbench): Phase 3 complete — performance optimizations"
```
