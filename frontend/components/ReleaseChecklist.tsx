"use client";

import { motion } from "framer-motion";
import {
  Check,
  X,
  AlertCircle,
  GitBranch,
  Shield,
  Zap,
  Users,
  TrendingUp,
} from "lucide-react";

interface ChecklistItem {
  id: string;
  label: string;
  description?: string;
  status: "complete" | "pending" | "blocked";
  signoff?: string;
  icon: typeof Check;
}

interface ReleaseChecklistProps {
  items?: ChecklistItem[];
  summary?: {
    completedCount: number;
    totalCount: number;
    blockedCount: number;
  };
}

const DEFAULT_ITEMS: ChecklistItem[] = [
  {
    id: "share-loop",
    label: "Share Loop with Comment Resolution",
    description: "Comment resolution workflow + event logging",
    status: "complete",
    signoff: "Product",
    icon: Check,
  },
  {
    id: "template-clone",
    label: "Template Clone Feature",
    description: "Gallery clone UI + backend endpoint tested",
    status: "complete",
    signoff: "Engineering",
    icon: Check,
  },
  {
    id: "quota-enforcement",
    label: "Quota Enforcement",
    description: "Token limits + daily reset + 429 responses",
    status: "complete",
    signoff: "Engineering",
    icon: Check,
  },
  {
    id: "usage-reporting",
    label: "Daily Usage Report API",
    description: "GET /api/quota/daily-report with aggregation",
    status: "complete",
    signoff: "Engineering",
    icon: Check,
  },
  {
    id: "load-test",
    label: "Load Test Baseline",
    description: "P95 latency < 2s under 1000 req/min",
    status: "complete",
    signoff: "QA",
    icon: Check,
  },
  {
    id: "security-review",
    label: "Security Audit",
    description: "OWASP Top 10 + input validation checklist",
    status: "complete",
    signoff: "Security",
    icon: Check,
  },
  {
    id: "p0-bugs",
    label: "P0 Bugs: Zero Open",
    description: "All critical issues resolved",
    status: "complete",
    signoff: "QA",
    icon: Check,
  },
  {
    id: "test-coverage",
    label: "Test Coverage ≥ 80%",
    description: "Unit + integration tests for new features",
    status: "complete",
    signoff: "Engineering",
    icon: Check,
  },
  {
    id: "monitoring",
    label: "Post-Launch Monitoring Runbook",
    description: "On-call procedures + alert rules configured",
    status: "complete",
    signoff: "Platform",
    icon: Check,
  },
  {
    id: "documentation",
    label: "Release Notes & API Docs",
    description: "Changelog + breaking changes documented",
    status: "complete",
    signoff: "Product",
    icon: Check,
  },
];

const STATUS_COLORS = {
  complete: "bg-emerald-500/20 text-emerald-300 border-emerald-500/30",
  pending: "bg-amber-500/20 text-amber-300 border-amber-500/30",
  blocked: "bg-red-500/20 text-red-300 border-red-500/30",
};

const STATUS_ICONS = {
  complete: Check,
  pending: AlertCircle,
  blocked: X,
};

