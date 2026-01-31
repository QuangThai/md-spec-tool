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
  "aria-label"?: string;
}

export function Select({
  value,
  onValueChange,
  options,
  placeholder,
  className,
  side = "bottom",
  "aria-label": ariaLabel,
}: SelectProps) {
  const [isOpen, setIsOpen] = React.useState(false);
  const [highlightedIndex, setHighlightedIndex] = React.useState(-1);
  const containerRef = React.useRef<HTMLDivElement>(null);
  const listRef = React.useRef<HTMLDivElement>(null);
  const buttonRef = React.useRef<HTMLButtonElement>(null);

  const selectedOption = options.find((opt) => opt.value === value);
  const isUp = side === "top";

  // Handle keyboard navigation
  React.useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      switch (event.key) {
        case "ArrowDown":
          event.preventDefault();
          setHighlightedIndex((prev) =>
            prev < options.length - 1 ? prev + 1 : 0
          );
          break;
        case "ArrowUp":
          event.preventDefault();
          setHighlightedIndex((prev) =>
            prev > 0 ? prev - 1 : options.length - 1
          );
          break;
        case "Enter":
          event.preventDefault();
          if (highlightedIndex >= 0) {
            onValueChange(options[highlightedIndex].value);
            setIsOpen(false);
            setHighlightedIndex(-1);
          }
          break;
        case "Escape":
          event.preventDefault();
          setIsOpen(false);
          setHighlightedIndex(-1);
          buttonRef.current?.focus();
          break;
        case "Tab":
          setIsOpen(false);
          setHighlightedIndex(-1);
          break;
        default:
          break;
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, highlightedIndex, options, onValueChange]);

  // Handle click outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
        setHighlightedIndex(-1);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Focus highlighted item when it changes
  React.useEffect(() => {
    if (highlightedIndex >= 0 && listRef.current) {
      const items = listRef.current.querySelectorAll("button");
      items[highlightedIndex]?.scrollIntoView({ block: "nearest" });
    }
  }, [highlightedIndex]);

  return (
    <div className={cn("relative w-full", className)} ref={containerRef}>
      <button
        ref={buttonRef}
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            setIsOpen(true);
          }
        }}
        className={cn(
          "flex h-12 w-full items-center justify-between rounded-xl border border-white/10 bg-black/40 px-6 py-4 text-[11px] font-bold uppercase tracking-wider text-white transition-all duration-300 outline-none hover:bg-white/5 focus:border-accent-orange/50 focus:ring-4 focus:ring-accent-orange/10",
          isOpen &&
            "border-accent-orange/50 bg-white/5 ring-4 ring-accent-orange/10"
        )}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        aria-label={
          ariaLabel ||
          `Select ${placeholder || "option"}: ${selectedOption?.label || placeholder || "Select option"}`
        }
      >
        <span className={cn(!selectedOption && "text-muted")}>
          {selectedOption
            ? selectedOption.label
            : placeholder || "Select option"}
        </span>
        <ChevronDown
          className={cn(
            "h-4 w-4 text-muted transition-transform duration-300",
            isOpen && "rotate-180 text-accent-orange"
          )}
          aria-hidden="true"
        />
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            ref={listRef}
            initial={{ opacity: 0, y: isUp ? -10 : 10, scale: 0.95 }}
            animate={{ opacity: 1, y: isUp ? -5 : 5, scale: 1 }}
            exit={{ opacity: 0, y: isUp ? -10 : 10, scale: 0.95 }}
            transition={{ duration: 0.2, ease: "easeOut" }}
            className={cn(
              "absolute z-50 w-full overflow-hidden rounded-xl border border-white/10 bg-[#0f0f0f] p-1 shadow-2xl backdrop-blur-3xl shadow-black/80",
              isUp ? "bottom-full mb-2" : "top-full mt-2"
            )}
            role="listbox"
            aria-label={ariaLabel || placeholder || "Select options"}
          >
            <div className="max-h-[300px] overflow-auto custom-scrollbar">
              {options.map((option, index) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => {
                    onValueChange(option.value);
                    setIsOpen(false);
                    setHighlightedIndex(-1);
                  }}
                  onMouseEnter={() => setHighlightedIndex(index)}
                  onMouseLeave={() => setHighlightedIndex(-1)}
                  className={cn(
                    "relative flex w-full items-center justify-between rounded-lg px-5 py-3 text-left text-[10px] font-bold uppercase tracking-widest transition-colors duration-200 hover:bg-white/10 focus:outline-none focus:ring-2 focus:ring-accent-orange/50",
                    highlightedIndex === index && "bg-white/10",
                    option.value === value
                      ? "bg-accent-orange/10 text-accent-orange"
                      : "text-white/70 hover:text-white"
                  )}
                  role="option"
                  aria-selected={option.value === value}
                >
                  <span>{option.label}</span>
                  {option.value === value && (
                    <motion.div
                      initial={{ scale: 0.5, opacity: 0 }}
                      animate={{ scale: 1, opacity: 1 }}
                      aria-hidden="true"
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
