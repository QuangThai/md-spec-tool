import MDFlowWorkbench from "@/components/MDFlowWorkbench";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Studio",
  description: "Advanced technical specification transformation engine.",
  alternates: {
    canonical: "/studio",
  },
  robots: {
    index: false,
    follow: false,
  },
  openGraph: {
    type: "website",
    url: "/studio",
    title: "Studio | MDFlow Studio",
    description: "Advanced technical specification transformation engine.",
    images: [
      {
        url: "https://md-spec-tool.vercel.app/opengraph-image",
        secureUrl: "https://md-spec-tool.vercel.app/opengraph-image",
        type: "image/png",
        width: 1200,
        height: 630,
        alt: "MDFlow Studio",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Studio | MDFlow Studio",
    description: "Advanced technical specification transformation engine.",
    images: ["https://md-spec-tool.vercel.app/opengraph-image"],
  },
};

export default function StudioPage() {
  return (
    <div className="max-w-7xl mx-auto">
      <MDFlowWorkbench />
    </div>
  );
}
