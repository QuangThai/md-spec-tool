import { useState, useCallback } from "react";
import { useDiffMDFlowMutation } from "@/lib/mdflowQueries";
import { useBodyScrollLock } from "@/lib/useBodyScrollLock";
import { toast } from "@/components/ui/Toast";

export function useDiffSnapshots() {
  const [showDiff, setShowDiff] = useState(false);
  const [snapshotA, setSnapshotA] = useState("");
  const [snapshotB, setSnapshotB] = useState("");
  const [currentDiff, setCurrentDiff] = useState<any>(null);

  useBodyScrollLock(showDiff);

  const diffMDFlowMutation = useDiffMDFlowMutation();

  const saveSnapshot = useCallback((output: string) => {
    if (!output) return;
    if (!snapshotA) {
      setSnapshotA(output);
      toast.success("Version A saved", "Run with new input, save again");
    } else if (!snapshotB) {
      setSnapshotB(output);
      toast.success("Version B saved", "Ready to compare");
    } else {
      setSnapshotB(output);
      toast.success("Version B updated");
    }
  }, [snapshotA, snapshotB]);

  const compareSnapshots = useCallback(async () => {
    if (diffMDFlowMutation.isPending) {
      return;
    }
    const diff = await diffMDFlowMutation.mutateAsync({
      before: snapshotA,
      after: snapshotB,
    });
    setCurrentDiff(diff);
    setShowDiff(true);
    if (diff && !diff.hunks?.length && diff.added_lines === 0 && diff.removed_lines === 0) {
      toast.info("No changes detected", "Versions may be identical");
    }
  }, [snapshotA, snapshotB, diffMDFlowMutation]);

  const clearSnapshots = useCallback(() => {
    setSnapshotA("");
    setSnapshotB("");
    toast.success("Snapshots cleared");
  }, []);

  return {
    showDiff,
    setShowDiff,
    snapshotA,
    snapshotB,
    currentDiff,
    compareLoading: diffMDFlowMutation.isPending,
    saveSnapshot,
    compareSnapshots,
    clearSnapshots,
  };
}
