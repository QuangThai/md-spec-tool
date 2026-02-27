"use client";

import { useShareSlug } from "@/hooks/useShareSlug";
import { LazyMotion, domAnimation, m } from "framer-motion";
import {
  AlertCircle,
  ArrowLeft,
  Check,
  Copy,
  Download,
  ExternalLink,
  FileText,
  MessageSquare,
  Terminal,
} from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Suspense, useCallback, useMemo, useState } from "react";

interface CommentFormState {
  author: string;
  message: string;
}

function ShareSlugErrorView({ error }: { error: string }) {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <LazyMotion features={domAnimation}>
        <m.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="max-w-md w-full p-8 rounded-2xl bg-white/5 border border-white/10 text-center"
        >
          <div className="w-16 h-16 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center justify-center mx-auto mb-6">
            <AlertCircle className="w-8 h-8 text-red-400" />
          </div>
          <h1 className="text-xl font-black text-white mb-2">Invalid Share Link</h1>
          <p className="text-sm text-white/60 mb-6">{error}</p>
          <Link
            href="/studio"
            className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-accent-orange hover:bg-accent-orange/90 text-white font-bold uppercase tracking-wider text-sm transition-[background-color,border-color,color]"
          >
            <ArrowLeft className="w-4 h-4" />
            Go to Studio
          </Link>
        </m.div>
      </LazyMotion>
    </div>
  );
}

function ShareSlugLoadingView() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="flex items-center gap-3 text-white/60">
        <div className="w-5 h-5 border-2 border-accent-orange/30 border-t-accent-orange rounded-full animate-spin" />
        <span className="text-sm font-medium">Loading shared content...</span>
      </div>
    </div>
  );
}

type ShareSlugSidebarProps = {
  shareData: NonNullable<ReturnType<typeof useShareSlug>["share"]>;
  createdDate: string;
  onToggleShare: (p: { is_public?: boolean; allow_comments?: boolean }) => Promise<unknown>;
  updatingShare: boolean;
  canComment: boolean;
  form: CommentFormState;
  setForm: React.Dispatch<React.SetStateAction<CommentFormState>>;
  onSubmitComment: () => Promise<void>;
  creatingComment: boolean;
  comments: { items: { id: string; author: string; message: string; resolved: boolean }[] };
  commentsLoading: boolean;
  onResolveComment: (commentId: string, resolved: boolean) => Promise<void>;
  onCopy: () => void;
  onDownload: () => void;
};

