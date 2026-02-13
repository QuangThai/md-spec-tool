import {
  convertGoogleSheet,
  convertPaste,
  convertTSV,
  convertXLSX,
  diffMDFlow,
  fetchGoogleSheet,
  getAISuggestions,
  getGoogleSheetSheets,
  getMDFlowTemplates,
  getTemplateContent,
  getTemplateInfo,
  getXLSXSheets,
  previewGoogleSheet,
  previewPaste,
  previewTemplate,
  previewTSV,
  previewXLSX,
  validatePaste,
} from '@/lib/mdflowApi';
import { fileKey, hashString, queryKeys, unwrap } from '@/lib/queryUtils';
import { ValidationRules } from '@/lib/types';
import { useMutation, useQuery } from '@tanstack/react-query';

// Default template to sort first if available
const DEFAULT_TEMPLATE_NAME = 'spec';

export function useMDFlowTemplatesQuery() {
  return useQuery({
    queryKey: queryKeys.mdflow.templates,
    queryFn: async () => unwrap(await getMDFlowTemplates()),
    staleTime: 5 * 60 * 1000,
    select: (data) => {
      const sorted = [...data.templates].sort((a, b) => {
        if (a.name === DEFAULT_TEMPLATE_NAME) return -1;
        if (b.name === DEFAULT_TEMPLATE_NAME) return 1;
        return (a.name || "").localeCompare(b.name || "");
      });
      return sorted;
    },
  });
}

export function useTemplateInfoQuery(enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.mdflow.templateInfo,
    queryFn: async () => unwrap(await getTemplateInfo()),
    enabled,
    staleTime: 5 * 60 * 1000,
  });
}

export function useTemplateContentQuery(name: string, enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.mdflow.templateContent(name),
    queryFn: async () => unwrap(await getTemplateContent(name)),
    enabled: enabled && Boolean(name),
  });
}

export function usePreviewPasteQuery(
  pasteText: string,
  enabled: boolean,
  template?: string
) {
  const hash = hashString(pasteText.trim());
  return useQuery({
    queryKey: queryKeys.mdflow.previewPaste(hash, template),
    queryFn: async () => unwrap(await previewPaste(pasteText, template)),
    enabled: enabled && pasteText.trim().length > 0,
    gcTime: 2 * 60 * 1000,
  });
}

export function usePreviewTSVQuery(
  file: File | null,
  enabled: boolean,
  template?: string
) {
  return useQuery({
    queryKey: queryKeys.mdflow.previewTSV(fileKey(file), template),
    queryFn: async () => unwrap(await previewTSV(file!, template)),
    enabled: enabled && Boolean(file),
    gcTime: 2 * 60 * 1000,
  });
}

export function usePreviewTSVMutation(template?: string) {
  return useMutation({
    mutationFn: async (file: File) => unwrap(await previewTSV(file, template)),
  });
}

export function usePreviewXLSXQuery(
  file: File | null,
  sheetName: string,
  enabled: boolean,
  template?: string
) {
  return useQuery({
    queryKey: queryKeys.mdflow.previewXLSX(fileKey(file), sheetName, template),
    queryFn: async () => unwrap(await previewXLSX(file!, sheetName, template)),
    enabled: enabled && Boolean(file) && Boolean(sheetName),
    gcTime: 2 * 60 * 1000,
  });
}

export function usePreviewXLSXMutation(template?: string) {
  return useMutation({
    mutationFn: async (payload: { file: File; sheetName?: string }) =>
      unwrap(await previewXLSX(payload.file, payload.sheetName, template)),
  });
}

/**
 * Preview for Google Sheet using dedicated backend endpoint.
 * Refetches when url, gid, template, or range changes.
 */
export function usePreviewGoogleSheetQuery(
  url: string,
  gid: string,
  enabled: boolean,
  template?: string,
  range?: string
) {
  return useQuery({
    queryKey: queryKeys.mdflow.previewGoogleSheet(url, gid, template, range),
    queryFn: async () => unwrap(await previewGoogleSheet(url, gid, template, range)),
    enabled: enabled && Boolean(url.trim()) && Boolean(gid),
    gcTime: 2 * 60 * 1000,
  });
}

export function useGetXLSXSheetsMutation() {
  return useMutation({
    mutationFn: async (file: File) => unwrap(await getXLSXSheets(file)),
  });
}

export function useConvertPasteMutation() {
  return useMutation({
    mutationFn: async (payload: { pasteText: string; template?: string; format?: string; columnOverrides?: Record<string, string> }) =>
      unwrap(await convertPaste(payload.pasteText, payload.template, payload.format, payload.columnOverrides)),
  });
}

export function useConvertXLSXMutation() {
  return useMutation({
    mutationFn: async (payload: { file: File; sheetName?: string; template?: string; format?: string; columnOverrides?: Record<string, string> }) =>
      unwrap(await convertXLSX(payload.file, payload.sheetName, payload.template, payload.format, payload.columnOverrides)),
  });
}

export function useConvertTSVMutation() {
  return useMutation({
    mutationFn: async (payload: { file: File; template?: string; format?: string; columnOverrides?: Record<string, string> }) =>
      unwrap(await convertTSV(payload.file, payload.template, payload.format, payload.columnOverrides)),
  });
}

export function useConvertGoogleSheetMutation() {
  return useMutation({
    mutationFn: async (payload: { url: string; template?: string; gid?: string; format?: string; range?: string; columnOverrides?: Record<string, string> }) =>
      unwrap(await convertGoogleSheet(payload.url, payload.template, payload.gid, payload.format, payload.range, payload.columnOverrides)),
  });
}

export function useValidatePasteMutation() {
  return useMutation({
    mutationFn: async (payload: { pasteText: string; rules: ValidationRules }) =>
      unwrap(await validatePaste(payload.pasteText, payload.rules)),
  });
}

export function usePreviewTemplateMutation() {
  return useMutation({
    mutationFn: async (payload: { templateContent: string; sampleData?: string }) =>
      unwrap(await previewTemplate(payload.templateContent, payload.sampleData)),
  });
}

export function useDiffMDFlowMutation() {
  return useMutation({
    mutationFn: async (payload: { before: string; after: string }) =>
      unwrap(await diffMDFlow(payload.before, payload.after)),
  });
}

export function useAISuggestionsMutation() {
  return useMutation({
    mutationFn: async (payload: { pasteText: string; template?: string }) =>
      unwrap(await getAISuggestions(payload.pasteText, payload.template)),
  });
}

export function useFetchGoogleSheetMutation() {
  return useMutation({
    mutationFn: async (payload: { url: string }) =>
      unwrap(await fetchGoogleSheet(payload.url)),
  });
}

export function useGetGoogleSheetSheetsMutation() {
  return useMutation({
    mutationFn: async (payload: { url: string }) =>
      unwrap(await getGoogleSheetSheets(payload.url)),
  });
}
