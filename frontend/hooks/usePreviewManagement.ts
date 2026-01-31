import { useCallback, useEffect } from "react";
import { previewPaste, previewTSV, previewXLSX } from "@/lib/mdflowApi";

interface PreviewManagementProps {
  mode: string;
  pasteText: string;
  file: File | null;
  selectedSheet: string;
  preview: any;
  previewLoading: boolean;
  showPreview: boolean;
  setPreview: (preview: any) => void;
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
  preview,
  previewLoading,
  showPreview,
  setPreview,
  setPreviewLoading,
  setShowPreview,
}: PreviewManagementProps) {
  // Auto-preview with debounce when paste text changes
  useEffect(() => {
    if (mode !== "paste" || !pasteText.trim()) {
      setPreview(null);
      setShowPreview(false);
      return;
    }

    const timer = setTimeout(async () => {
      setPreviewLoading(true);
      try {
        const result = await previewPaste(pasteText);
        if (result.data) {
          setPreview(result.data);
          setShowPreview(true);
        }
      } finally {
        setPreviewLoading(false);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [pasteText, mode, setPreview, setPreviewLoading, setShowPreview]);

  // Update preview when sheet selection changes for XLSX
  useEffect(() => {
    if (mode !== "xlsx" || !file || !selectedSheet) return;

    const fetchPreview = async () => {
      setPreviewLoading(true);
      try {
        const result = await previewXLSX(file, selectedSheet);
        if (result.data) {
          setPreview(result.data);
          setShowPreview(true);
        }
      } finally {
        setPreviewLoading(false);
      }
    };

    fetchPreview();
  }, [selectedSheet, file, mode, setPreview, setPreviewLoading, setShowPreview]);

  // Update preview when sheet selection changes for TSV
  useEffect(() => {
    if (mode !== "tsv" || !file) return;

    const fetchPreview = async () => {
      setPreviewLoading(true);
      try {
        const result = await previewTSV(file);
        if (result.data) {
          setPreview(result.data);
          setShowPreview(true);
        }
      } finally {
        setPreviewLoading(false);
      }
    };

    fetchPreview();
  }, [file, mode, setPreview, setPreviewLoading, setShowPreview]);

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
