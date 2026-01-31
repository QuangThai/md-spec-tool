import { AnimatePresence, motion } from "framer-motion";
import { Activity } from "lucide-react";
import { AISuggestion, MDFlowWarning } from "@/lib/mdflowApi";
import { AISuggestionsPanel } from "./AISuggestionsPanel";
import { WarningPanel } from "./WarningPanel";

interface TechnicalAnalysisProps {
  meta: {
    nodeCount?: number;
    total_rows?: number;
    headerCount?: number;
    header_row?: number;
    source_type?: string;
    parser?: string;
    source_url?: string;
  } | null;
  warnings: MDFlowWarning[];
  mdflowOutput: string | null;
  aiSuggestions: AISuggestion[];
  aiSuggestionsLoading: boolean;
  aiSuggestionsError: string | null;
  aiConfigured: boolean | null;
}

/**
 * TechnicalAnalysis - Displays stats, warnings, and AI suggestions
 * Shows idle state when no output, active state with analysis when available
 */
export function TechnicalAnalysis({
  meta,
  warnings,
  mdflowOutput,
  aiSuggestions,
  aiSuggestionsLoading,
  aiSuggestionsError,
  aiConfigured,
}: TechnicalAnalysisProps) {
  return (
    <AnimatePresence mode="wait">
      {!mdflowOutput ? (
        <motion.div
          key="idle"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="flex items-center justify-between gap-4"
        >
          <div className="flex items-center gap-3">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full rounded-full bg-accent-orange/40 animate-ping opacity-75" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-white/30" />
            </span>
            <span className="text-[10px] font-black uppercase tracking-[0.25em] text-white/40">
              Standby — run engine to see stats
            </span>
          </div>
          <div className="flex gap-1.5">
            <span className="w-10 h-1 rounded-full bg-white/10" />
            <span className="w-6 h-1 rounded-full bg-white/10" />
          </div>
        </motion.div>
      ) : (
        <motion.div
          key="active"
          initial={{ opacity: 0, y: 4 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.3 }}
          className="space-y-4"
        >
          {/* Stats row */}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Activity className="w-3.5 h-3.5 text-accent-orange/80" />
              <span className="text-[9px] font-black uppercase tracking-[0.25em] text-white/50">
                Stats
              </span>
            </div>
            <div className="flex gap-3">
              <div className="px-3 py-2 rounded-lg bg-white/5 border border-white/10 flex items-center gap-2">
                <span className="text-[8px] text-white/40 font-bold uppercase tracking-wider">
                  Rows
                </span>
                <span className="text-sm font-black font-mono text-white leading-none">
                  {meta?.nodeCount ?? meta?.total_rows ?? 0}
                </span>
              </div>
              <div className="px-3 py-2 rounded-lg bg-white/5 border border-white/10 flex items-center gap-2">
                <span className="text-[8px] text-white/40 font-bold uppercase tracking-wider">
                  Header
                </span>
                <span className="text-sm font-black font-mono text-white leading-none">
                  {meta?.headerCount ?? meta?.header_row ?? 0}
                </span>
              </div>
            </div>
          </div>

          {/* Enhanced Warning Panel */}
          {warnings && warnings.length > 0 && (
            <WarningPanel warnings={warnings} />
          )}

          {/* AI Suggestions Panel */}
          {(aiSuggestions.length > 0 || aiSuggestionsLoading || aiSuggestionsError) && (
            <AISuggestionsPanel
              suggestions={aiSuggestions}
              loading={aiSuggestionsLoading}
              error={aiSuggestionsError}
              configured={aiConfigured}
            />
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
}