function ShareSlugSidebar({
  shareData,
  createdDate,
  onToggleShare,
  updatingShare,
  canComment,
  form,
  setForm,
  onSubmitComment,
  creatingComment,
  comments,
  commentsLoading,
  onResolveComment,
  onCopy,
  onDownload,
}: ShareSlugSidebarProps) {
  return (
    <div className="space-y-4">
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
          <div className="flex items-center justify-between">
            <span className="text-xs text-white/50">Visibility</span>
            <span className="text-xs text-white/70 font-mono">
              {shareData.is_public ? "Public" : "Private"}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-xs text-white/50">Comments</span>
            <span className="text-xs text-white/70 font-mono">
              {shareData.allow_comments ? "Enabled" : "Disabled"}
            </span>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-white/10 bg-white/2 p-5">
        <h3 className="text-xs font-black text-white/50 uppercase tracking-widest mb-4">
          Manage Share
        </h3>
        <div className="space-y-2">
          <button
            onClick={() => onToggleShare({ is_public: !shareData.is_public })}
            disabled={updatingShare}
            className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-xs font-semibold text-white/70 hover:text-white transition-[background-color,border-color,color] disabled:opacity-50"
          >
            {shareData.is_public ? "Make Private" : "Make Public"}
          </button>
          <button
            onClick={() => onToggleShare({ allow_comments: !shareData.allow_comments })}
            disabled={updatingShare}
            className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-xs font-semibold text-white/70 hover:text-white transition-[background-color,border-color,color] disabled:opacity-50"
          >
            {shareData.allow_comments ? "Disable Comments" : "Enable Comments"}
          </button>
        </div>
      </div>

      {canComment && (
        <div className="rounded-xl border border-white/10 bg-white/2 p-5">
          <h3 className="text-xs font-black text-white/50 uppercase tracking-widest mb-4">
            Add Comment
          </h3>
          <div className="space-y-3">
            <input
              type="text"
              value={form.author}
              onChange={(e) => setForm((prev) => ({ ...prev, author: e.target.value }))}
              placeholder="Your name"
              aria-label="Your name"
              className="w-full rounded-lg bg-black/30 border border-white/10 px-3 py-2 text-xs text-white/80 focus:outline-none focus-visible:border-accent-orange/40 focus-visible:ring-2 focus-visible:ring-accent-orange/20"
            />
            <textarea
              value={form.message}
              onChange={(e) => setForm((prev) => ({ ...prev, message: e.target.value }))}
              placeholder="Write a comment…"
              aria-label="Comment message"
              className="w-full min-h-[80px] rounded-lg bg-black/30 border border-white/10 px-3 py-2 text-xs text-white/80 focus:outline-none focus-visible:border-accent-orange/40 focus-visible:ring-2 focus-visible:ring-accent-orange/20"
            />
            <button
              onClick={onSubmitComment}
              disabled={creatingComment || !form.message.trim()}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-accent-orange hover:bg-accent-orange/90 text-xs font-bold uppercase tracking-wider text-white transition-[background-color,border-color,color] disabled:opacity-40 disabled:cursor-not-allowed"
            >
              <MessageSquare className="w-3.5 h-3.5" />
              {creatingComment ? "Posting…" : "Post Comment"}
            </button>
          </div>
        </div>
      )}

      <div className="rounded-xl border border-white/10 bg-white/2 p-5">
        <h3 className="text-xs font-black text-white/50 uppercase tracking-widest mb-4">
          Comments
        </h3>
        <div className="space-y-3">
          {commentsLoading ? (
            <p className="text-xs text-white/40">Loading comments...</p>
          ) : comments.items.length ? (
            comments.items.map((comment) => (
              <div
                key={comment.id}
                className="rounded-lg border border-white/10 bg-black/20 p-3"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-bold text-white/70">{comment.author}</span>
                  <button
                    onClick={() => onResolveComment(comment.id, !comment.resolved)}
                    className={`text-[10px] font-bold uppercase tracking-wider ${
                      comment.resolved ? "text-green-400" : "text-white/40 hover:text-white/60"
                    }`}
                  >
                    {comment.resolved ? "Resolved" : "Resolve"}
                  </button>
                </div>
                <p
                  className={`text-xs text-white/70 ${comment.resolved ? "line-through opacity-60" : ""}`}
                >
                  {comment.message}
                </p>
              </div>
            ))
          ) : (
            <p className="text-xs text-white/40">No comments yet.</p>
          )}
        </div>
      </div>

      <div className="rounded-xl border border-blue-500/20 bg-blue-500/5 p-4">
        <div className="flex items-start gap-3">
          <FileText className="w-4 h-4 text-blue-400 mt-0.5 shrink-0" />
          <div>
            <p className="text-xs font-bold text-blue-400 mb-1">Link-only Access</p>
            <p className="text-xs text-blue-400/70 leading-relaxed">
              Anyone with this link can view and comment based on permission. Data is stored in
              memory on the server and will reset on restart.
            </p>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-white/10 bg-white/2 p-5">
        <h3 className="text-xs font-black text-white/50 uppercase tracking-widest mb-4">
          Actions
        </h3>
        <div className="space-y-2">
          <button
            onClick={onCopy}
            className="w-full flex items-center gap-3 px-4 py-3 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-medium text-white/80 hover:text-white transition-[background-color,border-color,color] cursor-pointer"
          >
            <Copy className="w-4 h-4" />
            Copy to Clipboard
          </button>
          <button
            onClick={onDownload}
            className="w-full flex items-center gap-3 px-4 py-3 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-medium text-white/80 hover:text-white transition-[background-color,border-color,color] cursor-pointer"
          >
            <Download className="w-4 h-4" />
            Download .md
          </button>
          <Link
            href="/studio"
            className="w-full flex items-center gap-3 px-4 py-3 rounded-lg bg-accent-orange/10 hover:bg-accent-orange/20 border border-accent-orange/20 text-sm font-medium text-accent-orange transition-[background-color,border-color,color]"
          >
            <Terminal className="w-4 h-4" />
            Open in Studio
          </Link>
        </div>
      </div>
    </div>
  );
}

