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

export async function convertPaste(
  pasteText: string,
  template?: string,
  format?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  return backendClient.safePost<MDFlowConvertResponse>('/api/mdflow/paste', {
    paste_text: pasteText,
    template: template || '',
    format: format || '',
  });
}

export async function convertXLSX(
  file: File,
  sheetName?: string,
  template?: string,
  format?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (sheetName) formData.append('sheet_name', sheetName);
  if (template) formData.append('template', template);
  if (format) formData.append('format', format);

  return backendClient.safePost<MDFlowConvertResponse>('/api/mdflow/xlsx', formData);
}

export async function convertTSV(
  file: File,
  template?: string,
  format?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (template) formData.append('template', template);
  if (format) formData.append('format', format);

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
    return (
      url.protocol === 'https:' &&
      url.hostname === 'docs.google.com' &&
      url.pathname.startsWith('/spreadsheets/')
    );
  } catch {
    return false;
  }
}

export async function fetchGoogleSheet(
  url: string,
  gid?: string
): Promise<ApiResult<GoogleSheetResponse>> {
  const body: { url: string; gid?: string } = { url };
  if (gid) body.gid = gid;
  return nextApiClient.safePost<GoogleSheetResponse>('/api/gsheet/fetch', body);
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
  format?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  const payload: { url: string; template: string; gid?: string; format?: string } = {
    url,
    template: template || '',
  };
  if (gid) payload.gid = gid;
  if (format) payload.format = format;

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
