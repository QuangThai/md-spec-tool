"use client";
import { useHistoryStore } from "@/lib/mdflowStore";
import { ConversionRecord } from "@/lib/types";
import { useOnboardingStore } from "@/lib/onboardingStore";
import { SHORTCUTS, formatShortcut } from "@/lib/useKeyboardShortcuts";
import { useBodyScrollLock } from "@/lib/useBodyScrollLock";
import { AnimatePresence, LazyMotion, domAnimation, m } from "framer-motion";
import { BookOpen, Check, Clock, Command, Copy, History, Keyboard, X } from "lucide-react";
import { useCallback, useState } from "react";

export default function HistoryModal({
  history,
  onClose,
  onSelect,
}: {
  history: ConversionRecord[];
  onClose: () => void;
  onSelect: (record: ConversionRecord) => void;
}) {
  useBodyScrollLock(true);
  const { clearHistory } = useHistoryStore();
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = useCallback(
    (record: ConversionRecord, e: React.MouseEvent) => {
      e.stopPropagation();
      navigator.clipboard.writeText(record.output);
      setCopiedId(record.id);
      setTimeout(() => setCopiedId(null), 2000);
    },
    []
  );

  return (
    <LazyMotion features={domAnimation}>
      <m.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
        className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4 overscroll-contain"
      >
        <m.div
          initial={{ scale: 0.95, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.95, opacity: 0 }}
          onClick={(e) => e.stopPropagation()}
          className="bg-black/80 backdrop-blur-xl border border-white/20 rounded-2xl shadow-2xl max-w-2xl w-full max-h-[70vh] flex flex-col overflow-hidden overscroll-contain"
        >
        <div className="flex items-center justify-between gap-4 px-6 py-4 border-b border-white/10 bg-white/3 shrink-0">
          <div className="flex items-center gap-3">
            <History className="w-4 h-4 text-accent-orange" />
            <span className="text-[11px] font-black uppercase tracking-[0.2em] text-white/80">
              Conversion History
            </span>
            <span className="text-[10px] text-white/40 font-mono">
              {history.length} records
            </span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => {
                clearHistory();
              }}
              className="text-[9px] text-white/40 hover:text-accent-red/80 transition-colors cursor-pointer font-bold uppercase"
            >
              Clear all
            </button>
            <button
              onClick={onClose}
              className="p-2 rounded-md hover:bg-white/10 transition-colors cursor-pointer text-white/60 hover:text-white"
              aria-label="Close"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-2">
          {history.length === 0 ? (
            <div className="text-center py-12 text-white/40">
              <Clock className="w-8 h-8 mx-auto mb-3 opacity-50" />
              <p className="text-[11px] font-bold uppercase tracking-wider">
                No history yet
              </p>
              <p className="text-[10px] mt-1">Conversions will appear here</p>
            </div>
          ) : (
            history.map((record, idx) => (
              <m.div
                key={record.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.05 }}
                className="group relative"
              >
                <div
                  onClick={() => onSelect(record)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      onSelect(record);
                    }
                  }}
                  className="w-full text-left p-4 rounded-xl bg-linear-to-br from-white/5 to-white/2 hover:from-white/10 hover:to-white/5 border border-white/5 hover:border-white/15 transition-all duration-300 cursor-pointer relative overflow-hidden"
                >
                  {/* Subtle gradient overlay on hover */}
                  <div className="absolute inset-0 bg-linear-to-r from-accent-orange/0 via-accent-orange/0 to-accent-orange/0 group-hover:from-accent-orange/5 group-hover:via-accent-orange/0 group-hover:to-transparent transition-all duration-500 pointer-events-none" />

                  <div className="relative z-10">
                    <div className="flex items-center justify-between gap-4 mb-3">
                      <div className="flex items-center gap-2.5 flex-wrap">
                        <m.span
                          whileHover={{ scale: 1.05 }}
                          className="text-[9px] font-black uppercase tracking-[0.15em] px-2.5 py-1.5 h-6 rounded-lg bg-linear-to-r from-accent-orange/20 to-accent-orange/10 border border-accent-orange/30 text-accent-orange shadow-[0_0_8px_rgba(242,123,47,0.15)] flex items-center leading-none whitespace-nowrap"
                        >
                          {record.mode}
                        </m.span>

                        {/* Premium Template Badge */}
                        <div className="relative">
                          <div className="absolute inset-0 bg-linear-to-r from-white/10 via-white/5 to-transparent rounded-lg blur-sm opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                          <span className="relative text-[9px] font-semibold uppercase tracking-[0.2em] px-3 py-1.5 h-6 rounded-lg bg-white/5 border border-white/10 text-white/70 group-hover:text-white/90 group-hover:border-white/20 transition-all duration-300 backdrop-blur-sm flex items-center leading-none whitespace-nowrap">
                            {record.template.replace(/-/g, " ")}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <span className="text-[9px] text-white/30 font-mono">
                          {new Date(record.timestamp).toLocaleString()}
                        </span>
                        <m.button
                          whileHover={{ scale: 1.1 }}
                          whileTap={{ scale: 0.95 }}
                          onClick={(e) => handleCopy(record, e)}
                          className="p-1.5 rounded-lg bg-white/5 hover:bg-accent-orange/20 border border-white/10 hover:border-accent-orange/30 text-white/50 hover:text-accent-orange transition-all duration-200 cursor-pointer"
                          title="Copy output"
                        >
                          {copiedId === record.id ? (
                            <Check className="w-3.5 h-3.5 text-accent-orange" />
                          ) : (
                            <Copy className="w-3.5 h-3.5" />
                          )}
                        </m.button>
                      </div>
                    </div>

                    <p className="text-[10px] text-white/60 font-mono truncate mb-1.5">
                      {record.inputPreview}
                    </p>
                    <div className="flex items-center gap-2">
                      <div className="h-px flex-1 bg-linear-to-r from-white/10 to-transparent" />
                      <p className="text-[9px] text-white/30 font-medium">
                        {record.meta?.total_rows ?? 0} rows
                      </p>
                    </div>
                  </div>
                </div>
              </m.div>
            ))
          )}
        </div>
        </m.div>
      </m.div>
    </LazyMotion>
  );
}


