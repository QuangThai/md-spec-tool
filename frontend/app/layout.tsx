import AppHeader from "@/components/AppHeader";
import ErrorBoundary from "@/components/ErrorBoundary";
import { Geist_Mono as GeistMono, Inter } from "next/font/google";
import Link from "next/link";
import React from "react";
import type { Metadata } from "next";
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
          dangerouslySetInnerHTML={{ __html: JSON.stringify(organizationJsonLd) }}
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

          <footer className="relative mt-20 border-t border-white/5 bg-[#020202] overflow-hidden">
            {/* Subtle top accent line */}
            <div
              className="absolute left-0 right-0 top-0 h-px z-10 opacity-60"
              style={{
                background:
                  "linear-gradient(90deg, transparent, rgba(242,123,47,0.4) 20%, rgba(242,123,47,0.6) 50%, rgba(242,123,47,0.4) 80%, transparent)",
              }}
            />
            <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_0%,rgba(242,123,47,0.04),transparent)]" />
            <div
              className="absolute inset-0 opacity-[0.02]"
              style={{
                backgroundImage:
                  "linear-gradient(#fff 1px, transparent 1px), linear-gradient(90deg, #fff 1px, transparent 1px)",
                backgroundSize: "24px 24px",
              }}
            />

            <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 relative z-20">
              {/* Compact CTA block */}
              <div className="py-8 sm:py-10 lg:py-12 flex flex-col items-center text-center space-y-3 sm:space-y-4 border-b border-white/5">
                <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-accent-orange/10 border border-accent-orange/20 text-accent-orange text-[9px] sm:text-[10px] font-bold uppercase tracking-widest">
                  <span className="w-1.5 h-1.5 rounded-full bg-accent-orange" />
                  System v1.2.0 Live
                </div>
                <h2 className="text-2xl sm:text-4xl md:text-5xl font-black text-white tracking-tighter leading-tight">
                  Ready to{" "}
                  <span className="text-transparent bg-clip-text bg-linear-to-br from-white to-white/70">
                    Scale Your Spec?
                  </span>
                </h2>
                <p className="max-w-md text-xs sm:text-sm lg:text-base text-muted font-medium">
                  Automate technical documentation with MDFlow's precision
                  engine.
                </p>
                <Link
                  href="/studio"
                  className="mt-1 h-9 sm:h-10 px-4 sm:px-6 rounded-lg bg-white text-black font-bold text-xs sm:text-sm tracking-wide hover:bg-white/95 hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 inline-flex items-center justify-center"
                >
                  Launch Studio
                </Link>
              </div>

              {/* Dense footer bar: logo · nav · status · legal */}
              <div className="py-4 sm:py-5 lg:py-6 flex flex-wrap items-center justify-between gap-3 sm:gap-4 lg:gap-6">
                <div className="flex items-center gap-6 sm:gap-8 min-w-0">
                  <Link
                    href="/"
                    className="flex items-center gap-1 shrink-0 opacity-90 hover:opacity-100 transition-opacity"
                    aria-label="MDFlow Home"
                  >
                    <svg
                      viewBox="0 0 160 40"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-10 sm:h-12 w-auto"
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

                <div className="flex items-center gap-4 sm:gap-6 shrink-0">
                  <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/5 border border-white/10">
                    <span className="relative flex h-1.5 w-1.5">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-500/80" />
                      <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
                    </span>
                    <span className="text-[10px] font-bold text-white/50 uppercase tracking-widest">
                      All Systems Normal
                    </span>
                  </div>
                  <p className="text-[10px] font-semibold uppercase tracking-widest text-muted">
                    © 2026 MDFlow Inc.
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
