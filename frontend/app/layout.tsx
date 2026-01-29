import AppHeader from "@/components/AppHeader";
import ErrorBoundary from "@/components/ErrorBoundary";
import { Command } from "lucide-react";
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

export const metadata = {
  title: "MDFlow Studio | Technical Specification Automation",
  description:
    "Standardize engineering knowledge with automated MDFlow generation.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${inter.variable} ${geistMono.variable}`}>
      <body className="relative bg-bg-mesh">
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
                <a
                  href="/studio"
                  className="mt-1 h-9 sm:h-10 px-4 sm:px-6 rounded-lg bg-white text-black font-bold text-xs sm:text-sm tracking-wide hover:bg-white/95 hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 inline-flex items-center justify-center"
                >
                  Launch Studio
                </a>
              </div>

              {/* Dense footer bar: logo · nav · status · legal */}
              <div className="py-4 sm:py-5 lg:py-6 flex flex-wrap items-center justify-between gap-3 sm:gap-4 lg:gap-6">
                <div className="flex items-center gap-6 sm:gap-8 min-w-0">
                  <a
                    href="/"
                    className="flex items-center gap-1 shrink-0 opacity-90 hover:opacity-100 transition-opacity"
                    aria-label="MDFlow Home"
                  >
                    <svg
                      viewBox="0 0 180 40"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-5 sm:h-6 lg:h-7 w-auto"
                    >
                      <defs>
                        <linearGradient
                          id="footer-premium-gold"
                          x1="0%"
                          y1="0%"
                          x2="100%"
                          y2="100%"
                        >
                          <stop offset="0%" stopColor="#F7CE68" />
                          <stop offset="50%" stopColor="#FBAB7E" />
                          <stop offset="100%" stopColor="#f27b2f" />
                        </linearGradient>
                      </defs>
                      <path
                        d="M14 26L14 14C14 14 14 10 18 10C22 10 22 14 22 14L22 26"
                        stroke="url(#footer-premium-gold)"
                        strokeWidth="3"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                      <path
                        d="M22 26L22 18C22 18 22 14 26 14C30 14 30 18 30 18L30 26"
                        stroke="url(#footer-premium-gold)"
                        strokeWidth="3"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeOpacity="0.8"
                      />
                      <path
                        d="M30 26L30 22C30 22 30 18 34 18C38 18 38 22 38 22L38 26"
                        stroke="url(#footer-premium-gold)"
                        strokeWidth="3"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeOpacity="0.6"
                      />
                      <circle cx="42" cy="12" r="2" fill="#F7CE68" />
                      <text
                        x="54"
                        y="27"
                        fill="white"
                        fontSize="22"
                        fontWeight="800"
                        letterSpacing="-0.02em"
                        style={{ fontFamily: "'Inter', sans-serif" }}
                      >
                        MDFlow
                      </text>
                    </svg>
                  </a>
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
