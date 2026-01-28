import Link from 'next/link';

export default function AppHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-slate-200/70 bg-gradient-to-b from-white/95 to-white/70 backdrop-blur">
      <div className="app-container flex flex-wrap items-center justify-between gap-4 py-4">
        <Link href="/" className="flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-slate-900 text-sm font-semibold text-white shadow-[0_10px_24px_rgba(15,23,42,0.2)]">
            MD
          </span>
          <div>
            <div className="flex items-center gap-2 text-lg font-semibold tracking-tight text-slate-900">
              MDFlow Studio
            </div>
            <div className="text-xs text-slate-500">Spreadsheet to MDFlow</div>
          </div>
        </Link>
        <div className="flex items-center gap-2">
          <a href="#converter" className="btn-ghost hidden sm:inline-flex">
            Converter
          </a>
          <a href="#output" className="btn-secondary hidden sm:inline-flex">
            Output
          </a>
          <a href="#converter" className="btn-primary">
            Start
          </a>
        </div>
      </div>
      <div className="h-px w-full bg-gradient-to-r from-transparent via-slate-200/70 to-transparent"></div>
    </header>
  );
}
