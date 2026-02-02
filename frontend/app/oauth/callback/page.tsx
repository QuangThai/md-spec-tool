"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";

function OAuthCallbackContent() {
  const params = useSearchParams();
  const [message, setMessage] = useState("Connecting Google account...");

  useEffect(() => {
    const code = params.get("code");
    const state = params.get("state");
    const error = params.get("error");

    if (error) {
      const errorMessage = `Google OAuth error: ${error}`;
      setMessage(errorMessage);
      window.opener?.postMessage(
        { type: "google-oauth-error", error: errorMessage },
        window.location.origin
      );
      return;
    }

    if (!code || !state) {
      const errorMessage = "Missing OAuth code or state";
      setMessage(errorMessage);
      window.opener?.postMessage(
        { type: "google-oauth-error", error: errorMessage },
        window.location.origin
      );
      return;
    }

    const exchange = async () => {
      try {
        const response = await fetch("/api/oauth/google/callback", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ code, state }),
        });

        if (!response.ok) {
          const data = await response.json().catch(() => ({}));
          throw new Error(data.error || "OAuth exchange failed");
        }

        setMessage("Connected. You can close this window.");
        window.opener?.postMessage(
          { type: "google-oauth-success" },
          window.location.origin
        );
        window.close();
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : "OAuth exchange failed";
        setMessage(errorMessage);
        window.opener?.postMessage(
          { type: "google-oauth-error", error: errorMessage },
          window.location.origin
        );
      }
    };

    void exchange();
  }, [params]);

  return <div className="text-sm text-white/80">{message}</div>;
}

export default function OAuthCallbackPage() {
  return (
    <main className="min-h-screen flex items-center justify-center bg-black text-white">
      <Suspense fallback={<div className="text-sm text-white/80">Loading...</div>}>
        <OAuthCallbackContent />
      </Suspense>
    </main>
  );
}
