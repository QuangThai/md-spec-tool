"use client";

import { AnimatePresence, motion } from "framer-motion";
import Link from "next/link";
import { useState } from "react";
import { Menu, X } from "lucide-react";

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

const navItems = [
  { label: "Studio", href: "/studio" },
  { label: "Batch", href: "/batch" },
  { label: "Docs", href: "/docs" },
];

export default function AppHeader() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <motion.header
      initial={{ y: -16, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.25, ease: [0.25, 0.46, 0.45, 0.94] }}
      className="sticky top-2 sm:top-3 lg:top-5 z-50 px-2 sm:px-4 lg:px-6"
    >
      <div className="mx-auto max-w-7xl">
        <div className="glass relative z-60 flex h-14 sm:h-12 lg:h-14 items-center justify-between px-3 sm:px-5 lg:px-8 rounded-xl sm:rounded-2xl border border-white/8 shadow-[0_4px_24px_-4px_rgba(0,0,0,0.25)] overflow-hidden">
          {/* Subtle brand glow */}
          <div className="absolute left-0 top-0 h-full w-32 bg-accent-orange/5 blur-[60px] pointer-events-none" />

          {/* Left: Logo (+ desktop nav from lg) */}
          <div className="flex items-center gap-6 lg:gap-10 z-10 min-w-0">
            <Link href="/" className="group cursor-pointer py-1 shrink-0" onClick={() => setMenuOpen(false)}>
              <div className="relative">
                <div className="absolute -inset-3 rounded-full bg-accent-orange/5 blur-lg opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                <MDFlowLogo className="h-6 sm:h-7 lg:h-8 w-auto opacity-95 group-hover:opacity-100 group-hover:scale-[1.02] transition-all duration-200" />
              </div>
            </Link>

            {/* Desktop nav */}
            <nav className="hidden lg:flex items-center gap-5 xl:gap-8">
              {navItems.map((item) => (
                <Link
                  key={item.label}
                  href={item.href}
                  className="text-[11px] font-semibold uppercase tracking-[0.15em] text-white/35 hover:text-accent-orange/90 transition-colors duration-200"
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>

          {/* Right: Menu button (mobile) / Status + Launch (desktop) */}
          <div className="flex items-center justify-end gap-2 sm:gap-4 lg:gap-6 z-10">
            {/* Mobile menu button - right side */}
            <motion.button
              type="button"
              onClick={() => setMenuOpen((o) => !o)}
              className="lg:hidden p-2 rounded-lg text-white/50 hover:text-white/90 hover:bg-white/5 transition-colors duration-200"
              whileTap={{ scale: 0.94 }}
              aria-label={menuOpen ? "Close menu" : "Open menu"}
            >
              <AnimatePresence mode="wait">
                {menuOpen ? (
                  <motion.span key="close" initial={{ rotate: -90, opacity: 0 }} animate={{ rotate: 0, opacity: 1 }} exit={{ rotate: 90, opacity: 0 }} transition={{ duration: 0.12 }}>
                    <X className="w-[18px] h-[18px]" />
                  </motion.span>
                ) : (
                  <motion.span key="menu" initial={{ rotate: 90, opacity: 0 }} animate={{ rotate: 0, opacity: 1 }} exit={{ rotate: -90, opacity: 0 }} transition={{ duration: 0.12 }}>
                    <Menu className="w-[18px] h-[18px]" />
                  </motion.span>
                )}
              </AnimatePresence>
            </motion.button>
            {/* System Status - Desktop Only */}
            <div className="hidden xl:flex items-center gap-2.5 bg-white/4 pl-3 pr-3.5 py-1.5 rounded-xl border border-white/6">
              <div className="relative flex h-1.5 w-1.5">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-orange/50 opacity-30" />
                <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-accent-orange/90" />
              </div>
              <span className="text-[9px] font-medium uppercase tracking-[0.2em] text-white/50 whitespace-nowrap">
                Optimal
              </span>
            </div>

            {/* Launch Studio */}
            <Link href="/studio" className="hidden sm:inline-block">
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="h-8 sm:h-9 lg:h-9 px-4 sm:px-5 lg:px-5 rounded-lg text-[10px] sm:text-[11px] font-semibold uppercase tracking-[0.12em] bg-accent-orange/90 hover:bg-accent-orange text-white shadow-[0_2px_12px_-2px_rgba(242,123,47,0.35)] cursor-pointer transition-colors duration-200"
              >
                Launch Studio
              </motion.button>
            </Link>
          </div>
        </div>

        {/* Mobile menu: compact, refined */}
        <AnimatePresence>
          {menuOpen && (
            <>
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.18 }}
                className="lg:hidden fixed inset-0 top-14 sm:top-12 z-40 bg-black/60 backdrop-blur-sm"
                onClick={() => setMenuOpen(false)}
              />
              <motion.div
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ type: "spring", damping: 30, stiffness: 350 }}
                className="lg:hidden absolute left-3 right-3 top-full mt-1.5 z-50 overflow-hidden rounded-xl border border-white/8 bg-surface/98 shadow-[0_16px_40px_-12px_rgba(0,0,0,0.5)]"
              >
                <div className="h-px bg-linear-to-r from-transparent via-accent-orange/25 to-transparent" />
                <nav className="py-2 px-0.5">
                  {navItems.map((item, i) => (
                    <motion.div
                      key={item.label}
                      initial={{ opacity: 0, x: -6 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0 }}
                      transition={{ delay: 0.02 * i, duration: 0.15 }}
                    >
                      <Link
                        href={item.href}
                        onClick={() => setMenuOpen(false)}
                        className="group flex items-center gap-2.5 px-3.5 py-2.5 rounded-lg text-[11px] font-medium uppercase tracking-[0.14em] text-white/60 hover:text-white/95 hover:bg-white/4 active:bg-white/6 transition-colors duration-150"
                      >
                        <span className="w-1 h-1 rounded-full bg-accent-orange/50 opacity-0 group-hover:opacity-100 transition-opacity shrink-0" aria-hidden />
                        {item.label}
                      </Link>
                    </motion.div>
                  ))}
                  <div className="my-2 mx-3 h-px bg-white/6" />
                  <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 0.08 }}
                    className="px-2.5 pb-2.5"
                  >
                    <Link href="/studio" onClick={() => setMenuOpen(false)} className="block">
                      <motion.button
                        whileTap={{ scale: 0.98 }}
                        className="w-full py-2.5 rounded-lg text-[11px] font-semibold uppercase tracking-[0.12em] text-white bg-accent-orange/90 hover:bg-accent-orange shadow-[0_2px_12px_-2px_rgba(242,123,47,0.4)] transition-colors duration-200"
                      >
                        Launch Studio
                      </motion.button>
                    </Link>
                  </motion.div>
                </nav>
              </motion.div>
            </>
          )}
        </AnimatePresence>
      </div>
    </motion.header>
  );
}
