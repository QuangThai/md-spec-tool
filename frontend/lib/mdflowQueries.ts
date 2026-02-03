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
  previewPaste,
  previewTemplate,
  previewTSV,
  previewXLSX,
  validatePaste,
} from "@/lib/mdflowApi";
import { ApiResult, ValidationRules } from "@/lib/types";
import { useMutation, useQuery } from "@tanstack/react-query";

const queryKeys = {
  templates: ["mdflow", "templates"] as const,
  templateInfo: ["mdflow", "template-info"] as const,
  templateContent: (name: string) => ["mdflow", "template", name] as const,
  previewPaste: (hash: string, template?: string) =>
    ["mdflow", "preview", "paste", hash, template ?? ""] as const,
  previewTSV: (fileKey: string, template?: string) =>
    ["mdflow", "preview", "tsv", fileKey, template ?? ""] as const,
  previewXLSX: (fileKey: string, sheet: string, template?: string) =>
    ["mdflow", "preview", "xlsx", fileKey, sheet, template ?? ""] as const,
  previewGoogleSheet: (url: string, gid: string, template?: string) =>
    ["mdflow", "preview", "gsheet", url, gid, template ?? ""] as const,
};

/**
 * Unwrap ApiResult - throw on error, return data on success
 */
function unwrap<T>(result: ApiResult<T>): T {
  if (result.error) {
    throw new Error(result.error);
  }
  if (!result.data) {
    throw new Error("No data returned");
  }
  return result.data;
}

function fileKey(file: File | null | undefined): string {
  if (!file) return "no-file";
  return `${file.name}:${file.size}:${file.lastModified}`;
}

function hashString(input: string): string {
  let hash = 5381;
  for (let i = 0; i < input.length; i += 1) {
    hash = (hash * 33) ^ input.charCodeAt(i);
  }
  return Math.abs(hash).toString(36);
}

export function useMDFlowTemplatesQuery() {
  return useQuery({
    queryKey: queryKeys.templates,
    queryFn: async () => unwrap(await getMDFlowTemplates()),
    staleTime: 5 * 60 * 1000,
    select: (data) => {
      const sorted = [...data.templates].sort((a, b) => {
        if (a.name === "test_spec_v1" || a.name === "default") return -1;
        if (b.name === "test_spec_v1" || b.name === "default") return 1;
        return (a.name || "").localeCompare(b.name || "");
      });
      return sorted;
    },
  });
}

export function useTemplateInfoQuery(enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.templateInfo,
    queryFn: async () => unwrap(await getTemplateInfo()),
    enabled,
    staleTime: 5 * 60 * 1000,
  });
}

export function useTemplateContentQuery(name: string, enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.templateContent(name),
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
    queryKey: queryKeys.previewPaste(hash, template),
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
    queryKey: queryKeys.previewTSV(fileKey(file), template),
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
    queryKey: queryKeys.previewXLSX(fileKey(file), sheetName, template),
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
 * Preview for Google Sheet: fetches sheet data by url+gid, then runs preview (same as paste).
 * Refetches when url, gid, or template changes so changing sheet shows the correct table.
 */
export function usePreviewGoogleSheetQuery(
  url: string,
  gid: string,
  enabled: boolean,
  template?: string
) {
  return useQuery({
    queryKey: queryKeys.previewGoogleSheet(url, gid, template),
    queryFn: async () => {
      const sheetResult = await fetchGoogleSheet(url, gid);
      const sheetData = unwrap(sheetResult);
      const previewResult = await previewPaste(sheetData.data, template);
      return unwrap(previewResult);
    },
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
    mutationFn: async (payload: { pasteText: string; template?: string; format?: string }) =>
      unwrap(await convertPaste(payload.pasteText, payload.template, payload.format)),
  });
}

export function useConvertXLSXMutation() {
  return useMutation({
    mutationFn: async (payload: { file: File; sheetName?: string; template?: string; format?: string }) =>
      unwrap(await convertXLSX(payload.file, payload.sheetName, payload.template, payload.format)),
  });
}

export function useConvertTSVMutation() {
  return useMutation({
    mutationFn: async (payload: { file: File; template?: string; format?: string }) =>
      unwrap(await convertTSV(payload.file, payload.template, payload.format)),
  });
}

export function useConvertGoogleSheetMutation() {
  return useMutation({
    mutationFn: async (payload: { url: string; template?: string; gid?: string; format?: string }) =>
      unwrap(await convertGoogleSheet(payload.url, payload.template, payload.gid, payload.format)),
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
      diffMDFlow(payload.before, payload.after),
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
