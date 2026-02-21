import { useCallback, useEffect, useState } from "react";
import type { InputMode, OutputFormat } from "@/lib/types";

type FailedAction = "preview" | "convert" | "other" | null;

interface UseWorkbenchUIStateParams {
  mode: InputMode;
  pasteText: string;
  format: OutputFormat;
}

export function useWorkbenchUIState({
  mode,
  pasteText,
  format,
}: UseWorkbenchUIStateParams) {
  const [showHistory, setShowHistory] = useState(false);
  const [showValidationConfigurator, setShowValidationConfigurator] =
    useState(false);
  const [showTemplateEditor, setShowTemplateEditor] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [showApiKeyInput, setShowApiKeyInput] = useState(false);
  const [apiKeyDraft, setApiKeyDraft] = useState("");
  const [includeMetadata, setIncludeMetadata] = useState(true);
  const [numberRows, setNumberRows] = useState(false);
  const [lastFailedAction, setLastFailedAction] = useState<FailedAction>(null);
  const [debouncedPasteText, setDebouncedPasteText] = useState("");

  useEffect(() => {
    if (mode !== "paste") {
      setDebouncedPasteText("");
      return;
    }

    const timer = setTimeout(() => {
      setDebouncedPasteText(pasteText);
    }, 500);

    return () => clearTimeout(timer);
  }, [mode, pasteText]);

  useEffect(() => {
    if (format !== "table" && numberRows) {
      setNumberRows(false);
    }
  }, [format, numberRows]);

  const toggleApiKeyInput = useCallback(() => {
    setShowApiKeyInput((prev) => !prev);
  }, []);

  const openTemplateEditor = useCallback(() => {
    setShowTemplateEditor(true);
  }, []);

  const openValidationConfigurator = useCallback(() => {
    setShowValidationConfigurator(true);
  }, []);

  const openHistory = useCallback(() => {
    setShowHistory(true);
  }, []);

  const openCommandPalette = useCallback(() => {
    setShowCommandPalette(true);
  }, []);

  return {
    showHistory,
    setShowHistory,
    showValidationConfigurator,
    setShowValidationConfigurator,
    showTemplateEditor,
    setShowTemplateEditor,
    showCommandPalette,
    setShowCommandPalette,
    showApiKeyInput,
    toggleApiKeyInput,
    apiKeyDraft,
    setApiKeyDraft,
    includeMetadata,
    setIncludeMetadata,
    numberRows,
    setNumberRows,
    lastFailedAction,
    setLastFailedAction,
    debouncedPasteText,
    openTemplateEditor,
    openValidationConfigurator,
    openHistory,
    openCommandPalette,
  };
}
