import {
  usePreviewPasteQuery,
  usePreviewTSVQuery,
  usePreviewXLSXQuery,
} from "@/lib/mdflowQueries";
import { InputMode, PreviewResponse } from "@/lib/types";
import { useCallback, useEffect, useState } from "react";

interface PreviewManagementProps {
  mode: InputMode;
  pasteText: string;
  file: File | null;
  selectedSheet: string;
  template?: string;
  preview: PreviewResponse | null;
  previewLoading: boolean;
  showPreview: boolean;
  setPreview: (preview: PreviewResponse | null) => void;
  setPreviewLoading: (loading: boolean) => void;
  setShowPreview: (show: boolean) => void;
}

/**
 * Custom hook for managing preview state and fetching
 * Handles debounced preview updates for paste mode and sheet changes
 */
export function usePreviewManagement({
  mode,
  pasteText,
  file,
  selectedSheet,
  template,
  preview,
  previewLoading,
  showPreview,
  setPreview,
  setPreviewLoading,
  setShowPreview,
}: PreviewManagementProps) {
  const [debouncedPasteText, setDebouncedPasteText] = useState("");
  const previewPasteQuery = usePreviewPasteQuery(
    debouncedPasteText,
    mode === "paste",
    template
  );
  const previewTSVQuery = usePreviewTSVQuery(file, mode === "tsv", template);
  const previewXLSXQuery = usePreviewXLSXQuery(
    file,
    selectedSheet,
    mode === "xlsx",
    template
  );

  // Auto-preview with debounce when paste text changes
  useEffect(() => {
    if (mode !== "paste") {
      setDebouncedPasteText("");
      return;
    }
    const timer = setTimeout(() => {
      setDebouncedPasteText(pasteText);
    }, 500);
    return () => clearTimeout(timer);
  }, [pasteText, mode]);

  useEffect(() => {
    if (mode !== "paste") return;
    if (!debouncedPasteText.trim()) {
      setPreview(null);
      setShowPreview(false);
      return;
    }
    if (previewPasteQuery.data) {
      setPreview(previewPasteQuery.data);
      setShowPreview(true);
    }
  }, [
    debouncedPasteText,
    mode,
    previewPasteQuery.data,
    setPreview,
    setShowPreview,
  ]);

  useEffect(() => {
    if (mode === "xlsx" && previewXLSXQuery.data) {
      setPreview(previewXLSXQuery.data);
      setShowPreview(true);
    }
  }, [mode, previewXLSXQuery.data, setPreview, setShowPreview]);

  useEffect(() => {
    if (mode === "tsv" && previewTSVQuery.data) {
      setPreview(previewTSVQuery.data);
      setShowPreview(true);
    }
  }, [mode, previewTSVQuery.data, setPreview, setShowPreview]);

  useEffect(() => {
    const isLoading =
      (mode === "paste" && previewPasteQuery.isFetching) ||
      (mode === "xlsx" && previewXLSXQuery.isFetching) ||
      (mode === "tsv" && previewTSVQuery.isFetching);
    setPreviewLoading(isLoading);
  }, [
    mode,
    previewPasteQuery.isFetching,
    previewTSVQuery.isFetching,
    previewXLSXQuery.isFetching,
    setPreviewLoading,
  ]);

  const togglePreview = useCallback(() => {
    setShowPreview(!showPreview);
  }, [showPreview, setShowPreview]);

  return {
    preview,
    previewLoading,
    showPreview,
    togglePreview,
  };
}
