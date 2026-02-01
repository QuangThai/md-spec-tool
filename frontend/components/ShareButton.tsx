import {
  generateShareURL,
  isShareDataTooLong,
  ShareData,
} from "@/lib/shareUtils";
import { Check, Share2 } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { Tooltip } from "./ui/Tooltip";

interface ShareButtonProps {
  mdflowOutput: string;
  template: string;
}

/**
 * ShareButton - Generates and copies shareable URLs
 * Shows tooltip with copy status
 */
export function ShareButton({ mdflowOutput, template }: ShareButtonProps) {
  const [copied, setCopied] = useState(false);
  const [copyFailed, setCopyFailed] = useState(false);
  const resetTimerRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (resetTimerRef.current) {
        window.clearTimeout(resetTimerRef.current);
      }
    };
  }, []);

  const shareData: ShareData = {
    mdflow: mdflowOutput,
    template,
    createdAt: Date.now(),
  };

  const isTooLong = isShareDataTooLong(shareData);
  const isDisabled = !mdflowOutput || isTooLong;

  const handleShare = useCallback(async () => {
    if (isDisabled) return;

    try {
      const url = generateShareURL(shareData);
      await navigator.clipboard.writeText(url);
      setCopied(true);

      if (resetTimerRef.current) window.clearTimeout(resetTimerRef.current);
      resetTimerRef.current = window.setTimeout(() => {
        setCopied(false);
        setCopyFailed(false);
      }, 2000);
    } catch (err) {
      console.error("Failed to copy share link", err);
      setCopyFailed(true);

      if (resetTimerRef.current) window.clearTimeout(resetTimerRef.current);
      resetTimerRef.current = window.setTimeout(() => {
        setCopied(false);
        setCopyFailed(false);
      }, 2000);
    }
  }, [shareData, isDisabled]);

  const tooltipText = isTooLong
    ? "Too large"
    : copied
    ? "Copied!"
    : copyFailed
    ? "Copy failed"
    : "Share";
  
  return (
    <Tooltip content={tooltipText}>
      <button
        type="button"
        onClick={handleShare}
        disabled={isDisabled}
        className={`p-1.5 sm:p-2 rounded-lg border transition-all ${
          isDisabled
            ? "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
            : "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
        }`}
      >
        {copied ? (
          <Check className="w-3.5 h-3.5 text-accent-orange" />
        ) : (
          <Share2 className="w-3.5 h-3.5" />
        )}
      </button>
    </Tooltip>
  );
}
