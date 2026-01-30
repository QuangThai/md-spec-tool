import MDFlowLanding from "@/components/MDFlowLanding";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Home",
  description:
    "Automate technical specifications from spreadsheets with MDFlow Studio.",
  alternates: {
    canonical: "/",
  },
  openGraph: {
    type: "website",
    url: "/",
    title: "MDFlow Studio | Technical Specification Automation",
    description:
      "Automate technical specifications from spreadsheets with MDFlow Studio.",
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
    title: "MDFlow Studio | Technical Specification Automation",
    description:
      "Automate technical specifications from spreadsheets with MDFlow Studio.",
    images: ["https://md-spec-tool.vercel.app/opengraph-image"],
  },
};

const homeJsonLd = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "MDFlow Studio",
  applicationCategory: "DeveloperApplication",
  operatingSystem: "Web",
  description:
    "Automate technical specifications from spreadsheets with MDFlow Studio.",
  url: "https://md-spec-tool.vercel.app",
  offers: {
    "@type": "Offer",
    price: "0",
    priceCurrency: "USD",
  },
};

const homeBreadcrumbJsonLd = {
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  itemListElement: [
    {
      "@type": "ListItem",
      position: 1,
      name: "Home",
      item: "https://md-spec-tool.vercel.app/",
    },
  ],
};

export default function Home() {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(homeJsonLd) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(homeBreadcrumbJsonLd) }}
      />
      <MDFlowLanding />
    </>
  );
}
