import type { Metadata } from "next";

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

export default function SharePage() {
  return <SharePageClient />;
}
