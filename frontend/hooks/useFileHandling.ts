import { useCallback, useState } from "react";
import { useGetXLSXSheetsMutation } from "@/lib/mdflowQueries";
import { useMDFlowActions, useMDFlowStore } from "@/lib/mdflowStore";

interface UseFileHandlingProps {
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
}

export interface UseFileHandlingReturn {
  dragOver: boolean;
  handleFileChange: (e: React.ChangeEvent<HTMLInputElement>) => Promise<void>;
  onDrop: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragLeave: () => void;
}

export function useFileHandling({
  setLastFailedAction,
}: UseFileHandlingProps): UseFileHandlingReturn {
  const mode = useMDFlowStore((state) => state.mode);
  const {
    setFile,
    setLoading,
    setError,
    setPreview,
    setSheets,
    setSelectedSheet,
  } = useMDFlowActions();

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
    handleFileChange,
    onDrop,
    onDragOver,
    onDragLeave,
  };
}
