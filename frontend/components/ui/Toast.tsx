"use client";

import { AnimatePresence, motion } from "framer-motion";
import { AlertCircle, AlertTriangle, Check, Info, X } from "lucide-react";
import { useEffect } from "react";
import { Toast as ToastType, useToastStore } from "@/lib/toastStore";

const easeOutExpo: [number, number, number, number] = [0.16, 1, 0.3, 1];

const icons = {
  success: Check,
  error: AlertCircle,
  info: Info,
  warning: AlertTriangle,
};

const iconColors = {
  success: "text-emerald-400",
  error: "text-red-400",
  info: "text-blue-400",
  warning: "text-amber-400",
};

const borderColors = {
  success: "border-emerald-500/30",
  error: "border-red-500/30",
  info: "border-blue-500/30",
  warning: "border-amber-500/30",
};

interface ToastItemProps {
  toast: ToastType;
  onDismiss: (id: string) => void;
}

function ToastItem({ toast, onDismiss }: ToastItemProps) {
  const Icon = icons[toast.type];

  useEffect(() => {
    const duration = toast.duration ?? 3000;
    const timer = setTimeout(() => {
      onDismiss(toast.id);
    }, duration);

    return () => clearTimeout(timer);
  }, [toast.id, toast.duration, onDismiss]);

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 24, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, x: 100, scale: 0.95 }}
      transition={{ duration: 0.3, ease: easeOutExpo }}
      className={`
        relative flex items-start gap-3 w-80 p-4 rounded-lg
        bg-zinc-900/95 backdrop-blur-sm border ${borderColors[toast.type]}
        shadow-xl shadow-black/20
      `}
      role="alert"
      aria-live="polite"
    >
      <div className={`shrink-0 mt-0.5 ${iconColors[toast.type]}`}>
        <Icon size={18} strokeWidth={2} />
      </div>

      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-white">{toast.message}</p>
        {toast.description && (
          <p className="mt-1 text-xs text-zinc-400">{toast.description}</p>
        )}
      </div>

      <button
        onClick={() => onDismiss(toast.id)}
        className="shrink-0 p-1 -m-1 rounded-md text-zinc-500 hover:text-zinc-300 hover:bg-white/5 transition-colors"
        aria-label="Dismiss notification"
      >
        <X size={14} />
      </button>
    </motion.div>
  );
}

export function ToastContainer() {
  const toasts = useToastStore((s) => s.toasts);
  const removeToast = useToastStore((s) => s.removeToast);

  return (
    <div
      className="fixed bottom-4 right-4 z-50 flex flex-col-reverse gap-2"
      aria-label="Notifications"
    >
      <AnimatePresence mode="popLayout">
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onDismiss={removeToast} />
        ))}
      </AnimatePresence>
    </div>
  );
}

export { toast } from "@/lib/toastStore";
