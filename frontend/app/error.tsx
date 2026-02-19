"use client";

import { AlertTriangle, RefreshCw } from "lucide-react";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="mx-auto flex min-h-[60vh] w-full max-w-2xl items-center justify-center px-4">
      <div className="w-full rounded-2xl border border-accent-red/30 bg-accent-red/10 p-6">
        <div className="mb-3 flex items-center gap-2 text-accent-red">
          <AlertTriangle className="h-5 w-5" />
          <h2 className="text-lg font-black tracking-tight">Something went wrong</h2>
        </div>
        <p className="text-sm text-red-200/90">
          {error?.message || "Unexpected runtime error."}
        </p>
        <button
          type="button"
          onClick={reset}
          className="mt-4 inline-flex h-10 items-center gap-2 rounded-lg border border-accent-red/40 bg-accent-red/20 px-4 text-xs font-bold uppercase tracking-wider text-red-100 hover:bg-accent-red/30"
        >
          <RefreshCw size={14} />
          Try Again
        </button>
      </div>
    </div>
  );
}
