"use client";

import dynamic from "next/dynamic";
import { Suspense } from "react";

// ✅ Dynamic import for heavy docs component (~1400 lines, markdown rendering)
const DocsContent = dynamic(() => import("@/components/DocsContent"), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center min-h-screen">
      <div className="animate-pulse text-white/50 text-sm">
        Loading documentation...
      </div>
    </div>
  ),
});

export function DocsPageClient() {
  return (
    <Suspense fallback={<div className="animate-pulse min-h-screen" />}>
      <DocsContent />
    </Suspense>
  );
}
