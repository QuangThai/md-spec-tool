import { DiffResponse } from './diffTypes';
import {
  AISuggestResponse,
  ApiResult,
  MDFlowConvertResponse,
  PreviewResponse,
  TemplateContentResponse,
  TemplateInfo,
  TemplatePreviewResponse,
  ValidationResult,
  ValidationRules
} from './types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface SheetsResponse {
  sheets: string[];
  active_sheet: string;
}

export interface TemplatesResponse {
  templates: string[];
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

/**
 * Helper to make API calls and handle errors consistently
 */
async function apiCall<T>(
  url: string,
  options?: RequestInit
): Promise<ApiResult<T>> {
  try {
    const response = await fetch(url, options);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return {
        error: errorData.error || `HTTP ${response.status}`,
      };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return {
      error: error instanceof Error ? error.message : 'Network error',
    };
  }
}

/**
 * Convert pasted TSV/CSV data to MDFlow markdown
 */
export async function convertPaste(
  pasteText: string,
  template?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  return apiCall<MDFlowConvertResponse>(`${API_URL}/api/mdflow/paste`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      paste_text: pasteText,
      template: template || '',
    }),
  });
}

/**
 * Convert XLSX file to MDFlow markdown
 */
export async function convertXLSX(
  file: File,
  sheetName?: string,
  template?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (sheetName) formData.append('sheet_name', sheetName);
  if (template) formData.append('template', template);

  return apiCall<MDFlowConvertResponse>(`${API_URL}/api/mdflow/xlsx`, {
    method: 'POST',
    body: formData,
  });
}

/**
 * Convert TSV file to MDFlow markdown
 */
export async function convertTSV(
  file: File,
  template?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (template) formData.append('template', template);

  return apiCall<MDFlowConvertResponse>(`${API_URL}/api/mdflow/tsv`, {
    method: 'POST',
    body: formData,
  });
}

/**
 * Get sheets from XLSX file
 */
export async function getXLSXSheets(
  file: File
): Promise<ApiResult<SheetsResponse>> {
  const formData = new FormData();
  formData.append('file', file);

  return apiCall<SheetsResponse>(`${API_URL}/api/mdflow/xlsx/sheets`, {
    method: 'POST',
    body: formData,
  });
}

/**
 * Get available MDFlow templates
 */
export async function getMDFlowTemplates(): Promise<ApiResult<TemplatesResponse>> {
  return apiCall<TemplatesResponse>(`${API_URL}/api/mdflow/templates`);
}

/**
 * Get diff between two MDFlow markdown documents
 * Returns null on error (preserving backward compatibility)
 */
export async function diffMDFlow(
  before: string,
  after: string
): Promise<DiffResponse | null> {
  const result = await apiCall<DiffResponse>(`${API_URL}/api/mdflow/diff`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ before, after }),
  });

  return result.error ? null : result.data || null;
}

/**
 * Preview pasted TSV/CSV data
 */
export async function previewPaste(
  pasteText: string
): Promise<ApiResult<PreviewResponse>> {
  return apiCall<PreviewResponse>(`${API_URL}/api/mdflow/preview`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      paste_text: pasteText,
    }),
  });
}

/**
 * Preview TSV file
 */
export async function previewTSV(
  file: File
): Promise<ApiResult<PreviewResponse>> {
  const formData = new FormData();
  formData.append('file', file);

  return apiCall<PreviewResponse>(`${API_URL}/api/mdflow/tsv/preview`, {
    method: 'POST',
    body: formData,
  });
}

/**
 * Preview XLSX file
 */
export async function previewXLSX(
  file: File,
  sheetName?: string
): Promise<ApiResult<PreviewResponse>> {
  const formData = new FormData();
  formData.append('file', file);
  if (sheetName) formData.append('sheet_name', sheetName);

  return apiCall<PreviewResponse>(`${API_URL}/api/mdflow/xlsx/preview`, {
    method: 'POST',
    body: formData,
  });
}

/**
 * Check if a text is a Google Sheets URL
 */
export function isGoogleSheetsURL(text: string): boolean {
  return text.includes('docs.google.com/spreadsheets');
}

/**
 * Fetch Google Sheet data
 */
export async function fetchGoogleSheet(
  url: string
): Promise<ApiResult<GoogleSheetResponse>> {
  return apiCall<GoogleSheetResponse>(`${API_URL}/api/mdflow/gsheet`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ url }),
  });
}

/**
 * Fetch Google Sheet tabs
 */
export async function getGoogleSheetSheets(
  url: string
): Promise<ApiResult<GoogleSheetSheetsResponse>> {
  return apiCall<GoogleSheetSheetsResponse>(`/api/gsheet/sheets`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ url }),
    credentials: 'include',
  });
}

/**
 * Convert Google Sheet to MDFlow markdown
 */
export async function convertGoogleSheet(
  url: string,
  template?: string,
  gid?: string
): Promise<ApiResult<MDFlowConvertResponse>> {
  const payload: { url: string; template: string; gid?: string } = {
    url,
    template: template || '',
  };
  if (gid) {
    payload.gid = gid;
  }
  return apiCall<MDFlowConvertResponse>(
    `/api/gsheet/convert`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
      credentials: 'include',
    }
  );
}

/**
 * Validate pasted data against custom validation rules
 */
export async function validatePaste(
  pasteText: string,
  rules: ValidationRules
): Promise<ApiResult<ValidationResult>> {
  return apiCall<ValidationResult>(`${API_URL}/api/mdflow/validate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      paste_text: pasteText,
      validation_rules: rules,
    }),
  });
}

/**
 * Get template info (variables and functions)
 */
export async function getTemplateInfo(): Promise<ApiResult<TemplateInfo>> {
  return apiCall<TemplateInfo>(`${API_URL}/api/mdflow/templates/info`);
}

/**
 * Get template content by name
 */
export async function getTemplateContent(
  name: string
): Promise<ApiResult<TemplateContentResponse>> {
  return apiCall<TemplateContentResponse>(
    `${API_URL}/api/mdflow/templates/${name}`
  );
}

/**
 * Preview template with sample data
 */
export async function previewTemplate(
  templateContent: string,
  sampleData?: string
): Promise<ApiResult<TemplatePreviewResponse>> {
  return apiCall<TemplatePreviewResponse>(
    `${API_URL}/api/mdflow/templates/preview`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        template_content: templateContent,
        sample_data: sampleData || '',
      }),
    }
  );
}

/**
 * Get AI suggestions for data improvement
 */
export async function getAISuggestions(
  pasteText: string,
  template?: string
): Promise<ApiResult<AISuggestResponse>> {
  return apiCall<AISuggestResponse>(`${API_URL}/api/mdflow/ai/suggest`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      paste_text: pasteText,
      template: template || 'default',
    }),
  });
}
