import type { Metadata } from "next";

import ShareSlugPageClient from "./ShareSlugPageClient";

type PageParams = {
  params: {
    key: string;
  };
};

export function generateMetadata({ params }: PageParams): Metadata {
  const canonical = `/s/${params.key}`;

  return {
    title: "Shared Spec | MDFlow Studio",
    description: "View and download shared MDFlow specifications.",
    alternates: {
      canonical,
    },
    robots: {
      index: false,
      follow: false,
    },
    openGraph: {
      type: "website",
      url: canonical,
      title: "Shared Spec | MDFlow Studio",
      description: "View and download shared MDFlow specifications.",
      images: [
        {
          url: "https://md-spec-tool.vercel.app/opengraph-image",
          secureUrl: "https://md-spec-tool.vercel.app/opengraph-image",
          type: "image/png",
          width: 1200,
          height: 630,
          alt: "MDFlow Shared Spec",
        },
      ],
      siteName: "MDFlow Studio",
    },
    twitter: {
      card: "summary_large_image",
      title: "Shared Spec | MDFlow Studio",
      description: "View and download shared MDFlow specifications.",
      images: ["https://md-spec-tool.vercel.app/opengraph-image"],
    },
  };
}

export default function ShareSlugPage() {
  return <ShareSlugPageClient />;
}
