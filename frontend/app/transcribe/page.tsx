import type { Metadata } from "next";
import { TranscribePageClient } from "./TranscribePageClient";

export const metadata: Metadata = {
  title: "Transcribe | MDFlow Studio",
  description:
    "Transcribe audio with timeline captions and split segments for quick editing.",
  alternates: {
    canonical: "/transcribe",
  },
  openGraph: {
    type: "website",
    url: "/transcribe",
    title: "Transcribe | MDFlow Studio",
    description:
      "Transcribe audio with timeline captions and split segments for quick editing.",
    siteName: "MDFlow Studio",
    images: [
      {
        url: "https://md-spec-tool.vercel.app/opengraph-image",
        secureUrl: "https://md-spec-tool.vercel.app/opengraph-image",
        type: "image/png",
        width: 1200,
        height: 630,
        alt: "MDFlow Studio Audio Transcribe",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Transcribe | MDFlow Studio",
    description:
      "Transcribe audio with timeline captions and split segments for quick editing.",
    images: ["https://md-spec-tool.vercel.app/opengraph-image"],
  },
};

export default function TranscribePage() {
  return <TranscribePageClient />;
}
