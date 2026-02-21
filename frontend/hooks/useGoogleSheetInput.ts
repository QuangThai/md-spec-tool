import { useEffect, useMemo, useRef, useState } from "react";
import { useGoogleAuth } from "@/hooks/useGoogleAuth";
import { useGetGoogleSheetSheetsMutation } from "@/lib/mdflowQueries";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import { toast } from "@/components/ui/Toast";
import { useMDFlowActions, useMDFlowStore } from "@/lib/mdflowStore";

export interface UseGoogleSheetInputReturn {
  gsheetLoading: boolean;
  gsheetRange: string;
  setGsheetRange: (range: string) => void;
  gsheetRangeValue: string;
  googleAuth: ReturnType<typeof useGoogleAuth>;
}

interface UseGoogleSheetInputParams {
  debouncedPasteText: string;
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
}

export function useGoogleSheetInput({
  debouncedPasteText,
  setLastFailedAction,
}: UseGoogleSheetInputParams): UseGoogleSheetInputReturn {
  const mode = useMDFlowStore((state) => state.mode);
  const gsheetTabs = useMDFlowStore((state) => state.gsheetTabs);
  const selectedGid = useMDFlowStore((state) => state.selectedGid);
  const { setGsheetTabs, setSelectedGid, setError } = useMDFlowActions();

  const [gsheetLoading, setGsheetLoading] = useState(false);
  const [gsheetRange, setGsheetRange] = useState("");

  const googleAuth = useGoogleAuth();
  const { mutateAsync: fetchGoogleSheetTabs } =
    useGetGoogleSheetSheetsMutation();

  // Derived value: computed gsheet range with tab name
  const gsheetRangeValue = useMemo(() => {
    const trimmed = gsheetRange.trim();
    if (!trimmed) return "";
    if (trimmed.includes("!")) return trimmed;
    const selectedTab = gsheetTabs.find((tab) => tab.gid === selectedGid);
    const title = selectedTab?.title?.trim();
    if (!title) return "";
    return `${title}!${trimmed}`;
  }, [gsheetRange, gsheetTabs, selectedGid]);

  // Effect: Load Google Sheet tabs when URL changes
  useEffect(() => {
    if (mode !== "paste") {
      setGsheetTabs([]);
      setSelectedGid("");
      setGsheetRange("");
      return;
    }

    const trimmed = debouncedPasteText.trim();
    if (!trimmed || !isGoogleSheetsURL(trimmed)) {
      setGsheetTabs([]);
      setSelectedGid("");
      setGsheetRange("");
      return;
    }

    let cancelled = false;
    const loadTabs = async () => {
      setGsheetLoading(true);
      setError(null);
      setLastFailedAction(null);
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
          setLastFailedAction("preview");
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
    setLastFailedAction,
  ]);

  // Effect: Handle Google auth error toast
  useEffect(() => {
    if (googleAuth.error) {
      toast.error("Google connection failed", googleAuth.error);
    }
  }, [googleAuth.error]);

  // Effect: Track Google auth connection success
  const prevConnectedRef = useRef(googleAuth.connected);
  useEffect(() => {
    if (!prevConnectedRef.current && googleAuth.connected) {
      // Just connected - clear any existing error and show success
      setError(null);
      setLastFailedAction(null);
      toast.success("Google connected", "You can now access private sheets");
    }
    prevConnectedRef.current = googleAuth.connected;
  }, [googleAuth.connected, setError, setLastFailedAction]);

  return {
    gsheetLoading,
    gsheetRange,
    setGsheetRange,
    gsheetRangeValue,
    googleAuth,
  };
}
