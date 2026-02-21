"use client";

import { useWorkbenchController } from "@/hooks/useWorkbenchController";
import { motion } from "framer-motion";
import { OnboardingTour } from "./OnboardingTour";
import { SourcePanel } from "./workbench/SourcePanel";
import { OutputPanel } from "./workbench/OutputPanel";
import { WorkbenchOverlays } from "./workbench/WorkbenchOverlays";

const stagger = {
  container: {
    animate: { transition: { staggerChildren: 0.05, delayChildren: 0.08 } },
  },
  item: {
    initial: { opacity: 0, y: 12 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.35, ease: [0.16, 1, 0.3, 1] },
  },
};

export default function MDFlowWorkbench() {
  const { sourcePanelProps, outputPanelProps, overlaysProps } =
    useWorkbenchController();

  return (
    <motion.div
      variants={stagger.container}
      initial="initial"
      animate="animate"
      className="flex flex-col gap-3 sm:gap-4 relative h-[calc(100vh-6rem)] sm:h-[calc(100vh-7rem)] lg:h-[calc(100vh-8rem)]"
    >
      {/* Onboarding Tour */}
      <OnboardingTour />

      {/* Main workspace: optimized for immediate visibility */}
      <div
        className="grid grid-cols-1 lg:grid-cols-2 gap-3 sm:gap-4 lg:gap-5 items-stretch flex-1 min-h-0"
        data-tour="welcome"
      >
        {/* Left: Source Panel */}
        <SourcePanel {...sourcePanelProps} />

        {/* Right: Output Panel */}
        <OutputPanel {...outputPanelProps} />
      </div>

      <WorkbenchOverlays {...overlaysProps} />
    </motion.div>
  );
}
