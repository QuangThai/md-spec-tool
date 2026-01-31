import type { ValidationPreset, ValidationRules } from "./types";

const STORAGE_KEY = "mdflow-validation-presets";

export function loadPresets(): ValidationPreset[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as ValidationPreset[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export function savePreset(preset: { name: string } & ValidationRules): ValidationPreset {
  const presets = loadPresets();
  const newPreset: ValidationPreset = {
    ...preset,
    id: `preset-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
    createdAt: Date.now(),
  };
  presets.unshift(newPreset);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(presets.slice(0, 20)));
  return newPreset;
}

export function deletePreset(id: string): void {
  const presets = loadPresets().filter((p) => p.id !== id);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(presets));
}

export const CANONICAL_FIELDS = [
  { value: "id", label: "ID" },
  { value: "feature", label: "Feature" },
  { value: "scenario", label: "Scenario" },
  { value: "instructions", label: "Instructions" },
  { value: "inputs", label: "Inputs" },
  { value: "expected", label: "Expected" },
  { value: "precondition", label: "Precondition" },
  { value: "priority", label: "Priority" },
  { value: "type", label: "Type" },
  { value: "status", label: "Status" },
  { value: "endpoint", label: "Endpoint" },
  { value: "notes", label: "Notes" },
  { value: "no", label: "No" },
  { value: "item_name", label: "Item Name" },
  { value: "item_type", label: "Item Type" },
  { value: "action", label: "Action" },
  { value: "navigation_destination", label: "Navigation Destination" },
] as const;

/** Validation presets aligned with MDFlow templates (default, feature-spec, test-plan, api-endpoint, spec-table) */
export const DEFAULT_PRESETS: { name: string; rules: ValidationRules }[] = [
  {
    name: "Default (Test Case)",
    rules: {
      required_fields: ["feature", "scenario", "expected"],
      format_rules: { id_pattern: "^[A-Z]{2,}-\\d+$" },
      cross_field: [
        { if_field: "id", then_field: "feature", message: "When ID is set, Feature is required" },
        { if_field: "instructions", then_field: "expected", message: "When Steps are set, Expected is required" },
      ],
    },
  },
  {
    name: "Feature Spec (User Story)",
    rules: {
      required_fields: ["feature", "scenario", "expected"],
      format_rules: null,
      cross_field: [
        { if_field: "precondition", then_field: "expected", message: "When Given is set, Then (expected) is required" },
      ],
    },
  },
  {
    name: "Test Plan",
    rules: {
      required_fields: ["id", "feature", "scenario", "expected"],
      format_rules: { id_pattern: "^[A-Z]{2,}-\\d+$" },
      cross_field: [{ if_field: "id", then_field: "feature", message: "When ID is set, Feature is required" }],
    },
  },
  {
    name: "API Endpoint",
    rules: {
      required_fields: ["endpoint", "type"],
      format_rules: { url_fields: ["endpoint"] },
      cross_field: [],
    },
  },
  {
    name: "Spec Table (UI)",
    rules: {
      required_fields: ["item_name", "item_type"],
      format_rules: null,
      cross_field: [
        { if_field: "action", then_field: "navigation_destination", message: "When Action is set, Navigation destination is required for navigation actions" },
      ],
    },
  },
];
