import { useState, useCallback } from "react";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import {
  useAISuggestionsMutation,
  useConvertGoogleSheetMutation,
  useConvertPasteMutation,
  useConvertTSVMutation,
  useConvertXLSXMutation,
  useDiffMDFlowMutation,
} from "@/lib/mdflowQueries";
import { useHistoryStore } from "@/lib/mdflowStore";
import {
  AISuggestion,
  InputMode,
  MDFlowMeta,
  MDFlowWarning,
} from "@/lib/types";
import { DiffResponse } from "@/lib/diffTypes";

interface ConversionFlowProps {
  mode: InputMode;
  pasteText: string;
  file: File | null;
  selectedSheet: string;
  template: string;
  columnOverrides: Record<string, string>;
  aiConfigured: boolean | null;
  setResult: (output: string, warnings: MDFlowWarning[], meta: MDFlowMeta) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setAISuggestions: (suggestions: AISuggestion[], configured: boolean) => void;
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
  const [currentDiff, setCurrentDiff] = useState<DiffResponse | null>(null);
  const [previousOutput, setPreviousOutput] = useState<string>("");
  const convertPasteMutation = useConvertPasteMutation();
  const convertXLSXMutation = useConvertXLSXMutation();
  const convertTSVMutation = useConvertTSVMutation();
  const convertGoogleSheetMutation = useConvertGoogleSheetMutation();
  const diffMDFlowMutation = useDiffMDFlowMutation();
  const aiSuggestionsMutation = useAISuggestionsMutation();

  const performConversion = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      let result;

        if (mode === "paste") {
          const trimmedPaste = pasteText.trim();
          if (isGoogleSheetsURL(trimmedPaste)) {
            // Google Sheets URL
            result = await convertGoogleSheetMutation.mutateAsync({
              url: trimmedPaste,
              template,
            });
          } else {
            // Paste TSV/CSV
            result = await convertPasteMutation.mutateAsync({
              pasteText,
              template,
            });
          }
        } else if (mode === "xlsx" && file) {
          result = await convertXLSXMutation.mutateAsync({
            file,
            sheetName: selectedSheet,
            template,
          });
        } else if (mode === "tsv" && file) {
          result = await convertTSVMutation.mutateAsync({
            file,
            template,
          });
        } else {
          setError("No input provided");
          setLoading(false);
          return;
        }

        if (result) {
          setResult(result.mdflow, result.warnings || [], result.meta || null);

          // Add to history
          addToHistory({
            mode,
            inputPreview: mode === "paste" ? pasteText.substring(0, 100) : file?.name || "file",
            template,
            output: result.mdflow,
            meta: result.meta || null,
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
    setResult,
    setLoading,
    setError,
    addToHistory,
    convertGoogleSheetMutation,
    convertPasteMutation,
    convertTSVMutation,
    convertXLSXMutation,
  ]);

  const handleDiff = useCallback(
    async (currentOutput: string) => {
      if (!previousOutput) {
        setCurrentDiff(null);
        return;
      }

      try {
        const diffResult = await diffMDFlowMutation.mutateAsync({
          before: previousOutput,
          after: currentOutput,
        });
        if (diffResult) {
          setCurrentDiff(diffResult);
        }
      } catch (err) {
        console.error("Diff failed:", err);
      }
    },
    [previousOutput, diffMDFlowMutation]
  );

  const generateAISuggestions = useCallback(
    async (output: string) => {
      if (!aiConfigured) return;

      setAISuggestionsLoading(true);
      try {
        const result = await aiSuggestionsMutation.mutateAsync({
          pasteText: output,
        });
        setAISuggestions(result.suggestions, result.configured);
        setAISuggestionsError(null);
      } catch (err) {
        setAISuggestionsError(
          err instanceof Error ? err.message : "Failed to get suggestions"
        );
      } finally {
        setAISuggestionsLoading(false);
      }
    },
    [
      aiConfigured,
      setAISuggestions,
      setAISuggestionsLoading,
      setAISuggestionsError,
      aiSuggestionsMutation,
    ]
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
