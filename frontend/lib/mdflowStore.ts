import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import {
  AISuggestion,
  ConversionRecord,
  InputMode,
  MDFlowMeta,
  MDFlowWarning,
  OutputFormat,
  PreviewResponse,
  ValidationRules,
} from './types';

/**
 * Actions interface - grouped for single selector access
 * These functions never change, so subscribing to all of them has no performance impact.
 */
export interface MDFlowActions {
  // Input actions
  setMode: (mode: InputMode) => void;
  setPasteText: (text: string) => void;
  setFile: (file: File | null) => void;
  setSheets: (sheets: string[]) => void;
  setSelectedSheet: (sheet: string) => void;
  setGsheetTabs: (tabs: { title: string; gid: string }[]) => void;
  setSelectedGid: (gid: string) => void;
  setTemplate: (template: string) => void;
  setFormat: (format: OutputFormat) => void;

  // Output actions
  setResult: (output: string, warnings: MDFlowWarning[], meta: MDFlowMeta) => void;

  // Validation actions
  setValidationRules: (rules: ValidationRules) => void;

  // UI actions
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  dismissWarning: (code: string) => void;
  clearDismissedWarnings: () => void;

  // Preview actions
  setPreview: (preview: PreviewResponse | null) => void;
  setPreviewLoading: (loading: boolean) => void;
  setShowPreview: (show: boolean) => void;
  setColumnOverride: (column: string, field: string) => void;
  clearColumnOverrides: () => void;

  // AI Suggestions actions
  setAISuggestions: (suggestions: AISuggestion[], configured: boolean) => void;
  setAISuggestionsLoading: (loading: boolean) => void;
  setAISuggestionsError: (error: string | null) => void;
  clearAISuggestions: () => void;

  // Lifecycle actions
  reset: () => void;
}

/**
 * State interface - the actual data values
 */
export interface MDFlowState {
  // Input state
  mode: InputMode;
  pasteText: string;
  file: File | null;
  sheets: string[];
  selectedSheet: string;
  gsheetTabs: { title: string; gid: string }[];
  selectedGid: string;
  template: string;
  format: OutputFormat;

  // Output state
  mdflowOutput: string;
  warnings: MDFlowWarning[];
  meta: MDFlowMeta | null;

  // Preview state
  preview: PreviewResponse | null;
  previewLoading: boolean;
  showPreview: boolean;
  columnOverrides: Record<string, string>;

  // Validation state
  validationRules: ValidationRules;

  // AI Suggestions state
  aiSuggestions: AISuggestion[];
  aiSuggestionsLoading: boolean;
  aiSuggestionsError: string | null;
  aiConfigured: boolean | null;

  // UI state
  loading: boolean;
  error: string | null;
  dismissedWarningCodes: Record<string, boolean>;
}

/**
 * Combined store interface
 */
export interface MDFlowStore extends MDFlowState {
  actions: MDFlowActions;
}

const initialState: MDFlowState = {
  mode: 'paste',
  pasteText: '',
  file: null,
  sheets: [],
  selectedSheet: '',
  gsheetTabs: [],
  selectedGid: '',
  template: 'spec',
  format: 'spec',
  mdflowOutput: '',
  warnings: [],
  meta: null,
  preview: null,
  previewLoading: false,
  showPreview: false,
  columnOverrides: {},
  validationRules: {
    required_fields: [],
    format_rules: null,
    cross_field: [],
  },
  aiSuggestions: [],
  aiSuggestionsLoading: false,
  aiSuggestionsError: null,
  aiConfigured: null,
  loading: false,
  error: null,
  dismissedWarningCodes: {},
};

// Separate persisted state for history
interface PersistedState {
  history: ConversionRecord[];
}

const persistedInitialState: PersistedState = {
  history: [],
};

// Create history store with persistence
export const useHistoryStore = create<PersistedState & {
  addToHistory: (record: Omit<ConversionRecord, 'id' | 'timestamp'>) => void;
  clearHistory: () => void;
}>()(
  persist(
    (set) => ({
      ...persistedInitialState,
      addToHistory: (record) => set((state) => ({
        history: [
          {
            ...record,
            id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            timestamp: Date.now(),
          },
          ...state.history,
        ].slice(0, 50), // Keep only last 50 records
      })),
      clearHistory: () => set({ history: [] }),
    }),
    {
      name: 'mdflow-history',
    }
  )
);

