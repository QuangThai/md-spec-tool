import { useCallback, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { Check, Share2 } from "lucide-react";
import { generateShareURL, isShareDataTooLong, ShareData } from "@/lib/shareUtils";

interface ShareButtonProps {
  mdflowOutput: string;
  template: string;
}

/**
 * ShareButton - Generates and copies shareable URLs
 * Shows tooltip with copy status
 */
export function ShareButton({ mdflowOutput, template }: ShareButtonProps) {
  const [showTooltip, setShowTooltip] = useState(false);
  const [copied, setCopied] = useState(false);

  const shareData: ShareData = {
    mdflow: mdflowOutput,
    template,
    createdAt: Date.now(),
  };

  const isTooLong = isShareDataTooLong(shareData);

  const handleShare = useCallback(() => {
    if (isTooLong) return;

    try {
      const url = generateShareURL(shareData);
      navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to generate share link", err);
    }
  }, [shareData, isTooLong]);

  const tooltipText = isTooLong ? "Too large" : copied ? "Copied!" : "Share";

  return (
    <div
      className="relative"
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
    >
      <button
        type="button"
        onClick={handleShare}
        disabled={isTooLong}
        className={`
          p-1.5 sm:p-2 rounded-lg border transition-all cursor-pointer
          ${isTooLong
            ? "bg-white/5 border-white/5 text-white/30 cursor-not-allowed"
            : "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
          }
        `}
      >
        {copied ? <Check className="w-3.5 h-3.5 text-accent-orange" /> : <Share2 className="w-3.5 h-3.5" />}
      </button>

      <AnimatePresence>
        {showTooltip && (
          <motion.div
            initial={{ opacity: 0, y: -4, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -4, scale: 0.95 }}
            transition={{ duration: 0.15 }}
            className="absolute left-1/2 -translate-x-1/2 top-full mt-1.5 z-50 px-2 py-1 rounded-md bg-black/95 backdrop-blur-sm border border-white/10 shadow-xl text-[9px] font-medium text-white/90 whitespace-nowrap pointer-events-none"
          >
            {tooltipText}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
