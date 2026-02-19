"use client";

import dynamic from "next/dynamic";
import { Suspense } from "react";

// ✅ Dynamic import for heavy audio component (~600 lines, multiple deps)
const AudioTranscribeStudio = dynamic(
  () => import("@/components/AudioTranscribeStudio"),
  {
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-pulse text-white/50 text-sm">
          Loading audio transcriber...
        </div>
      </div>
    ),
  }
);

export function TranscribePageClient() {
  return (
    <Suspense fallback={<div className="animate-pulse min-h-screen" />}>
      <AudioTranscribeStudio />
    </Suspense>
  );
}
