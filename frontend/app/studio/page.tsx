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
  },
  twitter: {
    card: "summary_large_image",
    title: "Studio | MDFlow Studio",
    description: "Advanced technical specification transformation engine.",
  },
};

export default function StudioPage() {
  return (
    <div className="space-y-4 sm:space-y-6 lg:space-y-8 px-4 sm:px-6 lg:px-8 max-w-[1600px] mx-auto">
      <MDFlowWorkbench />
    </div>
  );
}
