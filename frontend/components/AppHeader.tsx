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
      initial={{ y: -20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="sticky top-6 z-50 px-4 sm:px-6"
    >
      <div className="mx-auto max-w-7xl">
        <div className="glass relative flex h-20 items-center justify-between px-10 rounded-3xl border-white/10 shadow-3xl overflow-hidden ring-1 ring-white/10">
          {/* Subtle brand flow effect */}
          <div className="absolute left-0 top-0 h-full w-48 bg-accent-orange/10 blur-[80px] pointer-events-none" />

          <div className="flex items-center gap-12 z-10">
            {/* Ultra-Premium MDFlow Brand Logo */}
            <Link
              href="/"
              className="flex items-center gap-5 group cursor-pointer"
            >
              <div className="relative">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-black/60 border border-white/10 shadow-2xl transition-all group-hover:border-accent-orange/50 group-hover:scale-105 group-hover:bg-black/80">
                  <MDFlowLogo className="w-7 h-7" />
                </div>
                {/* Visual Glow */}
                <div className="absolute inset-0 bg-accent-orange/30 blur-2xl rounded-full opacity-0 group-hover:opacity-100 transition-all duration-700" />
              </div>

              <div className="flex items-center gap-4">
                <div className="flex flex-col">
                  <h1 className="text-xl font-black tracking-tighter text-white uppercase leading-none">
                    MD<span className="text-accent-orange">FLOW</span>
                  </h1>
                  <span className="text-[8px] font-black text-white/40 uppercase tracking-[0.5em] mt-1.5 ml-0.5">
                    Engineering_Studio
                  </span>
                </div>
              </div>
            </Link>

            <nav className="hidden lg:flex items-center gap-10">
              {[
                { label: "Home", href: "/" },
                { label: "Workbench", href: "/studio" },
                { label: "Docs", href: "/docs" },
              ].map((item) => (
                <Link
                  key={item.label}
                  href={item.href}
                  className="text-[11px] font-black uppercase tracking-[0.3em] text-white/40 hover:text-accent-orange transition-all hover:tracking-[0.4em]"
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>

          <div className="flex items-center gap-8 z-10">
            {/* System Status - Desktop Only */}
            <div className="hidden xl:flex items-center gap-4 bg-white/5 pl-4 pr-5 py-2.5 rounded-2xl border border-white/5 backdrop-blur-md">
              <div className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-orange opacity-40"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-accent-orange shadow-[0_0_12px_rgba(242,123,47,0.8)]"></span>
              </div>
              <span className="text-[9px] font-black uppercase tracking-[0.3em] text-accent-orange/90 whitespace-nowrap">
                Engine_Optimal
              </span>
            </div>

            <Link href="/studio">
              <motion.button
                whileHover={{ scale: 1.05, y: -1 }}
                whileTap={{ scale: 0.98 }}
                className="btn-primary h-12 px-8 text-[11px] font-black uppercase tracking-[0.2em] shadow-2xl shadow-accent-orange/20 cursor-pointer"
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
