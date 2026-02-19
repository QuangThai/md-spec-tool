import { useCallback, useState } from "react";
import { useGetXLSXSheetsMutation } from "@/lib/mdflowQueries";
import { PreviewResponse } from "@/lib/types";

interface UseFileHandlingProps {
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
  mode: "paste" | "xlsx" | "tsv";
  file: File | null;
  setFile: (file: File | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setPreview: (preview: PreviewResponse | null) => void;
  setSheets: (sheets: string[]) => void;
  setSelectedSheet: (sheet: string) => void;
}

interface UseFileHandlingReturn {
  dragOver: boolean;
  setDragOver: (over: boolean) => void;
  handleFileChange: (e: React.ChangeEvent<HTMLInputElement>) => Promise<void>;
  onDrop: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragLeave: () => void;
}

export function useFileHandling({
  setLastFailedAction,
  mode,
  file,
  setFile,
  setLoading,
  setError,
  setPreview,
  setSheets,
  setSelectedSheet,
}: UseFileHandlingProps): UseFileHandlingReturn {
  const [dragOver, setDragOver] = useState(false);
  const getSheetsMutation = useGetXLSXSheetsMutation();

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (!selectedFile) return;

      setFile(selectedFile);
      setLoading(true);
      setError(null);
      setLastFailedAction(null);
      setPreview(null);

      if (/\.tsv$/i.test(selectedFile.name)) {
        setLoading(false);
        return;
      }

      try {
        const result = await getSheetsMutation.mutateAsync(selectedFile);
        setSheets(result.sheets);
        setSelectedSheet(result.active_sheet);
      } catch (error) {
        setError(
          error instanceof Error ? error.message : "Failed to read sheets"
        );
        setLastFailedAction("other");
      } finally {
        setLoading(false);
      }
    },
    [
      setFile,
      setLoading,
      setError,
      setLastFailedAction,
      setSheets,
      setSelectedSheet,
      setPreview,
      getSheetsMutation,
    ]
  );

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const f = e.dataTransfer.files?.[0];
      if (!f) return;

      if (mode === "tsv" && /\.tsv$/i.test(f.name)) {
        setFile(f);
        setError(null);
        setLastFailedAction(null);
        setPreview(null);
        setLoading(true);
        setLoading(false);
        return;
      }

      if (mode === "xlsx" && /\.xlsx$/i.test(f.name)) {
        setFile(f);
        setLoading(true);
        setError(null);
        setLastFailedAction(null);
        setPreview(null);
        getSheetsMutation
          .mutateAsync(f)
          .then((result) => {
            setSheets(result.sheets);
            setSelectedSheet(result.active_sheet);
          })
          .catch((error) => {
            setError(
              error instanceof Error ? error.message : "Failed to read sheets"
            );
            setLastFailedAction("other");
          })
          .finally(() => {
            setLoading(false);
          });
      }
    },
    [
      mode,
      setFile,
      setLoading,
      setError,
      setLastFailedAction,
      setSheets,
      setSelectedSheet,
      setPreview,
      getSheetsMutation,
    ]
  );

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  }, []);

  const onDragLeave = useCallback(() => {
    setDragOver(false);
  }, []);

  return {
    dragOver,
    setDragOver,
    handleFileChange,
    onDrop,
    onDragOver,
    onDragLeave,
  };
}
