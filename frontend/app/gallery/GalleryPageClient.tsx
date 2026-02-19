"use client";

import { Skeleton } from "@/components/ui/Skeleton";
import { usePublicShares } from "@/hooks/usePublicShares";
import { motion } from "framer-motion";
import { ArrowRight, ExternalLink, Sparkles, Copy, Loader2, AlertCircle } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";

interface CloneShare {
  slug: string;
  title: string;
  template?: string;
}

export default function GalleryPageClient() {
  const { items, loading, error } = usePublicShares();
  const shareCount = items.length;
  const router = useRouter();
  const [cloneModal, setCloneModal] = useState<CloneShare | null>(null);
  const [newTitle, setNewTitle] = useState("");
  const [cloning, setCloning] = useState(false);
  const [cloneError, setCloneError] = useState("");

  const handleClone = async () => {
    if (!cloneModal) return;
    
    setCloning(true);
    setCloneError("");
    
    try {
      const response = await fetch("/api/mdflow/clone-template", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          source_share_slug: cloneModal.slug,
          new_title: newTitle || cloneModal.title,
          source_template: cloneModal.template,
        }),
      });
      
      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.error || `Clone failed (${response.status})`);
      }
      
      const data = await response.json();
      // Track clone event in telemetry
      if (typeof window !== "undefined" && (window as any).gtag) {
        (window as any).gtag("event", "template_cloned", {
          source_slug: cloneModal.slug,
          new_slug: data.slug,
        });
      }
      
      setCloneModal(null);
      setNewTitle("");
      // Redirect to new share
      router.push(data.redirect_url);
    } catch (err) {
      setCloneError(err instanceof Error ? err.message : "Clone failed");
    } finally {
      setCloning(false);
    }
  };

  return (
    <div className="min-h-[70vh] space-y-8">
      <div className="relative overflow-hidden rounded-3xl border border-white/10 bg-linear-to-br from-white/6 via-black/40 to-black/80 p-6 sm:p-8">
        <div className="absolute -top-20 -right-16 h-48 w-48 rounded-full bg-accent-orange/15 blur-3xl" />
        <div className="absolute -bottom-24 -left-10 h-56 w-56 rounded-full bg-blue-500/10 blur-3xl" />
        <div className="relative z-10 flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
          <div className="space-y-3">
            <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.3em] text-accent-orange/80">
              <Sparkles className="h-3 w-3" />
              Public Gallery
            </div>
            <h1 className="text-3xl sm:text-4xl font-black text-white">
              Shared Specs
            </h1>
            <p className="text-sm sm:text-base text-white/60 max-w-2xl">
              Explore community specs shared via link-only access. Save inspiration, compare templates, and jump into the studio in one click.
            </p>
          </div>
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="rounded-2xl border border-white/10 bg-black/40 px-4 py-3 text-xs text-white/70">
              <p className="text-[10px] uppercase tracking-widest text-white/40">
                Live Shares
              </p>
              <p className="text-xl font-black text-white">
                {loading ? "—" : shareCount}
              </p>
            </div>
            <Link
              href="/studio"
              className="inline-flex items-center justify-center gap-2 rounded-2xl bg-accent-orange px-5 py-3 text-xs font-bold uppercase tracking-widest text-white shadow-lg shadow-accent-orange/30 transition-all hover:bg-accent-orange/90"
            >
              Open Studio
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.3em] text-white/40 font-semibold">
            Latest Drops
          </p>
          <h2 className="text-xl font-black text-white">Browse specs</h2>
        </div>
        <Link
          href="/studio"
          className="hidden sm:flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-white/60 hover:text-white"
        >
          Create a share
          <ArrowRight className="w-3.5 h-3.5" />
        </Link>
      </div>

      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, idx) => (
            <div
              key={`skeleton-${idx}`}
              className="rounded-2xl border border-white/10 bg-white/4 p-4 space-y-3"
            >
              <div className="flex items-center justify-between">
                <Skeleton variant="text" className="h-3 w-20" />
                <Skeleton variant="text" className="h-3 w-16" />
              </div>
              <Skeleton variant="title" className="h-4 w-2/3" />
              <Skeleton variant="text" className="h-3 w-24" />
            </div>
          ))}
        </div>
      ) : error ? (
        <div className="rounded-2xl border border-red-500/20 bg-red-500/5 p-6 text-center space-y-2">
          <p className="text-sm font-semibold text-red-400">Failed to load public shares.</p>
          <p className="text-xs text-red-400/70">{error}</p>
        </div>
      ) : items.length ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map((item, idx) => (
            <motion.div
              key={`${item.slug}-${idx}`}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: idx * 0.03 }}
              className="group rounded-2xl border border-white/10 bg-white/4 p-4 flex flex-col gap-4 hover:border-white/20 hover:bg-white/6 transition-all"
            >
              <div className="flex items-center justify-between">
                <span className="text-[10px] uppercase tracking-widest text-white/40">
                  {item.template || "spec"}
                </span>
                <span className="text-[10px] text-white/40">
                  {new Date(item.created_at).toLocaleDateString()}
                </span>
              </div>
              <h3 className="text-sm font-bold text-white group-hover:text-white/90">
                {item.title || "Untitled Spec"}
              </h3>
              <div className="flex items-center justify-between mt-auto gap-2 flex-wrap">
                <span className="text-[10px] text-white/40">Link-only</span>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => {
                      setCloneModal({ slug: item.slug, title: item.title, template: item.template });
                      setNewTitle("");
                      setCloneError("");
                    }}
                    className="flex items-center gap-1.5 text-xs font-semibold text-accent-orange hover:text-accent-orange/90 transition-colors"
                    title="Clone this template"
                  >
                    Clone
                    <Copy className="w-3 h-3" />
                  </button>
                  <Link
                    href={`/s/${item.slug}`}
                    className="flex items-center gap-1.5 text-xs font-semibold text-accent-orange hover:text-accent-orange/90"
                  >
                    View
                    <ExternalLink className="w-3.5 h-3.5" />
                  </Link>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      ) : (
        <div className="rounded-2xl border border-white/10 bg-white/4 p-6 text-center space-y-3">
          <p className="text-sm font-semibold text-white/70">No public shares yet.</p>
          <p className="text-xs text-white/50">
            Create a spec in the studio and share it publicly to populate this space.
          </p>
          <Link
            href="/studio"
            className="inline-flex items-center justify-center gap-2 rounded-xl bg-accent-orange/20 px-4 py-2 text-xs font-bold uppercase tracking-widest text-accent-orange hover:bg-accent-orange/30"
          >
            Open Studio
            <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>
      )}

      {/* Clone Modal */}
      {cloneModal && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={() => !cloning && setCloneModal(null)}
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
        >
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-black/80 border border-white/20 rounded-2xl p-6 max-w-md w-full space-y-4"
          >
            <div className="space-y-1">
              <h2 className="text-lg font-bold text-white">Clone Template</h2>
              <p className="text-sm text-white/60">
                Create a copy of "{cloneModal.title}" to edit in your studio
              </p>
            </div>

            {cloneError && (
              <div className="flex items-start gap-3 rounded-lg bg-red-500/15 border border-red-500/30 p-3">
                <AlertCircle className="w-4 h-4 text-red-400 flex-shrink-0 mt-0.5" />
                <p className="text-xs text-red-300">{cloneError}</p>
              </div>
            )}

            <input
              type="text"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              placeholder={cloneModal.title}
              disabled={cloning}
              className="w-full px-3 py-2 rounded-lg bg-white/8 border border-white/15 text-white text-sm placeholder-white/40 focus:outline-none focus:border-accent-orange/50 focus:bg-white/12 disabled:opacity-50"
            />

            <div className="space-y-2 pt-2">
              <button
                onClick={handleClone}
                disabled={cloning}
                className="w-full flex items-center justify-center gap-2 rounded-lg bg-accent-orange px-4 py-2.5 text-sm font-bold text-white hover:bg-accent-orange/90 disabled:opacity-60 disabled:cursor-not-allowed transition-all"
              >
                {cloning ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Cloning...
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    Clone Now
                  </>
                )}
              </button>
              <button
                onClick={() => !cloning && setCloneModal(null)}
                disabled={cloning}
                className="w-full rounded-lg bg-white/8 border border-white/15 px-4 py-2.5 text-sm font-semibold text-white hover:bg-white/12 disabled:opacity-50 transition-all"
              >
                Cancel
              </button>
            </div>
          </motion.div>
        </motion.div>
      )}
    </div>
  );
}