export function ReleaseChecklist({
  items = DEFAULT_ITEMS,
  summary,
}: ReleaseChecklistProps) {
  const computedSummary =
    summary ||
    items.reduce(
      (acc, item) => {
        if (item.status === "complete") acc.completedCount++;
        if (item.status === "blocked") acc.blockedCount++;
        acc.totalCount++;
        return acc;
      },
      { completedCount: 0, blockedCount: 0, totalCount: 0 }
    );

  const isAllComplete =
    computedSummary.completedCount === computedSummary.totalCount &&
    computedSummary.blockedCount === 0;

  return (
    <section className="relative space-y-6">
      <div className="relative overflow-hidden rounded-3xl border border-white/10 bg-linear-to-br from-white/6 via-black/35 to-black/80 p-6 sm:p-8">
        <div className="pointer-events-none absolute -top-24 -right-16 h-64 w-64 rounded-full bg-emerald-500/10 blur-3xl" />
        <div className="relative z-10 space-y-4">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-2 flex-1">
              <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1">
                <Zap className="h-3 w-3 text-accent-orange" />
                <span className="text-[10px] font-semibold uppercase tracking-[0.2em] text-accent-orange/80">
                  MVP Release
                </span>
              </div>
              <h1 className="text-2xl sm:text-3xl font-black text-white">
                Release Checklist
              </h1>
              <p className="text-sm text-white/65 max-w-2xl">
                Epic 1-5 completion status. All items must be green for
                Go/No-Go decision.
              </p>
            </div>

            {isAllComplete && (
              <motion.div
                initial={{ scale: 0.9 }}
                animate={{ scale: 1 }}
                className="rounded-2xl bg-emerald-500/20 border border-emerald-500/30 px-4 py-3 text-center"
              >
                <div className="text-2xl font-black text-emerald-300">✓</div>
                <div className="text-[10px] uppercase tracking-wider text-emerald-300 font-semibold">
                  Ready to Launch
                </div>
              </motion.div>
            )}
          </div>

          {/* Progress Bar */}
          <div className="space-y-2 pt-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-white/70">Progress</span>
              <span className="font-bold text-white">
                {computedSummary.completedCount}/
                {computedSummary.totalCount}
              </span>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-white/10">
              <motion.div
                initial={{ width: 0 }}
                animate={{
                  width: `${(computedSummary.completedCount / computedSummary.totalCount) * 100}%`,
                }}
                className="h-full rounded-full bg-linear-to-r from-emerald-400 to-accent-orange"
                transition={{ duration: 0.6 }}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Checklist Items */}
      <div className="space-y-2">
        {items.map((item, idx) => {
          const StatusIcon = STATUS_ICONS[item.status];
          const colors = STATUS_COLORS[item.status];

          return (
            <motion.div
              key={item.id}
              initial={{ opacity: 0, x: -12 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: idx * 0.03 }}
              className={`rounded-xl border p-4 transition-all ${
                item.status === "complete"
                  ? "border-white/10 bg-white/4"
                  : item.status === "blocked"
                    ? "border-red-500/30 bg-red-500/5"
                    : "border-amber-500/30 bg-amber-500/5"
              }`}
            >
              <div className="flex items-start gap-4">
                <div
                  className={`mt-1 rounded-lg border p-2 ${colors}`}
                >
                  <StatusIcon className="h-4 w-4" />
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-2 mb-1">
                    <h3 className="font-bold text-white text-sm">
                      {item.label}
                    </h3>
                    {item.signoff && (
                      <span className="text-[10px] font-semibold uppercase tracking-wider text-white/50 whitespace-nowrap">
                        {item.signoff}
                      </span>
                    )}
                  </div>
                  {item.description && (
                    <p className="text-xs text-white/60">{item.description}</p>
                  )}
                </div>
              </div>
            </motion.div>
          );
        })}
      </div>

      {/* Go/No-Go Framework */}
      <div className="grid gap-4 lg:grid-cols-3">
        <div className="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-5">
          <div className="mb-3 flex items-center gap-2 text-white">
            <Check className="h-5 w-5 text-emerald-400" />
            <h3 className="font-bold text-sm uppercase tracking-wider">
              Go Criteria
            </h3>
          </div>
          <ul className="space-y-2 text-xs text-white/70">
            <li>• All items complete (100%)</li>
            <li>• Zero P0/P1 open bugs</li>
            <li>• Load test baseline met</li>
            <li>• Security review passed</li>
            <li>• All sign-offs received</li>
          </ul>
        </div>

        <div className="rounded-2xl border border-red-500/30 bg-red-500/10 p-5">
          <div className="mb-3 flex items-center gap-2 text-white">
            <X className="h-5 w-5 text-red-400" />
            <h3 className="font-bold text-sm uppercase tracking-wider">
              No-Go Criteria
            </h3>
          </div>
          <ul className="space-y-2 text-xs text-white/70">
            <li>• Incomplete checklist items</li>
            <li>• Open P0 bugs</li>
            <li>• Failed load test</li>
            <li>• Security issues found</li>
            <li>• Missing critical sign-offs</li>
          </ul>
        </div>

        <div className="rounded-2xl border border-white/10 bg-white/4 p-5">
          <div className="mb-3 flex items-center gap-2 text-white">
            <TrendingUp className="h-5 w-5 text-accent-orange" />
            <h3 className="font-bold text-sm uppercase tracking-wider">
              Post-Launch Plan
            </h3>
          </div>
          <ul className="space-y-2 text-xs text-white/70">
            <li>• 24/7 on-call rotation</li>
            <li>• Error budget: 99.5% uptime</li>
            <li>• Daily metrics review (1 week)</li>
            <li>• Gradual rollout: 25% → 100%</li>
            <li>• Rollback plan ready</li>
          </ul>
        </div>
      </div>

      {/* Decision Summary */}
      <div
        className={`rounded-2xl border p-6 ${
          isAllComplete
            ? "border-emerald-500/30 bg-emerald-500/10"
            : "border-amber-500/30 bg-amber-500/10"
        }`}
      >
        <div className="flex items-start gap-3">
          {isAllComplete ? (
            <Check className="h-6 w-6 text-emerald-400 shrink-0" />
          ) : (
            <AlertCircle className="h-6 w-6 text-amber-400 shrink-0" />
          )}
          <div>
            <h3
              className={`font-bold text-lg mb-1 ${
                isAllComplete ? "text-emerald-300" : "text-amber-300"
              }`}
            >
              {isAllComplete
                ? "✓ GO — Ready for Public Beta Launch"
                : `⚠ PENDING — ${computedSummary.blockedCount} Blockers`}
            </h3>
            <p className="text-sm text-white/70">
              {isAllComplete
                ? "All acceptance criteria met. Product and Engineering sign-offs complete. Monitoring runbook deployed. Ready for deployment to production."
                : `Review blocked items above before proceeding. Address all P0 issues and ensure all sign-offs are received.`}
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
