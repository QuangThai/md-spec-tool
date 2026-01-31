import { useState, useCallback } from "react";
import {
  convertGoogleSheet,
  convertPaste,
  convertTSV,
  convertXLSX,
  diffMDFlow,
  getAISuggestions,
} from "@/lib/mdflowApi";
import { useHistoryStore } from "@/lib/mdflowStore";

interface ConversionFlowProps {
  mode: string;
  pasteText: string;
  file: File | null;
  selectedSheet: string;
  template: string;
  columnOverrides: Record<string, string>;
  aiConfigured: boolean;
  setResult: (output: string, warnings: any[], meta: any) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setAISuggestions: (suggestions: any, configured: boolean) => void;
  setAISuggestionsLoading: (loading: boolean) => void;
  setAISuggestionsError: (error: string | null) => void;
}

/**
 * Custom hook for the main conversion flow
 * Handles conversions, diffs, and AI suggestions
 */
export function useConversionFlow({
  mode,
  pasteText,
  file,
  selectedSheet,
  template,
  columnOverrides,
  aiConfigured,
  setResult,
  setLoading,
  setError,
  setAISuggestions,
  setAISuggestionsLoading,
  setAISuggestionsError,
}: ConversionFlowProps) {
  const { addToHistory } = useHistoryStore();
  const [currentDiff, setCurrentDiff] = useState<any>(null);
  const [previousOutput, setPreviousOutput] = useState<string>("");

  const performConversion = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      let result;

      if (mode === "paste") {
        if (pasteText.trim().startsWith("http")) {
          // Google Sheets URL
          result = await convertGoogleSheet(pasteText, template);
        } else {
          // Paste TSV/CSV
          result = await convertPaste(pasteText, template);
        }
      } else if (mode === "xlsx" && file) {
        result = await convertXLSX(file, selectedSheet, template);
      } else if (mode === "tsv" && file) {
        result = await convertTSV(file, template);
      } else {
        setError("No input provided");
        setLoading(false);
        return;
      }

      if (result.error) {
        setError(result.error);
      } else if (result.data) {
        setResult(result.data.mdflow, result.data.warnings || [], result.data.meta || null);

        // Add to history
        addToHistory({
          mode: mode as "paste" | "xlsx" | "tsv",
          inputPreview: mode === "paste" ? pasteText.substring(0, 100) : file?.name || "file",
          template,
          output: result.data.mdflow,
          meta: result.data.meta || null,
        });
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Conversion failed");
    } finally {
      setLoading(false);
    }
  }, [
    mode,
    pasteText,
    file,
    selectedSheet,
    template,
    columnOverrides,
    setResult,
    setLoading,
    setError,
    addToHistory,
  ]);

  const handleDiff = useCallback(
    async (currentOutput: string) => {
      if (!previousOutput) {
        setCurrentDiff(null);
        return;
      }

      try {
        const diffResult = await diffMDFlow(previousOutput, currentOutput);
        if (diffResult) {
          setCurrentDiff(diffResult);
        }
      } catch (err) {
        console.error("Diff failed:", err);
      }
    },
    [previousOutput]
  );

  const generateAISuggestions = useCallback(
    async (output: string) => {
      if (!aiConfigured) return;

      setAISuggestionsLoading(true);
      try {
        const result = await getAISuggestions(output);
        if (result.data) {
          setAISuggestions(result.data, true);
          setAISuggestionsError(null);
        }
      } catch (err) {
        setAISuggestionsError(
          err instanceof Error ? err.message : "Failed to get suggestions"
        );
      } finally {
        setAISuggestionsLoading(false);
      }
    },
    [aiConfigured, setAISuggestions, setAISuggestionsLoading, setAISuggestionsError]
  );

  return {
    performConversion,
    handleDiff,
    generateAISuggestions,
    currentDiff,
    setCurrentDiff,
    previousOutput,
    setPreviousOutput,
  };
}
