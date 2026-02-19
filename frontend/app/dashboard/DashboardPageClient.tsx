"use client";

import { useTelemetryDashboardQuery } from "@/lib/mdflowQueries";
import { mapErrorToUserFacing } from "@/lib/errorUtils";
// import { ReleaseChecklist } from "@/components/ReleaseChecklist";
import { QuotaUsageCard } from "@/components/QuotaUsageCard";
import { AICostDashboard } from "@/components/AICostDashboard";
import { AlertTriangle, Gauge, LineChart, RefreshCw } from "lucide-react";
import { useMemo, useState } from "react";
import { Select } from "@/components/ui/Select";

function percent(value: number): string {
  return `${(value * 100).toFixed(1)}%`;
}

function compactNumber(value: number): string {
  return value.toLocaleString("en-US");
}

function widthFromRatio(value: number, max: number): string {
  if (max <= 0) {
    return "0%";
  }
  return `${Math.max(4, Math.min(100, (value / max) * 100))}%`;
}

export default function DashboardPageClient() {
  const [hours, setHours] = useState(24);
  const { data, isLoading, isFetching, error, refetch } = useTelemetryDashboardQuery(hours, true);
  const mappedError = useMemo(
    () =>
      error instanceof Error
        ? mapErrorToUserFacing(error.message)
        : null,
    [error]
  );

  const cards = useMemo(() => {
    if (!data) {
      return [];
    }
    return [
      {
        label: "Activation (10m)",
        value: percent(data.kpis.activation_rate_10m),
        tone: "text-emerald-300",
        barClass: "bg-gradient-to-r from-emerald-300 to-emerald-400",
      },
      {
        label: "Preview Success",
        value: percent(data.kpis.preview_success_rate),
        tone: "text-blue-300",
        barClass: "bg-gradient-to-r from-blue-300 to-sky-400",
      },
      {
        label: "Convert Success",
        value: percent(data.kpis.convert_success_rate),
        tone: "text-orange-300",
        barClass: "bg-gradient-to-r from-orange-300 to-accent-orange",
      },
      {
        label: "API 5xx Rate",
        value: percent(data.reliability.api_5xx_rate),
        tone: "text-red-300",
        barClass: "bg-gradient-to-r from-red-300 to-orange-300",
      },
    ];
  }, [data]);

  const funnelRows = useMemo(() => {
    if (!data) {
      return [];
    }

    return [
      { label: "Studio Opened", value: data.funnel.studio_opened },
      { label: "Input Provided", value: data.funnel.input_provided },
      { label: "Preview Succeeded", value: data.funnel.preview_succeeded },
      { label: "Convert Succeeded", value: data.funnel.convert_succeeded },
      { label: "Share Created", value: data.funnel.share_created },
    ];
  }, [data]);

  const maxFunnel = useMemo(
    () => funnelRows.reduce((max, row) => (row.value > max ? row.value : max), 0),
    [funnelRows]
  );

  const latencyRows = useMemo(() => {
    if (!data) {
      return [];
    }

    return [
      { label: "TTV Median", value: data.kpis.time_to_value_ms.median },
      { label: "TTV p75", value: data.kpis.time_to_value_ms.p75 },
      { label: "TTV p95", value: data.kpis.time_to_value_ms.p95 },
      { label: "p95 Preview API", value: data.reliability.p95_preview_latency_ms },
      { label: "p95 Convert API", value: data.reliability.p95_convert_latency_ms },
    ];
  }, [data]);

  const maxLatency = useMemo(
    () => latencyRows.reduce((max, row) => (row.value > max ? row.value : max), 0),
    [latencyRows]
  );

  const maxErrorCount = useMemo(() => {
    if (!data?.errors || data.errors.length === 0) {
      return 0;
    }
    return data.errors.reduce((max, row) => (row.count > max ? row.count : max), 0);
  }, [data]);

  return (
    <section className="relative space-y-6 overflow-hidden">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 -top-20 h-44 bg-[radial-gradient(circle_at_top,rgba(242,123,47,0.18),transparent_70%)]"
      />

      <div className="relative overflow-hidden rounded-3xl border border-white/10 bg-linear-to-br from-white/6 via-black/35 to-black/80 p-5 shadow-[0_18px_60px_-35px_rgba(242,123,47,0.6)] backdrop-blur sm:p-6">
        <div className="pointer-events-none absolute -top-24 -right-16 h-64 w-64 rounded-full bg-accent-orange/12 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-24 -left-10 h-64 w-64 rounded-full bg-blue-500/8 blur-3xl" />
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="space-y-2">
            <p className="text-[10px] font-semibold uppercase tracking-[0.3em] text-accent-orange/85">Observability</p>
            <h1 className="text-2xl font-black tracking-tight text-white">MVP Telemetry Dashboard</h1>
            <p className="max-w-2xl text-sm text-white/65">
              Theo doi conversion health, reliability va latency cua he thong theo thoi gian thuc.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Select
              value={hours.toString()}
              onValueChange={(value) => setHours(Number(value))}
              options={[
                { label: "Last 6h", value: "6" },
                { label: "Last 24h", value: "24" },
                { label: "Last 72h", value: "72" },
              ]}
              aria-label="Time window"
              placeholder="Select window"
              className="w-auto min-w-40"
            />
            <button
              type="button"
              onClick={() => void refetch()}
              disabled={isFetching}
              className="inline-flex h-11 cursor-pointer items-center gap-2 rounded-xl border border-accent-orange/40 bg-accent-orange/15 px-4 text-xs font-bold uppercase tracking-[0.16em] text-orange-100 transition-colors hover:bg-accent-orange/25 disabled:cursor-not-allowed disabled:opacity-70"
            >
              <RefreshCw size={14} className={isFetching ? "animate-spin" : ""} />
              Refresh
            </button>
          </div>
        </div>

        <div className="mt-4 flex flex-wrap gap-2 text-xs text-white/70">
          <div className="rounded-full border border-white/15 bg-black/25 px-3 py-1">Live telemetry</div>
          <div className="rounded-full border border-white/15 bg-black/25 px-3 py-1">Window: {hours}h</div>
          <div className="rounded-full border border-white/15 bg-black/25 px-3 py-1">
            {isFetching ? "Syncing" : "Synced"}
          </div>
        </div>
      </div>

      {error ? (
        <div className="rounded-2xl border border-red-500/30 bg-red-500/10 p-5 text-sm text-red-200">
          <div className="flex items-center gap-2">
            <AlertTriangle size={16} />
            <p className="font-semibold">{mappedError?.title || "Failed to load dashboard"}</p>
          </div>
          <p className="mt-1 text-red-300">{mappedError?.message || (error instanceof Error ? error.message : "unknown error")}</p>
          {mappedError?.requestId ? (
            <p className="mt-1 text-xs font-mono text-red-200/70">request_id={mappedError.requestId}</p>
          ) : null}
          {mappedError?.retryable ? (
            <button
              type="button"
              onClick={() => void refetch()}
              className="mt-3 inline-flex h-9 cursor-pointer items-center gap-1 rounded-md border border-red-300/30 bg-red-300/10 px-3 text-[11px] font-bold uppercase tracking-wider text-red-100 transition-colors hover:bg-red-300/20"
            >
              <RefreshCw size={12} />
              Retry
            </button>
          ) : null}
        </div>
      ) : null}

      {isLoading && !data ? (
        <div className="rounded-2xl border border-white/10 bg-white/5 p-6">
          <div className="space-y-3 animate-pulse">
            <div className="h-4 w-40 rounded bg-white/10" />
            <div className="h-3 w-64 rounded bg-white/10" />
            <div className="h-20 rounded-xl bg-white/10" />
          </div>
        </div>
      ) : null}

      {data ? (
        <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {cards.map((card, index) => (
              <div
                key={card.label}
                className="rounded-2xl border border-white/10 bg-linear-to-b from-white/8 to-white/3 p-4 shadow-[0_12px_26px_-20px_rgba(0,0,0,0.8)] transition-colors hover:border-white/20"
              >
                <div className="flex items-start justify-between gap-3">
                  <p className="text-[10px] uppercase tracking-[0.16em] text-white/55">{card.label}</p>
                  <span className="rounded-lg border border-white/15 bg-black/25 p-1.5 text-white/75">
                    {index === 0 ? <LineChart size={14} /> : null}
                    {index === 1 ? <Gauge size={14} /> : null}
                    {index === 2 ? <LineChart size={14} /> : null}
                    {index === 3 ? <AlertTriangle size={14} /> : null}
                  </span>
                </div>
                <p className={`mt-2 text-2xl font-black ${card.tone}`}>{card.value}</p>
                <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-white/10">
                  <div
                    className={`h-full rounded-full ${card.barClass}`}
                    style={{ width: card.value }}
                  />
                </div>
              </div>
            ))}
          </div>

          <div className="grid gap-4 lg:grid-cols-2">
            <div className="rounded-2xl border border-white/10 bg-white/4 p-5">
              <div className="mb-4 flex items-center justify-between gap-2 text-white">
                <div className="flex items-center gap-2">
                  <LineChart size={16} className="text-accent-orange" />
                  <h2 className="text-sm font-bold uppercase tracking-wider">Funnel</h2>
                </div>
                <span className="text-xs text-white/50">Step conversion</span>
              </div>
              <div className="space-y-3 text-sm text-white/80">
                {funnelRows.map((row) => (
                  <div key={row.label}>
                    <div className="mb-1 flex items-center justify-between text-xs">
                      <span className="text-white/70">{row.label}</span>
                      <span className="font-semibold text-white">{compactNumber(row.value)}</span>
                    </div>
                    <div className="h-2 overflow-hidden rounded-full bg-white/10">
                      <div
                        className="h-full rounded-full bg-linear-to-r from-accent-orange to-orange-300"
                        style={{ width: widthFromRatio(row.value, maxFunnel) }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-2xl border border-white/10 bg-white/4 p-5">
              <div className="mb-4 flex items-center justify-between gap-2 text-white">
                <div className="flex items-center gap-2">
                  <Gauge size={16} className="text-accent-orange" />
                  <h2 className="text-sm font-bold uppercase tracking-wider">TTV and Latency</h2>
                </div>
                <span className="text-xs text-white/50">Milliseconds</span>
              </div>
              <div className="space-y-3 text-sm text-white/80">
                {latencyRows.map((row) => (
                  <div key={row.label}>
                    <div className="mb-1 flex items-center justify-between text-xs">
                      <span className="text-white/70">{row.label}</span>
                      <span className="font-semibold text-white">{compactNumber(row.value)} ms</span>
                    </div>
                    <div className="h-2 overflow-hidden rounded-full bg-white/10">
                      <div
                        className="h-full rounded-full bg-linear-to-r from-emerald-300 to-accent-green"
                        style={{ width: widthFromRatio(row.value, maxLatency) }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="rounded-2xl border border-white/10 bg-white/4 p-5">
            <div className="mb-4 flex items-center justify-between gap-2 text-white">
              <div className="flex items-center gap-2">
                <AlertTriangle size={16} className="text-accent-orange" />
                <h2 className="text-sm font-bold uppercase tracking-wider">Top Error Events</h2>
              </div>
              <span className="text-xs text-white/50">By occurrence</span>
            </div>
            {!data.errors || data.errors.length === 0 ? (
              <p className="text-sm text-white/60">No error events in this window.</p>
            ) : (
              <div className="space-y-2">
                {data.errors.map((item) => (
                  <div
                    key={item.event_name}
                    className="rounded-lg border border-white/10 bg-black/20 px-3 py-2"
                  >
                    <div className="mb-1 flex items-center justify-between gap-3 text-sm">
                      <span className="truncate text-white/80">{item.event_name}</span>
                      <span className="font-bold text-white">{compactNumber(item.count)}</span>
                    </div>
                    <div className="h-1.5 overflow-hidden rounded-full bg-white/10">
                      <div
                        className="h-full rounded-full bg-linear-to-r from-rose-300 to-orange-300"
                        style={{ width: widthFromRatio(item.count, maxErrorCount) }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <AICostDashboard data={data?.ai_cost} />

          <div className="rounded-2xl border border-white/10 bg-linear-to-r from-black/70 via-black/55 to-accent-orange/12 p-4 text-xs text-white/65">
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
              <span>Window: last {data.window_hours} hours</span>
              <span>Generated at: {data.generated_at}</span>
              <span>Events: {compactNumber(data.totals.events_total)}</span>
              <span>Frontend: {compactNumber(data.totals.frontend_events)}</span>
              <span>Backend: {compactNumber(data.totals.backend_events)}</span>
            </div>
          </div>
          </>
          ) : null}

          <div className="mt-12 border-t border-white/10 pt-12 space-y-8">
            <div>
              <div className="mb-4">
                <p className="text-[10px] font-semibold uppercase tracking-[0.3em] text-blue-300/85">Usage Management</p>
                <h2 className="text-xl font-black tracking-tight text-white">Quota & Usage</h2>
              </div>
              <QuotaUsageCard showReportHistory={true} refreshInterval={30000} />
            </div>
          {/* <ReleaseChecklist /> */}
          </div>
          </section>
          );
          }
