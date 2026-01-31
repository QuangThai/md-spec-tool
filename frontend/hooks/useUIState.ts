import { useState, useCallback } from "react";

/**
 * Custom hook for managing all UI toggle states
 * Consolidates boolean state management for modal/panel visibility
 */
export function useUIState() {
  const [dragOver, setDragOver] = useState(false);
  const [showDiff, setShowDiff] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [showValidationConfigurator, setShowValidationConfigurator] = useState(false);
  const [showTemplateEditor, setShowTemplateEditor] = useState(false);
  const [copied, setCopied] = useState(false);

  const toggleDiff = useCallback(() => setShowDiff((prev) => !prev), []);
  const toggleHistory = useCallback(() => setShowHistory((prev) => !prev), []);
  const toggleValidationConfigurator = useCallback(
    () => setShowValidationConfigurator((prev) => !prev),
    []
  );
  const toggleTemplateEditor = useCallback(
    () => setShowTemplateEditor((prev) => !prev),
    []
  );

  const resetCopyState = useCallback(() => setCopied(false), []);
  const markCopied = useCallback(() => {
    setCopied(true);
    setTimeout(() => setCopied(false), 3000);
  }, []);

  return {
    // State
    dragOver,
    showDiff,
    showHistory,
    showValidationConfigurator,
    showTemplateEditor,
    copied,
    // Setters
    setDragOver,
    setShowDiff,
    setShowHistory,
    setShowValidationConfigurator,
    setShowTemplateEditor,
    setCopied,
    // Handlers
    toggleDiff,
    toggleHistory,
    toggleValidationConfigurator,
    toggleTemplateEditor,
    markCopied,
    resetCopyState,
  };
}
