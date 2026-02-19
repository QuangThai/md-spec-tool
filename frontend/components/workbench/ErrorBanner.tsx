"use client";

import React, { memo } from "react";
import { AlertCircle } from "lucide-react";
import { motion } from "framer-motion";

export interface ErrorBannerProps {
  error: string;
  mappedAppError: {
    title?: string;
    message?: string;
    requestId?: string;
    retryable?: boolean;
  } | null;
  lastFailedAction: "preview" | "convert" | "other" | null;
  loading: boolean;
  previewLoading: boolean;
  onRetry: () => void;
}

export const ErrorBanner = memo(function ErrorBanner({
  error,
  mappedAppError,
  lastFailedAction,
  loading,
  previewLoading,
  onRetry,
}: ErrorBannerProps) {
  if (!error) return null;

  const isRetrying = loading || previewLoading;
  const canRetry =
    mappedAppError?.retryable &&
    (lastFailedAction === "preview" || lastFailedAction === "convert");

  return (
    <motion.div
      initial={{ opacity: 0, y: -8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -8 }}
      className="mb-3 rounded-lg border border-accent-red/25 bg-accent-red/10 p-2.5 shrink-0"
    >
      <div className="flex items-start gap-2">
        <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-accent-red" />
        <div className="min-w-0 flex-1">
          <p className="text-[9px] font-bold uppercase tracking-wider text-accent-red">
            {mappedAppError?.title || "Request Failed"}
          </p>
          <p className="mt-1 text-[11px] text-accent-red/90">
            {mappedAppError?.message || error}
          </p>
          {mappedAppError?.requestId ? (
            <p className="mt-1 text-[10px] font-mono text-accent-red/70">
              request_id={mappedAppError.requestId}
            </p>
          ) : null}
          {canRetry ? (
            <button
              type="button"
              onClick={onRetry}
              disabled={isRetrying}
              className="mt-2 inline-flex h-7 items-center rounded-md border border-accent-red/35 bg-accent-red/15 px-2.5 text-[10px] font-bold uppercase tracking-wider text-accent-red hover:bg-accent-red/25 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {lastFailedAction === "convert" ? "Retry Convert" : "Retry Preview"}
            </button>
          ) : null}
        </div>
      </div>
    </motion.div>
  );
});

ErrorBanner.displayName = "ErrorBanner";
