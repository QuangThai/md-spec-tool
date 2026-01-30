import { create } from 'zustand';
import { MDFlowMeta } from './mdflowApi';

export type InputMode = 'paste' | 'xlsx' | 'tsv';

interface MDFlowStore {
  mode: InputMode;
  pasteText: string;
  file: File | null;
  sheets: string[];
  selectedSheet: string;
  template: string;
  
  mdflowOutput: string;
  warnings: string[];
  meta: MDFlowMeta | null;
  
  loading: boolean;
  error: string | null;
  
  setMode: (mode: InputMode) => void;
  setPasteText: (text: string) => void;
  setFile: (file: File | null) => void;
  setSheets: (sheets: string[]) => void;
  setSelectedSheet: (sheet: string) => void;
  setTemplate: (template: string) => void;
  setResult: (output: string, warnings: string[], meta: MDFlowMeta) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialState = {
  mode: 'paste' as InputMode,
  pasteText: '',
  file: null,
  sheets: [],
  selectedSheet: '',
  template: 'default',
  mdflowOutput: '',
  warnings: [],
  meta: null,
  loading: false,
  error: null,
};

export const useMDFlowStore = create<MDFlowStore>((set) => ({
  ...initialState,
  
  setMode: (mode) => set({ mode, error: null }),
  setPasteText: (pasteText) => set({ pasteText }),
  setFile: (file) => set({ file, sheets: [], selectedSheet: '' }),
  setSheets: (sheets) => set({ sheets, selectedSheet: sheets[0] || '' }),
  setSelectedSheet: (selectedSheet) => set({ selectedSheet }),
  setTemplate: (template) => set({ template }),
  setResult: (mdflowOutput, warnings, meta) => set({ mdflowOutput, warnings: warnings || [], meta }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  reset: () => set(initialState),
}));
