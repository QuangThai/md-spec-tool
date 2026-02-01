"use client";

import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

interface SkeletonProps {
  variant?: "text" | "title" | "button" | "card" | "table-row" | "avatar";
  className?: string;
  count?: number;
  animate?: boolean;
}

const variantStyles: Record<NonNullable<SkeletonProps["variant"]>, string> = {
  text: "h-4 w-full rounded",
  title: "h-6 w-3/4 rounded",
  button: "h-10 w-24 rounded-lg",
  card: "h-32 w-full rounded-xl",
  "table-row": "h-12 w-full rounded-lg",
  avatar: "h-10 w-10 rounded-full",
};

function ShimmerOverlay() {
  return (
    <motion.div
      className="absolute inset-0 -translate-x-full"
      style={{
        background:
          "linear-gradient(90deg, transparent 0%, rgba(255,255,255,0.1) 50%, transparent 100%)",
      }}
      animate={{ translateX: ["100%", "-100%"] }}
      transition={{
        duration: 1.5,
        ease: [0.16, 1, 0.3, 1],
        repeat: Infinity,
        repeatDelay: 0.5,
      }}
    />
  );
}

export function Skeleton({
  variant = "text",
  className,
  count = 1,
  animate = true,
}: SkeletonProps) {
  const items = Array.from({ length: count }, (_, i) => i);

  return (
    <>
      {items.map((i) => (
        <div
          key={i}
          className={cn(
            "relative overflow-hidden bg-white/5",
            variantStyles[variant],
            className
          )}
        >
          {animate && <ShimmerOverlay />}
        </div>
      ))}
    </>
  );
}

interface TableSkeletonProps {
  rows?: number;
  className?: string;
  animate?: boolean;
}

export function TableSkeleton({
  rows = 5,
  className,
  animate = true,
}: TableSkeletonProps) {
  return (
    <div className={cn("space-y-2", className)}>
      {Array.from({ length: rows }, (_, i) => (
        <div
          key={i}
          className="relative overflow-hidden bg-white/5 h-12 w-full rounded-lg"
        >
          {animate && <ShimmerOverlay />}
          <div className="absolute inset-0 flex items-center gap-4 px-4">
            <div className="h-4 w-8 bg-white/5 rounded" />
            <div className="h-4 flex-1 bg-white/5 rounded" />
            <div className="h-4 w-24 bg-white/5 rounded" />
            <div className="h-4 w-16 bg-white/5 rounded" />
          </div>
        </div>
      ))}
    </div>
  );
}

interface OutputSkeletonProps {
  lines?: number;
  className?: string;
  animate?: boolean;
}

export function OutputSkeleton({
  lines = 8,
  className,
  animate = true,
}: OutputSkeletonProps) {
  const lineWidths = [
    "w-3/4",
    "w-full",
    "w-5/6",
    "w-2/3",
    "w-full",
    "w-4/5",
    "w-1/2",
    "w-5/6",
    "w-3/4",
    "w-full",
  ];

  return (
    <div className={cn("space-y-3 p-4", className)}>
      <div className="relative overflow-hidden bg-white/5 h-6 w-1/3 rounded mb-4">
        {animate && <ShimmerOverlay />}
      </div>
      {Array.from({ length: lines }, (_, i) => (
        <div
          key={i}
          className={cn(
            "relative overflow-hidden bg-white/5 h-4 rounded",
            lineWidths[i % lineWidths.length]
          )}
        >
          {animate && <ShimmerOverlay />}
        </div>
      ))}
    </div>
  );
}
