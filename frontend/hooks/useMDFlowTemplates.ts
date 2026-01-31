import { useState, useEffect } from "react";
import { getMDFlowTemplates } from "@/lib/mdflowApi";

/**
 * Custom hook for managing MDFlow templates
 * Fetches templates from API and ensures "default" is always first
 */
export function useMDFlowTemplates() {
  const [templates, setTemplates] = useState<string[]>(["default"]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTemplates = async () => {
      setLoading(true);
      try {
        const res = await getMDFlowTemplates();
        if (res.data?.templates) {
          // Ensure "default" is always first
          const sorted = [...res.data.templates].sort((a, b) => {
            if (a === "default") return -1;
            if (b === "default") return 1;
            return 0;
          });
          setTemplates(sorted);
        }
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load templates");
      } finally {
        setLoading(false);
      }
    };

    fetchTemplates();
  }, []);

  return { templates, loading, error };
}
