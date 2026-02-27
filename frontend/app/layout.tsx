import AppShell from "@/app/AppShell";
import { Providers } from "@/app/providers";
import type { Metadata } from "next";
import React from "react";
import "../styles/globals.css";

export const metadata: Metadata = {
  metadataBase: new URL("https://md-spec-tool.vercel.app"),
  title: {
    default: "MDFlow Studio | Technical Specification Automation",
    template: "%s | MDFlow Studio",
  },
  description:
    "Standardize engineering knowledge with automated MDFlow generation.",
  alternates: {
    canonical: "/",
  },
  openGraph: {
    type: "website",
    url: "/",
    title: "MDFlow Studio | Technical Specification Automation",
    description:
      "Standardize engineering knowledge with automated MDFlow generation.",
    siteName: "MDFlow Studio",
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
      "Standardize engineering knowledge with automated MDFlow generation.",
    images: ["https://md-spec-tool.vercel.app/opengraph-image"],
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark" style={{ colorScheme: "dark" }}>
      <body className="relative bg-bg-mesh">
        <a
          href="#main"
          className="sr-only focus-visible:fixed focus-visible:left-4 focus-visible:top-4 focus-visible:z-100 focus-visible:m-0 focus-visible:w-auto focus-visible:h-auto focus-visible:overflow-visible focus-visible:[clip-path:unset] focus-visible:rounded focus-visible:bg-accent-orange focus-visible:px-4 focus-visible:py-2 focus-visible:text-sm focus-visible:font-bold focus-visible:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white"
        >
          Skip to main content
        </a>
        <Providers>
          <AppShell>{children}</AppShell>
        </Providers>
      </body>
    </html>
  );
}
