"use client";

import { motion } from "framer-motion";
import Link from "next/link";

const MDFlowLogo = ({ className }: { className?: string }) => (
  <svg
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    {/* Abstract 'M' and 'F' Flow Mark */}
    <path
      d="M4 18V6L12 14L20 6V18"
      stroke="url(#logo-gradient)"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M12 14L16 10"
      stroke="url(#logo-gradient)"
      strokeWidth="2.5"
      strokeLinecap="round"
    />
    <defs>
      <linearGradient
        id="logo-gradient"
        x1="4"
        y1="6"
        x2="20"
        y2="18"
        gradientUnits="userSpaceOnUse"
      >
        <stop stopColor="#f27b2f" />
        <stop offset="1" stopColor="#c37d0d" />
      </linearGradient>
    </defs>
  </svg>
);

export default function AppHeader() {
  return (
    <motion.header
      initial={{ y: -10, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="sticky top-6 z-50 px-4 sm:px-6"
    >
      <div className="mx-auto max-w-7xl">
        <div className="glass relative flex h-16 items-center justify-between px-8 rounded-2xl border-white/10 shadow-premium overflow-hidden">
          {/* Subtle brand flow effect */}
          <div className="absolute left-0 top-0 h-full w-32 bg-accent-orange/5 blur-3xl pointer-events-none" />

          <div className="flex items-center gap-10 z-10">
            {/* Ultra-Premium MDFlow Brand Logo */}
            <Link
              href="/"
              className="flex items-center gap-4 group cursor-pointer"
            >
              <div className="relative">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-black/40 border border-white/10 shadow-lg transition-all group-hover:border-accent-orange/50 group-hover:scale-105">
                  <MDFlowLogo className="w-6 h-6" />
                </div>
                {/* Visual Glow */}
                <div className="absolute inset-0 bg-accent-orange/20 blur-xl rounded-full opacity-0 group-hover:opacity-100 transition-all duration-500" />
              </div>

              <div className="flex items-center gap-4">
                <div className="flex flex-col">
                  <h1 className="text-lg font-black tracking-tighter text-white uppercase leading-none">
                    MD<span className="text-accent-orange">FLOW</span>
                  </h1>
                  <span className="text-[7px] font-bold text-white/50 uppercase tracking-[0.5em] mt-1 ml-0.5">
                    Engineering_Studio
                  </span>
                </div>
              </div>
            </Link>

            <nav className="hidden md:flex items-center gap-8">
              <Link
                href="/"
                className="text-[10px] font-bold uppercase tracking-[0.2em] text-muted hover:text-accent-orange transition-colors"
              >
                Home
              </Link>
              <Link
                href="/app"
                className="text-[10px] font-bold uppercase tracking-[0.2em] text-muted hover:text-accent-orange transition-colors"
              >
                Workbench
              </Link>
            </nav>
          </div>

          <div className="flex items-center gap-6 z-10">
            {/* System Status - Desktop Only */}
            <div className="hidden sm:flex items-center gap-3 bg-white/5 px-4 py-2 rounded-lg border border-white/5">
              <div className="relative flex h-1.5 w-1.5">
                <span className="animate-ping original-ping absolute inline-flex h-full w-full rounded-full bg-accent-orange opacity-40"></span>
                <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-accent-orange shadow-[0_0_10px_rgba(242,123,47,0.5)]"></span>
              </div>
              <span className="text-[8px] font-black uppercase tracking-[0.2em] text-accent-orange">
                Engine_Optimal
              </span>
            </div>

            <Link href="/app">
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.98 }}
                className="btn-primary h-10 px-6 text-[10px] uppercase tracking-widest shadow-xl shadow-accent-orange/15"
              >
                Launch App
              </motion.button>
            </Link>
          </div>
        </div>
      </div>
    </motion.header>
  );
}
