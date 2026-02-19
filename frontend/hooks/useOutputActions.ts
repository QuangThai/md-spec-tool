"use client";

import { useCallback, useState } from "react";

interface UseOutputActionsReturn {
  copied: boolean;
  setCopied: (value: boolean) => void;
  handleCopy: () => void;
  handleDownload: () => void;
}

/**
 * Hook for managing output actions: copying to clipboard and downloading
 * @param mdflowOutput - The markdown output content to copy/download
 * @returns Object containing copied state and action handlers
 */
export function useOutputActions(
  mdflowOutput: string
): UseOutputActionsReturn {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(mdflowOutput);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [mdflowOutput]);

  const handleDownload = useCallback(() => {
    const blob = new Blob([mdflowOutput], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "spec.mdflow.md";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [mdflowOutput]);

  return {
    copied,
    setCopied,
    handleCopy,
    handleDownload,
  };
}
