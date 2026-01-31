import { useCallback, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { AlertCircle, Check, Share2 } from "lucide-react";
import { generateShareURL, isShareDataTooLong, ShareData } from "@/lib/shareUtils";

interface ShareButtonProps {
  mdflowOutput: string;
  template: string;
}

/**
 * ShareButton - Generates and copies shareable URLs
 * Shows tooltip with copy status or error message
 */
export function ShareButton({ mdflowOutput, template }: ShareButtonProps) {
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

  return (
    <div className="relative">
      <button
        type="button"
        onClick={handleShare}
        onMouseEnter={() => !copied && setShowTooltip(true)}
        onMouseLeave={() => !copied && setShowTooltip(false)}
        disabled={isTooLong}
        className={`
          flex items-center justify-center h-8 px-3 rounded-lg text-[9px] font-bold uppercase tracking-wider transition-all cursor-pointer shrink-0
          ${isTooLong
            ? "bg-white/5 text-white/30 cursor-not-allowed border border-white/10"
            : "bg-white/10 hover:bg-white/20 text-white/80 hover:text-white border border-white/10 hover:border-white/20"
          }
        `}
        title={isTooLong ? "Content too large to share via URL" : "Share link"}
      >
        <Share2 className="w-3 h-3 shrink-0" />
      </button>

      <AnimatePresence>
        {showTooltip && (
          <motion.div
            initial={{ opacity: 0, y: 4, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 4, scale: 0.95 }}
            className="absolute right-0 top-full mt-2 z-50 w-56 rounded-xl bg-black/95 backdrop-blur-xl border border-white/20 shadow-2xl overflow-hidden"
          >
            <div className="p-3">
              {error ? (
                <div className="flex items-start gap-2">
                  <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                  <p className="text-[11px] text-red-400">{error}</p>
                </div>
              ) : copied ? (
                <div className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  <div>
                    <p className="text-[11px] font-bold text-white">Link copied!</p>
                    <p className="text-[9px] text-white/50 mt-0.5">
                      Share this URL with anyone
                    </p>
                  </div>
                </div>
              ) : (
                <div className="flex items-start gap-2">
                  <Share2 className="w-4 h-4 text-accent-orange shrink-0 mt-0.5" />
                  <div>
                    <p className="text-[11px] font-bold text-white">Share Link</p>
                    <p className="text-[9px] text-white/50 mt-0.5">
                      Click to copy shareable URL
                    </p>
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
