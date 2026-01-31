/**
 * Canonical fields for MDFlow
 * Used for mapping columns in preview tables
 */
export const CANONICAL_FIELDS = [
  "id",
  "feature",
  "scenario",
  "instructions",
  "inputs",
  "expected",
  "precondition",
  "priority",
  "type",
  "status",
  "endpoint",
  "notes",
  "no",
  "item_name",
  "item_type",
  "required_optional",
  "input_restrictions",
  "display_conditions",
  "action",
  "navigation_destination",
] as const;

export type CanonicalField = (typeof CANONICAL_FIELDS)[number];
