"use client";

import { Suspense } from "react";
import { SharePageContent } from "./SharePageContent";

export default function SharePageClient() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-black flex items-center justify-center">
          <div className="flex items-center gap-3 text-white/60">
            <div className="w-5 h-5 border-2 border-accent-orange/30 border-t-accent-orange rounded-full animate-spin" />
            <span className="text-sm font-medium">Loading…</span>
          </div>
        </div>
      }
    >
      <SharePageContent />
    </Suspense>
  );
}
