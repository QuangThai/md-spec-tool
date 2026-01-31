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
      <main className="max-w-7xl mx-auto px-3 sm:px-4 lg:px-6 xl:px-0 pt-8 sm:pt-10 lg:pt-12 min-h-[80vh]">
        {children}
      </main>

      <footer className="relative mt-16 border-t border-white/6 bg-linear-to-b from-bg-base/70 via-[#020202]/90 to-black overflow-hidden">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute -top-24 left-1/2 h-48 w-[520px] -translate-x-1/2 rounded-full bg-accent-orange/10 blur-[90px] opacity-60" />
        </div>

        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="py-8 sm:py-9 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6 border-b border-white/6">
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-5 min-w-0">
              <div className="inline-flex items-center gap-2 w-fit px-2.5 py-1 rounded-full bg-white/3 border border-white/10">
                <span className="relative w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.45)]" />
                <span className="text-[9px] font-semibold text-white/55 uppercase tracking-widest">
                  Live
                </span>
              </div>
              <div className="min-w-0">
                <p className="text-sm sm:text-base font-semibold text-white/95 tracking-tight">
                  Ready to scale your spec?
                </p>
                <p className="text-[11px] sm:text-xs text-white/45 mt-0.5">
                  MDFlow precision engine · docs in minutes
                </p>
              </div>
            </div>

            <div className="shrink-0">
              <div className="px-3 py-2 rounded-lg bg-white/2 border border-white/8">
                <p className="text-[11px] uppercase tracking-widest text-white/40">
                  Designed by
                </p>
                <p className="text-sm font-semibold tracking-tight text-white/90">
                  Quang Thai
                </p>
              </div>
            </div>
          </div>

          <div className="py-5 sm:py-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-4 sm:gap-6 min-w-0">
              <Link
                href="/"
                className="flex items-center gap-1 shrink-0 opacity-80 hover:opacity-100 transition-opacity duration-200"
                aria-label="MDFlow Home"
              >
                <svg
                  viewBox="0 0 140 32"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-7 sm:h-8 w-auto"
                  aria-label="MDFlow Logo"
                >
                  <defs>
                    <linearGradient id="footer-logo-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                      <stop offset="0%" stopColor="#F97316" />
                      <stop offset="50%" stopColor="#F59E0B" />
                      <stop offset="100%" stopColor="#EA580C" />
                    </linearGradient>
                    <linearGradient id="footer-logo-gradient-subtle" x1="0%" y1="0%" x2="100%" y2="0%">
                      <stop offset="0%" stopColor="#F97316" stopOpacity="0.9" />
                      <stop offset="100%" stopColor="#FB923C" stopOpacity="0.7" />
                    </linearGradient>
                  </defs>
                  <g opacity="0.85">
                    {/* Base M shape */}
                    <path
                      d="M4 24V10C4 8.89543 4.89543 8 6 8H8L14 18L20 8H22C23.1046 8 24 8.89543 24 10V24"
                      stroke="url(#footer-logo-gradient)"
                      strokeWidth="3"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      fill="none"
                    />
                    {/* Flow accent */}
                    <path
                      d="M27 8L31 16L27 24"
                      stroke="url(#footer-logo-gradient-subtle)"
                      strokeWidth="2.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      fill="none"
                    />
                    {/* Typography */}
                    <text
                      x="40"
                      y="22"
                      fontSize="17"
                      fontWeight="800"
                      letterSpacing="-0.03em"
                      style={{ fontFamily: "var(--font-inter), system-ui, sans-serif" }}
                    >
                      <tspan fill="white">MD</tspan>
                      <tspan fill="url(#footer-logo-gradient)">Flow</tspan>
                    </text>
                  </g>
                </svg>
              </Link>
              <nav
                className="hidden sm:flex items-center gap-0.5"
                aria-label="Footer navigation"
              >
                <Link
                  href="/studio"
                  className="px-3 py-1.5 text-xs font-medium uppercase tracking-widest text-white/50 hover:text-white/90 transition-colors duration-200 rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-orange/40"
                >
                  Studio
                </Link>
                <span className="text-white/10 mx-1" aria-hidden>
                  ·
                </span>
                <Link
                  href="/batch"
                  className="px-3 py-1.5 text-xs font-medium uppercase tracking-widest text-white/50 hover:text-white/90 transition-colors duration-200 rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-orange/40"
                >
                  Batch
                </Link>
                <span className="text-white/10 mx-1" aria-hidden>
                  ·
                </span>
                <Link
                  href="/docs"
                  className="px-3 py-1.5 text-xs font-medium uppercase tracking-widest text-white/50 hover:text-white/90 transition-colors duration-200 rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-orange/40"
                >
                  Docs
                </Link>
              </nav>
            </div>

            <div className="flex items-center gap-4 sm:gap-5 shrink-0">
              <div className="flex items-center gap-2 px-2.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20">
                <span className="relative w-1.5 h-1.5 shrink-0">
                  <span className="absolute inset-0 rounded-full bg-emerald-400 animate-pulse" />
                  <span className="absolute inset-0 rounded-full bg-emerald-500 shadow-[inset_0_0_4px_rgba(34,197,94,0.8)]" />
                </span>
                <span className="text-[8px] font-semibold text-emerald-200/70 uppercase tracking-widest">
                  Online
                </span>
              </div>
              <p className="text-[10px] font-medium uppercase tracking-widest text-white/30">
                © 2026 MDFlow
              </p>
            </div>
          </div>
        </div>
      </footer>
    </ErrorBoundary>
  );
}
