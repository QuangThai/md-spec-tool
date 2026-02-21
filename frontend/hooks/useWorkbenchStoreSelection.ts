import { useMDFlowTemplatesQuery } from "@/lib/mdflowQueries";
import {
  useHistoryStore,
  useMDFlowActions,
  useMDFlowStore,
  useOpenAIKeyStore,
  type MDFlowState,
} from "@/lib/mdflowStore";
import { useShallow } from "zustand/react/shallow";

const EMPTY_TEMPLATES: string[] = [];

export function useWorkbenchStoreSelection() {
  const state = useMDFlowStore(
    useShallow(
      (s): Omit<
        MDFlowState,
        "validationRules" | "dismissedWarningCodes" | "template"
      > => ({
        mode: s.mode,
        pasteText: s.pasteText,
        file: s.file,
        sheets: s.sheets,
        selectedSheet: s.selectedSheet,
        gsheetTabs: s.gsheetTabs,
        selectedGid: s.selectedGid,
        format: s.format,
        mdflowOutput: s.mdflowOutput,
        warnings: s.warnings,
        meta: s.meta,
        loading: s.loading,
        error: s.error,
        preview: s.preview,
        previewLoading: s.previewLoading,
        showPreview: s.showPreview,
        columnOverrides: s.columnOverrides,
        aiSuggestions: s.aiSuggestions,
        aiSuggestionsLoading: s.aiSuggestionsLoading,
        aiSuggestionsError: s.aiSuggestionsError,
        aiConfigured: s.aiConfigured,
      })
    )
  );

  const actions = useMDFlowActions();
  const history = useHistoryStore((s) => s.history);

  const openaiKey = useOpenAIKeyStore((s) => s.apiKey);
  const setOpenaiKey = useOpenAIKeyStore((s) => s.setApiKey);
  const clearOpenaiKey = useOpenAIKeyStore((s) => s.clearApiKey);

  const { data: templateList } = useMDFlowTemplatesQuery();
  const templates = templateList ?? EMPTY_TEMPLATES;

  return {
    state,
    actions,
    history,
    openaiKey,
    setOpenaiKey,
    clearOpenaiKey,
    templates,
  };
}
