import { useCallback } from "react";
import {
  useGetXLSXSheetsMutation,
  usePreviewTSVMutation,
  usePreviewXLSXMutation,
} from "@/lib/mdflowQueries";

interface FileHandlingProps {
  setFile: (file: File) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setSheets: (sheets: string[]) => void;
  setSelectedSheet: (sheet: string) => void;
  setPreview: (preview: any) => void;
  setShowPreview: (show: boolean) => void;
}

/**
 * Custom hook for handling file uploads and processing
 * Detects file type and fetches appropriate preview
 */
export function useFileHandling({
  setFile,
  setLoading,
  setError,
  setSheets,
  setSelectedSheet,
  setPreview,
  setShowPreview,
}: FileHandlingProps) {
  const getSheetsMutation = useGetXLSXSheetsMutation();
  const previewTSVMutation = usePreviewTSVMutation();
  const previewXLSXMutation = usePreviewXLSXMutation();

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (!selectedFile) return;

      setFile(selectedFile);
      setLoading(true);
      setError(null);
      setPreview(null);

      try {
        // Handle TSV files
        if (/\.tsv$/i.test(selectedFile.name)) {
          const previewResult = await previewTSVMutation.mutateAsync(selectedFile);
          setPreview(previewResult);
          setShowPreview(true);
          return;
        }

        // Handle XLSX files
        const result = await getSheetsMutation.mutateAsync(selectedFile);
        setSheets(result.sheets);
        setSelectedSheet(result.active_sheet);

        // Fetch XLSX preview for the active sheet
        const previewResult = await previewXLSXMutation.mutateAsync({
          file: selectedFile,
          sheetName: result.active_sheet,
        });
        setPreview(previewResult);
        setShowPreview(true);
      } catch (error) {
        setError(error instanceof Error ? error.message : "File processing failed");
      } finally {
        setLoading(false);
      }
    },
    [
      setFile,
      setLoading,
      setError,
      setSheets,
      setSelectedSheet,
      setPreview,
      setShowPreview,
      getSheetsMutation,
      previewTSVMutation,
      previewXLSXMutation,
    ]
  );

  return { handleFileChange };
}
