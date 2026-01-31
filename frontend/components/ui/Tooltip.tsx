"use client";

import { AnimatePresence, motion } from "framer-motion";
import { useEffect, useRef, useState } from "react";

interface TooltipProps {
  content: string;
  children: React.ReactNode;
  position?: "top" | "bottom";
  "aria-label"?: string;
  /** Allow keyboard interaction (focus to show tooltip) */
  interactive?: boolean;
}

export function Tooltip({
  content,
  children,
  position = "bottom",
  "aria-label": ariaLabel,
  interactive = false,
}: TooltipProps) {
  const [show, setShow] = useState(false);
  const tooltipId = useRef(`tooltip-${Math.random().toString(36).slice(2, 11)}`);
  const triggerRef = useRef<HTMLDivElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  // Focus trap: when tooltip shows, manage focus
  useEffect(() => {
    if (!show || !interactive) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        setShow(false);
        triggerRef.current?.focus();
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [show, interactive]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!interactive) return;

    switch (e.key) {
      case "Enter":
      case " ":
        e.preventDefault();
        setShow(!show);
        break;
      case "Escape":
        e.preventDefault();
        setShow(false);
        break;
      default:
        break;
    }
  };

  return (
    <div
      ref={triggerRef}
      className="relative inline-block"
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
      onKeyDown={handleKeyDown}
      onFocus={() => interactive && setShow(true)}
      onBlur={() => interactive && setShow(false)}
      role={interactive ? "button" : undefined}
      tabIndex={interactive ? 0 : undefined}
      aria-label={interactive ? ariaLabel || content : undefined}
      aria-describedby={show ? tooltipId.current : undefined}
    >
      {children}
      <AnimatePresence>
        {show && (
          <motion.div
            ref={tooltipRef}
            id={tooltipId.current}
            initial={{
              opacity: 0,
              y: position === "bottom" ? -4 : 4,
              scale: 0.95,
            }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{
              opacity: 0,
              y: position === "bottom" ? -4 : 4,
              scale: 0.95,
            }}
            transition={{ duration: 0.15 }}
            className={`
              absolute left-1/2 -translate-x-1/2 z-50 px-3 py-2 rounded-md
              bg-black/95 backdrop-blur-sm border border-white/20 shadow-xl
              text-[9px] font-medium text-white whitespace-nowrap
              pointer-events-none
              ${position === "bottom" ? "top-full mt-1.5" : "bottom-full mb-1.5"}
            `}
            role="tooltip"
            aria-hidden={!show}
          >
            {content}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
