import { DiffResponse } from './diffTypes';
import { backendClient, nextApiClient } from './httpClient';
import {
  AISuggestResponse,
  ApiResult,
  MDFlowConvertResponse,
  PreviewResponse,
  TemplateContentResponse,
  TemplateInfo,
  TemplateMetadata,
  TemplatePreviewResponse,
  TemplatesListResponse,
  ValidationResult,
  ValidationRules
} from './types';

export interface SheetsResponse {
  sheets: string[];
  active_sheet: string;
}



export interface GoogleSheetResponse {
  sheet_id: string;
  sheet_name?: string;
  data: string;
}

export interface GoogleSheetTab {
  title: string;
  gid: string;
}

export interface GoogleSheetSheetsResponse {
  sheets: GoogleSheetTab[];
  active_gid: string;
}

export interface ConvertRenderOptions {
  includeMetadata?: boolean;
  numberRows?: boolean;
}

export async function convertPaste(
  pasteText: string,
  template?: string,
  format?: string,
  columnOverrides?: Record<string, string>,
  options?: ConvertRenderOptions
): Promise<ApiResult<MDFlowConvertResponse>> {
  return backendClient.safePost<MDFlowConvertResponse>('/api/mdflow/paste', {
    paste_text: pasteText,
    template: template || '',
    format: format || '',
    column_overrides: columnOverrides && Object.keys(columnOverrides).length > 0
      ? columnOverrides
      : undefined,
    include_metadata: options?.includeMetadata,
    number_rows: options?.numberRows,
  });
}

export async function convertXLSX(
  file: File,
  sheetName?: string,
  template?: string,
  format?: string,
  columnOverrides?: Record<string, string>,
  options?: ConvertRenderOptions
): Promise<ApiResult<MDFlowConvertResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (sheetName) formData.append('sheet_name', sheetName);
  if (template) formData.append('template', template);
  if (format) formData.append('format', format);
  if (columnOverrides && Object.keys(columnOverrides).length > 0) {
    formData.append('column_overrides', JSON.stringify(columnOverrides));
  }
  if (typeof options?.includeMetadata === 'boolean') {
    formData.append('include_metadata', String(options.includeMetadata));
  }
  if (typeof options?.numberRows === 'boolean') {
    formData.append('number_rows', String(options.numberRows));
  }

  return backendClient.safePost<MDFlowConvertResponse>('/api/mdflow/xlsx', formData);
}

export async function convertTSV(
  file: File,
  template?: string,
  format?: string,
  columnOverrides?: Record<string, string>,
  options?: ConvertRenderOptions
): Promise<ApiResult<MDFlowConvertResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (template) formData.append('template', template);
  if (format) formData.append('format', format);
  if (columnOverrides && Object.keys(columnOverrides).length > 0) {
    formData.append('column_overrides', JSON.stringify(columnOverrides));
  }
  if (typeof options?.includeMetadata === 'boolean') {
    formData.append('include_metadata', String(options.includeMetadata));
  }
  if (typeof options?.numberRows === 'boolean') {
    formData.append('number_rows', String(options.numberRows));
  }

  return backendClient.safePost<MDFlowConvertResponse>('/api/mdflow/tsv', formData);
}

export async function getXLSXSheets(
  file: File
): Promise<ApiResult<SheetsResponse>> {
  const formData = new FormData();
  formData.append('file', file);

  return backendClient.safePost<SheetsResponse>('/api/mdflow/xlsx/sheets', formData);
}

export async function getMDFlowTemplates(): Promise<ApiResult<TemplatesListResponse>> {
  return backendClient.safeGet<TemplatesListResponse>('/api/mdflow/templates');
}

export async function diffMDFlow(
  before: string,
  after: string
): Promise<ApiResult<DiffResponse | null>> {
  return backendClient.safePost<DiffResponse | null>('/api/mdflow/diff', { before, after });
}

export async function previewPaste(
  pasteText: string,
  template?: string,
  format?: string,
  skipAI: boolean = true
): Promise<ApiResult<PreviewResponse>> {
  const params = new URLSearchParams();
  if (!skipAI) params.set('skip_ai', 'false');
  const query = params.toString();
  const url = `/api/mdflow/preview${query ? `?${query}` : ''}`;
  return backendClient.safePost<PreviewResponse>(url, {
    paste_text: pasteText,
    template: template || undefined,
    format: format || undefined,
  });
}

export async function previewTSV(
  file: File,
  template?: string,
  format?: string,
  skipAI: boolean = true
): Promise<ApiResult<PreviewResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (template) formData.append('template', template);
  if (format) formData.append('format', format);

  const params = new URLSearchParams();
  if (!skipAI) params.set('skip_ai', 'false');
  const query = params.toString();
  const url = `/api/mdflow/tsv/preview${query ? `?${query}` : ''}`;
  return backendClient.safePost<PreviewResponse>(url, formData);
}

