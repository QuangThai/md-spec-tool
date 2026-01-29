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
          <main className="max-w-7xl mx-auto py-12 sm:py-20 min-h-[80vh]">
            {children}
          </main>

          <footer className="mt-32 border-t border-white/5 bg-black/20 backdrop-blur-xl py-16">
            <div className="mx-auto max-w-7xl px-8 flex flex-col items-center justify-between gap-10 sm:flex-row">
              <div className="flex flex-col items-center gap-4 sm:items-start group transition-all">
                <div className="flex items-center gap-4">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-white/5 border border-white/10 shadow-premium transition-colors">
                    <Command className="w-5 h-5 text-white" />
                  </div>
                  <div className="flex flex-col">
                    <span className="text-sm font-bold tracking-tight text-white uppercase">
                      MD<span className="text-accent-orange">Flow</span> Studio
                    </span>
                    <span className="text-[9px] font-bold uppercase tracking-[0.2em] text-muted/80">
                      Specification_Automation
                    </span>
                  </div>
                </div>
              </div>

              <div className="flex gap-12">
                <div className="flex flex-col gap-2">
                  <span className="text-[9px] font-bold uppercase tracking-[0.3em] text-muted/60">
                    Platform
                  </span>
                  <p className="text-[10px] font-bold text-white uppercase tracking-widest bg-white/5 px-3 py-1 rounded-lg border border-white/5">
                    Next.js Studio
                  </p>
                </div>
                <div className="flex flex-col gap-2 text-right">
                  <span className="text-[9px] font-bold uppercase tracking-[0.3em] text-muted/60">
                    Engine_Ver
                  </span>
                  <p className="text-[10px] font-bold text-white uppercase tracking-widest bg-white/5 px-3 py-1 rounded-lg border border-white/5">
                    Alpha v1.2.0
                  </p>
                </div>
              </div>
            </div>

            <div className="mx-auto max-w-7xl px-8 mt-16 pt-8 border-t border-white/5 flex flex-col sm:flex-row justify-between items-center gap-6 text-[9px] font-bold uppercase tracking-[0.4em] text-muted/60">
              <p>© 2026 MD_Flow Cluster Engine</p>
              <p className="text-accent-orange/80">
                Precision_Automation_System
              </p>
            </div>
          </footer>
        </ErrorBoundary>
      </body>
    </html>
  );
}
