"use client";

import { useMDFlowTemplatesQuery, useTemplateInfoQuery } from '@/lib/mdflowQueries';

interface Template {
  name: string;
  description: string;
  format: string;
}

interface TemplateInfo {
  variables: Array<{
    name: string;
    type: string;
    description: string;
  }>;
  functions: Array<{
    name: string;
    signature: string;
    description: string;
  }>;
}

interface UseTemplatesSyncState {
  templates: Template[];
  templateInfo: TemplateInfo | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useTemplatesSync(
  enabled: boolean = true
): UseTemplatesSyncState {
  const templatesQuery = useMDFlowTemplatesQuery();
  const templateInfoQuery = useTemplateInfoQuery(enabled);

  const refetch = async () => {
    await Promise.all([
      templatesQuery.refetch(),
      templateInfoQuery.refetch(),
    ]);
  };

  const loading = templatesQuery.isLoading || templateInfoQuery.isLoading;
  const error = templatesQuery.error?.message || templateInfoQuery.error?.message || null;

  const templates: Template[] = enabled && templatesQuery.data
    ? templatesQuery.data.map(t => ({
        name: t.name,
        description: t.description,
        format: t.format,
      }))
    : [];

  const templateInfo: TemplateInfo | null = enabled && templateInfoQuery.data
    ? templateInfoQuery.data
    : null;

  return {
    templates,
    templateInfo,
    loading,
    error,
    refetch,
  };
}
