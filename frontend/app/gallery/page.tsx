import type { Metadata } from "next";

import GalleryPageClient from "./GalleryPageClient";

export const metadata: Metadata = {
  title: "Public Gallery | MDFlow Studio",
  description: "Browse public MDFlow specifications shared by the community.",
  alternates: {
    canonical: "/gallery",
  },
  openGraph: {
    type: "website",
    url: "/gallery",
    title: "Public Gallery | MDFlow Studio",
    description: "Browse public MDFlow specifications shared by the community.",
    images: [
      {
        url: "https://md-spec-tool.vercel.app/opengraph-image",
        secureUrl: "https://md-spec-tool.vercel.app/opengraph-image",
        type: "image/png",
        width: 1200,
        height: 630,
        alt: "MDFlow Studio Public Gallery",
      },
    ],
    siteName: "MDFlow Studio",
  },
  twitter: {
    card: "summary_large_image",
    title: "Public Gallery | MDFlow Studio",
    description: "Browse public MDFlow specifications shared by the community.",
    images: ["https://md-spec-tool.vercel.app/opengraph-image"],
  },
};

export default function GalleryPage() {
  return <GalleryPageClient />;
}