export async function previewXLSX(
  file: File,
  sheetName?: string,
  template?: string,
  format?: string,
  skipAI: boolean = true
): Promise<ApiResult<PreviewResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (sheetName) formData.append('sheet_name', sheetName);
  if (template) formData.append('template', template);
  if (format) formData.append('format', format);

  const params = new URLSearchParams();
  if (!skipAI) params.set('skip_ai', 'false');
  const query = params.toString();
  const url = `/api/mdflow/xlsx/preview${query ? `?${query}` : ''}`;
  return backendClient.safePost<PreviewResponse>(url, formData);
}

export function isGoogleSheetsURL(text: string): boolean {
  try {
    const url = new URL(text.trim());
    const hostname = url.hostname.toLowerCase();
    const allowedHosts = new Set([
      'docs.google.com',
      'spreadsheets.google.com',
    ]);
    return (
      (url.protocol === 'https:' || url.protocol === 'http:') &&
      allowedHosts.has(hostname) &&
      url.pathname.startsWith('/spreadsheets/')
    );
  } catch {
    return false;
  }
}

export async function fetchGoogleSheet(
  url: string,
  gid?: string,
  range?: string
): Promise<ApiResult<GoogleSheetResponse>> {
  const body: { url: string; gid?: string; range?: string } = { url };
  if (gid) body.gid = gid;
  if (range) body.range = range;
  return nextApiClient.safePost<GoogleSheetResponse>('/api/gsheet/fetch', body);
}

export async function previewGoogleSheet(
  url: string,
  gid?: string,
  template?: string,
  range?: string
): Promise<ApiResult<PreviewResponse>> {
  const body: { url: string; gid?: string; template?: string; range?: string } = { url };
  if (gid) body.gid = gid;
  if (template) body.template = template;
  if (range) body.range = range;
  return nextApiClient.safePost<PreviewResponse>('/api/gsheet/preview', body);
}

export async function getGoogleSheetSheets(
  url: string
): Promise<ApiResult<GoogleSheetSheetsResponse>> {
  return nextApiClient.safePost<GoogleSheetSheetsResponse>('/api/gsheet/sheets', { url });
}

export async function convertGoogleSheet(
  url: string,
  template?: string,
  gid?: string,
  format?: string,
  range?: string,
  selectedBlockId?: string,
  columnOverrides?: Record<string, string>,
  options?: ConvertRenderOptions
): Promise<ApiResult<MDFlowConvertResponse>> {
  const payload: {
    url: string;
    template: string;
    gid?: string;
    format?: string;
    range?: string;
    selected_block_id?: string;
    column_overrides?: Record<string, string>;
    include_metadata?: boolean;
    number_rows?: boolean;
  } = {
    url,
    template: template || '',
  };
  if (gid) payload.gid = gid;
  if (format) payload.format = format;
  if (range) payload.range = range;
  if (selectedBlockId) payload.selected_block_id = selectedBlockId;
  if (columnOverrides && Object.keys(columnOverrides).length > 0) {
    payload.column_overrides = columnOverrides;
  }
  if (typeof options?.includeMetadata === 'boolean') {
    payload.include_metadata = options.includeMetadata;
  }
  if (typeof options?.numberRows === 'boolean') {
    payload.number_rows = options.numberRows;
  }

  return nextApiClient.safePost<MDFlowConvertResponse>('/api/gsheet/convert', payload);
}

export async function validatePaste(
  pasteText: string,
  rules: ValidationRules
): Promise<ApiResult<ValidationResult>> {
  return backendClient.safePost<ValidationResult>('/api/mdflow/validate', {
    paste_text: pasteText,
    validation_rules: rules,
  });
}

export async function getTemplateInfo(): Promise<ApiResult<TemplateInfo>> {
  return backendClient.safeGet<TemplateInfo>('/api/mdflow/templates/info');
}

export async function getTemplateContent(
  name: string
): Promise<ApiResult<TemplateContentResponse>> {
  return backendClient.safeGet<TemplateContentResponse>(
    `/api/mdflow/templates/${encodeURIComponent(name)}`
  );
}

export async function previewTemplate(
  templateContent: string,
  sampleData?: string
): Promise<ApiResult<TemplatePreviewResponse>> {
  return backendClient.safePost<TemplatePreviewResponse>('/api/mdflow/templates/preview', {
    template_content: templateContent,
    sample_data: sampleData || '',
  });
}

export async function getAISuggestions(
  pasteText: string,
  template?: string
): Promise<ApiResult<AISuggestResponse>> {
  return backendClient.safePost<AISuggestResponse>('/api/mdflow/ai/suggest', {
    paste_text: pasteText,
    template: template || 'spec',
  });
}
