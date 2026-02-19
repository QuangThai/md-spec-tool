import { AISuggestion, MDFlowWarning } from "@/lib/types";
import { AnimatePresence, motion } from "framer-motion";
import { Activity } from "lucide-react";
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
    quality_report?: {
      strict_mode: boolean;
      validation_passed: boolean;
      validation_reason?: string;
      header_confidence: number;
      min_header_confidence: number;
      source_rows: number;
      converted_rows: number;
      row_loss_ratio: number;
      max_row_loss_ratio: number;
      header_count: number;
      mapped_columns: number;
      mapped_ratio: number;
      core_field_coverage?: Record<string, boolean>;
    };
    ai_mode?: string;
    ai_used?: boolean;
    ai_degraded?: boolean;
    ai_model?: string;
    ai_prompt_version?: string;
    ai_estimated_input_tokens?: number;
    ai_estimated_output_tokens?: number;
    ai_estimated_cost_usd?: number;
  } | null;
  warnings: MDFlowWarning[];
  mdflowOutput: string | null;
  aiSuggestions: AISuggestion[];
  aiSuggestionsLoading: boolean;
  aiSuggestionsError: string | null;
  aiConfigured: boolean | null;
  onRetryAISuggestions?: () => void;
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
  onRetryAISuggestions,
}: TechnicalAnalysisProps) {
  const qualityReport = meta?.quality_report;
  const hasAIMetadata = Boolean(
    meta?.ai_model ||
      meta?.ai_prompt_version ||
      meta?.ai_estimated_input_tokens ||
      meta?.ai_estimated_output_tokens ||
      meta?.ai_estimated_cost_usd
  );
  const coveredCoreFields = qualityReport?.core_field_coverage
    ? Object.values(qualityReport.core_field_coverage).filter(Boolean).length
    : 0;
  const totalCoreFields = qualityReport?.core_field_coverage
    ? Object.keys(qualityReport.core_field_coverage).length
    : 0;

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

          {qualityReport && (
            <div className="rounded-lg border border-white/10 bg-white/5 p-3">
              <div className="flex items-center justify-between gap-3">
                <span className="text-[9px] font-black uppercase tracking-[0.2em] text-white/60">
                  Quality Report
                </span>
                <span
                  className={`text-[9px] font-bold uppercase tracking-wider ${qualityReport.validation_passed ? "text-green-400/90" : "text-accent-gold/90"
                    }`}
                >
                  {qualityReport.validation_passed ? "Passed" : `Needs Review (${qualityReport.validation_reason || "unknown"})`}
                </span>
              </div>

              <div className="mt-2 grid grid-cols-2 md:grid-cols-4 gap-2">
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Header Conf</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    {qualityReport.header_confidence}% / {qualityReport.min_header_confidence}%
                  </div>
                </div>
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Row Loss</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    {Math.round(qualityReport.row_loss_ratio * 100)}% / {Math.round(qualityReport.max_row_loss_ratio * 100)}%
                  </div>
                </div>
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Rows</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    {qualityReport.converted_rows}/{qualityReport.source_rows}
                  </div>
                </div>
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Mapped</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    {qualityReport.mapped_columns}/{qualityReport.header_count} ({Math.round(qualityReport.mapped_ratio * 100)}%)
                  </div>
                </div>
              </div>

              {totalCoreFields > 0 && (
                <div className="mt-2 text-[9px] text-white/55 font-mono">
                  Core coverage: {coveredCoreFields}/{totalCoreFields}
                </div>
              )}
            </div>
          )}

          {hasAIMetadata && (
            <div className="rounded-lg border border-white/10 bg-white/5 p-3">
              <div className="flex items-center justify-between gap-3">
                <span className="text-[9px] font-black uppercase tracking-[0.2em] text-white/60">
                  AI Profile
                </span>
                <span className="text-[9px] font-bold uppercase tracking-wider text-white/70">
                  {meta?.ai_model || "unknown model"}
                </span>
              </div>
              <div className="mt-2 grid grid-cols-2 md:grid-cols-4 gap-2">
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Prompt</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    {meta?.ai_prompt_version || "n/a"}
                  </div>
                </div>
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Input Tokens</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    {meta?.ai_estimated_input_tokens ?? 0}
                  </div>
                </div>
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Output Tokens</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    {meta?.ai_estimated_output_tokens ?? 0}
                  </div>
                </div>
                <div className="rounded-md border border-white/10 px-2 py-1.5 bg-black/20">
                  <div className="text-[8px] uppercase tracking-wider text-white/45">Rough Cost</div>
                  <div className="text-[11px] font-mono font-bold text-white/90">
                    ${Number(meta?.ai_estimated_cost_usd ?? 0).toFixed(5)}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* AI Suggestions Panel */}
          {(aiSuggestions.length > 0 || aiSuggestionsLoading || aiSuggestionsError) && (
            <AISuggestionsPanel
              suggestions={aiSuggestions}
              loading={aiSuggestionsLoading}
              error={aiSuggestionsError}
              configured={aiConfigured}
              onRetry={onRetryAISuggestions}
            />
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
}
