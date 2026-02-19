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
    <html lang="en">
      <body className="relative bg-bg-mesh">
        <Providers>
          <AppShell>{children}</AppShell>
        </Providers>
      </body>
    </html>
  );
}
