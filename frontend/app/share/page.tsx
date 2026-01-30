"use client";

import { decodeShareData, ShareData } from "@/lib/shareUtils";
import { motion } from "framer-motion";
import {
  AlertCircle,
  ArrowLeft,
  Check,
  Copy,
  Download,
  ExternalLink,
  FileText,
  Share2,
  Terminal,
} from "lucide-react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useCallback, useEffect, useState } from "react";

function SharePageContent() {
  const searchParams = useSearchParams();
  const [shareData, setShareData] = useState<ShareData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [copiedLink, setCopiedLink] = useState(false);

  useEffect(() => {
    const encoded = searchParams.get("d");
    if (!encoded) {
      setError("No share data found in URL");
      return;
    }

    const decoded = decodeShareData(encoded);
    if (!decoded) {
      setError("Invalid or corrupted share link");
      return;
    }

    setShareData(decoded);
  }, [searchParams]);

  const handleCopy = useCallback(() => {
    if (!shareData) return;
    navigator.clipboard.writeText(shareData.mdflow);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [shareData]);

  const handleCopyLink = useCallback(() => {
    navigator.clipboard.writeText(window.location.href);
    setCopiedLink(true);
    setTimeout(() => setCopiedLink(false), 2000);
  }, []);

  const handleDownload = useCallback(() => {
    if (!shareData) return;
    const blob = new Blob([shareData.mdflow], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "shared-spec.mdflow.md";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [shareData]);

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="max-w-md w-full p-8 rounded-2xl bg-white/5 border border-white/10 text-center"
        >
          <div className="w-16 h-16 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center justify-center mx-auto mb-6">
            <AlertCircle className="w-8 h-8 text-red-400" />
          </div>
          <h1 className="text-xl font-black text-white mb-2">
            Invalid Share Link
          </h1>
          <p className="text-sm text-white/60 mb-6">{error}</p>
          <Link
            href="/studio"
            className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-accent-orange hover:bg-accent-orange/90 text-white font-bold uppercase tracking-wider text-sm transition-all"
          >
            <ArrowLeft className="w-4 h-4" />
            Go to Studio
          </Link>
        </motion.div>
      </div>
    );
  }

  if (!shareData) {
    return (
      <div className="min-h-screen  flex items-center justify-center">
        <div className="flex items-center gap-3 text-white/60">
          <div className="w-5 h-5 border-2 border-accent-orange/30 border-t-accent-orange rounded-full animate-spin" />
          <span className="text-sm font-medium">Loading shared content...</span>
        </div>
      </div>
    );
  }

  const createdDate = new Date(shareData.createdAt).toLocaleString();

  return (
    <div className="min-h-screen bg-white/2 rounded-2xl overflow-hidden">
      {/* Header */}
      <header className="border-b border-white/10 bg-white/2 backdrop-blur-xl sticky top-0 z-49">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              href="/studio"
              className="flex items-center gap-2 text-white/60 hover:text-white transition-colors"
            >
              <ArrowLeft className="w-4 h-4" />
              <span className="text-sm font-medium hidden sm:inline">
                Back to Studio
              </span>
            </Link>
            <div className="h-6 w-px bg-white/10" />
            <div className="flex items-center gap-2">
              <Share2 className="w-4 h-4 text-accent-orange" />
              <span className="text-sm font-black text-white uppercase tracking-wider">
                Shared Spec
              </span>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleCopyLink}
              className="flex items-center gap-2 px-3 py-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-medium text-white/70 hover:text-white transition-all cursor-pointer"
              title="Copy link"
            >
              {copiedLink ? (
                <Check className="w-4 h-4 text-green-400" />
              ) : (
                <ExternalLink className="w-4 h-4" />
              )}
              <span className="hidden sm:inline">
                {copiedLink ? "Copied!" : "Copy Link"}
              </span>
            </button>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="grid grid-cols-1 lg:grid-cols-[1fr_300px] gap-6"
        >
          {/* Output panel */}
          <div className="rounded-2xl border border-white/10 bg-white/2 overflow-hidden">
            <div className="flex items-center justify-between px-5 py-4 border-b border-white/10 bg-white/3">
              <div className="flex items-center gap-3">
                <div className="flex gap-1">
                  <span className="w-2 h-2 rounded-full bg-red-400/80" />
                  <span className="w-2 h-2 rounded-full bg-yellow-400/80" />
                  <span className="w-2 h-2 rounded-full bg-green-400/80" />
                </div>
                <span className="text-xs font-black text-white/60 uppercase tracking-widest">
                  MDFlow Output
                </span>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={handleCopy}
                  className="flex items-center justify-center w-8 h-8 rounded-lg bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-all cursor-pointer"
                  title={copied ? "Copied!" : "Copy"}
                >
                  {copied ? (
                    <Check className="w-4 h-4 text-green-400" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={handleDownload}
                  className="flex items-center justify-center w-8 h-8 rounded-lg bg-accent-orange hover:bg-accent-orange/90 text-white transition-all cursor-pointer"
                  title="Download"
                >
                  <Download className="w-4 h-4" />
                </button>
              </div>
            </div>

            <div className="p-5 max-h-[70vh] overflow-auto custom-scrollbar">
              <pre className="whitespace-pre-wrap wrap-break-word font-mono text-sm text-white/90 leading-relaxed">
                {shareData.mdflow}
              </pre>
            </div>
          </div>

          {/* Info sidebar */}
          <div className="space-y-4">
            {/* Metadata */}
            <div className="rounded-xl border border-white/10 bg-white/2 p-5">
              <h3 className="text-xs font-black text-white/50 uppercase tracking-widest mb-4">
                Share Info
              </h3>

              <div className="space-y-3">
                {shareData.template && (
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-white/50">Template</span>
                    <span className="text-xs font-bold text-white px-2 py-1 rounded-md uppercase bg-accent-orange/20">
                      {shareData.template}
                    </span>
                  </div>
                )}
                <div className="flex items-center justify-between">
                  <span className="text-xs text-white/50">Created</span>
                  <span className="text-xs text-white/70">{createdDate}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-white/50">Size</span>
                  <span className="text-xs text-white/70 font-mono">
                    {(shareData.mdflow.length / 1024).toFixed(1)} KB
                  </span>
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="rounded-xl border border-white/10 bg-white/2 p-5">
              <h3 className="text-xs font-black text-white/50 uppercase tracking-widest mb-4">
                Actions
              </h3>

              <div className="space-y-2">
                <button
                  onClick={handleCopy}
                  className="w-full flex items-center gap-3 px-4 py-3 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-medium text-white/80 hover:text-white transition-all cursor-pointer"
                >
                  <Copy className="w-4 h-4" />
                  Copy to Clipboard
                </button>
                <button
                  onClick={handleDownload}
                  className="w-full flex items-center gap-3 px-4 py-3 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-medium text-white/80 hover:text-white transition-all cursor-pointer"
                >
                  <Download className="w-4 h-4" />
                  Download .md
                </button>
                <Link
                  href="/studio"
                  className="w-full flex items-center gap-3 px-4 py-3 rounded-lg bg-accent-orange/10 hover:bg-accent-orange/20 border border-accent-orange/20 text-sm font-medium text-accent-orange transition-all"
                >
                  <Terminal className="w-4 h-4" />
                  Open in Studio
                </Link>
              </div>
            </div>

            {/* Note */}
            <div className="rounded-xl border border-blue-500/20 bg-blue-500/5 p-4">
              <div className="flex items-start gap-3">
                <FileText className="w-4 h-4 text-blue-400 mt-0.5 shrink-0" />
                <div>
                  <p className="text-xs font-bold text-blue-400 mb-1">
                    Stateless Share Link
                  </p>
                  <p className="text-xs text-blue-400/70 leading-relaxed">
                    This link contains all the data - no server storage needed.
                    The link will work forever as long as the URL is valid.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </main>
    </div>
  );
}

export default function SharePage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-black flex items-center justify-center">
          <div className="flex items-center gap-3 text-white/60">
            <div className="w-5 h-5 border-2 border-accent-orange/30 border-t-accent-orange rounded-full animate-spin" />
            <span className="text-sm font-medium">Loading...</span>
          </div>
        </div>
      }
    >
      <SharePageContent />
    </Suspense>
  );
}
