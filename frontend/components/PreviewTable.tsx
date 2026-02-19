import { memo, useState } from "react";
import { Table, ExternalLink } from "lucide-react";
import { PreviewResponse } from "@/lib/types";
import { CANONICAL_FIELDS } from "@/constants/mdflow";
import { Select } from "./ui/Select";

interface PreviewTableProps {
  preview: PreviewResponse;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;
  needsReview?: boolean;
  reviewApproved?: boolean;
  sourceUrl?: string;
  onSelectBlockRange?: (range: string) => void;
}

/**
 * PreviewTable - Displays preview data with column mapping selector
 * Shows first 4 rows by default, expandable to show all
 * ✅ Memoized to prevent re-renders on parent updates
 */
export const PreviewTable = memo(function PreviewTable({
  preview,
  columnOverrides,
  onColumnOverride,
  needsReview = false,
  reviewApproved = false,
  sourceUrl,
  onSelectBlockRange,
}: PreviewTableProps) {
  const [expanded, setExpanded] = useState(false);
  const maxCollapsedRows = 4;
  const hasMoreRows = preview.rows.length > maxCollapsedRows;
  const displayRows = expanded
    ? preview.rows
    : preview.rows.slice(0, maxCollapsedRows);
  const lowConfidenceColumns = new Set(
    preview.mapping_quality?.low_confidence_columns ?? []
  );
  const blockOptions = (preview.blocks ?? []).map((block) => ({
    label: `${block.language_hint} • ${block.range} • ${Math.round(block.english_score * 100)}% EN`,
    value: block.range,
  }));

  return (
    <div className="rounded-xl border border-white/10 bg-black/30 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2.5 bg-white/5 border-b border-white/10">
        <div className="flex items-center gap-2">
          <Table className="w-3.5 h-3.5 text-accent-orange/80" />
          <span className="text-[10px] font-black uppercase tracking-widest text-white/70">
            Preview
          </span>
          <span className="text-[9px] text-white/40 font-mono">
            {preview.total_rows} rows • {preview.headers.length} cols
          </span>
          {needsReview && !reviewApproved && (
            <span className="text-[9px] uppercase tracking-wider font-bold px-1.5 py-0.5 rounded bg-accent-gold/20 border border-accent-gold/30 text-accent-gold/80">
              Needs Review
            </span>
          )}
          {needsReview && reviewApproved && (
            <span className="text-[9px] uppercase tracking-wider font-bold px-1.5 py-0.5 rounded bg-green-500/20 border border-green-500/30 text-green-300">
              Reviewed
            </span>
          )}
          {preview.confidence < 70 && (
            <span className="text-[9px] text-accent-gold/80 font-medium">
              (low confidence: {preview.confidence}%)
            </span>
          )}
          {preview.mapping_quality?.recommended_format === "table" && (
            <span className="text-[9px] text-accent-gold/80 font-medium">
              (recommended output: table)
            </span>
          )}
          {sourceUrl && (
            <a
              href={sourceUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-[9px] text-green-400/80 hover:text-green-400 transition-colors font-medium"
              title={sourceUrl}
            >
              <ExternalLink className="w-3 h-3" />
              <span className="truncate max-w-[150px]">Google Sheet</span>
            </a>
          )}
          {blockOptions.length > 1 && onSelectBlockRange && (
            <Select
              value={preview.selected_block_range || blockOptions[0].value}
              onValueChange={onSelectBlockRange}
              options={blockOptions}
              size="compact"
              className="h-6 text-[9px] min-w-[220px]"
            />
          )}
        </div>
        {hasMoreRows && (
          <button
            type="button"
            onClick={() => setExpanded(!expanded)}
            className="text-[9px] text-accent-orange/70 hover:text-accent-orange cursor-pointer font-bold uppercase"
          >
            {expanded ? "Show less" : `Show all ${preview.rows.length}`}
          </button>
        )}
      </div>

      <div className="overflow-x-auto custom-scrollbar">
        <table className="w-full text-[11px]">
          <thead>
            <tr className="border-b border-white/10 bg-white/3">
              {preview.headers.map((header, i) => {
                const mappedField =
                  columnOverrides[header] ||
                  preview.column_mapping[header] ||
                  "";
                const isUnmapped =
                  !mappedField && preview.unmapped_columns.includes(header);
                const isLowConfidence = lowConfidenceColumns.has(header);
                const confidence =
                  preview.mapping_quality?.column_confidence?.[header];
                const reasons =
                  preview.mapping_quality?.column_reasons?.[header] ?? [];
                const displayHeader = header?.trim() || "Unmapped";

                const selectOptions = [
                  { label: "— unmapped —", value: "__unmapped__" },
                  ...CANONICAL_FIELDS.map((field) => ({
                    label: field.replace(/_/g, " "),
                    value: field,
                  })),
                ];

                return (
                  <th key={i} className="px-3 py-2 text-left">
                    <div className="space-y-1">
                      <span
                        className="font-bold text-white/90 block truncate max-w-[150px]"
                        title={displayHeader}
                      >
                        {displayHeader}
                      </span>
                      {isLowConfidence && reasons.length > 0 && (
                        <span
                          className="text-[9px] text-accent-gold/85 block"
                          title={reasons.join("; ")}
                        >
                          ! needs review
                        </span>
                      )}
                      {typeof confidence === "number" && mappedField && (
                        <span className="text-[9px] text-white/45 font-mono block">
                          {Math.round(confidence * 100)}%
                        </span>
                      )}
                      <Select
                        value={mappedField || "__unmapped__"}
                        onValueChange={(value) =>
                          onColumnOverride(header, value === "__unmapped__" ? "" : value)
                        }
                        options={selectOptions}
                        size="compact"
                        className={`
                          text-[9px] h-6 min-w-[100px] whitespace-nowrap
                          ${
                            isUnmapped || isLowConfidence
                              ? "border-accent-gold/40 text-accent-gold/80"
                              : "border-white/10 text-accent-orange/80"
                          }
                        `}
                        aria-label={`Map column ${displayHeader}`}
                      />
                    </div>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {displayRows.map((row, rowIdx) => (
              <tr
                key={rowIdx}
                className="border-b border-white/5 hover:bg-white/3"
              >
                {row.map((cell, cellIdx) => (
                  <td
                    key={cellIdx}
                    className="px-3 py-2 text-white/70 font-mono"
                  >
                    <span className="block truncate max-w-[200px]" title={cell}>
                      {cell || <span className="text-white/30">—</span>}
                    </span>
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {preview.total_rows > preview.preview_rows && (
        <div className="px-4 py-2 text-[9px] text-white/40 bg-white/3 border-t border-white/5">
          Showing {displayRows.length} of {preview.total_rows} rows
        </div>
      )}
    </div>
  );
});
