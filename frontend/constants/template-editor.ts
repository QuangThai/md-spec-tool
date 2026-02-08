export const STORAGE_KEY = "mdflow-custom-templates";

// Built-in output formats - ONLY 2 formats
export const OUTPUT_FORMATS = [
  { id: "spec", name: "Spec Document", description: "AGENTS.md compatible specification document" },
  { id: "table", name: "Simple Table", description: "Simple markdown table format" },
] as const;
