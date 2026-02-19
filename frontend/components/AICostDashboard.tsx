"use client";

import { useMemo } from "react";
import { DollarSign, Cpu, Coins } from "lucide-react";

interface AICostData {
  total_cost_usd: number;
  avg_cost_per_convert: number;
  total_input_tokens: number;
  total_output_tokens: number;
  total_ai_requests: number;
  cost_by_model: Array<{
    model: string;
    cost_usd: number;
    requests: number;
  }>;
}

interface AICostDashboardProps {
  data: AICostData | null | undefined;
}

function formatCost(value: number): string {
  return `$${value.toFixed(4)}`;
}

function widthFromRatio(value: number, max: number): string {
  if (max <= 0) {
    return "0%";
  }
  return `${Math.max(4, Math.min(100, (value / max) * 100))}%`;
}

export function AICostDashboard({ data }: AICostDashboardProps) {
  const maxModelCost = useMemo(() => {
    if (!data?.cost_by_model || data.cost_by_model.length === 0) {
      return 0;
    }
    return data.cost_by_model.reduce((max, row) => (row.cost_usd > max ? row.cost_usd : max), 0);
  }, [data]);

  const totalTokens = useMemo(() => {
    if (!data) {
      return 0;
    }
    return data.total_input_tokens + data.total_output_tokens;
  }, [data]);

  if (!data) {
    return null;
  }

  return (
    <div className="rounded-2xl border border-white/10 bg-white/4 p-5">
      <div className="mb-4 flex items-center justify-between gap-2 text-white">
        <div className="flex items-center gap-2">
          <DollarSign size={16} className="text-accent-orange" />
          <h2 className="text-sm font-bold uppercase tracking-wider">AI Cost & Token Usage</h2>
        </div>
        <span className="text-xs text-white/50">Estimated spend</span>
      </div>

      {data.total_ai_requests === 0 ? (
        <p className="text-sm text-white/60">No AI usage in this window.</p>
      ) : (
        <>
          <div className="mb-4 grid gap-3 sm:grid-cols-3">
            <div className="rounded-xl border border-white/10 bg-linear-to-b from-white/8 to-white/3 p-3">
              <div className="flex items-center gap-2">
                <DollarSign size={12} className="text-white/55" />
                <p className="text-[10px] uppercase tracking-[0.16em] text-white/55">Total Cost</p>
              </div>
              <p className="mt-1 text-xl font-black text-emerald-300">{formatCost(data.total_cost_usd)}</p>
            </div>
            <div className="rounded-xl border border-white/10 bg-linear-to-b from-white/8 to-white/3 p-3">
              <div className="flex items-center gap-2">
                <Coins size={12} className="text-white/55" />
                <p className="text-[10px] uppercase tracking-[0.16em] text-white/55">Avg Cost / Convert</p>
              </div>
              <p className="mt-1 text-xl font-black text-orange-300">{formatCost(data.avg_cost_per_convert)}</p>
            </div>
            <div className="rounded-xl border border-white/10 bg-linear-to-b from-white/8 to-white/3 p-3">
              <div className="flex items-center gap-2">
                <Cpu size={12} className="text-white/55" />
                <p className="text-[10px] uppercase tracking-[0.16em] text-white/55">Total Tokens</p>
              </div>
              <p className="mt-1 text-xl font-black text-blue-300">{totalTokens.toLocaleString()}</p>
              <p className="mt-0.5 text-[10px] text-white/40">
                In: {data.total_input_tokens.toLocaleString()} · Out: {data.total_output_tokens.toLocaleString()}
              </p>
            </div>
          </div>

          {data.cost_by_model.length > 0 ? (
            <div className="space-y-3 text-sm text-white/80">
              <p className="text-[10px] uppercase tracking-[0.16em] text-white/55">Cost by Model</p>
              {data.cost_by_model.map((row) => (
                <div key={row.model}>
                  <div className="mb-1 flex items-center justify-between text-xs">
                    <span className="text-white/70">{row.model}</span>
                    <span className="font-semibold text-white">
                      {formatCost(row.cost_usd)} · {row.requests.toLocaleString()} req
                    </span>
                  </div>
                  <div className="h-2 overflow-hidden rounded-full bg-white/10">
                    <div
                      className="h-full rounded-full bg-linear-to-r from-emerald-300 to-accent-green"
                      style={{ width: widthFromRatio(row.cost_usd, maxModelCost) }}
                    />
                  </div>
                </div>
              ))}
            </div>
          ) : null}
        </>
      )}
    </div>
  );
}
