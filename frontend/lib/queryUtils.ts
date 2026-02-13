import { ApiResult } from './types';

export function unwrap<T>(result: ApiResult<T>): T {
  if (result.error) {
    throw new Error(result.error);
  }
  if (!result.data) {
    throw new Error('No data returned');
  }
  return result.data;
}

export function fileKey(file: File | null | undefined): string {
  if (!file) return 'no-file';
  return `${file.name}:${file.size}:${file.lastModified}`;
}

export function hashString(input: string): string {
  let hash = 5381;
  for (let i = 0; i < input.length; i += 1) {
    hash = (hash * 33) ^ input.charCodeAt(i);
  }
  return Math.abs(hash).toString(36);
}

export const queryKeys = {
  mdflow: {
    templates: ['mdflow', 'templates'] as const,
    templateInfo: ['mdflow', 'template-info'] as const,
    templateContent: (name: string) => ['mdflow', 'template', name] as const,
    previewPaste: (hash: string, template?: string) =>
      ['mdflow', 'preview', 'paste', hash, template ?? ''] as const,
    previewTSV: (fileKey: string, template?: string) =>
      ['mdflow', 'preview', 'tsv', fileKey, template ?? ''] as const,
    previewXLSX: (fileKey: string, sheet: string, template?: string) =>
      ['mdflow', 'preview', 'xlsx', fileKey, sheet, template ?? ''] as const,
    previewGoogleSheet: (url: string, gid: string, template?: string, range?: string) =>
      ['mdflow', 'preview', 'gsheet', url, gid, template ?? '', range ?? ''] as const,
  },
  share: {
    all: ['share'] as const,
    detail: (key: string) => ['share', key] as const,
    comments: (key: string) => ['share', key, 'comments'] as const,
    publicList: ['share', 'public'] as const,
  },
  oauth: {
    googleStatus: ['oauth', 'google', 'status'] as const,
  },
} as const;
