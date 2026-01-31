import { useCallback } from "react";
import { getXLSXSheets, previewTSV, previewXLSX } from "@/lib/mdflowApi";

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
          const previewResult = await previewTSV(selectedFile);
          if (previewResult.data) {
            setPreview(previewResult.data);
            setShowPreview(true);
          }
          return;
        }

        // Handle XLSX files
        const result = await getXLSXSheets(selectedFile);

        if (result.error) {
          setError(result.error);
        } else if (result.data) {
          setSheets(result.data.sheets);
          setSelectedSheet(result.data.active_sheet);

          // Fetch XLSX preview for the active sheet
          const previewResult = await previewXLSX(
            selectedFile,
            result.data.active_sheet
          );
          if (previewResult.data) {
            setPreview(previewResult.data);
            setShowPreview(true);
          }
        }
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
    ]
  );

  return { handleFileChange };
}
