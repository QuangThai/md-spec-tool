"use client";

import { AnimatePresence, motion } from "framer-motion";
import { createPortal } from "react-dom";
import { useEffect, useLayoutEffect, useRef, useState } from "react";

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
  const [mounted, setMounted] = useState(false);
  const [tooltipStyle, setTooltipStyle] = useState<React.CSSProperties>({});
  const tooltipId = useRef(`tooltip-${Math.random().toString(36).slice(2, 11)}`);
  const triggerRef = useRef<HTMLDivElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  // Focus trap: when tooltip shows, manage focus
  useEffect(() => {
    setMounted(true);
  }, []);

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

  useLayoutEffect(() => {
    if (!show) return;

    const updatePosition = () => {
      if (!triggerRef.current || !tooltipRef.current) return;

      const rect = triggerRef.current.getBoundingClientRect();
      const tooltipRect = tooltipRef.current.getBoundingClientRect();
      const padding = 8;
      const offset = 6;

      let top =
        position === "bottom"
          ? rect.bottom + offset
          : rect.top - offset - tooltipRect.height;
      let left = rect.left + rect.width / 2 - tooltipRect.width / 2;

      left = Math.max(
        padding,
        Math.min(left, window.innerWidth - tooltipRect.width - padding),
      );
      top = Math.max(
        padding,
        Math.min(top, window.innerHeight - tooltipRect.height - padding),
      );

      setTooltipStyle({ top: `${top}px`, left: `${left}px` });
    };

    updatePosition();
    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, true);
    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition, true);
    };
  }, [show, position, content]);

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
      {mounted &&
        createPortal(
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
                style={tooltipStyle}
                className="fixed z-200 px-3 py-2 rounded-md bg-black/95 backdrop-blur-sm border border-white/20 shadow-xl text-[9px] font-medium text-white whitespace-nowrap pointer-events-none"
                role="tooltip"
                aria-hidden={!show}
              >
                {content}
              </motion.div>
            )}
          </AnimatePresence>,
          document.body,
        )}
    </div>
  );
}
