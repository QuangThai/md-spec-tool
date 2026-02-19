"use client";

import { useQuotaStatus } from "@/hooks/useQuotaStatus";
import { AlertCircle, RefreshCw, Zap } from "lucide-react";
import { useEffect, useState } from "react";
import { Tooltip } from "./ui/Tooltip";

interface QuotaStatusProps {
  onQuotaExceeded?: () => void;
  compact?: boolean;
}

const TOKENS_WARNING_THRESHOLD = 10000; // Show warning when < 10K tokens

/**
 * QuotaStatus - Displays remaining API tokens and quota status
 * Shows warning when approaching limit and error when exceeded
 */
export function QuotaStatus({ onQuotaExceeded, compact = false }: QuotaStatusProps) {
  const { quota, loading, error } = useQuotaStatus({
    enabled: true,
    pollingInterval: 30000, // Refresh every 30 seconds
  });

  const [hasWarned, setHasWarned] = useState(false);

  // Call callback when quota exceeded
  useEffect(() => {
    if (quota?.status === "exceeded" && !hasWarned) {
      setHasWarned(true);
      onQuotaExceeded?.();
    }
  }, [quota?.status, hasWarned, onQuotaExceeded]);

  if (!quota && !loading) {
    return null; // Don't show if no data and not loading
  }

  const percentUsed = quota
    ? (quota.used_tokens / quota.limit_tokens) * 100
    : 0;
  const isWarning = quota && quota.remaining_tokens < TOKENS_WARNING_THRESHOLD;
  const isExceeded = quota?.status === "exceeded";

  // Compact mode: just icon + number
  if (compact) {
    return (
      <Tooltip content={`${quota?.remaining_tokens || 0} tokens remaining`}>
        <div className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-white/5 border border-white/10">
          <Zap className={`w-3.5 h-3.5 ${isExceeded ? "text-red-400" : isWarning ? "text-amber-400" : "text-green-400"}`} />
          <span className="text-[10px] font-mono text-white/70">
            {loading ? "..." : Math.floor(quota?.remaining_tokens || 0).toLocaleString()}
          </span>
        </div>
      </Tooltip>
    );
  }

  // Full mode: detailed quota display
  return (
    <div className={`rounded-lg border p-3 ${
      isExceeded
        ? "bg-red-500/10 border-red-500/30"
        : isWarning
        ? "bg-amber-500/10 border-amber-500/30"
        : "bg-white/5 border-white/10"
    }`}>
      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          {isExceeded ? (
            <AlertCircle className="w-4 h-4 text-red-400" />
          ) : (
            <Zap className={`w-4 h-4 ${isWarning ? "text-amber-400" : "text-green-400"}`} />
          )}
          <span className="text-sm font-semibold text-white">Daily Quota</span>
        </div>
        {loading && <RefreshCw className="w-3.5 h-3.5 text-white/40 animate-spin" />}
      </div>

      {error && (
        <p className="text-xs text-white/60 mb-2">Error loading quota</p>
      )}

      {quota && !error && (
        <>
          {/* Progress bar */}
          <div className="space-y-1.5 mb-2">
            <div className="h-2 bg-black/40 rounded-full overflow-hidden border border-white/10">
              <div
                className={`h-full transition-all ${
                  isExceeded
                    ? "bg-red-500"
                    : isWarning
                    ? "bg-amber-500"
                    : "bg-green-500"
                }`}
                style={{ width: `${Math.min(percentUsed, 100)}%` }}
              />
            </div>

            {/* Stats */}
            <div className="flex items-center justify-between text-[11px] font-mono">
              <span className="text-white/70">
                {Math.floor(quota.used_tokens).toLocaleString()} /{" "}
                {Math.floor(quota.limit_tokens).toLocaleString()}
              </span>
              <span className={`font-semibold ${
                isExceeded
                  ? "text-red-400"
                  : isWarning
                  ? "text-amber-400"
                  : "text-green-400"
              }`}>
                {Math.floor(quota.remaining_tokens).toLocaleString()} left
              </span>
            </div>
          </div>

          {/* Status message */}
          {isExceeded && (
            <p className="text-xs text-red-400">Daily quota exceeded. Resets at midnight UTC.</p>
          )}
          {isWarning && !isExceeded && (
            <p className="text-xs text-amber-400">Approaching quota limit ({TOKENS_WARNING_THRESHOLD.toLocaleString()} tokens warning)</p>
          )}

          {/* Conversions counter */}
          <div className="text-xs text-white/50 mt-2 pt-2 border-t border-white/10">
            {quota.daily_conversions} conversion{quota.daily_conversions !== 1 ? "s" : ""} today
          </div>
        </>
      )}
    </div>
  );
}
