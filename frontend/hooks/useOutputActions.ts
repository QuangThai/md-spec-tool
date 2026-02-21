"use client";

import { useCallback, useState } from "react";
import { useMDFlowStore } from "@/lib/mdflowStore";

interface UseOutputActionsReturn {
  copied: boolean;
  handleCopy: () => void;
  handleDownload: () => void;
}

export function useOutputActions(): UseOutputActionsReturn {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    const mdflowOutput = useMDFlowStore.getState().mdflowOutput;
    navigator.clipboard.writeText(mdflowOutput);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, []);

  const handleDownload = useCallback(() => {
    const mdflowOutput = useMDFlowStore.getState().mdflowOutput;
    const blob = new Blob([mdflowOutput], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "spec.mdflow.md";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, []);

  return {
    copied,
    handleCopy,
    handleDownload,
  };
}
