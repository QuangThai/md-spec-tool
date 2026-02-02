"use client";

import { cn } from "@/lib/utils";
import { Check, ChevronDown } from "lucide-react";
import * as SelectPrimitive from "@radix-ui/react-select";

interface SelectProps {
  value: string;
  onValueChange: (value: string) => void;
  options: readonly { label: string; value: string }[];
  placeholder?: string;
  className?: string;
  side?: "top" | "bottom";
  size?: "default" | "compact";
  "aria-label"?: string;
}

export function Select({
  value,
  onValueChange,
  options,
  placeholder,
  className,
  side = "bottom",
  size = "default",
  "aria-label": ariaLabel,
}: SelectProps) {
  const selectedOption = options.find((opt) => opt.value === value);
  const isCompact = size === "compact";

  return (
    <SelectPrimitive.Root value={value} onValueChange={onValueChange}>
      <SelectPrimitive.Trigger
        className={cn(
          "inline-flex min-w-[120px] items-center justify-between border border-white/10 bg-black/40 text-white transition-all duration-200 outline-none hover:bg-white/5 focus:border-accent-orange/50 focus:ring-4 focus:ring-accent-orange/10",
          isCompact
            ? "h-7 rounded-lg px-2.5 text-[10px] font-semibold tracking-normal"
            : "h-12 rounded-xl px-6 py-4 text-[11px] font-bold uppercase tracking-wider",
          className
        )}
        aria-label={
          ariaLabel ||
          `Select ${placeholder || "option"}: ${selectedOption?.label || placeholder || "Select option"}`
        }
      >
        <SelectPrimitive.Value
          placeholder={placeholder || "Select option"}
          className={cn(!selectedOption && "text-muted")}
        />
        <SelectPrimitive.Icon>
          <ChevronDown
            className={cn(
              isCompact ? "h-3 w-3 text-muted" : "h-4 w-4 text-muted"
            )}
            aria-hidden="true"
          />
        </SelectPrimitive.Icon>
      </SelectPrimitive.Trigger>

      <SelectPrimitive.Portal>
        <SelectPrimitive.Content
          side={side}
          sideOffset={6}
          align="start"
          position="popper"
          className={cn(
            "z-100 w-(--radix-select-trigger-width) overflow-hidden border border-white/10 bg-[#0f0f0f] p-1 shadow-2xl backdrop-blur-3xl shadow-black/80",
            isCompact ? "rounded-lg" : "rounded-xl"
          )}
        >
          <SelectPrimitive.Viewport className="max-h-[300px] custom-scrollbar">
            {options.map((option) => (
              <SelectPrimitive.Item
                key={option.value}
                value={option.value}
                className={cn(
                  "relative flex w-full cursor-pointer items-center justify-between rounded-md text-left transition-colors duration-150 outline-none data-highlighted:bg-white/10 data-highlighted:text-white",
                  isCompact
                    ? "px-3 py-2 text-[10px] font-semibold tracking-normal"
                    : "px-5 py-3 text-[10px] font-bold uppercase tracking-widest",
                  "text-white/70 data-[state=checked]:bg-accent-orange/10 data-[state=checked]:text-accent-orange"
                )}
              >
                <SelectPrimitive.ItemText>{option.label}</SelectPrimitive.ItemText>
                <SelectPrimitive.ItemIndicator>
                  <Check className="h-3.5 w-3.5" />
                </SelectPrimitive.ItemIndicator>
              </SelectPrimitive.Item>
            ))}
          </SelectPrimitive.Viewport>
        </SelectPrimitive.Content>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  );
}