/**
 * Main MDFlow store
 * 
 * State and actions are now separated:
 * - State values can be subscribed to individually using atomic selectors
 * - Actions are grouped in a single `actions` object for efficient access
 * 
 * Usage:
 * ```typescript
 * // Subscribe to individual state values (recommended)
 * const mode = useMDFlowStore((s) => s.mode);
 * const output = useMDFlowStore((s) => s.mdflowOutput);
 * 
 * // Get all actions with single selector (no performance impact)
 * const { setMode, setPasteText, reset } = useMDFlowStore((s) => s.actions);
 * ```
 * 
 * History is managed separately by useHistoryStore (persisted store)
 */
export const useMDFlowStore = create<MDFlowStore>((set) => ({
  ...initialState,

  actions: {
    // Input actions
    setMode: (mode) => set({
      mode,
      error: null,
      preview: null,
      showPreview: false,
      pasteText: '',
      file: null,
      sheets: [],
      selectedSheet: '',
    }),
    setPasteText: (pasteText) => set({ pasteText }),
    setFile: (file) => set({ file, sheets: [], selectedSheet: '', preview: null, showPreview: false }),
    setSheets: (sheets) => set({ sheets, selectedSheet: sheets[0] || '' }),
    setSelectedSheet: (selectedSheet) => set({ selectedSheet }),
    setGsheetTabs: (gsheetTabs) => set({ gsheetTabs, selectedGid: gsheetTabs[0]?.gid || '' }),
    setSelectedGid: (selectedGid) => set({ selectedGid }),
    setTemplate: (template) => set({ template }),
    setFormat: (format) => set({ format }),

    // Output actions
    setResult: (mdflowOutput, warnings, meta) => set({ mdflowOutput, warnings: warnings || [], meta }),

    // Validation actions
    setValidationRules: (validationRules) => set({ validationRules }),

    // UI actions
    setLoading: (loading) => set({ loading }),
    setError: (error) => set({ error }),
    dismissWarning: (code) => set((state) => ({
      dismissedWarningCodes: { ...state.dismissedWarningCodes, [code]: true },
    })),
    clearDismissedWarnings: () => set({ dismissedWarningCodes: {} }),

    // Preview actions
    setPreview: (preview) => set({ preview }),
    setPreviewLoading: (previewLoading) => set({ previewLoading }),
    setShowPreview: (showPreview) => set({ showPreview }),
    setColumnOverride: (column, field) => set((state) => ({
      columnOverrides: { ...state.columnOverrides, [column]: field },
    })),
    clearColumnOverrides: () => set({ columnOverrides: {} }),

    // AI Suggestions actions
    setAISuggestions: (aiSuggestions, aiConfigured) => set({ aiSuggestions, aiConfigured, aiSuggestionsError: null }),
    setAISuggestionsLoading: (aiSuggestionsLoading) => set({ aiSuggestionsLoading }),
    setAISuggestionsError: (aiSuggestionsError) => set({ aiSuggestionsError }),
    clearAISuggestions: () => set({ aiSuggestions: [], aiSuggestionsError: null }),

    // Lifecycle actions
    reset: () => set({ ...initialState }),
  },
}));

/**
 * Custom hook to get all actions at once.
 * Since actions never change, this is safe and performant.
 */
export const useMDFlowActions = () => useMDFlowStore((s) => s.actions);

// =============================================================================
// OpenAI BYOK (Bring Your Own Key) Store
// Persisted separately in localStorage so the key survives page reloads.
// =============================================================================

interface OpenAIKeyState {
  apiKey: string;
  setApiKey: (key: string) => void;
  clearApiKey: () => void;
}

export const useOpenAIKeyStore = create<OpenAIKeyState>()(
  persist(
    (set) => ({
      apiKey: '',
      setApiKey: (apiKey: string) => set({ apiKey }),
      clearApiKey: () => set({ apiKey: '' }),
    }),
    {
      name: 'mdflow-openai-key',
    }
  )
);
