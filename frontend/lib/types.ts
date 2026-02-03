/**
 * Comprehensive type definitions for MDFlow application
 * Eliminates all 'any' types and provides type-safe interfaces
 */

// ============ API Response Types ============
export type WarningSeverity = 'info' | 'warn' | 'error';
export type WarningCategory = 'input' | 'detect' | 'header' | 'mapping' | 'rows' | 'render';

export interface MDFlowWarning {
  code: string;
  message: string;
  severity: WarningSeverity;
  category: WarningCategory;
  hint?: string;
  details?: Record<string, unknown>;
}

export interface MDFlowMeta {
  sheet_name?: string;
  header_row: number;
  column_map: Record<string, number>;
  unmapped_columns?: string[];
  total_rows: number;
  rows_by_feature?: Record<string, number>;
}

export interface MDFlowConvertResponse {
  mdflow: string;
  warnings: MDFlowWarning[];
  meta: MDFlowMeta;
  format?: string;
  template?: string;
}

export interface PreviewResponse {
  headers: string[];
  rows: string[][];
  total_rows: number;
  preview_rows: number;
  header_row: number;
  confidence: number;
  column_mapping: Record<string, string>;
  unmapped_columns: string[];
  input_type: 'table' | 'markdown';
}

// ============ Diff Types ============
// Diff types moved to diffTypes.ts for proper type alignment with DiffViewer

// ============ Validation Types ============
export interface ValidationFormatRules {
  id_pattern?: string;
  date_format?: string;
  email_fields?: string[];
  url_fields?: string[];
}

export interface ValidationCrossFieldRule {
  if_field: string;
  then_field: string;
  message?: string;
}

export interface ValidationRules {
  required_fields: string[];
  format_rules?: ValidationFormatRules | null;
  cross_field?: ValidationCrossFieldRule[];
}

export interface ValidationPreset {
  id: string;
  name: string;
  createdAt: number;
  required_fields: string[];
  format_rules?: ValidationFormatRules | null;
  cross_field?: ValidationCrossFieldRule[];
}

export interface ValidationResult {
  valid: boolean;
  warnings: MDFlowWarning[];
}

// ============ AI Suggestions Types ============
export type AISuggestionType =
  | 'missing_field'
  | 'vague_description'
  | 'incomplete_steps'
  | 'formatting'
  | 'coverage';

export interface AISuggestion {
  type: AISuggestionType;
  severity: 'info' | 'warn' | 'error';
  message: string;
  row_ref?: number;
  field?: string;
  suggestion: string;
}

export interface AISuggestResponse {
  suggestions: AISuggestion[];
  error?: string;
  configured: boolean;
}

// ============ Template Types ============
export interface TemplateVariable {
  name: string;
  type: string;
  description: string;
}

export interface TemplateFunction {
  name: string;
  signature: string;
  description: string;
}

export interface TemplateMetadata {
  name: string;
  description: string;
  format: string;
}

export interface TemplateInfo {
  variables: TemplateVariable[];
  functions: TemplateFunction[];
}

export interface TemplateContentResponse {
  name: string;
  content: string;
}

export interface TemplatePreviewResponse {
  output: string;
  error?: string;
  warnings: MDFlowWarning[];
}

export interface TemplatesListResponse {
  templates: TemplateMetadata[];
}

// ============ UI State Types ============
export type InputMode = 'paste' | 'xlsx' | 'tsv';

export interface ConversionRecord {
  id: string;
  timestamp: number;
  mode: InputMode;
  template: string;
  inputPreview: string;
  output: string;
  meta: MDFlowMeta | null;
}

export interface CustomTemplate {
  id: string;
  name: string;
  content: string;
  createdAt: number;
  updatedAt: number;
}

// ============ API Response Wrapper ============
export interface ApiResult<T> {
  data?: T;
  error?: string;
}

// ============ Component Props Types ============
export interface SelectOption {
  label: string;
  value: string;
}

export type TooltipPosition = 'top' | 'bottom' | 'left' | 'right';
