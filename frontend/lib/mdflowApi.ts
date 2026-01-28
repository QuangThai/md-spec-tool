const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

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
  warnings: string[];
  meta: MDFlowMeta;
}

export interface SheetsResponse {
  sheets: string[];
  active_sheet: string;
}

export interface TemplatesResponse {
  templates: string[];
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
