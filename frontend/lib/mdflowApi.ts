import { DiffResponse } from './diffTypes';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

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
}

export interface SheetsResponse {
  sheets: string[];
  active_sheet: string;
}

export interface TemplatesResponse {
  templates: string[];
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

export async function convertPaste(
  pasteText: string,
  template?: string
): Promise<{ data?: MDFlowConvertResponse; error?: string }> {
  try {
    const response = await fetch(`${API_URL}/api/mdflow/paste`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        paste_text: pasteText,
        template: template || '',
      }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function convertXLSX(
  file: File,
  sheetName?: string,
  template?: string
): Promise<{ data?: MDFlowConvertResponse; error?: string }> {
  try {
    const formData = new FormData();
    formData.append('file', file);
    if (sheetName) formData.append('sheet_name', sheetName);
    if (template) formData.append('template', template);

    const response = await fetch(`${API_URL}/api/mdflow/xlsx`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function convertTSV(
  file: File,
  template?: string
): Promise<{ data?: MDFlowConvertResponse; error?: string }> {
  try {
    const formData = new FormData();
    formData.append('file', file);
    if (template) formData.append('template', template);

    const response = await fetch(`${API_URL}/api/mdflow/tsv`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function getXLSXSheets(
  file: File
): Promise<{ data?: SheetsResponse; error?: string }> {
  try {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch(`${API_URL}/api/mdflow/xlsx/sheets`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function getMDFlowTemplates(): Promise<{
  data?: TemplatesResponse;
  error?: string;
}> {
  try {
    const response = await fetch(`${API_URL}/api/mdflow/templates`);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function diffMDFlow(
  before: string,
  after: string
): Promise<DiffResponse | null> {
  try {
    const response = await fetch(`${API_URL}/api/mdflow/diff`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ before, after }),
    });

    if (!response.ok) {
      console.error(`Diff failed: ${response.status}`);
      return null;
    }

    return response.json();
  } catch (error) {
    console.error('Diff error:', error);
    return null;
  }
}

export async function previewPaste(
  pasteText: string
): Promise<{ data?: PreviewResponse; error?: string }> {
  try {
    const response = await fetch(`${API_URL}/api/mdflow/preview`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        paste_text: pasteText,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function previewTSV(
  file: File
): Promise<{ data?: PreviewResponse; error?: string }> {
  try {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch(`${API_URL}/api/mdflow/tsv/preview`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function previewXLSX(
  file: File,
  sheetName?: string
): Promise<{ data?: PreviewResponse; error?: string }> {
  try {
    const formData = new FormData();
    formData.append('file', file);
    if (sheetName) formData.append('sheet_name', sheetName);

    const response = await fetch(`${API_URL}/api/mdflow/xlsx/preview`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

// Google Sheets integration
export interface GoogleSheetResponse {
  sheet_id: string;
  sheet_name?: string;
  data: string;
}

export function isGoogleSheetsURL(text: string): boolean {
  return text.includes('docs.google.com/spreadsheets');
}

export async function fetchGoogleSheet(
  url: string
): Promise<{ data?: GoogleSheetResponse; error?: string }> {
  try {
    const response = await fetch(`${API_URL}/api/mdflow/gsheet`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}

export async function convertGoogleSheet(
  url: string,
  template?: string
): Promise<{ data?: MDFlowConvertResponse; error?: string }> {
  try {
    const response = await fetch(`${API_URL}/api/mdflow/gsheet/convert`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        url,
        template: template || '',
      }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return { error: errorData.error || `HTTP ${response.status}` };
    }

    const data = await response.json();
    return { data };
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Network error' };
  }
}
