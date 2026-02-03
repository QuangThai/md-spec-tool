"use client";

import { useGoogleAuth } from "@/hooks/useGoogleAuth";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import {
  useAISuggestionsMutation,
  useConvertGoogleSheetMutation,
  useConvertPasteMutation,
  useConvertTSVMutation,
  useConvertXLSXMutation,
  useDiffMDFlowMutation,
  useGetGoogleSheetSheetsMutation,
  useGetXLSXSheetsMutation,
  useMDFlowTemplatesQuery,
  usePreviewGoogleSheetQuery,
  usePreviewPasteQuery,
  usePreviewTSVQuery,
  usePreviewXLSXQuery,
} from "@/lib/mdflowQueries";
import {
  useHistoryStore,
  useMDFlowStore,
  type MDFlowStore,
} from "@/lib/mdflowStore";
import { createShare } from "@/lib/shareApi";
import { ConversionRecord } from "@/lib/types";
import { useBodyScrollLock } from "@/lib/useBodyScrollLock";
import { useKeyboardShortcuts } from "@/lib/useKeyboardShortcuts";
import { AnimatePresence, motion } from "framer-motion";
import {
  AlertCircle,
  Check,
  Copy,
  Download,
  Eye,
  EyeOff,
  FileCode,
  FileSpreadsheet,
  FileText,
  GitCompare,
  History,
  KeyRound,
  Link2,
  RefreshCcw,
  Save,
  Share2,
  ShieldCheck,
  Terminal,
  Zap,
} from "lucide-react";
import dynamic from "next/dynamic";
import { useCallback, useEffect, useRef, useState } from "react";
import { useShallow } from "zustand/react/shallow";
import { CommandPalette } from "./CommandPalette";
import HistoryModal, { KeyboardShortcutsTooltip } from "./HistoryModal";
import { OnboardingTour } from "./OnboardingTour";
import { PreviewTable } from "./PreviewTable";
import { ShareButton } from "./ShareButton";
import { TechnicalAnalysis } from "./TechnicalAnalysis";
import { TemplateCards } from "./TemplateCards";
import { Select } from "./ui/Select";
import { OutputSkeleton } from "./ui/Skeleton";
import { toast, ToastContainer } from "./ui/Toast";
import { Tooltip } from "./ui/Tooltip";

const DiffViewer = dynamic(
  () => import("./DiffViewer").then((mod) => mod.DiffViewer),
  { ssr: false }
);

const TemplateEditor = dynamic(
  () => import("./TemplateEditor").then((mod) => mod.TemplateEditor),
  { ssr: false }
);

