import AppHeader from "@/components/AppHeader";
import ErrorBoundary from "@/components/ErrorBoundary";
import Link from "next/link";
import React from "react";

const organizationJsonLd = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "MDFlow Studio",
  url: "https://md-spec-tool.vercel.app",
};

interface AppShellProps {
  children: React.ReactNode;
}

export default function AppShell({ children }: AppShellProps) {
  return (
    <ErrorBoundary>
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

      <AppHeader />
      <main className="max-w-7xl mx-auto pt-8 sm:pt-12 lg:pt-20 min-h-[80vh]">
        {children}
      </main>

      <footer className="relative mt-20 border-t border-white/5 bg-linear-to-b from-[#020202]/60 via-[#020202]/80 to-[#000000]/90 backdrop-blur-lg overflow-hidden">
        {/* Top accent: double gradient line */}
        <div
          className="absolute left-0 right-0 top-0 h-0.5 z-10"
          style={{
            background:
              "linear-gradient(90deg, transparent 0%, rgba(242,123,47,0.3) 25%, rgba(242,123,47,0.5) 50%, rgba(242,123,47,0.3) 75%, transparent 100%)",
          }}
        />
        <div
          className="absolute left-0 right-0 top-0.5 h-px z-10 opacity-40"
          style={{
            background:
              "linear-gradient(90deg, transparent 0%, rgba(249,115,22,0.2) 50%, transparent 100%)",
          }}
        />
        <div
          className="absolute inset-0 opacity-[0.008]"
          style={{
            backgroundImage:
              "linear-gradient(#fff 1px, transparent 1px), linear-gradient(90deg, #fff 1px, transparent 1px)",
            backgroundSize: "20px 20px",
          }}
        />

        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 relative z-20">
          {/* Single-row CTA strip: badge + copy + button — compact, one line on lg */}
          <div className="py-6 sm:py-7 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 sm:gap-6 border-b border-white/3">
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-5 min-w-0">
              <div className="inline-flex items-center gap-1.5 w-fit px-2.5 py-1 rounded-sm bg-accent-orange/5 border border-accent-orange/20 hover:border-accent-orange/40 transition-colors">
                <span className="relative w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0 shadow-[0_0_6px_rgba(52,211,153,0.5)]" />
                <span className="text-[8px] font-semibold text-white/50 uppercase tracking-widest">
                  Stream
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
            <div className="shrink-0 group relative inline-flex items-center justify-center">
              <div className="absolute -inset-3 rounded-lg bg-linear-to-r from-accent-orange/0 via-accent-orange/5 to-accent-orange/0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 blur-md" />
              <div className="relative px-4 py-2">
                <p className="text-sm font-light tracking-wide text-white/60 group-hover:text-white/80 transition-colors duration-300">
                  Designed by
                </p>
                <p className="text-base font-semibold tracking-tight text-white/95 group-hover:text-accent-orange transition-colors duration-300">
                  Quang Thai
                </p>
              </div>
            </div>
          </div>

          {/* Dense footer bar: logo · nav · status · legal */}
          <div className="py-4 sm:py-5 flex flex-wrap items-center justify-between gap-3 sm:gap-6">
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
                className="hidden sm:flex items-center gap-0.5 text-muted"
                aria-label="Footer navigation"
              >
                <Link
                  href="/studio"
                  className="px-3 py-1.5 text-xs font-medium uppercase tracking-widest text-white/50 hover:text-accent-orange hover:bg-white/2 transition-all rounded-sm"
                >
                  Studio
                </Link>
                <span className="text-white/10 mx-1" aria-hidden>
                  ·
                </span>
                <Link
                  href="/batch"
                  className="px-3 py-1.5 text-xs font-medium uppercase tracking-widest text-white/50 hover:text-accent-orange hover:bg-white/2 transition-all rounded-sm"
                >
                  Batch
                </Link>
                <span className="text-white/10 mx-1" aria-hidden>
                  ·
                </span>
                <Link
                  href="/docs"
                  className="px-3 py-1.5 text-xs font-medium uppercase tracking-widest text-white/50 hover:text-accent-orange hover:bg-white/2 transition-all rounded-sm"
                >
                  Docs
                </Link>
              </nav>
            </div>

            <div className="flex items-center gap-4 sm:gap-5 shrink-0">
              <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-sm bg-emerald-500/8 border border-emerald-500/20">
                <span className="relative w-1.5 h-1.5 shrink-0 overflow-hidden rounded-full">
                  <span className="absolute inset-0 rounded-full bg-emerald-400 animate-pulse" />
                  <span className="absolute inset-0 rounded-full bg-emerald-500 shadow-[inset_0_0_4px_rgba(34,197,94,0.8)]" />
                </span>
                <span className="text-[7px] font-semibold text-emerald-200/60 uppercase tracking-widest">
                  Online
                </span>
              </div>
              <p className="text-[8px] font-medium uppercase tracking-widest text-white/30">
                © 2026 MDFlow
              </p>
            </div>
          </div>
        </div>
      </footer>
    </ErrorBoundary>
  );
}
