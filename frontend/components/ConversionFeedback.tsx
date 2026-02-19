"use client";

import { emitTelemetryEvent } from "@/lib/telemetry";
import { AnimatePresence, motion } from "framer-motion";
import { X } from "lucide-react";
import { useCallback, useRef, useState } from "react";

interface ConversionFeedbackProps {
  visible: boolean;
  inputSource?: string;
  onDismiss: () => void;
}

type Rating = "okay" | "good" | "great";

const RATINGS: { value: Rating; emoji: string; label: string }[] = [
  { value: "okay", emoji: "😐", label: "Okay" },
  { value: "good", emoji: "😊", label: "Good" },
  { value: "great", emoji: "🤩", label: "Great" },
];

const SPRING_TRANSITION = { type: "spring", damping: 25, stiffness: 300 } as const;

export function ConversionFeedback({
  visible,
  inputSource,
  onDismiss,
}: ConversionFeedbackProps) {
  const [selectedRating, setSelectedRating] = useState<Rating | null>(null);
  const [comment, setComment] = useState("");
  const shownRef = useRef(false);

  const handleDismiss = useCallback(() => {
    emitTelemetryEvent("feedback_dismissed", { status: "cancel" });
    onDismiss();
  }, [onDismiss]);

  const handleSubmit = useCallback(() => {
    if (!selectedRating) return;
    emitTelemetryEvent("feedback_submitted", {
      status: "success",
      rating: selectedRating,
      has_comment: comment.length > 0,
      input_source: inputSource as "paste" | "xlsx" | "gsheet" | "tsv" | undefined,
    });
    onDismiss();
  }, [selectedRating, comment, inputSource, onDismiss]);

  // Only show once per session
  if (visible && !shownRef.current) {
    shownRef.current = true;
  }
  const shouldRender = visible && shownRef.current;

  return (
    <AnimatePresence>
      {shouldRender ? (
        <motion.div
          key="conversion-feedback"
          initial={{ opacity: 0, y: 40, scale: 0.95 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 20, scale: 0.95 }}
          transition={SPRING_TRANSITION}
          className="fixed bottom-6 right-6 z-50 w-80 rounded-2xl border border-white/10 bg-linear-to-br from-white/8 to-white/3 p-5 shadow-[0_18px_60px_-35px_rgba(242,123,47,0.4)] backdrop-blur-xl"
        >
          {/* Close button */}
          <button
            onClick={handleDismiss}
            className="absolute top-3 right-3 rounded-lg p-1 text-white/40 transition-colors hover:bg-white/10 hover:text-white/70"
          >
            <X size={16} />
          </button>

          {/* Title */}
          <p className="mb-3 text-sm font-medium text-white/90">
            How was this conversion?
          </p>

          {/* Rating buttons */}
          <div className="mb-4 flex gap-2">
            {RATINGS.map(({ value, emoji, label }) => (
              <button
                key={value}
                onClick={() => setSelectedRating(value)}
                className={`flex-1 rounded-full border px-3 py-2 text-sm transition-all ${
                  selectedRating === value
                    ? "border-orange-400/40 bg-accent-orange/15 text-orange-100 shadow-[0_0_12px_-3px_rgba(242,123,47,0.35)]"
                    : "border-white/10 bg-white/4 text-white/70 hover:border-white/20 hover:bg-white/8"
                }`}
              >
                <span className="mr-1">{emoji}</span>
                {label}
              </button>
            ))}
          </div>

          {/* Comment area (shown after rating) */}
          {selectedRating ? (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              transition={{ duration: 0.2 }}
            >
              <textarea
                value={comment}
                onChange={(e) => setComment(e.target.value.slice(0, 200))}
                placeholder="Any additional feedback? (optional)"
                rows={2}
                className="mb-3 w-full resize-none rounded-xl border border-white/10 bg-white/4 px-3 py-2 text-sm text-white/90 placeholder-white/30 outline-none transition-colors focus:border-white/20 focus:bg-white/6"
              />
              <div className="flex items-center justify-between">
                <span className="text-xs text-white/30">
                  {comment.length}/200
                </span>
                <button
                  onClick={handleSubmit}
                  className="rounded-lg bg-accent-orange/15 px-4 py-1.5 text-sm font-medium text-orange-100 transition-all hover:bg-accent-orange/25 hover:shadow-[0_0_16px_-4px_rgba(242,123,47,0.4)]"
                >
                  Submit
                </button>
              </div>
            </motion.div>
          ) : null}
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}
