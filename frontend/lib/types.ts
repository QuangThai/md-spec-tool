/**
 * Comprehensive type definitions for MDFlow application
 * Eliminates all 'any' types and provides type-safe interfaces
 */

// ============ Output Format Types ============
export type OutputFormat = 'spec' | 'table';

// ============ Column Mapping Types ============
export interface ColumnMapping {
  canonical_name: string;
  source_header: string;
  column_index: number;
  confidence: number;
  reasoning?: string;
}

export interface ExtraColumn {
  name: string;
  semantic_role: string;
  column_index: number;
}

export interface MappingMeta {
  detected_type: string;
  source_language: string;
  total_columns: number;
  mapped_columns: number;
  avg_confidence: number;
}

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

  // AI metadata (optional)
  ai_mode?: 'off' | 'shadow' | 'on';
  ai_used?: boolean;
  ai_degraded?: boolean;
  ai_avg_confidence?: number;
  ai_mapped_columns?: number;
  ai_unmapped_columns?: number;
  quality_report?: {
    strict_mode: boolean;
    validation_passed: boolean;
    validation_reason?: string;
    header_confidence: number;
    min_header_confidence: number;
    source_rows: number;
    converted_rows: number;
    row_loss_ratio: number;
    max_row_loss_ratio: number;
    header_count: number;
    mapped_columns: number;
    mapped_ratio: number;
    core_field_coverage?: Record<string, boolean>;
  };
}

export interface ConversionResponse {
  markdown: string;
  format: OutputFormat;
  meta: {
    title: string;
    total_items: number;
    schema_version: string;
    ai_mode: 'off' | 'shadow' | 'on';
    degraded?: boolean;
  };
  column_mappings: {
    canonical_fields: ColumnMapping[];
    extra_columns: ExtraColumn[];
    meta: MappingMeta;
  };
  warnings?: string[];
}

export interface MDFlowConvertResponse {
  mdflow: string;
  warnings: MDFlowWarning[];
  meta: MDFlowMeta;
  format?: OutputFormat;
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
  mapping_quality?: {
    score: number;
    header_score: number;
    mapped_ratio: number;
    core_coverage: number;
    core_mapped: number;
    recommended_format: OutputFormat;
    low_confidence_columns: string[];
    column_confidence: Record<string, number>;
    column_reasons?: Record<string, string[]>;
  };
  blocks?: Array<{
    id: string;
    range: string;
    total_rows: number;
    total_columns: number;
    language_hint: 'english' | 'japanese' | 'mixed' | 'unknown';
    english_score: number;
    header_row: number;
    confidence: number;
    mapping_quality?: {
      score: number;
      header_score: number;
      mapped_ratio: number;
      core_coverage: number;
      core_mapped: number;
      recommended_format: OutputFormat;
      low_confidence_columns: string[];
      column_confidence: Record<string, number>;
      column_reasons?: Record<string, string[]>;
    };
  }>;
  selected_block_id?: string;
  selected_block_range?: string;
  input_type: 'table' | 'markdown';
  ai_available: boolean;
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

// ============ API Error Types ============
export class ApiError extends Error {
  constructor(
    message: string,
    public status?: number,
    public code?: string,
    public details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'ApiError';
  }

  static fromResponse(status: number, body?: { error?: string; code?: string; details?: Record<string, unknown> }): ApiError {
    const message = body?.error || `HTTP ${status}`;
    return new ApiError(message, status, body?.code, body?.details);
  }

  get isUnauthorized(): boolean {
    return this.status === 401;
  }

  get isForbidden(): boolean {
    return this.status === 403;
  }

  get isNotFound(): boolean {
    return this.status === 404;
  }

  get isNetworkError(): boolean {
    return !this.status;
  }
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
