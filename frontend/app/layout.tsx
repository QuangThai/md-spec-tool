import AppHeader from "@/components/AppHeader";
import ErrorBoundary from "@/components/ErrorBoundary";
import { Command } from "lucide-react";
import { Geist_Mono as GeistMono, Inter } from "next/font/google";
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
          <main className="max-w-7xl mx-auto py-12 sm:pt-20 min-h-[80vh]">
            {children}
          </main>

          <footer className="relative mt-20 border-t border-white/5 bg-bg-base overflow-hidden">
            {/* Premium Ambient Backlight */}
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(242,123,47,0.08),transparent_50%)]" />

            <div className="mx-auto max-w-7xl px-8 py-10 relative z-10">
              <div className="flex flex-col md:flex-row items-center justify-between gap-10">
                {/* Brand Identity - Clean & Premium */}
                <div className="flex items-center gap-5 group">
                  <div className="relative">
                    <div className="absolute -inset-1 rounded-2xl bg-linear-to-br from-accent-orange/40 to-transparent blur-md opacity-40 group-hover:opacity-80 transition-opacity duration-700" />
                    <div className="relative flex h-14 w-14 items-center justify-center rounded-2xl bg-[#0A0A0A] border border-white/10 shadow-2xl ring-1 ring-white/5 group-hover:border-accent-orange/30 transition-all">
                      <svg
                        viewBox="0 0 24 24"
                        className="w-7 h-7"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          d="M4 18V6L12 14L20 6V18"
                          stroke="#F27B2F"
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          className="drop-shadow-[0_0_10px_rgba(242,123,47,0.5)]"
                        />
                        <path
                          d="M12 14L16 10"
                          stroke="#F27B2F"
                          strokeWidth="2"
                          strokeLinecap="round"
                          className="drop-shadow-[0_0_10px_rgba(242,123,47,0.5)]"
                        />
                      </svg>
                    </div>
                  </div>

                  <div className="flex flex-col">
                    <h2 className="text-2xl font-black tracking-tight text-white uppercase leading-none">
                      MDFLOW{" "}
                      <span className="text-transparent bg-clip-text bg-linear-to-r from-accent-orange via-accent-gold to-white">
                        STUDIO
                      </span>
                    </h2>
                    <div className="flex items-center gap-2 mt-2">
                      <span className="h-px w-8 bg-white/10" />
                      <span className="text-[10px] font-medium uppercase tracking-[0.3em] text-white/40">
                        Specification Engine
                      </span>
                    </div>
                  </div>
                </div>

                {/* Condensed & Eye-Catching Status Indicators */}
                <div className="flex items-center gap-4 p-1.5 rounded-full border border-white/5 bg-white/2 backdrop-blur-sm">
                  <div className="flex items-center gap-3 px-5 py-2.5 rounded-full bg-white/5 border border-white/5 group hover:border-accent-orange/30 transition-colors cursor-default">
                    <div className="h-1.5 w-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)] animate-pulse" />
                    <span className="text-[10px] font-bold text-white/70 uppercase tracking-widest group-hover:text-white transition-colors">
                      System Operational
                    </span>
                  </div>
                  <div className="hidden sm:flex items-center gap-3 px-5 py-2.5 rounded-full hover:bg-white/5 transition-colors cursor-default">
                    <span className="text-[10px] font-bold text-white/40 uppercase tracking-widest">
                      v1.2.0 Alpha
                    </span>
                  </div>
                </div>
              </div>

              {/* Minimal Copyright */}
              <div className="flex justify-center">
                <p className="text-[10px] font-medium text-white/20 uppercase tracking-[0.2em] hover:text-white/40 transition-colors cursor-default">
                  © 2026 MDFlow Inc. All Rights Reserved.
                </p>
              </div>
            </div>
          </footer>
        </ErrorBoundary>
      </body>
    </html>
  );
}
