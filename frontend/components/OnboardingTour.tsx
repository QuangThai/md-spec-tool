"use client";

import { useOnboardingStore, ONBOARDING_STEPS } from "@/lib/onboardingStore";
import { AnimatePresence, motion } from "framer-motion";
import {
  ArrowLeft,
  ArrowRight,
  BookOpen,
  CheckCircle2,
  Sparkles,
  X,
} from "lucide-react";
import { useEffect, useState, useCallback } from "react";

export function OnboardingTour() {
  const {
    hasSeenTour,
    isActive,
    currentStep,
    startTour,
    nextStep,
    prevStep,
    skipTour,
    completeTour,
  } = useOnboardingStore();

  const [targetRect, setTargetRect] = useState<DOMRect | null>(null);
  const [showWelcome, setShowWelcome] = useState(false);

  // Show welcome modal for first-time users
  useEffect(() => {
    if (!hasSeenTour && !isActive) {
      const timer = setTimeout(() => setShowWelcome(true), 1000);
      return () => clearTimeout(timer);
    }
  }, [hasSeenTour, isActive]);

  // Update target element position
  useEffect(() => {
    if (!isActive) return;

    const step = ONBOARDING_STEPS[currentStep];
    const updatePosition = () => {
      const element = document.querySelector(step.target);
      if (element) {
        const rect = element.getBoundingClientRect();
        setTargetRect(rect);
      } else {
        setTargetRect(null);
      }
    };

    updatePosition();
    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, { passive: true });

    // Re-check position on interval (for dynamic elements)
    const interval = setInterval(updatePosition, 500);

    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition);
      clearInterval(interval);
    };
  }, [isActive, currentStep]);

  const handleStartTour = useCallback(() => {
    setShowWelcome(false);
    startTour();
  }, [startTour]);

  const handleSkipWelcome = useCallback(() => {
    setShowWelcome(false);
    skipTour();
  }, [skipTour]);

  const step = ONBOARDING_STEPS[currentStep];
  const isLastStep = currentStep === ONBOARDING_STEPS.length - 1;
  const isFirstStep = currentStep === 0;

  // Calculate tooltip position
  const getTooltipStyle = () => {
    if (typeof window === "undefined") {
      return { top: "50%", left: "50%", transform: "translate(-50%, -50%)" };
    }

    const padding = 16;
    const baseWidth = 360;
    const tooltipWidth = Math.min(baseWidth, window.innerWidth - padding * 2);
    const tooltipHeight = 220;
    const isMobile = window.innerWidth < 640;

    if (!targetRect) {
      return {
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -50%)",
        width: `${tooltipWidth}px`,
      };
    }

    if (isMobile) {
      return {
        left: `${padding}px`,
        right: `${padding}px`,
        bottom: `${padding}px`,
        top: "auto",
        width: `calc(100vw - ${padding * 2}px)`,
        maxWidth: `${baseWidth}px`,
      };
    }

    let top = 0;
    let left = 0;

    switch (step.position) {
      case "top":
        top = targetRect.top - tooltipHeight - padding;
        left = targetRect.left + targetRect.width / 2 - tooltipWidth / 2;
        break;
      case "bottom":
        top = targetRect.bottom + padding;
        left = targetRect.left + targetRect.width / 2 - tooltipWidth / 2;
        break;
      case "left":
        top = targetRect.top + targetRect.height / 2 - tooltipHeight / 2;
        left = targetRect.left - tooltipWidth - padding;
        break;
      case "right":
        top = targetRect.top + targetRect.height / 2 - tooltipHeight / 2;
        left = targetRect.right + padding;
        break;
    }

    // Keep tooltip in viewport
    left = Math.max(padding, Math.min(left, window.innerWidth - tooltipWidth - padding));
    top = Math.max(padding, Math.min(top, window.innerHeight - tooltipHeight - padding));

    return { top: `${top}px`, left: `${left}px`, width: `${tooltipWidth}px` };
  };

  return (
    <>
      {/* Welcome Modal for First-Time Users */}
      <AnimatePresence>
        {showWelcome && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/60 backdrop-blur-sm z-100 flex items-center justify-center p-4"
          >
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.9, opacity: 0 }}
              className="bg-linear-to-br from-gray-900 via-black to-gray-900 border border-white/20 rounded-3xl shadow-2xl max-w-lg w-full overflow-hidden"
            >
              {/* Header with gradient */}
              <div className="relative p-8 pb-6 overflow-hidden">
                <div className="absolute inset-0 bg-linear-to-br from-accent-orange/20 via-transparent to-transparent" />
                <div className="absolute top-4 right-4">
                  <button
                    onClick={handleSkipWelcome}
                    className="p-2 rounded-lg hover:bg-white/10 transition-colors cursor-pointer text-white/50 hover:text-white"
                    aria-label="Close"
                  >
                    <X className="w-5 h-5" />
                  </button>
                </div>
                
                <div className="relative z-10 flex items-center gap-4 mb-4">
                  <div className="p-3 rounded-2xl bg-accent-orange/20 border border-accent-orange/30">
                    <Sparkles className="w-8 h-8 text-accent-orange" />
                  </div>
                  <div>
                    <h2 className="text-xl font-black text-white tracking-tight">
                      Welcome to MDFlow Studio
                    </h2>
                    <p className="text-sm text-white/60 mt-0.5">
                      First time here? Let me show you around.
                    </p>
                  </div>
                </div>
              </div>

              {/* Content */}
              <div className="px-8 pb-6">
                <p className="text-sm text-white/70 leading-relaxed mb-6">
                  MDFlow Studio transforms your spreadsheet data into structured, 
                  AI-ready Markdown specifications. Perfect for test plans, feature specs, 
                  and API documentation.
                </p>

                <div className="space-y-3 mb-6">
                  <div className="flex items-start gap-3 p-3 rounded-xl bg-white/5 border border-white/10">
                    <CheckCircle2 className="w-5 h-5 text-green-400 mt-0.5 shrink-0" />
                    <div>
                      <p className="text-sm font-semibold text-white">Paste or Upload</p>
                      <p className="text-xs text-white/50">Support for Excel, TSV, CSV, and Google Sheets</p>
                    </div>
                  </div>
                  <div className="flex items-start gap-3 p-3 rounded-xl bg-white/5 border border-white/10">
                    <CheckCircle2 className="w-5 h-5 text-green-400 mt-0.5 shrink-0" />
                    <div>
                      <p className="text-sm font-semibold text-white">Smart Detection</p>
                      <p className="text-xs text-white/50">Auto-detect columns and map to spec fields</p>
                    </div>
                  </div>
                  <div className="flex items-start gap-3 p-3 rounded-xl bg-white/5 border border-white/10">
                    <CheckCircle2 className="w-5 h-5 text-green-400 mt-0.5 shrink-0" />
                    <div>
                      <p className="text-sm font-semibold text-white">Multiple Templates</p>
                      <p className="text-xs text-white/50">Choose from 5 output formats for different use cases</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Actions */}
              <div className="px-8 block pb-8 lg:flex lg:gap-3">
                <button
                  onClick={handleSkipWelcome}
                  className="flex-1 w-full lg:w-auto px-4 py-3 text-sm font-bold uppercase tracking-wider rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white transition-all cursor-pointer mb-3 lg:mb-0"
                >
                  Skip
                </button>
                <button
                  onClick={handleStartTour}
                  className="flex-1 w-full lg:w-auto px-4 py-3 text-sm font-bold uppercase tracking-wider rounded-xl bg-accent-orange hover:bg-accent-orange/90 text-white shadow-lg shadow-accent-orange/25 transition-all cursor-pointer flex items-center justify-center gap-2"
                >
                  <BookOpen className="w-4 h-4" />
                  Take the Tour
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Tour Overlay */}
      <AnimatePresence>
        {isActive && (
          <>
            {/* Dimmed overlay with spotlight */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-90 pointer-events-none"
              style={{
                background: targetRect
                  ? `radial-gradient(ellipse ${targetRect.width + 60}px ${targetRect.height + 60}px at ${targetRect.left + targetRect.width / 2}px ${targetRect.top + targetRect.height / 2}px, transparent 0%, rgba(0,0,0,0.75) 100%)`
                  : "rgba(0,0,0,0.75)",
              }}
            />

            {/* Click blocker except for target */}
            <div
              className="fixed inset-0 z-91"
              onClick={(e) => {
                // Allow clicks on the target element
                if (targetRect) {
                  const x = e.clientX;
                  const y = e.clientY;
                  if (
                    x >= targetRect.left &&
                    x <= targetRect.right &&
                    y >= targetRect.top &&
                    y <= targetRect.bottom
                  ) {
                    return;
                  }
                }
                e.preventDefault();
                e.stopPropagation();
              }}
            />

            {/* Tooltip */}
            <motion.div
              key={currentStep}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              style={getTooltipStyle()}
              className="fixed z-92 max-h-[calc(100vh-32px)] overflow-y-auto onboarding-tooltip"
            >
              <div className="bg-linear-to-br from-gray-900 via-black to-gray-900 border border-white/20 rounded-2xl shadow-2xl overflow-hidden">
                {/* Progress bar */}
                <div className="h-1 bg-white/10">
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{
                      width: `${((currentStep + 1) / ONBOARDING_STEPS.length) * 100}%`,
                    }}
                    className="h-full bg-accent-orange"
                  />
                </div>

                <div className="p-5">
                  {/* Step indicator */}
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-[10px] font-bold uppercase tracking-wider text-accent-orange">
                      Step {currentStep + 1} of {ONBOARDING_STEPS.length}
                    </span>
                    <button
                      onClick={skipTour}
                      className="text-[10px] text-white/40 hover:text-white/70 cursor-pointer uppercase font-bold tracking-wider"
                    >
                      Skip tour
                    </button>
                  </div>

                  {/* Content */}
                  <h3 className="text-base font-black text-white mb-2">
                    {step.title}
                  </h3>
                  <p className="text-sm text-white/70 leading-relaxed mb-5">
                    {step.description}
                  </p>

                  {/* Navigation */}
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <button
                      onClick={prevStep}
                      disabled={isFirstStep}
                      className={`
                        flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg text-[11px] font-bold uppercase tracking-wider transition-all cursor-pointer w-full sm:w-auto
                        ${isFirstStep
                          ? "opacity-30 cursor-not-allowed text-white/50"
                          : "bg-white/5 hover:bg-white/10 text-white/70 hover:text-white"
                        }
                      `}
                    >
                      <ArrowLeft className="w-3.5 h-3.5" />
                      Back
                    </button>

                    <div className="flex justify-center gap-1.5 sm:justify-start">
                      {ONBOARDING_STEPS.map((_, i) => (
                        <div
                          key={i}
                          className={`
                            w-2 h-2 rounded-full transition-all
                            ${i === currentStep
                              ? "bg-accent-orange w-4"
                              : i < currentStep
                              ? "bg-accent-orange/50"
                              : "bg-white/20"
                            }
                          `}
                        />
                      ))}
                    </div>

                    <button
                      onClick={isLastStep ? completeTour : nextStep}
                      className="flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg text-[11px] font-bold uppercase tracking-wider bg-accent-orange hover:bg-accent-orange/90 text-white shadow-md shadow-accent-orange/25 transition-all cursor-pointer w-full sm:w-auto"
                    >
                      {isLastStep ? (
                        <>
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          Done
                        </>
                      ) : (
                        <>
                          Next
                          <ArrowRight className="w-3.5 h-3.5" />
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            </motion.div>

            {/* Target highlight ring */}
            {targetRect && (
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="fixed z-89 pointer-events-none"
                style={{
                  top: targetRect.top - 8,
                  left: targetRect.left - 8,
                  width: targetRect.width + 16,
                  height: targetRect.height + 16,
                  borderRadius: "16px",
                  border: "2px solid rgba(242,123,47,0.6)",
                  boxShadow: "0 0 0 4px rgba(242,123,47,0.15), 0 0 30px rgba(242,123,47,0.3)",
                }}
              />
            )}
          </>
        )}
      </AnimatePresence>
    </>
  );
}

// Button to restart tour (for settings/help)
export function RestartTourButton() {
  const { resetTour, startTour } = useOnboardingStore();

  const handleClick = () => {
    resetTour();
    setTimeout(startTour, 100);
  };

  return (
    <button
      onClick={handleClick}
      className="flex items-center gap-2 px-3 py-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-[10px] font-bold uppercase tracking-wider text-white/70 hover:text-white transition-all cursor-pointer"
    >
      <BookOpen className="w-3.5 h-3.5" />
      Restart Tour
    </button>
  );
}
