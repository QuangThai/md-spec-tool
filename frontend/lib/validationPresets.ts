import type { TemplateMetadata, ValidationPreset, ValidationRules } from "./types";

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

/**
 * Canonical fields synced with backend validator.go:141-189
 * These fields map to SpecRow struct fields in the Go backend
 */
export const CANONICAL_FIELDS = [
  // Core fields
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
  // UI Spec fields
  { value: "no", label: "No" },
  { value: "item_name", label: "Item Name" },
  { value: "item_type", label: "Item Type" },
  { value: "required_optional", label: "Required/Optional" },
  { value: "input_restrictions", label: "Input Restrictions" },
  { value: "display_conditions", label: "Display Conditions" },
  { value: "action", label: "Action" },
  { value: "navigation_destination", label: "Navigation Destination" },
] as const;

/**
 * Default validation rules for known template/format types
 * These provide sensible defaults based on the template's purpose
 */
const TEMPLATE_VALIDATION_RULES: Record<string, ValidationRules> = {
  // Spec document format - structured specifications
  spec: {
    required_fields: ["feature", "scenario", "expected"],
    format_rules: { id_pattern: "^[A-Z]{2,}-\\d+$" },
    cross_field: [
      { if_field: "id", then_field: "feature", message: "When ID is set, Feature is required" },
      { if_field: "instructions", then_field: "expected", message: "When Steps are set, Expected is required" },
      { if_field: "precondition", then_field: "expected", message: "When Precondition is set, Expected is required" },
    ],
  },
  // Table format - simple markdown tables
  table: {
    required_fields: [],
    format_rules: null,
    cross_field: [],
  },
};

/**
 * Fallback validation rules for unknown templates
 */
const DEFAULT_VALIDATION_RULES: ValidationRules = {
  required_fields: [],
  format_rules: null,
  cross_field: [],
};

/**
 * Generate validation presets from templates fetched via API
 * This ensures presets stay in sync with available templates
 */
export function generatePresetsFromTemplates(
  templates: TemplateMetadata[]
): { name: string; rules: ValidationRules }[] {
  if (!templates || templates.length === 0) {
    // Return minimal fallback if no templates available
    return [
      { name: "Default", rules: DEFAULT_VALIDATION_RULES },
    ];
  }

  return templates.map((template) => {
    // Try to find matching rules by template name or format
    const rules = 
      TEMPLATE_VALIDATION_RULES[template.name] ||
      TEMPLATE_VALIDATION_RULES[template.format] ||
      DEFAULT_VALIDATION_RULES;

    return {
      name: template.description || template.name,
      rules,
    };
  });
}

/**
 * Get validation rules for a specific template/format
 */
export function getValidationRulesForTemplate(templateName: string): ValidationRules {
  return TEMPLATE_VALIDATION_RULES[templateName] || DEFAULT_VALIDATION_RULES;
}
