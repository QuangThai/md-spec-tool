# Phase 2: Extract UI Components

> **Prerequisite**: Phase 1 complete (all hooks extracted, orchestrator clean)  
> **Commit after**: Yes (one commit for all components)  
> **Build check**: `cd frontend && npm run build && npm test`

---

## Goal

Split ~1050 lines of JSX into composable, memoizable components under `frontend/components/workbench/`.

**Rules**:
- No business logic in components — only presentation + event forwarding
- Props typed with explicit interfaces
- Components receive data via props, not direct store subscriptions (exception: Zustand store values already read in orchestrator)

---

## Extraction order

Extract bottom-up (leaf components first, then composing parents):

### Step 1: Leaf components (no children to extract)

#### 1a. `ErrorBanner` (~40 lines, L1271–L1307)

```tsx
interface ErrorBannerProps {
  error: string;
  mappedAppError: { title?: string; message?: string; requestId?: string; retryable?: boolean } | null;
  lastFailedAction: "preview" | "convert" | "other" | null;
  loading: boolean;
  previewLoading: boolean;
  onRetry: () => void;
}
```

#### 1b. `ReviewGateBanner` (~75 lines, L1309–L1383)

```tsx
interface ReviewGateBannerProps {
  reviewRequiredColumns: string[];
  reviewedColumns: Record<string, boolean>;
  reviewRemainingCount: number;
  onToggleColumn: (column: string) => void;
  onMarkAll: () => void;
  onConfirmReview: () => void;
  canConfirm: boolean;
}
```

#### 1c. `OutputContent` (~25 lines, L1993–L2019)

```tsx
interface OutputContentProps {
  loading: boolean;
  mdflowOutput: string;
}
```

#### 1d. `DiffModal` (L2043–L2080)

```tsx
interface DiffModalProps {
  showDiff: boolean;
  currentDiff: any;
  onClose: () => void;
}
```

### Step 2: Composite leaf components

#### 2a. `ApiKeyPanel` (~70 lines, L1199–L1267)

```tsx
interface ApiKeyPanelProps {
  show: boolean;
  openaiKey: string;
  apiKeyDraft: string;
  onDraftChange: (v: string) => void;
  onSave: () => void;
  onClear: () => void;
}
```

#### 2b. `PasteInput` (~170 lines, L1385–L1561)

```tsx
interface PasteInputProps {
  pasteText: string;
  onPasteTextChange: (v: string) => void;
  isInputGsheetUrl: boolean;
  gsheetTabs: Array<{ title: string; gid: string }>;
  selectedGid: string;
  onSelectGid: (v: string) => void;
  gsheetLoading: boolean;
  gsheetRange: string;
  onGsheetRangeChange: (v: string) => void;
  googleAuth: { connected: boolean; loading: boolean; login: () => void; logout: () => void };
  preview: any;
  showPreview: boolean;
  onTogglePreview: () => void;
  previewLoading: boolean;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
  onRefetchGsheetPreview: () => void;
}
```

#### 2c. `FileUploadInput` (~160 lines, L1562–L1726)

```tsx
interface FileUploadInputProps {
  mode: "xlsx" | "tsv";
  file: File | null;
  sheets: string[];
  selectedSheet: string;
  onSelectSheet: (v: string) => void;
  dragOver: boolean;
  onDragOver: (e: React.DragEvent) => void;
  onDragLeave: () => void;
  onDrop: (e: React.DragEvent) => void;
  onFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  preview: any;
  showPreview: boolean;
  previewLoading: boolean;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
}
```

### Step 3: Panel headers & footers

#### 3a. `SourcePanelHeader` (~70 lines, L1131–L1197)

```tsx
interface SourcePanelHeaderProps {
  mode: "paste" | "xlsx" | "tsv";
  onModeChange: (mode: "paste" | "xlsx" | "tsv") => void;
  openaiKey: string;
  showApiKeyInput: boolean;
  onToggleApiKey: () => void;
  onOpenTemplateEditor: () => void;
  onOpenValidation: () => void;
}
```

#### 3b. `WorkbenchFooter` (~65 lines, L1730–L1794)

```tsx
interface WorkbenchFooterProps {
  format: string;
  onFormatChange: (v: string) => void;
  onConvert: () => void;
  loading: boolean;
  disabled: boolean;
}
```

#### 3c. `OutputToolbar` (~155 lines, L1808–L1991)

```tsx
interface OutputToolbarProps {
  // Output state
  mdflowOutput: string;
  meta: any;
  // Review state
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
  reviewGateReason: string | undefined;
  // Copy
  copied: boolean;
  onCopy: () => void;
  // Snapshots
  snapshotA: string;
  snapshotB: string;
  onSaveSnapshot: () => void;
  onCompareSnapshots: () => void;
  onClearSnapshots: () => void;
  // Export
  onDownload: () => void;
  // Share
  format: string;
  // History
  historyCount: number;
  onShowHistory: () => void;
}
```

### Step 4: Panel composites

#### 4a. `SourcePanel` → composes Header + ApiKey + Error + ReviewGate + PasteInput/FileUploadInput + Footer

#### 4b. `OutputPanel` → composes OutputToolbar + OutputContent + TechnicalAnalysis footer

---

## File structure

```
frontend/components/workbench/
├── SourcePanel.tsx          (composes all source sub-components)
├── SourcePanelHeader.tsx
├── ApiKeyPanel.tsx
├── ErrorBanner.tsx
├── ReviewGateBanner.tsx
├── PasteInput.tsx
├── FileUploadInput.tsx
├── WorkbenchFooter.tsx
├── OutputPanel.tsx          (composes output sub-components)
├── OutputToolbar.tsx
├── OutputContent.tsx
└── DiffModal.tsx
```

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] All visual elements identical to before
- [ ] No business logic leaked into components

## Commit

```
git add frontend/components/workbench/ frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): Phase 2 complete — extract UI components"
```
