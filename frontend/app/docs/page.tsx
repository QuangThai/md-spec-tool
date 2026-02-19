import type { Metadata } from "next";
import { DocsPageClient } from "./DocsPageClient";

export const metadata: Metadata = {
  title: "Documentation | MDFlow Studio",
  description:
    "Guides and API reference for MDFlow Studio: conversion pipeline, AI/BYOK mapping, Google Sheets OAuth, templates, validation, diff, sharing, and collaboration APIs.",
  alternates: {
    canonical: "/docs",
  },
  keywords: [
    "MDFlow",
    "documentation",
    "API reference",
    "architecture",
    "features",
    "BYOK",
    "Google OAuth",
    "batch processing",
    "share API",
    "collaboration",
    "AI suggestions",
    "technical specifications",
    "Excel to Markdown",
  ],
  openGraph: {
    type: "article",
    url: "/docs",
    title: "Documentation | MDFlow Studio",
    description:
      "Guides and API reference for MDFlow Studio: conversion pipeline, AI/BYOK mapping, Google Sheets OAuth, templates, validation, diff, sharing, and collaboration APIs.",
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
      "Guides and API reference for MDFlow Studio: AI/BYOK mapping, Google Sheets OAuth, templates, validation, diff, and sharing.",
    images: ["https://md-spec-tool.vercel.app/docs/opengraph-image"],
    creator: "@mdflow",
  },
};

const docsJsonLd = {
  "@context": "https://schema.org",
  "@type": "TechArticle",
  headline: "MDFlow Studio - Complete Documentation",
  description:
    "Documentation covering architecture, API endpoints, conversion/preview flows, AI/BYOK behavior, Google Sheets OAuth integration, templates, validation, diff, and sharing.",
  image: "https://md-spec-tool.vercel.app/docs/opengraph-image",
  keywords: [
    "MDFlow",
    "technical specifications",
    "Excel to Markdown",
    "API documentation",
    "BYOK",
    "Google OAuth",
    "share links",
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
      <DocsPageClient />
    </>
  );
}
