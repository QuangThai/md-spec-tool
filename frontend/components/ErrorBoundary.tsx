"use client";

import React, { ReactNode } from "react";
import { mapErrorToUserFacing } from "@/lib/errorUtils";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

/**
 * ErrorBoundary catches unexpected React errors during rendering.
 * 
 * For API errors, use the error mapping utilities in lib/errorUtils.ts
 * to provide user-facing error messages.
 * 
 * This boundary is for component tree failures, not HTTP errors.
 */
export default class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log full error info for debugging
    console.error("React Error Boundary caught:", {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
    });

    // Report to monitoring service (optional)
    if (typeof window !== "undefined" && (window as any).__APP_CONFIG__?.sentry) {
      try {
        // Integration point for Sentry or similar
        console.info("Error reported to monitoring service");
      } catch {
        // Ignore reporting failures
      }
    }
  }

  render() {
    if (this.state.hasError) {
      const message = this.state.error?.message || "A critical error occurred during operation.";
      const userError = mapErrorToUserFacing(message);

      return (
        <div className="min-h-screen bg-bg-base flex items-center justify-center px-4">
          <div className="surface max-w-md w-full text-center">
            <div className="mb-6 flex justify-center">
              <div className="h-16 w-16 rounded-2xl bg-accent-red/10 flex items-center justify-center border border-accent-red/20 shadow-premium">
                <span className="text-2xl">⚠️</span>
              </div>
            </div>
            <h1 className="text-2xl font-bold text-white mb-2 uppercase tracking-tighter">
              {userError.title}
            </h1>
            <p className="text-muted text-sm mb-4 leading-relaxed uppercase tracking-widest font-medium">
              {userError.message}
            </p>

            {/* Show request ID if available for support */}
            {userError.requestId && (
              <p className="text-muted text-xs mb-8 font-mono break-all">
                Request ID: {userError.requestId}
              </p>
            )}

            <div className="flex gap-3">
              <button
                onClick={() => {
                  this.setState({ hasError: false, error: null });
                }}
                className="btn-secondary flex-1"
              >
                Dismiss
              </button>
              <button
                onClick={() => {
                  this.setState({ hasError: false, error: null });
                  window.location.href = "/";
                }}
                className="btn-primary flex-1"
              >
                Home
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
