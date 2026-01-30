import AppHeader from "@/components/AppHeader";
import ErrorBoundary from "@/components/ErrorBoundary";
import type { Metadata } from "next";
import { Geist_Mono as GeistMono, Inter } from "next/font/google";
import Link from "next/link";
import React from "react";
import "../styles/globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

const geistMono = GeistMono({
  subsets: ["latin"],
  variable: "--font-geist-mono",
  display: "swap",
});

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

const organizationJsonLd = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "MDFlow Studio",
  url: "https://md-spec-tool.vercel.app",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${inter.variable} ${geistMono.variable}`}>
      <body className="relative bg-bg-mesh">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(organizationJsonLd),
          }}
        />
        {/* Advanced Ambient Glows */}
        <div className="aurora-bg">
          <div className="aurora-blob left-[-10%] top-[-10%] h-[800px] w-[800px] bg-accent-orange/15 animate-pulse-soft" />
          <div className="aurora-blob right-[-5%] top-[10%] h-[600px] w-[600px] bg-accent-gold/5 [animation-delay:2s]" />
          <div className="aurora-blob bottom-[-10%] left-[20%] h-[900px] w-[900px] bg-accent-orange/10 [animation-delay:4s]" />
        </div>

        <ErrorBoundary>
          <AppHeader />
          <main className="max-w-7xl mx-auto pt-8 sm:pt-12 lg:pt-20 min-h-[80vh]">
            {children}
          </main>

          <footer className="relative mt-16 border-t border-white/5 bg-[#020202]/80 backdrop-blur-sm overflow-hidden">
            {/* Top accent: thin gradient line */}
            <div
              className="absolute left-0 right-0 top-0 h-px z-10"
              style={{
                background:
                  "linear-gradient(90deg, transparent 0%, rgba(242,123,47,0.35) 30%, rgba(242,123,47,0.5) 50%, rgba(242,123,47,0.35) 70%, transparent 100%)",
              }}
            />
            <div
              className="absolute inset-0 opacity-[0.015]"
              style={{
                backgroundImage:
                  "linear-gradient(#fff 1px, transparent 1px), linear-gradient(90deg, #fff 1px, transparent 1px)",
                backgroundSize: "20px 20px",
              }}
            />

            <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 relative z-20">
              {/* Single-row CTA strip: badge + copy + button — compact, one line on lg */}
              <div className="py-5 sm:py-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 sm:gap-6 border-b border-white/5">
                <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-5 min-w-0">
                  <div className="inline-flex items-center gap-1.5 w-fit px-2 py-0.5 rounded-md bg-accent-orange/8 border border-accent-orange/15">
                    <span className="w-1 h-1 rounded-full bg-emerald-500 shrink-0" />
                    <span className="text-[9px] font-bold text-white/60 uppercase tracking-wider">
                      v1.2.0
                    </span>
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm sm:text-base font-semibold text-white/95 tracking-tight">
                      Ready to scale your spec?
                    </p>
                    <p className="text-[11px] sm:text-xs text-muted mt-0.5">
                      MDFlow precision engine · docs in minutes
                    </p>
                  </div>
                </div>
                <Link
                  href="/studio"
                  className="shrink-0 h-9 px-5 rounded-lg bg-white/20 text-white font-bold text-xs tracking-wide hover:bg-white/95 hover:shadow-lg hover:shadow-white/10 active:scale-[0.98] transition-all duration-200 inline-flex items-center justify-center"
                >
                  Launch Studio
                </Link>
              </div>

              {/* Dense footer bar: logo · nav · status · legal */}
              <div className="py-3 sm:py-4 flex flex-wrap items-center justify-between gap-3 sm:gap-4">
                <div className="flex items-center gap-4 sm:gap-6 min-w-0">
                  <Link
                    href="/"
                    className="flex items-center gap-1 shrink-0 opacity-90 hover:opacity-100 transition-opacity"
                    aria-label="MDFlow Home"
                  >
                    <svg
                      viewBox="0 0 160 40"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-8 sm:h-9 w-auto"
                      aria-label="MDFlow Logo"
                    >
                      <defs>
                        <linearGradient
                          id="footer-flow-gradient"
                          x1="0%"
                          y1="0%"
                          x2="100%"
                          y2="0%"
                        >
                          <stop offset="0%" stopColor="#F59E0B" />
                          <stop offset="100%" stopColor="#F97316" />
                        </linearGradient>
                      </defs>
                      <g opacity="0.9">
                        <path
                          d="M10 30V14C10 10.6863 12.6863 8 16 8C19.3137 8 22 10.6863 22 14V22C22 23.1046 22.8954 24 24 24C25.1046 24 26 23.1046 26 22V14C26 10.6863 28.6863 8 32 8C35.3137 8 38 10.6863 38 14V30"
                          stroke="url(#footer-flow-gradient)"
                          strokeWidth="5"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        />
                        <circle cx="24" cy="31" r="2.5" fill="#F97316" />
                        <text
                          x="54"
                          y="28"
                          fill="white"
                          fontSize="22"
                          fontWeight="700"
                          letterSpacing="-0.02em"
                          style={{ fontFamily: "'Inter', sans-serif" }}
                        >
                          MDFlow
                        </text>
                      </g>
                    </svg>
                  </Link>
                  <nav
                    className="flex items-center gap-1 text-muted"
                    aria-label="Footer navigation"
                  >
                    <Link
                      href="/studio"
                      className="px-2 py-1 text-xs font-semibold uppercase tracking-wider text-white/70 hover:text-accent-orange transition-colors rounded"
                    >
                      Studio
                    </Link>
                    <span className="text-white/20" aria-hidden>
                      ·
                    </span>
                    <Link
                      href="/docs"
                      className="px-2 py-1 text-xs font-semibold uppercase tracking-wider text-white/70 hover:text-accent-orange transition-colors rounded"
                    >
                      Docs
                    </Link>
                  </nav>
                </div>

                <div className="flex items-center gap-3 sm:gap-4 shrink-0">
                  <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-white/5 border border-white/10">
                    <span className="relative flex h-1.5 w-1.5 shrink-0">
                      <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
                    </span>
                    <span className="text-[9px] font-semibold text-white/45 uppercase tracking-wider">
                      Systems OK
                    </span>
                  </div>
                  <p className="text-[9px] font-medium uppercase tracking-wider text-muted">
                    © 2026 MDFlow
                  </p>
                </div>
              </div>
            </div>
          </footer>
        </ErrorBoundary>
      </body>
    </html>
  );
}
