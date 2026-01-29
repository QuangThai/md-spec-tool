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

          <footer className="relative border-t border-white/5 bg-black/40 backdrop-blur-2xl py-20 overflow-hidden">
            {/* Ambient Footer Glow */}
            <div className="absolute inset-x-0 bottom-0 h-64 bg-accent-orange/5 blur-3xl pointer-events-none" />

            <div className="mx-auto max-w-7xl px-8 relative z-10">
              <div className="flex flex-col items-center justify-between gap-12 lg:flex-row lg:items-end">
                {/* Brand Identity */}
                <div className="flex flex-col items-center lg:items-start gap-6 group">
                  <div className="flex items-center gap-4">
                    <div className="relative">
                      <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-black/60 border border-white/10 shadow-2xl transition-all group-hover:border-accent-orange/50">
                        <svg
                          viewBox="0 0 24 24"
                          className="w-7 h-7"
                          fill="none"
                          xmlns="http://www.w3.org/2000/svg"
                        >
                          <path
                            d="M4 18V6L12 14L20 6V18"
                            stroke="#f27b2f"
                            strokeWidth="2.5"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                          <path
                            d="M12 14L16 10"
                            stroke="#f27b2f"
                            strokeWidth="2.5"
                            strokeLinecap="round"
                          />
                        </svg>
                      </div>
                      <div className="absolute inset-0 bg-accent-orange/20 blur-xl rounded-full opacity-0 group-hover:opacity-100 transition-all duration-500" />
                    </div>

                    <div className="flex flex-col">
                      <span className="text-xl font-black tracking-tighter text-white uppercase leading-none">
                        MD<span className="text-accent-orange">Flow</span>{" "}
                        Studio
                      </span>
                      <span className="text-[10px] font-bold uppercase tracking-[0.4em] text-white/40 mt-1.5 ml-0.5">
                        Specification_Automation
                      </span>
                    </div>
                  </div>
                </div>

                {/* Technical Metadata */}
                <div className="flex flex-wrap justify-center gap-8 lg:gap-16">
                  <div className="flex flex-col gap-3">
                    <span className="text-[9px] font-black uppercase tracking-[0.4em] text-white/30 text-center lg:text-left">
                      Platform_Stack
                    </span>
                    <div className="inline-flex items-center px-4 py-2 rounded-xl bg-white/5 border border-white/10 backdrop-blur-md">
                      <p className="text-[10px] font-black text-white/80 uppercase tracking-widest font-mono">
                        Next_JS_v16
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-col gap-3 text-right">
                    <span className="text-[9px] font-black uppercase tracking-[0.4em] text-white/30 text-center lg:text-right">
                      Engine_Version
                    </span>
                    <div className="inline-flex items-center px-4 py-2 rounded-xl bg-accent-orange/5 border border-accent-orange/20 shadow-lg shadow-accent-orange/5">
                      <p className="text-[10px] font-black text-accent-orange uppercase tracking-widest font-mono">
                        Alpha v1.2.0
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Bottom Copyright Strip */}
              <div className="mt-20 pt-10 border-t border-white/5 flex flex-col sm:flex-row justify-between items-center gap-8">
                <div className="flex items-center gap-8 text-[9px] font-bold uppercase tracking-[0.4em] text-white/30">
                  <p>© 2026 MD_Flow Cluster Engine</p>
                  <div className="hidden sm:block h-3 w-px bg-white/10" />
                  <p className="hidden sm:block">Legal_Standard_Compliant</p>
                </div>

                <div className="flex items-center gap-3 px-4 py-2 rounded-full border border-white/5 bg-white/2">
                  <div className="h-1.5 w-1.5 rounded-full bg-accent-orange shadow-[0_0_8px_rgba(242,123,47,0.5)] animate-pulse" />
                  <span className="text-[9px] font-black uppercase tracking-[0.3em] text-accent-orange/80">
                    Precision_Automation_System
                  </span>
                </div>
              </div>
            </div>
          </footer>
        </ErrorBoundary>
      </body>
    </html>
  );
}
