'use client';

import { useState, useEffect } from 'react';

interface TemplateInfo {
  name: string;
  formats: string[];
  description?: string;
}

/**
 * Hook for managing output formats
 * Fetches supported formats from /api/mdflow/templates
 */
export function useMDFlowTemplates() {
  const [formats, setFormats] = useState<string[]>([]);
  const [templates, setTemplates] = useState<TemplateInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTemplates = async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/mdflow/templates');
        if (!response.ok) {
          throw new Error(`Failed to fetch templates: ${response.status}`);
        }
        const data = await response.json();
        
        // Extract unique formats from all templates
        const uniqueFormats = Array.from(
          new Set((data.templates?.flatMap((t: TemplateInfo) => t.formats) || []) as string[])
        ) as string[];
        
        setTemplates(data.templates || []);
        setFormats(uniqueFormats);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
        setFormats([]); // fallback to empty
        setTemplates([]);
      } finally {
        setLoading(false);
      }
    };

    fetchTemplates();
  }, []);

  return {
    formats,
    templates,
    loading,
    error,
  };
}
