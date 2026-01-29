"use client";

import { cn } from "@/lib/utils";
import { AnimatePresence, motion } from "framer-motion";
import { Check, ChevronDown } from "lucide-react";
import * as React from "react";

interface SelectProps {
  value: string;
  onValueChange: (value: string) => void;
  options: { label: string; value: string }[];
  placeholder?: string;
  className?: string;
  side?: "top" | "bottom";
}

export function Select({
  value,
  onValueChange,
  options,
  placeholder,
  className,
  side = "bottom",
}: SelectProps) {
  const [isOpen, setIsOpen] = React.useState(false);
  const containerRef = React.useRef<HTMLDivElement>(null);

  const selectedOption = options.find((opt) => opt.value === value);
  const isUp = side === "top";

  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div className={cn("relative w-full", className)} ref={containerRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "flex h-12 w-full items-center justify-between rounded-xl border border-white/10 bg-black/40 px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-white transition-all duration-300 outline-none hover:bg-white/5 focus:border-accent-orange/50 focus:ring-4 focus:ring-accent-orange/10",
          isOpen &&
            "border-accent-orange/50 bg-white/5 ring-4 ring-accent-orange/10",
        )}
      >
        <span className={cn(!selectedOption && "text-muted")}>
          {selectedOption
            ? selectedOption.label
            : placeholder || "Select option"}
        </span>
        <ChevronDown
          className={cn(
            "h-4 w-4 text-muted transition-transform duration-300",
            isOpen && "rotate-180 text-accent-orange",
          )}
        />
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: isUp ? -10 : 10, scale: 0.95 }}
            animate={{ opacity: 1, y: isUp ? -5 : 5, scale: 1 }}
            exit={{ opacity: 0, y: isUp ? -10 : 10, scale: 0.95 }}
            transition={{ duration: 0.2, ease: "easeOut" }}
            className={cn(
              "absolute z-50 w-full overflow-hidden rounded-xl border border-white/10 bg-[#0f0f0f] p-1 shadow-2xl backdrop-blur-3xl shadow-black/80",
              isUp ? "bottom-full mb-2" : "top-full mt-2",
            )}
          >
            <div className="max-h-[300px] overflow-auto custom-scrollbar">
              {options.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => {
                    onValueChange(option.value);
                    setIsOpen(false);
                  }}
                  className={cn(
                    "relative flex w-full items-center justify-between rounded-lg px-5 py-3 text-left text-[10px] font-bold uppercase tracking-widest transition-colors duration-200 hover:bg-white/10",
                    option.value === value
                      ? "bg-accent-orange/10 text-accent-orange"
                      : "text-white/70 hover:text-white",
                  )}
                >
                  <span>{option.label}</span>
                  {option.value === value && (
                    <motion.div
                      initial={{ scale: 0.5, opacity: 0 }}
                      animate={{ scale: 1, opacity: 1 }}
                    >
                      <Check className="h-3.5 w-3.5" />
                    </motion.div>
                  )}
                </button>
              ))}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