const ValidationConfigurator = dynamic(
  () =>
    import("./ValidationConfigurator").then(
      (mod) => mod.ValidationConfigurator
    ),
  { ssr: false }
);

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
  const {
    mode,
    pasteText,
    file,
    sheets,
    selectedSheet,
    gsheetTabs,
    selectedGid,
    template,
    format,
    mdflowOutput,
    warnings,
    meta,
    loading,
    error,
    preview,
    previewLoading,
    showPreview,
    columnOverrides,
    aiSuggestions,
    aiSuggestionsLoading,
    aiSuggestionsError,
    aiConfigured,
  } = useMDFlowStore(
    useShallow((state: MDFlowStore) => ({
      mode: state.mode,
      pasteText: state.pasteText,
      file: state.file,
      sheets: state.sheets,
      selectedSheet: state.selectedSheet,
      gsheetTabs: state.gsheetTabs,
      selectedGid: state.selectedGid,
      template: state.template,
      format: state.format,
      mdflowOutput: state.mdflowOutput,
      warnings: state.warnings,
      meta: state.meta,
      loading: state.loading,
      error: state.error,
      preview: state.preview,
      previewLoading: state.previewLoading,
      showPreview: state.showPreview,
      columnOverrides: state.columnOverrides,
      aiSuggestions: state.aiSuggestions,
      aiSuggestionsLoading: state.aiSuggestionsLoading,
      aiSuggestionsError: state.aiSuggestionsError,
      aiConfigured: state.aiConfigured,
    }))
  );

  const setMode = useMDFlowStore((state) => state.setMode);
  const setPasteText = useMDFlowStore((state) => state.setPasteText);
  const setFile = useMDFlowStore((state) => state.setFile);
  const setSheets = useMDFlowStore((state) => state.setSheets);
  const setSelectedSheet = useMDFlowStore((state) => state.setSelectedSheet);
  const setGsheetTabs = useMDFlowStore((state) => state.setGsheetTabs);
  const setSelectedGid = useMDFlowStore((state) => state.setSelectedGid);
  const setTemplate = useMDFlowStore((state) => state.setTemplate);
  const setFormat = useMDFlowStore((state) => state.setFormat);
  const setResult = useMDFlowStore((state) => state.setResult);
  const setLoading = useMDFlowStore((state) => state.setLoading);
  const setError = useMDFlowStore((state) => state.setError);
  const setPreview = useMDFlowStore((state) => state.setPreview);
  const setPreviewLoading = useMDFlowStore((state) => state.setPreviewLoading);
  const setShowPreview = useMDFlowStore((state) => state.setShowPreview);
  const setColumnOverride = useMDFlowStore((state) => state.setColumnOverride);
  const setAISuggestions = useMDFlowStore((state) => state.setAISuggestions);
  const setAISuggestionsLoading = useMDFlowStore(
    (state) => state.setAISuggestionsLoading
  );
  const setAISuggestionsError = useMDFlowStore(
    (state) => state.setAISuggestionsError
  );
  const clearAISuggestions = useMDFlowStore(
    (state) => state.clearAISuggestions
  );
  const reset = useMDFlowStore((state) => state.reset);

  const addToHistory = useHistoryStore((state) => state.addToHistory);
  const history = useHistoryStore((state) => state.history);

  const [copied, setCopied] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [showDiff, setShowDiff] = useState(false);
  const [previousOutput, setPreviousOutput] = useState<string>("");
  const [currentDiff, setCurrentDiff] = useState<any>(null);
  const [showHistory, setShowHistory] = useState(false);
  const [showValidationConfigurator, setShowValidationConfigurator] =
    useState(false);
  const [showTemplateEditor, setShowTemplateEditor] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [gsheetLoading, setGsheetLoading] = useState(false);
  const [creatingShare, setCreatingShare] = useState(false);
  const [shareTitle, setShareTitle] = useState("");
  const [shareSlug, setShareSlug] = useState("");
  const [shareVisibility, setShareVisibility] = useState<"public" | "private">(
    "public"
  );
  const [shareAllowComments, setShareAllowComments] = useState(true);
  const [showShareOptions, setShowShareOptions] = useState(false);
  const [shareSlugError, setShareSlugError] = useState<string | null>(null);
  const shareOptionsRef = useRef<HTMLDivElement>(null);
  const [debouncedPasteText, setDebouncedPasteText] = useState("");
  const googleAuth = useGoogleAuth();

  useBodyScrollLock(showDiff);

  const { data: templateList = [] } = useMDFlowTemplatesQuery();
  const templates = templateList;
  const getSheetsMutation = useGetXLSXSheetsMutation();
  const { mutateAsync: fetchGoogleSheetTabs } =
    useGetGoogleSheetSheetsMutation();
  const convertPasteMutation = useConvertPasteMutation();
  const convertXLSXMutation = useConvertXLSXMutation();
  const convertTSVMutation = useConvertTSVMutation();
  const convertGoogleSheetMutation = useConvertGoogleSheetMutation();
  const diffMDFlowMutation = useDiffMDFlowMutation();
  const aiSuggestionsMutation = useAISuggestionsMutation();

  const isGsheetUrl = isGoogleSheetsURL(debouncedPasteText.trim());
  const isInputGsheetUrl = isGoogleSheetsURL(pasteText.trim());
  const previewPasteQuery = usePreviewPasteQuery(
    debouncedPasteText,
    mode === "paste" && debouncedPasteText.trim().length > 0 && !isGsheetUrl,
    template
  );
  const previewTSVQuery = usePreviewTSVQuery(file, mode === "tsv", template);
  const previewXLSXQuery = usePreviewXLSXQuery(
    file,
    selectedSheet,
    mode === "xlsx",
    template
  );
  const previewGoogleSheetQuery = usePreviewGoogleSheetQuery(
    debouncedPasteText.trim(),
    selectedGid,
    mode === "paste" && isGsheetUrl && gsheetTabs.length > 0 && Boolean(selectedGid),
    template
  );

  // Reset store when leaving Studio so data is not shown when user comes back
  useEffect(() => {
    return () => reset();
  }, [reset]);

  // Debounce paste text for preview queries
  useEffect(() => {
    if (mode !== "paste") {
      setDebouncedPasteText("");
      return;
    }

    const timer = setTimeout(() => {
      setDebouncedPasteText(pasteText);
    }, 500);

    return () => clearTimeout(timer);
  }, [pasteText, mode]);

  useEffect(() => {
    if (mode !== "paste") {
      setGsheetTabs([]);
      setSelectedGid("");
      return;
    }

    const trimmed = debouncedPasteText.trim();
    if (!trimmed || !isGoogleSheetsURL(trimmed)) {
      setGsheetTabs([]);
      setSelectedGid("");
      return;
    }

    let cancelled = false;
    const loadTabs = async () => {
      setGsheetLoading(true);
      setError(null);
      try {
        const result = await fetchGoogleSheetTabs({
          url: trimmed,
        });
        if (cancelled) return;
        setGsheetTabs(result.sheets);
        setSelectedGid(result.active_gid);
      } catch (error) {
        if (cancelled) return;
        setGsheetTabs([]);
        setSelectedGid("");
        const message =
          error instanceof Error
            ? error.message
            : "Failed to read Google Sheets tabs";
        if (!message.toLowerCase().includes("not configured")) {
          setError(message);
        }
      } finally {
        if (!cancelled) {
          setGsheetLoading(false);
        }
      }
    };

    loadTabs();
    return () => {
      cancelled = true;
    };
  }, [
    debouncedPasteText,
    mode,
    googleAuth.connected,
    fetchGoogleSheetTabs,
    setError,
    setGsheetTabs,
    setSelectedGid,
  ]);

  useEffect(() => {
    if (mode !== "paste") return;
    if (!debouncedPasteText.trim()) {
      setPreview(null);
      setShowPreview(false);
      return;
    }
    if (isGsheetUrl) {
      if (previewGoogleSheetQuery.data) {
        setPreview(previewGoogleSheetQuery.data);
        setShowPreview(true);
      } else {
        setPreview(null);
        setShowPreview(false);
      }
      return;
    }
    if (previewPasteQuery.data) {
      setPreview(previewPasteQuery.data);
      setShowPreview(true);
    }
  }, [
    debouncedPasteText,
    mode,
    isGsheetUrl,
    previewPasteQuery.data,
    previewGoogleSheetQuery.data,
    setPreview,
    setShowPreview,
  ]);

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (!selectedFile) return;

      setFile(selectedFile);
      setLoading(true);
      setError(null);
      setPreview(null);

      if (/\.tsv$/i.test(selectedFile.name)) {
        setLoading(false);
        return;
      }

      try {
        const result = await getSheetsMutation.mutateAsync(selectedFile);
        setSheets(result.sheets);
        setSelectedSheet(result.active_sheet);
      } catch (error) {
        setError(
          error instanceof Error ? error.message : "Failed to read sheets"
        );
      } finally {
        setLoading(false);
      }
    },
    [
      setFile,
      setLoading,
      setError,
      setSheets,
      setSelectedSheet,
      setPreview,
      getSheetsMutation,
    ]
  );

  useEffect(() => {
    if (mode === "xlsx" && previewXLSXQuery.data) {
      setPreview(previewXLSXQuery.data);
      setShowPreview(true);
    }
  }, [mode, previewXLSXQuery.data, setPreview, setShowPreview]);

  useEffect(() => {
    if (mode === "tsv" && previewTSVQuery.data) {
      setPreview(previewTSVQuery.data);
      setShowPreview(true);
    }
  }, [mode, previewTSVQuery.data, setPreview, setShowPreview]);

  useEffect(() => {
    if (googleAuth.error) {
      toast.error("Google connection failed", googleAuth.error);
    }
  }, [googleAuth.error]);

  // Track previous connected state to detect successful connection
  const prevConnectedRef = useRef(googleAuth.connected);
  useEffect(() => {
    if (!prevConnectedRef.current && googleAuth.connected) {
      // Just connected - clear any existing error and show success
      setError(null);
      toast.success("Google connected", "You can now access private sheets");
    }
    prevConnectedRef.current = googleAuth.connected;
  }, [googleAuth.connected, setError]);

  useEffect(() => {
    const isLoading =
      (mode === "paste" && previewPasteQuery.isFetching) ||
      (mode === "paste" && isGsheetUrl && previewGoogleSheetQuery.isFetching) ||
      (mode === "xlsx" && previewXLSXQuery.isFetching) ||
      (mode === "tsv" && previewTSVQuery.isFetching);
    setPreviewLoading(isLoading);
  }, [
    mode,
    isGsheetUrl,
    previewPasteQuery.isFetching,
    previewGoogleSheetQuery.isFetching,
    previewTSVQuery.isFetching,
    previewXLSXQuery.isFetching,
    setPreviewLoading,
  ]);

  const handleConvert = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      let result;
      let inputPreview = "";
      if (mode === "paste") {
        if (!pasteText.trim()) {
          setError("Missing source data");
          return;
        }

        // Check if it's a Google Sheets URL
        if (isGoogleSheetsURL(pasteText.trim())) {
          result = await convertGoogleSheetMutation.mutateAsync({
            url: pasteText.trim(),
            template: "default",
            gid: selectedGid,
            format,
          });
          const selectedTab = gsheetTabs.find((tab) => tab.gid === selectedGid);
          const tabLabel = selectedTab?.title || selectedGid;
          inputPreview = tabLabel
            ? `Google Sheet: ${pasteText.trim().slice(0, 60)}... (${tabLabel})`
            : `Google Sheet: ${pasteText.trim().slice(0, 60)}...`;
        } else {
          result = await convertPasteMutation.mutateAsync({
            pasteText,
            template: "default",
            format,
          });
          inputPreview =
            pasteText.slice(0, 200) + (pasteText.length > 200 ? "..." : "");
        }
      } else if (mode === "xlsx") {
        if (!file) {
          setError("No file uploaded");
          return;
        }
        result = await convertXLSXMutation.mutateAsync({
          file,
          sheetName: selectedSheet,
          template: "default",
          format,
        });
        inputPreview = `${file.name}${selectedSheet ? ` (${selectedSheet})` : ""
          }`;
      } else {
        if (!file) {
          setError("No file uploaded");
          return;
        }
        result = await convertTSVMutation.mutateAsync({
          file,
          template: "default",
          format,
        });
        inputPreview = file.name;
      }

      if (result) {
        setResult(result.mdflow, result.warnings, result.meta);
        // Add to history (use format as display template name)
        addToHistory({
          mode,
          template: format,
          inputPreview,
          output: result.mdflow,
          meta: result.meta,
        });
        toast.success(
          "Conversion complete",
          `${result.meta?.total_rows || 0} rows processed`
        );
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Conversion failed";
      setError(errorMessage);
      toast.error("Conversion failed", errorMessage);
    } finally {
      setLoading(false);
    }
  }, [
    mode,
    pasteText,
    file,
    selectedSheet,
    selectedGid,
    template,
    format,
    setLoading,
    setError,
    setResult,
    addToHistory,
    gsheetTabs,
    convertGoogleSheetMutation,
    convertPasteMutation,
    convertTSVMutation,
    convertXLSXMutation,
  ]);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(mdflowOutput);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [mdflowOutput]);

  const handleDownload = useCallback(() => {
    const blob = new Blob([mdflowOutput], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "spec.mdflow.md";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [mdflowOutput]);

  const handleCreateShare = useCallback(async () => {
    if (!mdflowOutput || creatingShare) return;

    const trimmedSlug = shareSlug.trim();
    if (trimmedSlug && !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(trimmedSlug)) {
      setShareSlugError(
        "Slug must be lowercase letters, numbers, and hyphens only"
      );
      return;
    }

    setShareSlugError(null);

    setCreatingShare(true);
    const result = await createShare({
      mdflow: mdflowOutput,
      template: format,
      title: shareTitle.trim(),
      slug: shareSlug.trim() || undefined,
      is_public: shareVisibility === "public",
      allow_comments: shareAllowComments,
      permission: shareAllowComments ? "comment" : "view",
    });

    if (result.error || !result.data) {
      if (result.error?.toLowerCase().includes("slug")) {
        setShareSlugError(result.error);
      }
      toast.error(
        "Share failed",
        result.error || "Unable to create share link"
      );
      setCreatingShare(false);
      return;
    }

    const shareUrl = `${window.location.origin}/s/${result.data.slug || result.data.token
      }`;
    try {
      await navigator.clipboard.writeText(shareUrl);
      toast.success("Share link copied", "Public link ready to share");
    } catch (error) {
      toast.error("Copy failed", "Could not copy share link");
    }
    setCreatingShare(false);
    setShowShareOptions(false);
  }, [
    creatingShare,
    mdflowOutput,
    format,
    shareTitle,
    shareSlug,
    shareVisibility,
    shareAllowComments,
  ]);

  useEffect(() => {
    if (!showShareOptions) return;
    const handleClickOutside = (event: MouseEvent) => {
      if (!shareOptionsRef.current) return;
      if (!shareOptionsRef.current.contains(event.target as Node)) {
        setShowShareOptions(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [showShareOptions]);

  const handleGetAISuggestions = useCallback(async () => {
    if (!pasteText.trim() || aiSuggestionsLoading) return;

    setAISuggestionsLoading(true);
    setAISuggestionsError(null);
    clearAISuggestions();

    try {
      const result = await aiSuggestionsMutation.mutateAsync({
        pasteText,
        template,
      });
      setAISuggestions(result.suggestions, result.configured);
      if (result.error) {
        setAISuggestionsError(result.error);
      }
    } catch (error) {
      setAISuggestionsError(
        error instanceof Error ? error.message : "Failed to get suggestions"
      );
    } finally {
      setAISuggestionsLoading(false);
    }
  }, [
    pasteText,
    template,
    aiSuggestionsLoading,
    setAISuggestionsLoading,
    setAISuggestionsError,
    setAISuggestions,
    clearAISuggestions,
    aiSuggestionsMutation,
  ]);

  // Keyboard shortcuts via hook
  useKeyboardShortcuts({
    commandPalette: () => setShowCommandPalette(true),
    convert: handleConvert,
    copy: () => {
      if (mdflowOutput) {
        handleCopy();
        toast.success("Copied to clipboard");
      }
    },
    export: () => {
      if (mdflowOutput) {
        handleDownload();
        toast.success("Downloaded spec.mdflow.md");
      }
    },
    togglePreview: () => setShowPreview(!showPreview),
    showShortcuts: () => { }, // Handled by KeyboardShortcutsTooltip
    escape: () => {
      if (showCommandPalette) setShowCommandPalette(false);
      else if (showHistory) setShowHistory(false);
      else if (showDiff) setShowDiff(false);
      else if (showTemplateEditor) setShowTemplateEditor(false);
      else if (showValidationConfigurator) setShowValidationConfigurator(false);
    },
  });

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const f = e.dataTransfer.files?.[0];
      if (!f) return;

      if (mode === "tsv" && /\.tsv$/i.test(f.name)) {
        setFile(f);
        setError(null);
        setPreview(null);
        setLoading(true);
        setLoading(false);
        return;
      }

      if (mode === "xlsx" && /\.(xlsx|xls)$/i.test(f.name)) {
        setFile(f);
        setLoading(true);
        setError(null);
        setPreview(null);
        getSheetsMutation
          .mutateAsync(f)
          .then((result) => {
            setSheets(result.sheets);
            setSelectedSheet(result.active_sheet);
          })
          .catch((error) => {
            setError(
              error instanceof Error ? error.message : "Failed to read sheets"
            );
          })
          .finally(() => {
            setLoading(false);
          });
      }
    },
    [
      mode,
      setFile,
      setLoading,
      setError,
      setSheets,
      setSelectedSheet,
      setPreview,
      getSheetsMutation,
    ]
  );

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
        {/* Left: Source & config — compact header with integrated controls */}
        <motion.div
          variants={stagger.item}
          className="flex flex-col min-h-0 h-full overflow-hidden"
        >
          <section className="p-0 flex flex-col h-full min-h-0 border border-white/10 bg-black/30 backdrop-blur-xl relative overflow-hidden rounded-xl sm:rounded-2xl">
            <div className="studio-grain" aria-hidden />
            <div className="relative z-10 flex flex-col h-full min-h-0">
              {/* Compact header with mode toggle */}
              <div className="flex items-center justify-between gap-2 px-3 sm:px-4 py-2.5 sm:py-3 border-b border-white/5 bg-white/2 shrink-0">
                <div
                  className="flex bg-black/40 rounded-lg border border-white/5 shrink-0"
                  data-tour="input-mode"
                >
                  {[
                    { key: "paste", label: "Paste" },
                    { key: "xlsx", label: "Excel" },
                    { key: "tsv", label: "TSV" },
                  ].map((m) => (
                    <button
                      key={m.key}
                      type="button"
                      onClick={() => {
                        setMode(m.key as "paste" | "xlsx" | "tsv");
                        setFile(null);
                      }}
                      className={`
                        px-3 sm:px-4 py-1.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md transition-all duration-200
                        ${mode === m.key
                          ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                          : "text-muted hover:text-white hover:bg-white/5"
                        }
                      `}
                    >
                      {m.label}
                    </button>
                  ))}
                </div>

                {/* Quick actions */}
                <div className="flex items-center gap-1.5">
                  <Tooltip content="Template Editor">
                    <button
                      type="button"
                      onClick={() => setShowTemplateEditor(true)}
                      className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                    >
                      <FileCode className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                  <Tooltip content="Validation Rules">
                    <button
                      type="button"
                      onClick={() => setShowValidationConfigurator(true)}
                      className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                    >
                      <ShieldCheck className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                </div>
              </div>

              <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-3 sm:px-4 py-3 custom-scrollbar bg-black/3">
                <AnimatePresence mode="wait" initial={false}>
                  {error && (
                    <motion.div
                      initial={{ opacity: 0, y: -8 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -8 }}
                      className="mb-3 p-2.5 bg-accent-red/10 border border-accent-red/25 rounded-lg flex items-center gap-2 text-accent-red text-[9px] font-bold uppercase tracking-wider shrink-0"
                    >
                      <AlertCircle className="w-3.5 h-3.5 shrink-0" /> {error}
                    </motion.div>
                  )}
                </AnimatePresence>

                <AnimatePresence mode="wait">
                  {mode === "paste" ? (
                    <motion.div
                      key="paste"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="h-full flex flex-col min-h-0"
                    >
                      {/* Compact status bar */}
                      <div className="flex flex-wrap items-center gap-2 text-[9px] uppercase font-bold text-muted/50 mb-2 shrink-0">
                        {isGoogleSheetsURL(pasteText.trim()) && (
                          <span className="flex items-center gap-1 text-green-400/80 bg-green-400/10 px-2 py-0.5 rounded">
                            <Link2 className="w-3 h-3" />
                            Google Sheet
                          </span>
                        )}
                        {preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 &&
                          !isGoogleSheetsURL(pasteText.trim()) && (
                            <button
                              type="button"
                              onClick={() => setShowPreview(!showPreview)}
                              className="flex items-center gap-1 text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer bg-accent-orange/10 px-2 py-0.5 rounded"
                            >
                              {showPreview ? (
                                <EyeOff className="w-3 h-3" />
                              ) : (
                                <Eye className="w-3 h-3" />
                              )}
                              {showPreview ? "Hide" : "Show"} Preview
                            </button>
                          )}
                        {previewLoading && (
                          <span className="flex items-center gap-1 text-accent-orange/60">
                            <RefreshCcw className="w-3 h-3 animate-spin" />
                            Analyzing...
                          </span>
                        )}
                        {isGoogleSheetsURL(pasteText.trim()) && gsheetLoading && (
                          <span className="flex items-center gap-1 text-blue-400/70">
                            <RefreshCcw className="w-3 h-3 animate-spin" />
                            Loading sheets...
                          </span>
                        )}
                      </div>

                      {isGoogleSheetsURL(pasteText.trim()) && gsheetTabs.length > 0 && (
                        <div className="mb-3 shrink-0">
                          <Select
                            value={selectedGid}
                            onValueChange={setSelectedGid}
                            options={gsheetTabs.map((tab) => ({
                              label: tab.title,
                              value: tab.gid,
                            }))}
                            placeholder="Choose sheet"
                            size="compact"
                            className="w-auto min-w-40"
                          />
                        </div>
                      )}
                      {isInputGsheetUrl && (
                        <div className={`mb-3 flex flex-wrap items-center gap-2 rounded-lg px-3 py-2 text-[10px] text-white/70 border ${googleAuth.connected
                          ? "border-green-500/20 bg-green-500/10"
                          : "border-accent-orange/20 bg-accent-orange/10"
                          }`}>
                          {googleAuth.connected ? (
                            <Check className="h-3 w-3 shrink-0 text-green-400" />
                          ) : (
                            <AlertCircle className="h-3 w-3 shrink-0 text-accent-orange/80" />
                          )}
                          <span className="flex-1 min-w-40">
                            {googleAuth.connected
                              ? "Google connected. You can access private sheets without sharing."
                              : "Private sheet? Connect Google to access without sharing."}
                          </span>
                          {googleAuth.connected ? (
                            <button
                              type="button"
                              onClick={googleAuth.logout}
                              className="inline-flex items-center gap-1 rounded-md border border-white/10 bg-white/5 px-2 py-1 text-[9px] font-bold uppercase tracking-wider text-white/70 hover:text-white transition-colors"
                            >
                              Disconnect
                            </button>
                          ) : (
                            <button
                              type="button"
                              onClick={googleAuth.login}
                              disabled={googleAuth.loading}
                              className="inline-flex items-center gap-1 rounded-md border border-accent-orange/30 bg-accent-orange/20 px-2 py-1 text-[9px] font-bold uppercase tracking-wider text-white/90 hover:bg-accent-orange/30 transition-colors"
                            >
                              <KeyRound className="h-3 w-3" />
                              Connect Google
                            </button>
                          )}
                        </div>
                      )}

                      {/* Preview Table - Collapsible */}
                      <AnimatePresence>
                        {showPreview &&
                          preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 && (
                            <motion.div
                              initial={{ opacity: 0, height: 0 }}
                              animate={{ opacity: 1, height: "auto" }}
                              exit={{ opacity: 0, height: 0 }}
                              className="mb-3 shrink-0 max-h-[30vh] overflow-auto custom-scrollbar"
                              data-tour="preview-table"
                            >
                              <PreviewTable
                                preview={preview}
                                columnOverrides={columnOverrides}
                                onColumnOverride={setColumnOverride}
                                sourceUrl={undefined}
                              />
                            </motion.div>
                          )}
                        {preview && preview.input_type === "markdown" && (
                          <motion.div
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: "auto" }}
                            exit={{ opacity: 0, height: 0 }}
                            className="mb-3 shrink-0"
                          >
                            <div className="rounded-lg border border-blue-500/20 bg-blue-500/5 px-3 py-2 flex items-center gap-2">
                              <FileText className="w-3.5 h-3.5 text-blue-400/80 shrink-0" />
                              <span className="text-[9px] font-bold text-blue-400/90 uppercase tracking-wider">
                                Markdown detected - passthrough mode
                              </span>
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>

                      <textarea
                        value={pasteText}
                        onChange={(e) => setPasteText(e.target.value)}
                        placeholder="Paste your table data here (TSV, CSV, or Google Sheets URL)…"
                        className="input flex-1 font-mono text-[12px] leading-relaxed resize-none border-white/5 bg-black/30 focus:bg-black/40 focus:border-accent-orange/30 custom-scrollbar min-h-30 rounded-lg"
                        aria-label="Paste TSV or CSV data"
                        data-tour="paste-area"
                      />
                    </motion.div>
                  ) : (
                    <motion.div
                      key={mode}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className={`h-full flex flex-col gap-4 min-h-0 ${!file ? "justify-center items-center" : "justify-start"
                        }`}
                    >
                      {/* File drop zone - centered when no file, shrink when file uploaded */}
                      <div
                        onDragOver={(e) => {
                          e.preventDefault();
                          setDragOver(true);
                        }}
                        onDragLeave={() => setDragOver(false)}
                        onDrop={onDrop}
                        className={`
                          relative rounded-2xl border-2 border-dashed transition-all duration-300 cursor-pointer w-full shrink-0
                          ${file ? "p-4" : "p-8 sm:p-12 max-w-lg"}
                          ${dragOver
                            ? "border-accent-orange/50 bg-accent-orange/10 scale-[1.02]"
                            : file
                              ? "border-accent-orange/30 bg-accent-orange/5"
                              : "border-white/20 hover:border-accent-orange/40 hover:bg-white/5"
                          }
                        `}
                      >
                        <input
                          type="file"
                          accept={mode === "tsv" ? ".tsv" : ".xlsx,.xls"}
                          onChange={handleFileChange}
                          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                          aria-label={
                            mode === "tsv"
                              ? "Upload TSV file"
                              : "Upload Excel file"
                          }
                        />
                        <div
                          className={`flex items-center gap-4 ${file ? "justify-start" : "justify-center flex-col"
                            }`}
                        >
                          <div
                            className={`
                              rounded-2xl flex items-center justify-center transition-all
                              ${file
                                ? "h-12 w-12 bg-accent-orange/20"
                                : "h-16 w-16 bg-white/10"
                              }
                            `}
                          >
                            {file ? (
                              <Check className="w-6 h-6 text-accent-orange" />
                            ) : (
                              <FileSpreadsheet
                                className={`w-8 h-8 ${dragOver
                                  ? "text-accent-orange"
                                  : "text-white/40"
                                  }`}
                              />
                            )}
                          </div>
                          <div className={file ? "text-left" : "text-center"}>
                            {file ? (
                              <>
                                <p className="text-sm font-bold text-white truncate max-w-62.5">
                                  {file.name}
                                </p>
                                <p className="text-xs text-white/50 font-mono">
                                  {(file.size / 1024).toFixed(1)} KB
                                </p>
                              </>
                            ) : (
                              <>
                                <p className="text-sm font-black text-white uppercase tracking-widest">
                                  {dragOver
                                    ? "Drop file here"
                                    : mode === "tsv"
                                      ? "Upload .TSV"
                                      : "Upload .XLSX or .XLS"}
                                </p>
                                <p className="text-xs text-white/50 mt-1">
                                  Click or drag & drop
                                </p>
                              </>
                            )}
                          </div>
                        </div>
                      </div>

                      {/* Sheet selector */}
                      {mode === "xlsx" && sheets.length > 0 && (
                        <div className="shrink-0">
                          <Select
                            value={selectedSheet}
                            onValueChange={setSelectedSheet}
                            options={sheets.map((s) => ({
                              label: s,
                              value: s,
                            }))}
                            placeholder="Choose sheet"
                            size="compact"
                            className="w-auto min-w-30"
                          />
                        </div>
                      )}

                      {/* File Preview Table - takes remaining space */}
                      <AnimatePresence>
                        {file &&
                          preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 && (
                            <motion.div
                              initial={{ opacity: 0 }}
                              animate={{ opacity: 1 }}
                              exit={{ opacity: 0 }}
                              className="flex-1 min-h-0 flex flex-col"
                            >
                              <div className="flex items-center justify-between mb-2 shrink-0">
                                <span className="text-[10px] text-white/50 uppercase font-bold tracking-wider">
                                  Data Preview
                                </span>
                                <button
                                  type="button"
                                  onClick={() => setShowPreview(!showPreview)}
                                  className="flex items-center gap-1.5 text-[10px] text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer font-bold uppercase"
                                >
                                  {showPreview ? (
                                    <EyeOff className="w-3.5 h-3.5" />
                                  ) : (
                                    <Eye className="w-3.5 h-3.5" />
                                  )}
                                  {showPreview ? "Hide" : "Show"}
                                </button>
                              </div>
                              {showPreview && (
                                <div className="flex-1 min-h-0 overflow-auto custom-scrollbar rounded-lg border border-white/10">
                                  <PreviewTable
                                    preview={preview}
                                    columnOverrides={columnOverrides}
                                    onColumnOverride={setColumnOverride}
                                  />
                                </div>
                              )}
                              {previewLoading && (
                                <div className="flex items-center gap-2 text-[10px] text-accent-orange/60 mt-2 shrink-0">
                                  <RefreshCcw className="w-3.5 h-3.5 animate-spin" />
                                  Loading preview...
                                </div>
                              )}
                            </motion.div>
                          )}
                      </AnimatePresence>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              {/* Compact footer with template, format & run */}
              <div className="px-3 sm:px-4 py-2.5 sm:py-3 border-t border-white/5 bg-white/2 shrink-0">
                <div
                  className="flex items-center gap-2 sm:gap-3"
                  data-tour="template-selector"
                >
                  {/* Template dropdown - collapsible on mobile */}
                  <div className="flex-1 min-w-0">
                    <TemplateCards
                      templates={templates}
                      selected={format}
                      onSelect={setFormat}
                      compact
                    />
                  </div>

                  {/* Run button */}
                  <div className="shrink-0" data-tour="run-button">
                    {(() => {
                      const isDisabled =
                        loading ||
                        (mode === "paste" && !pasteText.trim()) ||
                        ((mode === "xlsx" || mode === "tsv") && !file);
                      const isMac =
                        typeof navigator !== "undefined" &&
                        navigator.platform.toUpperCase().indexOf("MAC") >= 0;
                      const modKey = isMac ? "⌘" : "Ctrl";

                      return (
                        <motion.button
                          type="button"
                          whileHover={!isDisabled ? { scale: 1.02 } : {}}
                          whileTap={!isDisabled ? { scale: 0.98 } : {}}
                          onClick={handleConvert}
                          disabled={isDisabled || loading}
                          className={`
                            h-9 sm:h-10 px-4 sm:px-6
                            uppercase tracking-wider text-[10px] sm:text-xs font-bold rounded-lg
                            flex items-center justify-center gap-2
                            transition-all duration-200
                            ${isDisabled
                              ? "bg-white/5 border border-white/10 text-white/30 cursor-not-allowed"
                              : "btn-primary shadow-lg shadow-accent-orange/20 cursor-pointer hover:shadow-xl hover:shadow-accent-orange/30"
                            }
                          `}
                          title={
                            isDisabled
                              ? mode === "paste"
                                ? "Paste data"
                                : "Upload file"
                              : `${modKey}+Enter`
                          }
                        >
                          {loading ? (
                            <RefreshCcw className="w-3.5 h-3.5 animate-spin" />
                          ) : (
                            <Zap className="w-3.5 h-3.5" />
                          )}
                          <span className="hidden xs:inline">
                            {loading ? "Running" : "Run"}
                          </span>
                        </motion.button>
                      );
                    })()}
                  </div>
                </div>
              </div>
            </div>
          </section>
        </motion.div>

        {/* Right: Output — compact and efficient */}
        <motion.div
          variants={stagger.item}
          className="flex flex-col min-h-0 h-full overflow-hidden"
          data-tour="output-panel"
        >
          <div className="p-0 flex flex-col h-full min-h-0 border border-white/10 bg-black/30 backdrop-blur-xl relative overflow-hidden rounded-xl sm:rounded-2xl">
            <div className="studio-grain" aria-hidden />
            <div className="relative z-10 flex flex-col h-full min-h-0">
              {/* Header - synced with Source section */}
              <div className="flex items-center justify-between gap-2 px-3 sm:px-4 py-2.5 sm:py-3 border-b border-white/5 bg-white/2 shrink-0">
                <div className="flex items-center gap-2 min-w-0">
                  <div className="flex gap-0.5 shrink-0">
                    <span
                      className={`w-1.5 h-1.5 rounded-full bg-red-400/80`}
                    />
                    <span className="w-1.5 h-1.5 rounded-full bg-yellow-400/80" />
                    <span className="w-1.5 h-1.5 rounded-full bg-green-400/80" />
                  </div>
                  <span className="text-[9px] sm:text-[10px] font-bold uppercase tracking-wider text-white/70">
                    Output
                  </span>
                  {mdflowOutput && meta && (
                    <span className="text-[8px] sm:text-[9px] hidden sm:inline text-muted/50 font-mono">
                      {meta.total_rows || 0} rows
                    </span>
                  )}
                </div>
                {/* Action buttons - always visible, disabled when no output */}
                <div className="flex items-center gap-1.5 shrink-0">
                  <Tooltip content={copied ? "Copied!" : "Copy"}>
                    <button
                      type="button"
                      onClick={handleCopy}
                      disabled={!mdflowOutput}
                      className={`p-1.5 sm:p-2 rounded-lg border transition-all ${mdflowOutput
                        ? "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
                        : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
                        }`}
                    >
                      {copied ? (
                        <Check className="w-3.5 h-3.5 text-accent-orange" />
                      ) : (
                        <Copy className="w-3.5 h-3.5" />
                      )}
                    </button>
                  </Tooltip>
                  <Tooltip content="Save snapshot">
                    <button
                      type="button"
                      onClick={() => {
                        if (mdflowOutput) {
                          setPreviousOutput(mdflowOutput);
                          setCopied(true);
                          setTimeout(() => setCopied(false), 1500);
                        }
                      }}
                      disabled={!mdflowOutput}
                      className={`p-1.5 sm:p-2 rounded-lg border transition-all ${mdflowOutput
                        ? "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
                        : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
                        }`}
                    >
                      <Save className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                  {previousOutput && (
                    <Tooltip content="Compare">
                      <button
                        type="button"
                        onClick={async () => {
                          if (mdflowOutput) {
                            const diff = await diffMDFlowMutation.mutateAsync({
                              before: previousOutput,
                              after: mdflowOutput,
                            });
                            setCurrentDiff(diff);
                            setShowDiff(true);
                          }
                        }}
                        disabled={!mdflowOutput}
                        className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                      >
                        <GitCompare className="w-3.5 h-3.5" />
                      </button>
                    </Tooltip>
                  )}
                  <Tooltip content="Export">
                    <button
                      type="button"
                      disabled={!mdflowOutput}
                      className={`p-1.5 sm:p-2 rounded-lg border transition-all ${mdflowOutput
                        ? "bg-accent-orange/90 hover:bg-accent-orange border-accent-orange/50 text-white"
                        : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
                        }`}
                      onClick={() => {
                        if (mdflowOutput) {
                          const blob = new Blob([mdflowOutput], {
                            type: "text/markdown",
                          });
                          const url = URL.createObjectURL(blob);
                          const a = document.createElement("a");
                          a.href = url;
                          a.download = "spec.mdflow.md";
                          a.click();
                          URL.revokeObjectURL(url);
                        }
                      }}
                    >
                      <Download className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                  <ShareButton
                    mdflowOutput={mdflowOutput}
                    template={template}
                  />
                  {history.length > 0 && (
                    <Tooltip content="History">
                      <button
                        type="button"
                        onClick={() => setShowHistory(true)}
                        className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                      >
                        <History className="w-3.5 h-3.5" />
                      </button>
                    </Tooltip>
                  )}
                </div>
              </div>

              {/* Output content */}
              <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-3 sm:px-4 py-3 custom-scrollbar">
                {loading ? (
                  <OutputSkeleton />
                ) : mdflowOutput ? (
                  <motion.pre
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 0.2 }}
                    className="whitespace-pre-wrap wrap-break-word font-mono text-[11px] sm:text-[12px] leading-relaxed text-white/90 selection:bg-accent-orange/30"
                  >
                    {mdflowOutput}
                  </motion.pre>
                ) : (
                  <div className="h-full flex flex-col items-center justify-center text-center py-6">
                    <div className="rounded-xl bg-white/5 border border-white/5 p-4 mb-3">
                      <Terminal className="w-8 h-8 text-white/20" />
                    </div>
                    <p className="text-[10px] font-bold uppercase tracking-widest text-white/40">
                      Output will appear here
                    </p>
                    <p className="text-[9px] text-muted/50 mt-1">
                      Paste data and run to generate
                    </p>
                  </div>
                )}
              </div>

              {/* Compact stats footer - only show when there's output */}
              {(mdflowOutput ||
                warnings.length > 0 ||
                aiSuggestions.length > 0) && (
                  <div className="border-t border-white/10 bg-white/2 px-3 sm:px-4 py-2 sm:py-2.5 shrink-0">
                    <TechnicalAnalysis
                      meta={meta}
                      warnings={warnings}
                      mdflowOutput={mdflowOutput}
                      aiSuggestions={aiSuggestions}
                      aiSuggestionsLoading={aiSuggestionsLoading}
                      aiSuggestionsError={aiSuggestionsError}
                      aiConfigured={aiConfigured}
                    />
                  </div>
                )}
            </div>
          </div>
        </motion.div>
      </div>

      {/* Diff Viewer Modal */}
      <AnimatePresence>
        {showDiff && currentDiff && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setShowDiff(false)}
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
          >
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              onClick={(e) => e.stopPropagation()}
              className="bg-black/60 backdrop-blur-xl border border-white/20 rounded-2xl shadow-2xl max-w-4xl w-full max-h-[80vh] flex flex-col overflow-hidden"
            >
              <div className="flex items-center justify-between gap-4 px-6 py-3 border-b border-white/10 bg-white/3 shrink-0">
                <div className="flex items-center gap-3">
                  <span className="text-[10px] font-black uppercase tracking-[0.25em] text-white/80">
                    MDFlow Diff Viewer
                  </span>
                </div>
                <button
                  onClick={() => setShowDiff(false)}
                  className="p-2 h-auto -mr-2 rounded-md hover:bg-white/20 transition-colors cursor-pointer text-white/60 hover:text-white"
                  aria-label="Close"
                >
                  ✕
                </button>
              </div>
              <div className="flex-1 min-h-0 overflow-auto custom-scrollbar">
                <DiffViewer diff={currentDiff} />
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* History Modal */}
      <AnimatePresence>
        {showHistory && (
          <HistoryModal
            history={history}
            onClose={() => setShowHistory(false)}
            onSelect={(record: ConversionRecord) => {
              setResult(record.output, [], record.meta!);
              setShowHistory(false);
            }}
          />
        )}
      </AnimatePresence>

      {/* Validation Rules Configurator */}
      <ValidationConfigurator
        open={showValidationConfigurator}
        onClose={() => setShowValidationConfigurator(false)}
        showValidateAction={true}
      />

      {/* Template Editor */}
      <TemplateEditor
        isOpen={showTemplateEditor}
        onClose={() => setShowTemplateEditor(false)}
        currentSampleData={pasteText || undefined}
      />

      {/* Keyboard shortcuts tooltip */}
      <div className="fixed bottom-4 right-4 z-40">
        <KeyboardShortcutsTooltip />
      </div>

      {/* Command Palette */}
      <CommandPalette
        open={showCommandPalette}
        onOpenChange={setShowCommandPalette}
        onConvert={handleConvert}
        onCopy={() => {
          if (mdflowOutput) {
            handleCopy();
            toast.success("Copied to clipboard");
          }
        }}
        onExport={() => {
          if (mdflowOutput) {
            handleDownload();
            toast.success("Downloaded spec.mdflow.md");
          }
        }}
        onTogglePreview={() => setShowPreview(!showPreview)}
        onShowHistory={() => setShowHistory(true)}
        onOpenTemplateEditor={() => setShowTemplateEditor(true)}
        onOpenValidation={() => setShowValidationConfigurator(true)}
        templates={templates}
        currentTemplate={format}
        onSelectTemplate={setFormat}
        hasOutput={Boolean(mdflowOutput)}
      />

      {/* Toast notifications */}
      <ToastContainer />
    </motion.div>
  );
}
