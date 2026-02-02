"use client";

import { useEffect } from "react";

export default function OAuthCallbackError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Notify opener about error
    window.opener?.postMessage(
      { type: "google-oauth-error", error: error.message || "OAuth callback error" },
      window.location.origin
    );
  }, [error.message]);

  return (
    <main className="min-h-screen flex items-center justify-center bg-black text-white">
      <div className="text-center max-w-sm">
        <h1 className="text-2xl font-bold mb-2">OAuth Error</h1>
        <p className="text-sm text-red-400 mb-6 wrap-break-word">{error.message || "An error occurred during authentication"}</p>
        <button
          onClick={reset}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-sm font-medium transition"
        >
          Try Again
        </button>
      </div>
    </main>
  );
}
