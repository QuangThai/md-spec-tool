import { useState } from "react";
import { Table, ExternalLink } from "lucide-react";
import { PreviewResponse } from "@/lib/types";
import { CANONICAL_FIELDS } from "@/constants/mdflow";
import { Select } from "./ui/Select";

interface PreviewTableProps {
  preview: PreviewResponse;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;
  sourceUrl?: string;
}

/**
 * PreviewTable - Displays preview data with column mapping selector
 * Shows first 4 rows by default, expandable to show all
 */
export function PreviewTable({
  preview,
  columnOverrides,
  onColumnOverride,
  sourceUrl,
}: PreviewTableProps) {
  const [expanded, setExpanded] = useState(false);
  const maxCollapsedRows = 4;
  const hasMoreRows = preview.rows.length > maxCollapsedRows;
  const displayRows = expanded
    ? preview.rows
    : preview.rows.slice(0, maxCollapsedRows);

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
          {preview.confidence < 70 && (
            <span className="text-[9px] text-accent-gold/80 font-medium">
              (low confidence: {preview.confidence}%)
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
                            isUnmapped
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
}
