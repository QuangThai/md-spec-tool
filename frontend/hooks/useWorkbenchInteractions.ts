import { useCallback } from "react";
import { useKeyboardShortcuts } from "@/lib/useKeyboardShortcuts";
import { toast } from "@/components/ui/Toast";

interface UseWorkbenchInteractionsParams {
  showCommandPalette: boolean;
  setShowCommandPalette: (show: boolean) => void;
  showHistory: boolean;
  setShowHistory: (show: boolean) => void;
  showDiff: boolean;
  setShowDiff: (show: boolean) => void;
  showTemplateEditor: boolean;
  setShowTemplateEditor: (show: boolean) => void;
  showValidationConfigurator: boolean;
  setShowValidationConfigurator: (show: boolean) => void;
  openCommandPalette: () => void;
  handleConvert: () => Promise<void>;
  togglePreview: () => void;
  mdflowOutput: string;
  reviewGateReason: string | undefined;
  copyOutput: () => void;
  exportOutput: () => void;
}

export function useWorkbenchInteractions({
  showCommandPalette,
  setShowCommandPalette,
  showHistory,
  setShowHistory,
  showDiff,
  setShowDiff,
  showTemplateEditor,
  setShowTemplateEditor,
  showValidationConfigurator,
  setShowValidationConfigurator,
  openCommandPalette,
  handleConvert,
  togglePreview,
  mdflowOutput,
  reviewGateReason,
  copyOutput,
  exportOutput,
}: UseWorkbenchInteractionsParams) {
  const handleShortcutCopy = useCallback(() => {
    if (!mdflowOutput) return;
    copyOutput();
    toast.success("Copied to clipboard");
  }, [copyOutput, mdflowOutput]);

  const handleShortcutExport = useCallback(() => {
    if (!mdflowOutput) return;
    exportOutput();
    toast.success("Downloaded spec.mdflow.md");
  }, [exportOutput, mdflowOutput]);

  const handleCommandPaletteCopy = useCallback(() => {
    if (reviewGateReason) {
      toast.error("Review required", reviewGateReason);
      return;
    }
    if (!mdflowOutput) return;
    copyOutput();
    toast.success("Copied to clipboard");
  }, [copyOutput, mdflowOutput, reviewGateReason]);

  const handleCommandPaletteExport = useCallback(() => {
    if (reviewGateReason) {
      toast.error("Review required", reviewGateReason);
      return;
    }
    if (!mdflowOutput) return;
    exportOutput();
    toast.success("Downloaded spec.mdflow.md");
  }, [exportOutput, mdflowOutput, reviewGateReason]);

  const handleEscape = useCallback(() => {
    if (showCommandPalette) setShowCommandPalette(false);
    else if (showHistory) setShowHistory(false);
    else if (showDiff) setShowDiff(false);
    else if (showTemplateEditor) setShowTemplateEditor(false);
    else if (showValidationConfigurator) setShowValidationConfigurator(false);
  }, [
    setShowCommandPalette,
    setShowDiff,
    setShowHistory,
    setShowTemplateEditor,
    setShowValidationConfigurator,
    showCommandPalette,
    showDiff,
    showHistory,
    showTemplateEditor,
    showValidationConfigurator,
  ]);

  useKeyboardShortcuts({
    commandPalette: openCommandPalette,
    convert: handleConvert,
    copy: handleShortcutCopy,
    export: handleShortcutExport,
    togglePreview,
    showShortcuts: () => {},
    escape: handleEscape,
  });

  return {
    handleCommandPaletteCopy,
    handleCommandPaletteExport,
  };
}
