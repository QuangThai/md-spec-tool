"use client";

import { motion } from "framer-motion";
import Link from "next/link";

const MDFlowLogo = ({ className }: { className?: string }) => (
  <svg
    viewBox="0 0 180 40"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <defs>
      <linearGradient id="premium-gold" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#F7CE68" />
        <stop offset="50%" stopColor="#FBAB7E" />
        <stop offset="100%" stopColor="#f27b2f" />
      </linearGradient>
      <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
        <feGaussianBlur stdDeviation="2" result="blur" />
        <feComposite in="SourceGraphic" in2="blur" operator="over" />
      </filter>
    </defs>

    {/* Premium Abstract Mark: Quantized Flow 'M' */}
    <path
      d="M14 26L14 14C14 14 14 10 18 10C22 10 22 14 22 14L22 26"
      stroke="url(#premium-gold)"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M22 26L22 18C22 18 22 14 26 14C30 14 30 18 30 18L30 26"
      stroke="url(#premium-gold)"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.8"
    />
    <path
      d="M30 26L30 22C30 22 30 18 34 18C38 18 38 22 38 22L38 26"
      stroke="url(#premium-gold)"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.6"
    />

    {/* Accent Dot */}
    <circle cx="42" cy="12" r="2" fill="#F7CE68" className="animate-pulse" />

    {/* Integrated Logotype - Modern Wide Sans */}
    <text
      x="54"
      y="27"
      fill="white"
      fontSize="24"
      fontWeight="800"
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
            {/* Ultra-Premium MDFlow Brand Logo (Integrated) */}
            <Link href="/" className="group cursor-pointer py-2">
              <div className="relative">
                <div className="absolute -inset-4 rounded-full bg-linear-to-r from-accent-gold/20 via-accent-orange/10 to-transparent blur-xl opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                <MDFlowLogo className="w-auto h-7 sm:h-8 lg:h-10 drop-shadow-lg group-hover:scale-105 transition-transform duration-300" />
              </div>
            </Link>

            <nav className="hidden lg:flex items-center gap-6 xl:gap-10">
              {[
                { label: "Home", href: "/" },
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
