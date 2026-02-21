"use client";

import { ConversionFeedback } from "@/components/ConversionFeedback";
import { ConversionRecord, OutputFormat } from "@/lib/types";
import { AnimatePresence } from "framer-motion";
import dynamic from "next/dynamic";
import HistoryModal, { KeyboardShortcutsTooltip } from "../HistoryModal";
import { CommandPalette } from "../CommandPalette";
import { ToastContainer } from "../ui/Toast";
import type { DiffResponse } from "@/lib/diffTypes";

const TemplateEditor = dynamic(
  () => import("../TemplateEditor").then((mod) => mod.TemplateEditor),
  { ssr: false }
);

const ValidationConfigurator = dynamic(
  () =>
    import("../ValidationConfigurator").then(
      (mod) => mod.ValidationConfigurator
    ),
  { ssr: false }
);

const DiffModal = dynamic(() => import("./DiffModal").then((mod) => mod.DiffModal), {
  ssr: false,
});

export interface WorkbenchOverlaysProps {
  showDiff: boolean;
  currentDiff: DiffResponse | null;
  onCloseDiff: () => void;
  showHistory: boolean;
  history: ConversionRecord[];
  onCloseHistory: () => void;
  onSelectHistory: (record: ConversionRecord) => void;
  showValidationConfigurator: boolean;
  onCloseValidationConfigurator: () => void;
  showTemplateEditor: boolean;
  onCloseTemplateEditor: () => void;
  currentSampleData?: string;
  showCommandPalette: boolean;
  onShowCommandPaletteChange: (open: boolean) => void;
  onConvert: () => Promise<void>;
  onCopy: () => void;
  onExport: () => void;
  onTogglePreview: () => void;
  onShowHistory: () => void;
  onOpenTemplateEditor: () => void;
  onOpenValidation: () => void;
  templates: Array<string | { name: string; description?: string }>;
  currentTemplate: OutputFormat;
  onSelectTemplate: (format: OutputFormat) => void;
  hasOutput: boolean;
  showFeedback: boolean;
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  onDismissFeedback: () => void;
}

export function WorkbenchOverlays({
  showDiff,
  currentDiff,
  onCloseDiff,
  showHistory,
  history,
  onCloseHistory,
  onSelectHistory,
  showValidationConfigurator,
  onCloseValidationConfigurator,
  showTemplateEditor,
  onCloseTemplateEditor,
  currentSampleData,
  showCommandPalette,
  onShowCommandPaletteChange,
  onConvert,
  onCopy,
  onExport,
  onTogglePreview,
  onShowHistory,
  onOpenTemplateEditor,
  onOpenValidation,
  templates,
  currentTemplate,
  onSelectTemplate,
  hasOutput,
  showFeedback,
  inputSource,
  onDismissFeedback,
}: WorkbenchOverlaysProps) {
  return (
    <>
      <DiffModal showDiff={showDiff} currentDiff={currentDiff} onClose={onCloseDiff} />

      <AnimatePresence>
        {showHistory ? (
          <HistoryModal
            history={history}
            onClose={onCloseHistory}
            onSelect={onSelectHistory}
          />
        ) : null}
      </AnimatePresence>

      <ValidationConfigurator
        open={showValidationConfigurator}
        onClose={onCloseValidationConfigurator}
        showValidateAction={true}
      />

      <TemplateEditor
        isOpen={showTemplateEditor}
        onClose={onCloseTemplateEditor}
        currentSampleData={currentSampleData}
      />

      <div className="fixed bottom-4 right-4 z-40">
        <KeyboardShortcutsTooltip />
      </div>

      <CommandPalette
        open={showCommandPalette}
        onOpenChange={onShowCommandPaletteChange}
        onConvert={onConvert}
        onCopy={onCopy}
        onExport={onExport}
        onTogglePreview={onTogglePreview}
        onShowHistory={onShowHistory}
        onOpenTemplateEditor={onOpenTemplateEditor}
        onOpenValidation={onOpenValidation}
        templates={templates}
        currentTemplate={currentTemplate}
        onSelectTemplate={onSelectTemplate}
        hasOutput={hasOutput}
      />

      <ConversionFeedback
        visible={showFeedback}
        inputSource={inputSource}
        onDismiss={onDismissFeedback}
      />

      <ToastContainer />
    </>
  );
}
