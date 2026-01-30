import DocsContent from "@/components/DocsContent";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Docs",
  description:
    "MDFlow documentation covering architecture, parsing logic, and templates.",
  alternates: {
    canonical: "/docs",
  },
  openGraph: {
    type: "article",
    url: "/docs",
    title: "Docs | MDFlow Studio",
    description:
      "MDFlow documentation covering architecture, parsing logic, and templates.",
    images: [
      {
        url: "https://md-spec-tool.vercel.app/docs/opengraph-image",
        width: 1200,
        height: 630,
        alt: "MDFlow Docs",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Docs | MDFlow Studio",
    description:
      "MDFlow documentation covering architecture, parsing logic, and templates.",
    images: ["https://md-spec-tool.vercel.app/docs/opengraph-image"],
  },
};

const docsJsonLd = {
  "@context": "https://schema.org",
  "@type": "TechArticle",
  headline: "MDFlow Documentation",
  description:
    "MDFlow documentation covering architecture, parsing logic, and templates.",
  author: {
    "@type": "Organization",
    name: "MDFlow Studio",
    url: "https://md-spec-tool.vercel.app",
  },
  publisher: {
    "@type": "Organization",
    name: "MDFlow Studio",
    url: "https://md-spec-tool.vercel.app",
  },
  url: "https://md-spec-tool.vercel.app/docs",
};

const docsBreadcrumbJsonLd = {
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  itemListElement: [
    {
      "@type": "ListItem",
      position: 1,
      name: "Home",
      item: "https://md-spec-tool.vercel.app/",
    },
    {
      "@type": "ListItem",
      position: 2,
      name: "Docs",
      item: "https://md-spec-tool.vercel.app/docs",
    },
  ],
};

export default function DocsPage() {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(docsJsonLd) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(docsBreadcrumbJsonLd) }}
      />
      <DocsContent />
    </>
  );
}
