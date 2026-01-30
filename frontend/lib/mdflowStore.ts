import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { MDFlowMeta, MDFlowWarning, PreviewResponse, ValidationRules } from './mdflowApi';

export type InputMode = 'paste' | 'xlsx' | 'tsv';

// History record for conversion history feature
export interface ConversionRecord {
  id: string;
  timestamp: number;
  mode: InputMode;
  template: string;
  inputPreview: string;
  output: string;
  meta: MDFlowMeta | null;
}

// Custom template record
export interface CustomTemplate {
  id: string;
  name: string;
  content: string;
  createdAt: number;
  updatedAt: number;
}

interface MDFlowStore {
  mode: InputMode;
  pasteText: string;
  file: File | null;
  sheets: string[];
  selectedSheet: string;
  template: string;
  
  mdflowOutput: string;
  warnings: MDFlowWarning[];
  meta: MDFlowMeta | null;
  
  // Preview state
  preview: PreviewResponse | null;
  previewLoading: boolean;
  showPreview: boolean;
  columnOverrides: Record<string, string>;
  
  // History state
  history: ConversionRecord[];
  
  // Validation rules (custom validation configurator)
  validationRules: ValidationRules;
  
  loading: boolean;
  error: string | null;
  
  dismissedWarningCodes: Record<string, boolean>;
  
  setMode: (mode: InputMode) => void;
  setPasteText: (text: string) => void;
  setFile: (file: File | null) => void;
  setSheets: (sheets: string[]) => void;
  setSelectedSheet: (sheet: string) => void;
  setTemplate: (template: string) => void;
  setResult: (output: string, warnings: MDFlowWarning[], meta: MDFlowMeta) => void;
  setValidationRules: (rules: ValidationRules) => void;
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
  
  // History actions
  addToHistory: (record: Omit<ConversionRecord, 'id' | 'timestamp'>) => void;
  clearHistory: () => void;
  
  reset: () => void;
}

const initialState = {
  mode: 'paste' as InputMode,
  pasteText: '',
  file: null as File | null,
  sheets: [] as string[],
  selectedSheet: '',
  template: 'default',
  mdflowOutput: '',
  warnings: [] as MDFlowWarning[],
  meta: null as MDFlowMeta | null,
  preview: null as PreviewResponse | null,
  previewLoading: false,
  showPreview: false,
  columnOverrides: {} as Record<string, string>,
  validationRules: {
    required_fields: [],
    format_rules: null,
    cross_field: [],
  } as ValidationRules,
  loading: false,
  error: null as string | null,
  dismissedWarningCodes: {} as Record<string, boolean>,
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

export const useMDFlowStore = create<MDFlowStore>((set) => ({
  ...initialState,
  history: [],
  
  setMode: (mode) => set({ mode, error: null, preview: null, showPreview: false }),
  setPasteText: (pasteText) => set({ pasteText }),
  setFile: (file) => set({ file, sheets: [], selectedSheet: '', preview: null, showPreview: false }),
  setSheets: (sheets) => set({ sheets, selectedSheet: sheets[0] || '' }),
  setSelectedSheet: (selectedSheet) => set({ selectedSheet }),
  setTemplate: (template) => set({ template }),
  setResult: (mdflowOutput, warnings, meta) => set({ mdflowOutput, warnings: warnings || [], meta }),
  setValidationRules: (validationRules) => set({ validationRules }),
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
  
  // History actions (delegated to persisted store)
  addToHistory: () => {},
  clearHistory: () => {},
  
  reset: () => set({ ...initialState, history: [], validationRules: initialState.validationRules }),
}));
