"use client";

import { motion } from "framer-motion";
import Link from "next/link";

const MDFlowLogo = ({ className }: { className?: string }) => (
  <svg
    viewBox="0 0 160 40"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
    aria-label="MDFlow Logo"
  >
    <defs>
      <linearGradient id="flow-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
        <stop offset="0%" stopColor="#F59E0B" />
        <stop offset="100%" stopColor="#F97316" />
      </linearGradient>
      <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
        <feGaussianBlur stdDeviation="2" result="coloredBlur" />
        <feComposite in="coloredBlur" in2="SourceGraphic" operator="in" />
        <feMerge>
          <feMergeNode in="coloredBlur" />
          <feMergeNode in="SourceGraphic" />
        </feMerge>
      </filter>
    </defs>

    {/* Fluid 'M' Symbol */}
    <g filter="url(#glow)">
      <path
        d="M10 30V14C10 10.6863 12.6863 8 16 8C19.3137 8 22 10.6863 22 14V22C22 23.1046 22.8954 24 24 24C25.1046 24 26 23.1046 26 22V14C26 10.6863 28.6863 8 32 8C35.3137 8 38 10.6863 38 14V30"
        stroke="url(#flow-gradient)"
        strokeWidth="5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {/* Subtle accent dot for 'spec' precision */}
      <circle cx="24" cy="31" r="2.5" fill="#F97316" />
    </g>

    {/* Text: Modern & Clean */}
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
  </svg>
);

export default function AppHeader() {
  return (
    <motion.header
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="sticky top-3 sm:top-4 lg:top-6 z-50 px-3 sm:px-4 lg:px-6"
    >
      <div className="mx-auto max-w-7xl">
        <div className="glass relative flex h-14 sm:h-16 lg:h-20 items-center justify-between px-4 sm:px-6 lg:px-10 rounded-2xl sm:rounded-3xl border-white/10 shadow-3xl overflow-hidden ring-1 ring-white/10">
          {/* Subtle brand flow effect */}
          <div className="absolute left-0 top-0 h-full w-48 bg-accent-orange/10 blur-[80px] pointer-events-none" />

          <div className="flex items-center gap-4 sm:gap-8 lg:gap-12 z-10">
            {/* Ultra-Premium MDFlow Brand Logo */}
            <Link href="/" className="group cursor-pointer py-2">
              <div className="relative">
                <div className="absolute -inset-4 rounded-full bg-linear-to-r from-accent-gold/20 via-accent-orange/10 to-transparent blur-xl opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                <MDFlowLogo className="h-8 sm:h-9 lg:h-11 w-auto drop-shadow-lg group-hover:scale-110 transition-transform duration-300" />
              </div>
            </Link>

            <nav className="hidden lg:flex items-center gap-6 xl:gap-10">
              {[
                { label: "Studio", href: "/studio" },
                { label: "Batch", href: "/batch" },
                { label: "Docs", href: "/docs" },
              ].map((item) => (
                <Link
                  key={item.label}
                  href={item.href}
                  className="text-[10px] lg:text-[11px] font-black uppercase tracking-[0.25em] lg:tracking-[0.3em] text-white/40 hover:text-accent-orange transition-all hover:tracking-[0.4em]"
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>

          <div className="flex items-center gap-3 sm:gap-6 lg:gap-8 z-10">
            {/* System Status - Desktop Only */}
            <div className="hidden xl:flex items-center gap-4 bg-white/5 pl-4 pr-5 py-2.5 rounded-2xl border border-white/5 backdrop-blur-md">
              <div className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-orange opacity-40"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-accent-orange shadow-[0_0_12px_rgba(242,123,47,0.8)]"></span>
              </div>
              <span className="text-[9px] font-black uppercase tracking-[0.3em] text-accent-orange/90 whitespace-nowrap">
                Engine Optimal
              </span>
            </div>

            <Link href="/studio">
              <motion.button
                whileHover={{ scale: 1.05, y: -1 }}
                whileTap={{ scale: 0.98 }}
                className="btn-primary h-9 sm:h-10 lg:h-12 px-4 sm:px-6 lg:px-8 text-[10px] sm:text-[11px] font-black uppercase tracking-[0.18em] sm:tracking-[0.2em] shadow-2xl shadow-accent-orange/20 cursor-pointer"
              >
                Launch Studio
              </motion.button>
            </Link>
          </div>
        </div>
      </div>
    </motion.header>
  );
}
