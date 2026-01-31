import { useState, useCallback } from "react";
import { generateShareURL, isShareDataTooLong, ShareData } from "@/lib/shareUtils";

interface ShareLinkProps {
  mdflowOutput: string;
  template: string;
}

/**
 * Custom hook for share link functionality
 * Generates shareable URLs and manages copy-to-clipboard state
 */
export function useShareLink({ mdflowOutput, template }: ShareLinkProps) {
  const [showTooltip, setShowTooltip] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const shareData: ShareData = {
    mdflow: mdflowOutput,
    template,
    createdAt: Date.now(),
  };

  const isTooLong = isShareDataTooLong(shareData);

  const handleShare = useCallback(() => {
    if (isTooLong) {
      setError("Content too large for URL sharing. Try exporting instead.");
      setShowTooltip(true);
      return;
    }

    try {
      const url = generateShareURL(shareData);
      navigator.clipboard.writeText(url);
      setCopied(true);
      setError(null);
      setShowTooltip(true);
      setTimeout(() => {
        setCopied(false);
        setShowTooltip(false);
      }, 3000);
    } catch (err) {
      setError("Failed to generate share link");
      setShowTooltip(true);
    }
  }, [shareData, isTooLong]);

  return {
    showTooltip,
    setShowTooltip,
    copied,
    error,
    isTooLong,
    handleShare,
  };
}