function ShareSlugContent() {
  const params = useParams<{ key: string }>();
  const shareKey = params?.key ?? "";

  const {
    share: shareData,
    shareLoading,
    shareError,
    comments,
    commentsLoading,
    createComment,
    creatingComment,
    updateComment,
    updateShare,
    updatingShare,
  } = useShareSlug(shareKey);
  const [form, setForm] = useState<CommentFormState>({ author: "", message: "" });
  const [copied, setCopied] = useState(false);
  const [copiedLink, setCopiedLink] = useState(false);

  const canComment = shareData?.allow_comments && shareData.permission === "comment";

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

  const handleSubmit = useCallback(async () => {
    if (!shareKey || !canComment || !form.message.trim()) return;
    await createComment({
      author: form.author.trim() || "Anonymous",
      message: form.message.trim(),
    });
    setForm((prev) => ({ ...prev, message: "" }));
  }, [canComment, createComment, form.author, form.message, shareKey]);

  const handleResolve = useCallback(
    async (commentId: string, resolved: boolean) => {
      if (!shareKey) return;
      await updateComment({ commentId, resolved });
    },
    [shareKey, updateComment]
  );

  const createdDate = useMemo(() => {
    if (!shareData) return "";
    return new Date(shareData.created_at).toLocaleString();
  }, [shareData]);

  const handleToggleShare = useCallback(
    async (payload: { is_public?: boolean; allow_comments?: boolean }) => {
      if (!shareKey) return;
      await updateShare(payload);
    },
    [shareKey, updateShare]
  );

  if (shareError) return <ShareSlugErrorView error={shareError} />;
  if (shareLoading || !shareData) return <ShareSlugLoadingView />;

  return (
    <div className="min-h-screen bg-white/2 rounded-2xl overflow-hidden">
      <header className="border-b border-white/10 bg-white/2 backdrop-blur-xl sticky top-0 z-40">
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
              <ExternalLink className="w-4 h-4 text-accent-orange" />
              <span className="text-sm font-black text-white uppercase tracking-wider">
                Shared Spec
              </span>
            </div>
          </div>

          <div className="flex items-center">
            <button
              onClick={handleCopyLink}
              className="flex items-center gap-2 px-3 py-1 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-sm font-medium text-white/70 hover:text-white transition-[background-color,border-color,color] cursor-pointer"
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

      <main className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight mb-6">
          {shareData.title || "Shared Specification"}
        </h1>
        <LazyMotion features={domAnimation}>
          <m.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-6"
          >
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
                  className="flex items-center justify-center w-8 h-8 rounded-lg bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-[background-color,border-color,color] cursor-pointer"
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
                  className="flex items-center justify-center w-8 h-8 rounded-lg bg-accent-orange hover:bg-accent-orange/90 text-white transition-[background-color,border-color,color] cursor-pointer"
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

          <ShareSlugSidebar
            shareData={shareData}
            createdDate={createdDate}
            onToggleShare={handleToggleShare}
            updatingShare={updatingShare}
            canComment={canComment}
            form={form}
            setForm={setForm}
            onSubmitComment={handleSubmit}
            creatingComment={creatingComment}
            comments={comments}
            commentsLoading={commentsLoading}
            onResolveComment={handleResolve}
            onCopy={handleCopy}
            onDownload={handleDownload}
          />
          </m.div>
        </LazyMotion>
      </main>
    </div>
  );
}

export default function ShareSlugPageClient() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-black flex items-center justify-center">
          <div className="flex items-center gap-3 text-white/60">
            <div className="w-5 h-5 border-2 border-accent-orange/30 border-t-accent-orange rounded-full animate-spin" />
            <span className="text-sm font-medium">Loading…</span>
          </div>
        </div>
      }
    >
      <ShareSlugContent />
    </Suspense>
  );
}
