import type { Metadata } from "next";

import BatchPageClient from "./BatchPageClient";

export const metadata: Metadata = {
  title: "Batch Processing | MDFlow Studio",
  description:
    "Process multiple spreadsheets in one run and download all outputs as a ZIP.",
  alternates: {
    canonical: "/batch",
  },
  keywords: [
    "MDFlow",
    "batch processing",
    "Excel to Markdown",
    "CSV to Markdown",
    "TSV to Markdown",
    "ZIP export",
  ],
  openGraph: {
    type: "website",
    url: "/batch",
    title: "Batch Processing | MDFlow Studio",
    description:
      "Process multiple spreadsheets in one run and download all outputs as a ZIP.",
    images: [
      {
        url: "https://md-spec-tool.vercel.app/batch/opengraph-image",
        secureUrl: "https://md-spec-tool.vercel.app/batch/opengraph-image",
        type: "image/png",
        width: 1200,
        height: 630,
        alt: "MDFlow Batch Processing",
      },
    ],
    siteName: "MDFlow Studio",
  },
  twitter: {
    card: "summary_large_image",
    title: "Batch Processing | MDFlow Studio",
    description:
      "Process multiple spreadsheets in one run and download all outputs as a ZIP.",
    images: ["https://md-spec-tool.vercel.app/batch/opengraph-image"],
  },
};

export default function BatchPage() {
  return <BatchPageClient />;
}
