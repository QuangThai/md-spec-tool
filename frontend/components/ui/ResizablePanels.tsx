"use client";

import { useState, useRef, useCallback, useEffect, ReactNode } from "react";
import { motion } from "framer-motion";
import { GripVertical } from "lucide-react";

interface ResizablePanelsProps {
  leftPanel: ReactNode;
  rightPanel: ReactNode;
  defaultLeftWidth?: number; // percentage
  minLeftWidth?: number; // percentage
  maxLeftWidth?: number; // percentage
  className?: string;
}

export function ResizablePanels({
  leftPanel,
  rightPanel,
  defaultLeftWidth = 55,
  minLeftWidth = 30,
  maxLeftWidth = 70,
  className = "",
}: ResizablePanelsProps) {
  const [leftWidth, setLeftWidth] = useState(defaultLeftWidth);
  const [isDragging, setIsDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !containerRef.current) return;

      const containerRect = containerRef.current.getBoundingClientRect();
      const newWidth = ((e.clientX - containerRect.left) / containerRect.width) * 100;
      
      // Clamp to min/max
      const clampedWidth = Math.min(Math.max(newWidth, minLeftWidth), maxLeftWidth);
      setLeftWidth(clampedWidth);
    },
    [isDragging, minLeftWidth, maxLeftWidth]
  );

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  useEffect(() => {
    if (isDragging) {
      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);
      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";
    }

    return () => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    };
  }, [isDragging, handleMouseMove, handleMouseUp]);

  return (
    <div
      ref={containerRef}
      className={`flex h-full w-full ${className}`}
    >
      {/* Left Panel */}
      <div
        className="h-full overflow-hidden flex flex-col min-w-0"
        style={{ width: `${leftWidth}%` }}
      >
        {leftPanel}
      </div>

      {/* Resize Handle */}
      <div
        onMouseDown={handleMouseDown}
        className={`
          relative shrink-0 w-2 cursor-col-resize group
          flex items-center justify-center
          transition-colors duration-150
          ${isDragging ? "bg-accent-orange/30" : "hover:bg-white/10"}
        `}
      >
        <motion.div
          animate={{
            opacity: isDragging ? 1 : 0.3,
            scale: isDragging ? 1.2 : 1,
          }}
          className="absolute inset-y-0 w-1 left-1/2 -translate-x-1/2 rounded-full bg-white/20 group-hover:bg-accent-orange/50 transition-colors"
        />
        <div className="absolute top-1/2 -translate-y-1/2 p-1 rounded bg-black/50 border border-white/10 opacity-0 group-hover:opacity-100 transition-opacity">
          <GripVertical className="w-3 h-3 text-white/60" />
        </div>
      </div>

      {/* Right Panel */}
      <div
        className="h-full overflow-hidden flex flex-col min-w-0 flex-1"
      >
        {rightPanel}
      </div>
    </div>
  );
}
