import DocsContent from "@/components/DocsContent";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Documentation | MDFlow Studio",
  description:
    "Complete MDFlow documentation covering architecture, API reference, features, input formats, conversion engine, AI suggestions, batch processing, and deployment guide.",
  alternates: {
    canonical: "/docs",
  },
  keywords: [
    "MDFlow",
    "documentation",
    "API reference",
    "architecture",
    "features",
    "batch processing",
    "AI suggestions",
    "technical specifications",
    "Excel to Markdown",
  ],
  openGraph: {
    type: "article",
    url: "/docs",
    title: "Documentation | MDFlow Studio",
    description:
      "Complete MDFlow documentation: architecture, API reference, features guide, input formats, conversion engine, templates, AI suggestions, batch processing, Google Sheets integration, diff & comparison, and deployment.",
    images: [
      {
        url: "https://md-spec-tool.vercel.app/docs/opengraph-image",
        secureUrl: "https://md-spec-tool.vercel.app/docs/opengraph-image",
        type: "image/png",
        width: 1200,
        height: 630,
        alt: "MDFlow Documentation",
      },
    ],
    siteName: "MDFlow Studio",
  },
  twitter: {
    card: "summary_large_image",
    title: "Documentation | MDFlow Studio",
    description:
      "Complete MDFlow documentation: architecture, API, features, batch processing, AI suggestions, and deployment guide.",
    images: ["https://md-spec-tool.vercel.app/docs/opengraph-image"],
    creator: "@mdflow",
  },
};

const docsJsonLd = {
  "@context": "https://schema.org",
  "@type": "TechArticle",
  headline: "MDFlow Studio - Complete Documentation",
  description:
    "Complete documentation covering system architecture, API reference, features guide, batch processing, AI suggestions, Google Sheets integration, diff & comparison tools, conversion engine mechanics, templates, input formats, and deployment guide.",
  image: "https://md-spec-tool.vercel.app/docs/opengraph-image",
  keywords: [
    "MDFlow",
    "technical specifications",
    "Excel to Markdown",
    "API documentation",
    "batch processing",
    "AI suggestions",
    "Google Sheets",
  ],
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
  datePublished: "2024-01-01",
  dateModified: new Date().toISOString().split("T")[0],
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
      name: "Studio",
      item: "https://md-spec-tool.vercel.app/studio",
    },
    {
      "@type": "ListItem",
      position: 3,
      name: "Documentation",
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
