import { useMDFlowTemplatesQuery } from "@/lib/mdflowQueries";

/**
 * Custom hook for managing MDFlow templates
 * Fetches templates from API and ensures "default" is always first
 */
export function useMDFlowTemplates() {
  const { data: templates = ["default"], isLoading, error } =
    useMDFlowTemplatesQuery();

  return {
    templates,
    loading: isLoading,
    error: error ? error.message : null,
  };
}
