'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { uploadFile, fetchAPI } from '@/lib/api';
import { useAuthStore, useConversionStore, TableData } from '@/lib/store';

export default function ConverterPage() {
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [preview, setPreview] = useState<TableData | null>(null);
  const [templates, setTemplates] = useState<any[]>([]);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>('default');
  const [converting, setConverting] = useState(false);
  const [markdown, setMarkdown] = useState<string | null>(null);
  const [isMounted, setIsMounted] = useState(false);

  const router = useRouter();
  const { isLoggedIn, token } = useAuthStore();
  const { setTableData, setMarkdown: storeMarkdown } = useConversionStore();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!isMounted) return;
    if (!isLoggedIn) {
      router.push('/auth/login');
      return;
    }
  }, [isMounted, isLoggedIn, router]);

  useEffect(() => {
    if (!isMounted) return;
    if (isLoggedIn) {
      loadTemplates();
    }
  }, [isMounted, isLoggedIn]);

  const loadTemplates = async () => {
    const result = await fetchAPI<any[]>('/templates');
    if (result.data && Array.isArray(result.data)) {
      setTemplates(result.data);
    }
  };

  if (!isMounted) {
    return <div className="surface p-6">Loading...</div>;
  }

  if (!isLoggedIn) {
    return <p>Redirecting to login...</p>;
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (!selectedFile) return;

    setFile(selectedFile);
    setLoading(true);
    setError('');
    setMarkdown(null);

    const result = await uploadFile('/import/excel', selectedFile, token || undefined);

    setLoading(false);

    if (result.error) {
      setError(result.error);
    } else {
      setPreview(result.data);
      setTableData(result.data);
    }
  };

  const handleConvert = async () => {
    if (!preview) return;

    setConverting(true);
    setError('');

    const result = await fetchAPI<{ markdown: string }>('/convert/markdown', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        table_data: preview,
        template_id: selectedTemplateId === 'default' ? null : selectedTemplateId,
      }),
    });

    setConverting(false);

    if (result.error) {
      setError(result.error);
    } else if (result.data) {
      setMarkdown(result.data.markdown);
      storeMarkdown(result.data.markdown);
    }
  };

  const handleSaveSpec = async () => {
    if (!markdown) return;

    const title = prompt('Enter document title:');
    if (!title) return;

    setConverting(true);
    const result = await fetchAPI<any>('/spec', {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        title,
        content: markdown,
        template_id: selectedTemplateId === 'default' ? null : selectedTemplateId,
      }),
    });

    setConverting(false);

    if (result.error) {
      setError(result.error);
    } else if (result.data?.id) {
      alert('Document saved successfully!');
      router.push('/documents');
    }
  };

  return (
    <div className="space-y-8">
      <div className="surface p-8">
        <div className="flex flex-col gap-2">
          <p className="pill">Converter</p>
          <h2 className="text-3xl font-semibold text-slate-900">Excel to Markdown</h2>
          <p className="text-sm text-slate-600">
            Upload, preview, and generate a clean spec without manual formatting.
          </p>
        </div>
      </div>

      {error && (
        <div className="rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 text-sm text-rose-700">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="space-y-6">
          <div className="surface p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">Step 1 - Upload Excel</h3>
              <span className="text-xs text-slate-400">.xlsx or .xls</span>
            </div>
            <div className="mt-4 flex flex-col gap-3">
              <input
                type="file"
                accept=".xlsx,.xls"
                onChange={handleFileUpload}
                disabled={loading}
                className="input file:mr-3 file:rounded-full file:border-0 file:bg-slate-900 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-white"
              />
              {loading && <p className="text-sm text-slate-500">Uploading and parsing...</p>}
              {!loading && !preview && (
                <p className="text-sm text-slate-500">We will show a preview once the file is uploaded.</p>
              )}
            </div>
          </div>

          {preview && (
            <div className="surface p-6">
              <h3 className="text-lg font-semibold text-slate-900">Step 2 - Choose template</h3>
              <p className="mt-2 text-sm text-slate-500">Pick a Markdown structure for the output.</p>

              <div className="mt-4 space-y-4">
                <select
                  value={selectedTemplateId}
                  onChange={(e) => setSelectedTemplateId(e.target.value)}
                  className="input"
                >
                  <option value="default">Default Template</option>
                  {templates.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name}
                    </option>
                  ))}
                </select>

                <button onClick={handleConvert} disabled={converting} className="btn-primary w-full">
                  {converting ? 'Converting...' : 'Generate Markdown'}
                </button>

                {markdown && (
                  <button onClick={handleSaveSpec} className="btn-secondary w-full">
                    Save Document
                  </button>
                )}
              </div>
            </div>
          )}
        </div>

        <div className="space-y-6">
          {preview && (
            <div className="surface p-6">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h4 className="text-lg font-semibold text-slate-900">Table Preview</h4>
                <span className="text-xs text-slate-400">{preview.sheet_name}</span>
              </div>
              <div className="mt-4 overflow-x-auto rounded-xl border border-slate-200/70 bg-white">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50">
                    <tr>
                      {preview.headers.map((header: string) => (
                        <th key={header} className="border-b border-slate-200/70 px-3 py-2 text-left font-semibold text-slate-600">
                          {header}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {preview.rows.slice(0, 6).map((row: any, idx: number) => (
                      <tr key={idx} className="odd:bg-white even:bg-slate-50/60">
                        {preview.headers.map((header: string) => (
                          <td key={header} className="border-b border-slate-200/70 px-3 py-2 text-slate-700">
                            {row[header]}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {preview.rows.length > 6 && (
                <p className="mt-3 text-sm text-slate-500">+ {preview.rows.length - 6} more rows</p>
              )}
            </div>
          )}

          {markdown && (
            <div className="surface p-6">
              <div className="flex items-center justify-between">
                <h4 className="text-lg font-semibold text-slate-900">Markdown Output</h4>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(markdown);
                    alert('Copied to clipboard!');
                  }}
                  className="btn-ghost"
                >
                  Copy
                </button>
              </div>
              <div className="mt-4 max-h-72 overflow-y-auto rounded-xl border border-slate-200/70 bg-slate-50 p-4">
                <pre className="text-sm whitespace-pre-wrap text-slate-700">{markdown}</pre>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
