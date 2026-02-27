import type { Metadata } from "next";
import { Suspense } from "react";

import SharePageClient from "./SharePageClient";

export const metadata: Metadata = {
  title: "Shared Spec | MDFlow Studio",
  description:
    "View and download shared MDFlow specifications directly in the browser.",
  alternates: {
    canonical: "/share",
  },
  robots: {
    index: false,
    follow: false,
  },
  openGraph: {
    type: "website",
    url: "/share",
    title: "Shared Spec | MDFlow Studio",
    description:
      "View and download shared MDFlow specifications directly in the browser.",
    images: [
      {
        url: "https://md-spec-tool.vercel.app/opengraph-image",
        secureUrl: "https://md-spec-tool.vercel.app/opengraph-image",
        type: "image/png",
        width: 1200,
        height: 630,
        alt: "Shared MDFlow Spec",
      },
    ],
    siteName: "MDFlow Studio",
  },
  twitter: {
    card: "summary_large_image",
    title: "Shared Spec | MDFlow Studio",
    description:
      "View and download shared MDFlow specifications directly in the browser.",
    images: ["https://md-spec-tool.vercel.app/opengraph-image"],
  },
};

function SharePageFallback() {
  return (
    <div className="min-h-screen bg-black flex items-center justify-center">
      <div className="flex items-center gap-3 text-white/60">
        <div className="w-5 h-5 border-2 border-orange-400/30 border-t-orange-400 rounded-full animate-spin" />
        <span className="text-sm font-medium">Loading…</span>
      </div>
    </div>
  );
}

export default function SharePage() {
  return (
    <Suspense fallback={<SharePageFallback />}>
      <SharePageClient />
    </Suspense>
  );
}
