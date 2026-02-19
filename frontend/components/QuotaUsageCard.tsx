"use client";

import { AlertCircle, TrendingUp, Zap } from "lucide-react";
import { useQuotaStatus } from "@/hooks/useQuotaStatus";
import { useDailyReport } from "@/hooks/useDailyReport";

interface QuotaUsageCardProps {
  showReportHistory?: boolean;
  refreshInterval?: number; // in milliseconds, 0 = no refresh
}

export function QuotaUsageCard({
  showReportHistory = true,
  refreshInterval = 30000,
}: QuotaUsageCardProps) {
  const { quota: quotaStatus, loading: quotaLoading, error, refetch } = useQuotaStatus({
    pollingInterval: refreshInterval,
  });
  const { reports: reportData } = useDailyReport({
    enabled: showReportHistory,
  });

  if (quotaLoading && !quotaStatus) {
    return (
      <div className="rounded-2xl border border-white/10 bg-linear-to-br from-white/5 via-black/30 to-black/70 p-6">
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <p className="text-xs font-semibold uppercase tracking-wider text-white/50">
              Quota Usage
            </p>
            <div className="h-6 w-32 animate-pulse rounded bg-white/10" />
          </div>
          <div className="text-3xl font-bold text-white/30">—</div>
        </div>
      </div>
    );
  }

  if (error && !quotaStatus) {
    return (
      <div className="rounded-2xl border border-red-500/30 bg-red-500/5 p-6">
        <div className="flex items-center gap-3">
          <AlertCircle size={20} className="text-red-400" aria-hidden="true" />
          <div>
            <p className="font-semibold text-red-300">Quota Error</p>
            <p className="text-sm text-red-200/70">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (!quotaStatus) {
    return null;
  }

  // Guard against divide-by-zero
  const usagePercent =
    quotaStatus.limit_tokens > 0
      ? (quotaStatus.used_tokens / quotaStatus.limit_tokens) * 100
      : 0;
  const isExceeded = quotaStatus.status === "exceeded";
  const resetDate = new Date(quotaStatus.reset_at);
  const now = new Date();
  const hoursUntilReset = Math.max(
    0,
    Math.ceil((resetDate.getTime() - now.getTime()) / (1000 * 60 * 60)),
  );

  return (
    <div className="space-y-4">
      <div className="rounded-2xl border border-white/10 bg-linear-to-br from-white/5 via-black/30 to-black/70 p-6 shadow-lg">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
              <div className="rounded-lg bg-blue-500/20 p-2">
                <Zap size={18} className="text-blue-400" aria-hidden="true" />
              </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-wider text-white/50">
                Daily Quota
              </p>
              <p className="text-sm font-medium text-white">Token Usage</p>
            </div>
          </div>
          <div className="text-right">
            <p className="text-2xl font-bold text-white">
              {quotaStatus.remaining_tokens.toLocaleString()}
            </p>
            <p className="text-xs text-white/60">remaining</p>
          </div>
        </div>

        {/* Progress Bar */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-xs">
            <span className="text-white/70">
              {quotaStatus.used_tokens.toLocaleString()} /{" "}
              {quotaStatus.limit_tokens.toLocaleString()} tokens
            </span>
            <span
              className={
                isExceeded ? "text-red-400 font-semibold" : "text-blue-400"
              }
            >
              {usagePercent.toFixed(1)}%
            </span>
          </div>
          <div
            className="h-2 w-full overflow-hidden rounded-full bg-white/10"
            role="progressbar"
            aria-valuenow={Math.min(usagePercent, 100)}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label="Token usage"
          >
            <div
              className={`h-full transition-all duration-300 ${
                usagePercent > 90
                  ? "bg-linear-to-r from-orange-500 to-red-500"
                  : usagePercent > 75
                    ? "bg-linear-to-r from-yellow-500 to-orange-500"
                    : "bg-linear-to-r from-blue-500 to-cyan-500"
              }`}
              style={{ width: `${Math.max(2, Math.min(100, usagePercent))}%` }}
            />
          </div>
        </div>

        {/* Stats */}
        <div className="mt-4 grid grid-cols-2 gap-4">
          <div className="rounded-lg bg-white/5 px-3 py-2">
            <p className="text-xs text-white/60">Conversions Today</p>
            <p className="text-lg font-bold text-white">
              {quotaStatus.daily_conversions}
            </p>
          </div>
          <div className="rounded-lg bg-white/5 px-3 py-2">
            <p className="text-xs text-white/60">Resets In</p>
            <p className="text-lg font-bold text-white">{hoursUntilReset}h</p>
          </div>
        </div>

        {/* Status Badge */}
        <div className="mt-4">
          {isExceeded ? (
              <div className="flex items-center gap-2 rounded-lg bg-red-500/15 px-3 py-2 text-xs font-semibold text-red-300">
                <AlertCircle size={14} aria-hidden="true" />
                Daily quota exceeded - contact support
              </div>
            ) : usagePercent > 80 ? (
              <div className="flex items-center gap-2 rounded-lg bg-yellow-500/15 px-3 py-2 text-xs font-semibold text-yellow-300">
                <AlertCircle size={14} aria-hidden="true" />
                Approaching quota limit ({(100 - usagePercent).toFixed(0)}%
                remaining)
              </div>
            ) : (
              <div className="flex items-center gap-2 rounded-lg bg-blue-500/15 px-3 py-2 text-xs font-semibold text-blue-300">
                <TrendingUp size={14} aria-hidden="true" />
                Quota available
              </div>
            )}
        </div>
      </div>

      {/* Daily Report History */}
      {showReportHistory && reportData.length > 0 && (
        <div className="rounded-2xl border border-white/10 bg-linear-to-br from-white/5 via-black/30 to-black/70 p-6">
          <h3 className="mb-4 flex items-center gap-2 text-sm font-semibold text-white">
            <TrendingUp size={16} className="text-blue-400" aria-hidden="true" />
            7-Day Usage History
          </h3>
          <div className="space-y-2">
            {reportData.map((report) => (
              <div
                key={report.date}
                className="flex items-center justify-between rounded-lg bg-white/5 px-3 py-2 text-xs"
              >
                <span className="text-white/70">
                  {new Date(report.date).toLocaleDateString()}
                </span>
                <div className="flex items-center gap-4">
                  <span className="text-white">
                    {report.tokens_used.toLocaleString()} tokens
                  </span>
                  <span className="text-blue-400">
                    {report.conversions_count} conversions
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