export function KeyboardShortcutsTooltip() {
  const [show, setShow] = useState(false);
  const { resetTour, startTour } = useOnboardingStore();

  const shortcuts = [
    { keys: formatShortcut(SHORTCUTS.COMMAND_PALETTE), action: "Command Palette", icon: Command },
    { keys: formatShortcut(SHORTCUTS.CONVERT), action: "Convert" },
    { keys: formatShortcut(SHORTCUTS.COPY), action: "Copy output" },
    { keys: formatShortcut(SHORTCUTS.EXPORT), action: "Export" },
    { keys: formatShortcut(SHORTCUTS.TOGGLE_PREVIEW), action: "Toggle Preview" },
  ];

  const handleRestartTour = () => {
    setShow(false);
    resetTour();
    setTimeout(startTour, 100);
  };

  return (
    <div
      className="relative"
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
    >
      <button
        type="button"
        onClick={() => setShow(!show)}
        className="p-2.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white/40 hover:text-white/60 transition-all cursor-pointer"
        aria-label="Keyboard shortcuts"
      >
        <Keyboard className="w-4 h-4" />
      </button>

      <AnimatePresence>
        {show && (
          <LazyMotion features={domAnimation}>
            <m.div
              initial={{ opacity: 0, y: 8, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: 8, scale: 0.95 }}
              className="absolute bottom-full right-0 mb-2 w-52 rounded-xl bg-black/90 backdrop-blur-xl border border-white/20 shadow-2xl overflow-hidden"
            >
            <div className="px-3 py-2 border-b border-white/10 bg-white/5">
              <span className="text-[9px] font-black uppercase tracking-widest text-white/60">
                Shortcuts
              </span>
            </div>
            <div className="p-2 space-y-1">
              {shortcuts.map((s) => (
                <div
                  key={s.keys}
                  className="flex items-center justify-between px-2 py-1.5 rounded-lg hover:bg-white/5"
                >
                  <span className="text-[10px] text-white/70">{s.action}</span>
                  <kbd className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-white/10 text-white/60">
                    {s.keys}
                  </kbd>
                </div>
              ))}
            </div>
            <div className="p-2 border-t border-white/10">
              <button
                onClick={handleRestartTour}
                className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-accent-orange/10 hover:bg-accent-orange/20 border border-accent-orange/20 text-[10px] font-bold uppercase tracking-wider text-accent-orange hover:text-accent-orange transition-all cursor-pointer"
              >
                <BookOpen className="w-3.5 h-3.5" />
                Restart Tour
              </button>
            </div>
            </m.div>
          </LazyMotion>
        )}
      </AnimatePresence>
    </div>
  );
}
