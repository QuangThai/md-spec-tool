"use client";

import { AnimatePresence, motion } from "framer-motion";
import { useState } from "react";

interface TooltipProps {
  content: string;
  children: React.ReactNode;
  position?: "top" | "bottom";
}

export function Tooltip({ content, children, position = "bottom" }: TooltipProps) {
  const [show, setShow] = useState(false);

  return (
    <div
      className="relative"
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
    >
      {children}
      <AnimatePresence>
        {show && (
          <motion.div
            initial={{ opacity: 0, y: position === "bottom" ? -4 : 4, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: position === "bottom" ? -4 : 4, scale: 0.95 }}
            transition={{ duration: 0.15 }}
            className={`
              absolute left-1/2 -translate-x-1/2 z-50 px-2 py-1 rounded-md
              bg-black/95 backdrop-blur-sm border border-white/10 shadow-xl
              text-[9px] font-medium text-white/90 whitespace-nowrap
              pointer-events-none
              ${position === "bottom" ? "top-full mt-1.5" : "bottom-full mb-1.5"}
            `}
          >
            {content}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
