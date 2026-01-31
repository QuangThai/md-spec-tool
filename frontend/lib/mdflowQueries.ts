import { useMutation, useQuery } from "@tanstack/react-query";
import {
  convertGoogleSheet,
  convertPaste,
  convertTSV,
  convertXLSX,
  diffMDFlow,
  fetchGoogleSheet,
  getAISuggestions,
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

const queryKeys = {
  templates: ["mdflow", "templates"] as const,
  templateInfo: ["mdflow", "template-info"] as const,
  templateContent: (name: string) => ["mdflow", "template", name] as const,
  previewPaste: (hash: string) => ["mdflow", "preview", "paste", hash] as const,
  previewTSV: (fileKey: string) => ["mdflow", "preview", "tsv", fileKey] as const,
  previewXLSX: (fileKey: string, sheet: string) =>
    ["mdflow", "preview", "xlsx", fileKey, sheet] as const,
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
        if (a === "default") return -1;
        if (b === "default") return 1;
        return 0;
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

export function usePreviewPasteQuery(pasteText: string, enabled: boolean) {
  const hash = hashString(pasteText.trim());
  return useQuery({
    queryKey: queryKeys.previewPaste(hash),
    queryFn: async () => unwrap(await previewPaste(pasteText)),
    enabled: enabled && pasteText.trim().length > 0,
    gcTime: 2 * 60 * 1000,
  });
}

export function usePreviewTSVQuery(file: File | null, enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.previewTSV(fileKey(file)),
    queryFn: async () => unwrap(await previewTSV(file!)),
    enabled: enabled && Boolean(file),
    gcTime: 2 * 60 * 1000,
  });
}

export function usePreviewTSVMutation() {
  return useMutation({
    mutationFn: async (file: File) => unwrap(await previewTSV(file)),
  });
}

export function usePreviewXLSXQuery(
  file: File | null,
  sheetName: string,
  enabled: boolean
) {
  return useQuery({
    queryKey: queryKeys.previewXLSX(fileKey(file), sheetName),
    queryFn: async () => unwrap(await previewXLSX(file!, sheetName)),
    enabled: enabled && Boolean(file) && Boolean(sheetName),
    gcTime: 2 * 60 * 1000,
  });
}

export function usePreviewXLSXMutation() {
  return useMutation({
    mutationFn: async (payload: { file: File; sheetName?: string }) =>
      unwrap(await previewXLSX(payload.file, payload.sheetName)),
  });
}

export function useGetXLSXSheetsMutation() {
  return useMutation({
    mutationFn: async (file: File) => unwrap(await getXLSXSheets(file)),
  });
}

export function useConvertPasteMutation() {
  return useMutation({
    mutationFn: async (payload: { pasteText: string; template?: string }) =>
      unwrap(await convertPaste(payload.pasteText, payload.template)),
  });
}

export function useConvertXLSXMutation() {
  return useMutation({
    mutationFn: async (payload: { file: File; sheetName?: string; template?: string }) =>
      unwrap(await convertXLSX(payload.file, payload.sheetName, payload.template)),
  });
}

export function useConvertTSVMutation() {
  return useMutation({
    mutationFn: async (payload: { file: File; template?: string }) =>
      unwrap(await convertTSV(payload.file, payload.template)),
  });
}

export function useConvertGoogleSheetMutation() {
  return useMutation({
    mutationFn: async (payload: { url: string; template?: string }) =>
      unwrap(await convertGoogleSheet(payload.url, payload.template)),
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
