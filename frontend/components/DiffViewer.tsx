"use client";

import { useState } from "react";
import { DiffResponse } from "@/lib/diffTypes";
import { Copy, CheckCircle } from "lucide-react";

interface DiffViewerProps {
  diff: DiffResponse;
}

export function DiffViewer({ diff }: DiffViewerProps) {
  const [viewMode, setViewMode] = useState<"inline" | "sidebyside">(
    "inline"
  );
  const [copied, setCopied] = useState(false);

  const getLineClass = (type: string) => {
    switch (type) {
      case "add":
        return "bg-emerald-500/10 text-emerald-300 border-l-emerald-500";
      case "remove":
        return "bg-rose-500/10 text-rose-300 border-l-rose-500";
      default:
        return "text-white/70 border-l-white/10";
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(diff.text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const totalLines = diff.added_lines + diff.removed_lines;
  const changePercentage =
    totalLines > 0 ? Math.round((diff.added_lines / totalLines) * 100) : 0;

  return (
    <div className="w-full h-full flex flex-col bg-linear-to-br from-white/2 to-white/1 border border-white/10  overflow-hidden">
      {/* Header */}
      <div className="border-b border-white/10 bg-white/3 backdrop-blur-md px-6 py-5 shrink-0">
        <div className="flex items-center justify-between gap-4 mb-4">
          <div className="flex-1">
            <h3 className="text-[10px] font-black uppercase tracking-[0.3em] text-white/60 mb-2">
              Change Summary
            </h3>
            <div className="flex items-center gap-6">
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-black text-emerald-400">
                  {diff.added_lines}
                </span>
                <span className="text-[10px] uppercase tracking-wider text-white/50">
                  additions
                </span>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-black text-rose-400">
                  {diff.removed_lines}
                </span>
                <span className="text-[10px] uppercase tracking-wider text-white/50">
                  deletions
                </span>
              </div>
              <div className="h-1.5 flex-1 max-w-xs bg-white/5 rounded-full overflow-hidden">
                <div
                  className="h-full bg-linear-to-r from-emerald-500 to-rose-500 transition-all duration-500"
                  style={{ width: `${changePercentage}%` }}
                />
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <button
              onClick={copyToClipboard}
              className="p-2.5 rounded-lg bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-all border border-white/10 hover:border-white/20 cursor-pointer group"
              title="Copy diff"
            >
              {copied ? (
                <CheckCircle className="w-4 h-4 text-emerald-400" />
              ) : (
                <Copy className="w-4 h-4 group-hover:scale-110 transition-transform" />
              )}
            </button>
          </div>
        </div>

        {/* View Mode Tabs */}
        <div className="flex items-center gap-2 w-fit">
          <button
            onClick={() => setViewMode("inline")}
            className={`px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-wider cursor-pointer transition-colors ${
              viewMode === "inline"
                ? "bg-accent-orange text-white"
                : "border border-white/20 text-white/50 hover:text-white/70 hover:border-white/40"
            }`}
          >
            Inline
          </button>
          <button
            onClick={() => setViewMode("sidebyside")}
            className={`px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-wider cursor-pointer transition-colors ${
              viewMode === "sidebyside"
                ? "bg-accent-orange text-white"
                : "border border-white/20 text-white/50 hover:text-white/70 hover:border-white/40"
            }`}
          >
            Side-by-Side
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 min-h-0 overflow-hidden bg-black/40 rounded-b-2xl">
        {(!diff.hunks?.length && diff.added_lines === 0 && diff.removed_lines === 0) ? (
          <div className="h-full flex flex-col items-center justify-center text-center px-6 py-12">
            <p className="text-[12px] font-bold text-white/60 uppercase tracking-wider">
              No changes detected
            </p>
            <p className="text-[11px] text-white/40 mt-1.5">
              Versions may be identical
            </p>
          </div>
        ) : (
          <>
            {viewMode === "inline" && (
              <InlineDiffView diff={diff} getLineClass={getLineClass} />
            )}

            {viewMode === "sidebyside" && (
              <SideBySideDiffView diff={diff} getLineClass={getLineClass} />
            )}
          </>
        )}
      </div>
    </div>
  );
}

function InlineDiffView({ diff, getLineClass }: any) {
  return (
    <div className="h-full overflow-y-auto overflow-x-auto custom-scrollbar rounded-b-2xl">
      <div className="font-mono text-xs leading-relaxed">
        <div className="sticky top-0 z-10 bg-white/3 backdrop-blur-md border-b border-white/10 px-4 py-3 flex gap-4 text-white/50 rounded-t-2xl shrink-0">
          <span className="flex items-center gap-1">
            <span className="w-4 h-4 rounded bg-rose-500/20 border border-rose-500/50" />
            before
          </span>
          <span className="flex items-center gap-1">
            <span className="w-4 h-4 rounded bg-emerald-500/20 border border-emerald-500/50" />
            after
          </span>
        </div>

        {diff.hunks.map((hunk: any, hunkIdx: number) => (
          <div key={hunkIdx}>
            <div className="sticky top-12 z-10 bg-black/60 backdrop-blur-md px-4 py-2.5 border-b border-white/5 text-cyan-300/80 text-[11px] font-bold uppercase tracking-wider">
              @@ -{hunk.old_start},{hunk.old_count} +{hunk.new_start},
              {hunk.new_count} @@
            </div>

            {hunk.lines.map((line: any, lineIdx: number) => (
              <div
                key={lineIdx}
                className={`${getLineClass(
                  line.type
                )} border-l-2 px-4 py-1.5 hover:bg-white/5 transition-colors`}
              >
                <span className="inline-block w-6 text-right text-white/40 mr-3 select-none text-[10px]">
                  {line.type === "add"
                    ? "+"
                    : line.type === "remove"
                    ? "−"
                    : " "}
                </span>
                <code className="break-all">{line.content}</code>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

function SideBySideDiffView({ diff }: any) {
  // Group lines: remove+add pairs on same row, unpaired on empty columns
  const groupLines = (hunk: any) => {
    const pairs: Array<[any, any]> = [];
    let i = 0;

    while (i < hunk.lines.length) {
      const line = hunk.lines[i];
      const isRemove = line.type === "remove";
      const isAdd = line.type === "add";

      if (isRemove) {
        // Check if next line is "add" - if so, pair them
        if (i + 1 < hunk.lines.length && hunk.lines[i + 1].type === "add") {
          pairs.push([line, hunk.lines[i + 1]]);
          i += 2;
        } else {
          // Unpaired removal
          pairs.push([line, null]);
          i++;
        }
      } else if (isAdd) {
        // Unpaired addition (should not happen if remove handled above)
        pairs.push([null, line]);
        i++;
      } else {
        // Context line or anything else - show on both sides
        pairs.push([line, line]);
        i++;
      }
    }

    return pairs;
  };

  return (
    <div className="h-full flex flex-col overflow-hidden rounded-b-2xl">
      {/* Header */}
      <div className="flex border-b border-white/10 bg-white/2 shrink-0 rounded-t-2xl">
        <div className="flex-1 px-4 py-3 border-r border-white/5">
          <span className="text-[10px] font-black uppercase tracking-wider text-rose-400/80">
            Removed
          </span>
        </div>
        <div className="flex-1 px-4 py-3">
          <span className="text-[10px] font-black uppercase tracking-wider text-emerald-400/80">
            Added
          </span>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 min-h-0 overflow-y-auto custom-scrollbar">
        <div className="font-mono text-[11px] leading-relaxed">
          {diff.hunks.map((hunk: any, hunkIdx: number) => {
            const pairs = groupLines(hunk);

            return (
              <div key={`hunk-${hunkIdx}`}>
                {/* Hunk header */}
                <div className="sticky top-0 z-10 flex bg-black/60 backdrop-blur-md border-b border-white/5 shrink-0">
                  <div className="flex-1 px-4 py-2 border-r border-white/5">
                    <span className="text-cyan-300/70 text-[10px] font-bold">
                      @@ -{hunk.old_start},{hunk.old_count}
                    </span>
                  </div>
                  <div className="flex-1 px-4 py-2">
                    <span className="text-cyan-300/70 text-[10px] font-bold">
                      +{hunk.new_start},{hunk.new_count} @@
                    </span>
                  </div>
                </div>

                {/* Lines */}
                {pairs.map((pair: any, pairIdx: number) => {
                  const [removeLine, addLine] = pair;
                  const isContext = removeLine?.type === "context";

                  return (
                    <div
                      key={`pair-${hunkIdx}-${pairIdx}`}
                      className={`flex border-b border-white/5 hover:bg-white/3 transition-colors ${
                        isContext ? "bg-white/1" : ""
                      }`}
                    >
                      {/* Left column - Removed */}
                      <div className="flex-1 border-r border-white/5 px-4 py-1.5 min-w-0">
                        {removeLine ? (
                          <div
                            className={
                              isContext
                                ? "text-white/60"
                                : "bg-rose-500/15 border border-rose-500/30 rounded px-2 py-0.5 text-rose-300"
                            }
                          >
                            <code className="break-all">{removeLine.content}</code>
                          </div>
                        ) : (
                          <div className="h-6" />
                        )}
                      </div>

                      {/* Right column - Added */}
                      <div className="flex-1 px-4 py-1.5 min-w-0">
                        {addLine ? (
                          <div
                            className={
                              isContext
                                ? "text-white/60"
                                : "bg-emerald-500/15 border border-emerald-500/30 rounded px-2 py-0.5 text-emerald-300"
                            }
                          >
                            <code className="break-all">{addLine.content}</code>
                          </div>
                        ) : (
                          <div className="h-6" />
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
