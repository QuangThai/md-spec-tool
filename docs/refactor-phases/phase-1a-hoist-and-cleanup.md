# Phase 1a: Hoist Shared State + Delete Dead Code

> **Prerequisite**: Nothing  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## Step 1: Delete dead share code

The following state/logic is **never referenced in JSX** — the UI uses `<ShareButton />` component (L1974–L1978) which self-manages share logic.

**Delete these lines from `MDFlowWorkbench.tsx`:**

| What | Line(s) |
|------|---------|
| `creatingShare` state | L199 |
| `shareTitle` state | L200 |
| `shareSlug` state | L201 |
| `shareVisibility` state | L202–L204 |
| `shareAllowComments` state | L205 |
| `showShareOptions` state | L206 |
| `shareSlugError` state | L207 |
| `shareOptionsRef` ref | L208 |
| `handleCreateShare` callback | L879–L946 |
| click-outside effect | L948–L958 |

**⚠️ After deletion, all subsequent line numbers in the plan will shift. Re-verify before continuing.**

## Step 2: Hoist shared state to orchestrator level

These already exist at the right place (orchestrator level) — just **mark them mentally** as "hoisted, will be passed to hooks":

```tsx
// Already at L216 (post-delete line will shift) — stays in orchestrator
const [lastFailedAction, setLastFailedAction] = useState<"preview" | "convert" | "other" | null>(null);

// Already at L209 — stays in orchestrator
const [debouncedPasteText, setDebouncedPasteText] = useState("");
```

The debounce effect (L334–L345) stays in orchestrator:

```tsx
useEffect(() => {
  if (mode !== "paste") { setDebouncedPasteText(""); return; }
  const timer = setTimeout(() => setDebouncedPasteText(pasteText), 500);
  return () => clearTimeout(timer);
}, [pasteText, mode]);
```

Derived values (L244–L247) stay in orchestrator:

```tsx
const isGsheetUrl = isGoogleSheetsURL(debouncedPasteText.trim());
const isInputGsheetUrl = isGoogleSheetsURL(pasteText.trim());
const inputSource = mode === "paste" ? (isInputGsheetUrl ? "gsheet" : "paste") : mode;
```

## Step 3: Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: app loads, paste → preview → convert → copy all still work
- [ ] `<ShareButton />` still works (it's self-contained)

## Commit

```
git add frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): delete dead share code, mark hoisted state"
```
