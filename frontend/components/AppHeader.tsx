"use client";

import { motion } from "framer-motion";
import Image from "next/image";
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
      {/* Prism Facet 1: Light Gold */}
      <linearGradient id="prism-light" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#FFF7ED" />
        <stop offset="100%" stopColor="#FCD34D" />
      </linearGradient>

      {/* Prism Facet 2: Mid Gold */}
      <linearGradient id="prism-mid" x1="0%" y1="0%" x2="100%" y2="0%">
        <stop offset="0%" stopColor="#F59E0B" />
        <stop offset="100%" stopColor="#D97706" />
      </linearGradient>

      {/* Prism Facet 3: Dark Gold (Shadow) */}
      <linearGradient id="prism-dark" x1="100%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#B45309" />
        <stop offset="100%" stopColor="#78350F" />
      </linearGradient>

      <filter id="crisp-shadow" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow
          dx="0"
          dy="1"
          stdDeviation="1"
          floodColor="#000"
          floodOpacity="0.4"
        />
      </filter>
    </defs>

    <g filter="url(#crisp-shadow)">
      {/* 3D Isometric 'M' Symbol */}
      <g transform="translate(4, 4)">
        {/* Left Leg: Dark Side */}
        <path d="M6 32V12L16 6L16 26L6 32Z" fill="url(#prism-dark)" />
        {/* Left Leg: Light Top */}
        <path d="M6 12L16 6L22 10L12 16L6 12Z" fill="url(#prism-light)" />
        {/* Left Leg: Face */}
        <path d="M16 26L22 22V10L16 6V26Z" fill="url(#prism-mid)" />

        {/* V-Connector: Deep Shadow */}
        <path d="M22 10L28 14L22 22Z" fill="#78350F" />

        {/* Right Leg: Dark Side */}
        <path d="M38 32V12L28 6L28 26L38 32Z" fill="url(#prism-dark)" />
        {/* Right Leg: Light Top */}
        <path d="M38 12L28 6L22 10L32 16L38 12Z" fill="url(#prism-light)" />
        {/* Right Leg: Face */}
        <path d="M28 26L22 22V10L28 6V26Z" fill="url(#prism-mid)" />
      </g>

      {/* Text: Technical & Precise */}
      <text
        x="52"
        y="29"
        fill="white"
        fontSize="22"
        fontWeight="800"
        letterSpacing="-0.01em"
        style={{ fontFamily: "'Inter', sans-serif" }}
      >
        MDFlow
      </text>
    </g>
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
