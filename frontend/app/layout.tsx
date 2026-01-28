import React from 'react';
import '../styles/globals.css';
import ErrorBoundary from '@/components/ErrorBoundary';
import AppHeader from '@/components/AppHeader';
import { Sora, Fraunces, Space_Grotesk as SpaceGrotesk } from 'next/font/google';

const sora = Sora({
  subsets: ['latin'],
  variable: '--font-sora',
  display: 'swap',
});

const fraunces = Fraunces({
  subsets: ['latin'],
  variable: '--font-fraunces',
  display: 'swap',
});

const spaceGrotesk = SpaceGrotesk({
  subsets: ['latin'],
  variable: '--font-space-grotesk',
  display: 'swap',
});

export const metadata = {
  title: 'MD-Spec-Tool',
  description: 'Convert Excel to Markdown Specifications',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${spaceGrotesk.variable} ${sora.variable} ${fraunces.variable} `}>
      <body>
        <ErrorBoundary>
          <AppHeader />
          <main className="app-container py-10 sm:py-12 min-h-[70vh]">
            {children}
          </main>
          <footer className="border-t border-slate-200/70 bg-white/70">
            <div className="app-container flex flex-col items-center justify-between gap-2 py-6 text-center text-sm text-slate-500 sm:flex-row sm:text-left">
              <p>(c) 2026 MD Spec Tool</p>
              <p>Built with Quang Thai</p>
            </div>
          </footer>
        </ErrorBoundary>
      </body>
    </html>
  );
}
