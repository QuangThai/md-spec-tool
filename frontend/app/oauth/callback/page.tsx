import { Suspense } from "react";
import OAuthCallbackClient from "./OAuthCallbackClient";

export const metadata = {
  title: "Google Sign-in | MDFlow Studio",
  description: "Completing Google account connection for MDFlow Studio.",
};

export default function OAuthCallbackPage() {
  return (
    <main className="min-h-screen flex items-center justify-center bg-black text-white">
      <Suspense fallback={<div className="text-sm text-white/80">Loading…</div>}>
        <OAuthCallbackClient />
      </Suspense>
    </main>
  );
}
