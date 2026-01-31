import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import {
  AISuggestion,
  ConversionRecord,
  InputMode,
  MDFlowMeta,
  MDFlowWarning,
  PreviewResponse,
  ValidationRules,
} from './types';

/**
 * Main app state store
 * Does NOT include history (use useHistoryStore for that)
 */
export interface MDFlowStore {
  // Input state
  mode: InputMode;
  pasteText: string;
  file: File | null;
  sheets: string[];
  selectedSheet: string;
  template: string;

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

  // Actions: Input
  setMode: (mode: InputMode) => void;
  setPasteText: (text: string) => void;
  setFile: (file: File | null) => void;
  setSheets: (sheets: string[]) => void;
  setSelectedSheet: (sheet: string) => void;
  setTemplate: (template: string) => void;

  // Actions: Output
  setResult: (output: string, warnings: MDFlowWarning[], meta: MDFlowMeta) => void;

  // Actions: Validation
  setValidationRules: (rules: ValidationRules) => void;

  // Actions: UI
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  dismissWarning: (code: string) => void;
  clearDismissedWarnings: () => void;

  // Actions: Preview
  setPreview: (preview: PreviewResponse | null) => void;
  setPreviewLoading: (loading: boolean) => void;
  setShowPreview: (show: boolean) => void;
  setColumnOverride: (column: string, field: string) => void;
  clearColumnOverrides: () => void;

  // Actions: AI Suggestions
  setAISuggestions: (suggestions: AISuggestion[], configured: boolean) => void;
  setAISuggestionsLoading: (loading: boolean) => void;
  setAISuggestionsError: (error: string | null) => void;
  clearAISuggestions: () => void;

  // Actions: Lifecycle
  reset: () => void;
}

const initialState: Omit<MDFlowStore, 'setMode' | 'setPasteText' | 'setFile' | 'setSheets' | 'setSelectedSheet' | 'setTemplate' | 'setResult' | 'setValidationRules' | 'setLoading' | 'setError' | 'dismissWarning' | 'clearDismissedWarnings' | 'setPreview' | 'setPreviewLoading' | 'setShowPreview' | 'setColumnOverride' | 'clearColumnOverrides' | 'setAISuggestions' | 'setAISuggestionsLoading' | 'setAISuggestionsError' | 'clearAISuggestions' | 'reset'> = {
  mode: 'paste',
  pasteText: '',
  file: null,
  sheets: [],
  selectedSheet: '',
  template: 'default',
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
 * Manages input, output, preview, validation, and AI suggestion state
 * 
 * History is managed separately by useHistoryStore (persisted store)
 * Use Zustand selectors to prevent unnecessary re-renders:
 * 
 * const output = useMDFlowStore((s) => s.mdflowOutput);
 * const setTemplate = useMDFlowStore((s) => s.setTemplate);
 */
export const useMDFlowStore = create<MDFlowStore>((set) => ({
  ...initialState,

  // Input actions
  setMode: (mode) => set({ mode, error: null, preview: null, showPreview: false }),
  setPasteText: (pasteText) => set({ pasteText }),
  setFile: (file) => set({ file, sheets: [], selectedSheet: '', preview: null, showPreview: false }),
  setSheets: (sheets) => set({ sheets, selectedSheet: sheets[0] || '' }),
  setSelectedSheet: (selectedSheet) => set({ selectedSheet }),
  setTemplate: (template) => set({ template }),

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
  reset: () => set({ ...initialState, validationRules: initialState.validationRules }),
}));
