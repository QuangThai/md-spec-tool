import type { useDiffSnapshots } from "@/hooks/useDiffSnapshots";
import type { useOutputActions } from "@/hooks/useOutputActions";
import type { useReviewGate } from "@/hooks/useReviewGate";
import type { MDFlowActions, MDFlowState } from "@/lib/mdflowStore";
import type { ConversionRecord, InputMode, OutputFormat } from "@/lib/types";
import type { mapErrorToUserFacing } from "@/lib/errorUtils";

export type FailedAction = "preview" | "convert" | "other" | null;

export interface WorkbenchDataFlows {
  mode: InputMode;
  pasteText: string;
  file: File | null;
  sheets: string[];
  selectedSheet: string;
  gsheetTabs: Array<{ title: string; gid: string }>;
  selectedGid: string;
  format: OutputFormat;
  mdflowOutput: string;
  warnings: MDFlowState["warnings"];
  meta: MDFlowState["meta"];
  loading: boolean;
  error: string | null;
  preview: MDFlowState["preview"];
  previewLoading: boolean;
  showPreview: boolean;
  columnOverrides: Record<string, string>;
  aiSuggestions: MDFlowState["aiSuggestions"];
  aiSuggestionsLoading: boolean;
  aiSuggestionsError: string | null;
  aiConfigured: boolean;
  setMode: MDFlowActions["setMode"];
  setPasteText: MDFlowActions["setPasteText"];
  setFile: MDFlowActions["setFile"];
  setSelectedSheet: MDFlowActions["setSelectedSheet"];
  setSelectedGid: MDFlowActions["setSelectedGid"];
  setFormat: MDFlowActions["setFormat"];
  setResult: MDFlowActions["setResult"];
  setShowPreview: MDFlowActions["setShowPreview"];
  history: ConversionRecord[];
  openaiKey: string;
  setOpenaiKey: (key: string) => void;
  clearOpenaiKey: () => void;
  templates: Array<string | { name: string; description?: string }>;
  showHistory: boolean;
  setShowHistory: (show: boolean) => void;
  showValidationConfigurator: boolean;
  setShowValidationConfigurator: (show: boolean) => void;
  showTemplateEditor: boolean;
  setShowTemplateEditor: (show: boolean) => void;
  showCommandPalette: boolean;
  setShowCommandPalette: (show: boolean) => void;
  showApiKeyInput: boolean;
  toggleApiKeyInput: () => void;
  apiKeyDraft: string;
  setApiKeyDraft: (draft: string) => void;
  lastFailedAction: FailedAction;
  openTemplateEditor: () => void;
  openValidationConfigurator: () => void;
  openHistory: () => void;
  openCommandPalette: () => void;
  diff: ReturnType<typeof useDiffSnapshots>;
  gsheetLoading: boolean;
  gsheetRange: string;
  setGsheetRange: (value: string) => void;
  googleAuth: {
    connected: boolean;
    loading: boolean;
    login: () => void;
    logout: () => void;
  };
  dragOver: boolean;
  handleFileChange: (e: React.ChangeEvent<HTMLInputElement>) => Promise<void>;
  onDrop: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragLeave: () => void;
  isInputGsheetUrl: boolean;
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  review: ReturnType<typeof useReviewGate>;
  canConfirmReviewGate: boolean;
  mappedAppError: ReturnType<typeof mapErrorToUserFacing> | null;
  handleRetryPreview: () => void;
  refetchGoogleSheetPreview: () => void;
  output: ReturnType<typeof useOutputActions>;
  handleConvert: () => Promise<void>;
  handleGetAISuggestions: () => Promise<void>;
  showFeedback: boolean;
  setShowFeedback: (show: boolean) => void;
}

export interface WorkbenchActionBundle {
  retryAISuggestions: () => void;
  saveSnapshot: () => void;
  compareSnapshots: () => void;
  clearSnapshots: () => void;
  handleRetryFailedAction: () => Promise<void>;
  togglePreview: () => void;
  handleCommandPaletteCopy: () => void;
  handleCommandPaletteExport: () => void;
  googleSheetInputProps: {
    gsheetLoading: boolean;
    gsheetRange: string;
    setGsheetRange: (value: string) => void;
    setSelectedGid: (gid: string) => void;
    googleAuth: {
      connected: boolean;
      loading: boolean;
      login: () => void;
      logout: () => void;
    };
    gsheetTabs: Array<{ title: string; gid: string }>;
    selectedGid: string;
    onRefetchGsheetPreview: () => void;
  };
  fileHandlingProps: {
    dragOver: boolean;
    handleFileChange: (e: React.ChangeEvent<HTMLInputElement>) => Promise<void>;
    onDrop: (e: React.DragEvent) => void;
    onDragOver: (e: React.DragEvent) => void;
    onDragLeave: () => void;
  };
}

export type WorkbenchDomainModel = WorkbenchDataFlows & WorkbenchActionBundle;
