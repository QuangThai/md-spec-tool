'use client';

import { useState, useEffect, useCallback } from 'react';
import { useMDFlowStore } from '@/lib/mdflowStore';
import {
  convertPaste,
  convertXLSX,
  getXLSXSheets,
  getMDFlowTemplates,
} from '@/lib/mdflowApi';

const highlights = [
  { title: 'No login', desc: 'Run instantly without an account.' },
  { title: 'Sheet aware', desc: 'Pick the correct worksheet quickly.' },
  { title: 'Clean export', desc: 'Copy or download MDFlow fast.' },
];

const steps = [
  { title: 'Import', desc: 'Paste or upload your sheet.' },
  { title: 'Configure', desc: 'Select sheet and template.' },
  { title: 'Generate', desc: 'Get MDFlow output instantly.' },
];

export default function MDFlowLanding() {
  const {
    mode,
    pasteText,
    file,
    sheets,
    selectedSheet,
    template,
    mdflowOutput,
    warnings,
    meta,
    loading,
    error,
    setMode,
    setPasteText,
    setFile,
    setSheets,
    setSelectedSheet,
    setTemplate,
    setResult,
    setLoading,
    setError,
    reset,
  } = useMDFlowStore();

  const [templates, setTemplates] = useState<string[]>(['default']);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    getMDFlowTemplates().then((res) => {
      if (res.data?.templates) {
        setTemplates(res.data.templates);
      }
    });
  }, []);

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (!selectedFile) return;

      setFile(selectedFile);
      setLoading(true);
      setError(null);

      const result = await getXLSXSheets(selectedFile);

      setLoading(false);

      if (result.error) {
        setError(result.error);
      } else if (result.data) {
        setSheets(result.data.sheets);
        setSelectedSheet(result.data.active_sheet);
      }
    },
    [setFile, setLoading, setError, setSheets, setSelectedSheet]
  );

  const handleConvert = useCallback(async () => {
    setLoading(true);
    setError(null);

    let result;

    if (mode === 'paste') {
      if (!pasteText.trim()) {
        setError('Please paste some data first');
        setLoading(false);
        return;
      }
      result = await convertPaste(pasteText, template);
    } else {
      if (!file) {
        setError('Please upload a file first');
        setLoading(false);
        return;
      }
      result = await convertXLSX(file, selectedSheet, template);
    }

    setLoading(false);

    if (result.error) {
      setError(result.error);
    } else if (result.data) {
      setResult(result.data.mdflow, result.data.warnings, result.data.meta);
    }
  }, [
    mode,
    pasteText,
    file,
    selectedSheet,
    template,
    setLoading,
    setError,
    setResult,
  ]);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(mdflowOutput);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [mdflowOutput]);

  const handleDownload = useCallback(() => {
    const blob = new Blob([mdflowOutput], { type: 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'spec.mdflow.md';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [mdflowOutput]);

  return (
    <div className="flex flex-col gap-12">
      <section className="relative overflow-hidden rounded-3xl border border-slate-200/70 bg-white/70 p-8 shadow-[0_20px_50px_rgba(15,23,42,0.08)] sm:p-12">
        <div className="pointer-events-none absolute -right-24 -top-20 h-60 w-60 rounded-full bg-sky-200/35 blur-3xl animate-float-slow" />
        <div className="pointer-events-none absolute -bottom-16 left-8 h-52 w-52 rounded-full bg-amber-200/35 blur-3xl animate-float-slow" />
        <div className="relative grid gap-10 lg:grid-cols-[1.2fr_0.8fr] lg:items-center animate-fade-up">
          <div className="flex flex-col gap-6">
            <span className="pill">MDFlow Studio</span>
            <div className="space-y-4">
              <h1 className="text-4xl font-semibold leading-tight text-slate-900 sm:text-5xl">
                Spreadsheet to MDFlow, without the clutter.
              </h1>
              <p className="text-lg text-slate-600">
                Paste a table or upload Excel, pick a template, and generate a clean MDFlow spec.
              </p>
            </div>
            <div className="flex flex-wrap gap-3">
              <a href="#converter" className="btn-primary">
                Start converting
              </a>
              <a href="#output" className="btn-ghost">
                Jump to output
              </a>
            </div>
            <div className="grid gap-3 sm:grid-cols-3">
              {highlights.map((item) => (
                <div key={item.title} className="rounded-2xl border border-slate-200/70 bg-white/80 p-4">
                  <p className="text-sm font-semibold text-slate-900">{item.title}</p>
                  <p className="mt-1 text-sm text-slate-500">{item.desc}</p>
                </div>
              ))}
            </div>
          </div>
          <div className="surface-muted p-6 sm:p-8">
            <div className="flex items-center justify-between">
              <p className="label">Conversion flow</p>
              <span className="text-xs text-slate-400">3 steps</span>
            </div>
            <div className="mt-6 grid gap-4">
              {steps.map((step, index) => (
                <div key={step.title} className="flex items-start gap-3">
                  <span className="mt-1 inline-flex h-7 w-7 items-center justify-center rounded-full bg-slate-900 text-xs font-semibold text-white">
                    {index + 1}
                  </span>
                  <div>
                    <p className="text-sm font-semibold text-slate-900">{step.title}</p>
                    <p className="text-sm text-slate-500">{step.desc}</p>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-6 rounded-2xl border border-slate-200/70 bg-white/80 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Output</p>
              <p className="mt-1 text-sm font-semibold text-slate-900">MDFlow spec with metadata</p>
              <p className="mt-1 text-sm text-slate-500">Warnings and row totals included.</p>
            </div>
          </div>
        </div>
      </section>

      <section id="converter" className="space-y-6">
        <div className="surface p-6 sm:p-8">
          <div className="flex flex-col gap-2">
            <span className="pill">MDFlow Converter</span>
            <h2 className="text-3xl font-semibold text-slate-900">Convert your sheet</h2>
            <p className="text-sm text-slate-600">Import data, pick a template, generate MDFlow.</p>
          </div>
        </div>

        {error && (
          <div className="rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 text-sm text-rose-700">
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div className="surface p-6">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-sm font-semibold text-slate-900">Input</p>
                <p className="text-xs text-slate-500">Paste table data or upload a file.</p>
              </div>
              <div className="flex rounded-full bg-slate-100 p-1">
                <button
                  onClick={() => setMode('paste')}
                  className={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${
                    mode === 'paste'
                      ? 'bg-slate-900 text-white'
                      : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  Paste
                </button>
                <button
                  onClick={() => setMode('xlsx')}
                  className={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${
                    mode === 'xlsx'
                      ? 'bg-slate-900 text-white'
                      : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  Upload
                </button>
              </div>
            </div>

            {mode === 'paste' ? (
              <div className="mt-4">
                <label className="label mb-2 block">Paste table data</label>
                <textarea
                  value={pasteText}
                  onChange={(e) => setPasteText(e.target.value)}
                  placeholder="Copy cells from Google Sheet and paste here..."
                  className="input min-h-[220px] w-full font-mono text-sm"
                />
                <p className="mt-2 text-xs text-slate-500">
                  Supports TSV (from Google Sheets) and CSV formats.
                </p>
              </div>
            ) : (
              <div className="mt-4 space-y-4">
                <div>
                  <label className="label mb-2 block">Upload Excel file</label>
                  <input
                    type="file"
                    accept=".xlsx,.xls"
                    onChange={handleFileChange}
                    className="input w-full file:mr-3 file:rounded-full file:border-0 file:bg-slate-900 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-white"
                  />
                </div>

                {sheets.length > 0 && (
                  <div>
                    <label className="label mb-2 block">Select sheet</label>
                    <select
                      value={selectedSheet}
                      onChange={(e) => setSelectedSheet(e.target.value)}
                      className="input w-full"
                    >
                      {sheets.map((sheet) => (
                        <option key={sheet} value={sheet}>
                          {sheet}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
              </div>
            )}

            <div className="mt-6 grid gap-4 border-t border-slate-200/70 pt-4">
              <div>
                <label className="label mb-2 block">Output template</label>
                <select
                  value={template}
                  onChange={(e) => setTemplate(e.target.value)}
                  className="input w-full"
                >
                  {templates.map((t) => (
                    <option key={t} value={t}>
                      {t.charAt(0).toUpperCase() + t.slice(1).replace(/-/g, ' ')}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex flex-wrap gap-3 mt-10">
                <button onClick={handleConvert} disabled={loading} className="btn-primary flex-1">
                  {loading ? 'Converting...' : 'Convert to MDFlow'}
                </button>
                <button onClick={reset} className="btn-ghost px-6">
                  Reset
                </button>
              </div>
            </div>
          </div>

          <div id="output" className="space-y-6">
            {mdflowOutput && (
              <>
                <div className="surface p-6">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <h3 className="text-lg font-semibold text-slate-900">MDFlow Output</h3>
                    <div className="flex gap-2">
                      <button
                        onClick={handleCopy}
                        className="btn-ghost text-sm"
                        title="Copy to clipboard"
                      >
                        {copied ? 'Copied' : 'Copy'}
                      </button>
                      <button
                        onClick={handleDownload}
                        className="btn-ghost text-sm"
                        title="Download as .md file"
                      >
                        Download
                      </button>
                    </div>
                  </div>
                  <div className="mt-4 max-h-[500px] overflow-y-auto rounded-xl border border-slate-200/70 bg-slate-50 p-4">
                    <pre className="whitespace-pre-wrap text-sm text-slate-700 font-mono">
                      {mdflowOutput}
                    </pre>
                  </div>
                </div>

                {warnings && warnings.length > 0 && (
                  <div className="surface-muted p-4">
                    <h4 className="text-sm font-semibold text-amber-700">Warnings</h4>
                    <ul className="mt-2 space-y-1">
                      {warnings.map((warning, i) => (
                        <li key={i} className="text-sm text-amber-600">
                          • {warning}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {meta && (
                  <div className="surface p-6">
                    <h4 className="text-sm font-semibold text-slate-900">Conversion details</h4>
                    <dl className="mt-3 space-y-2 text-sm">
                      <div className="flex justify-between">
                        <dt className="text-slate-500">Total rows:</dt>
                        <dd className="font-medium text-slate-900">{meta.total_rows}</dd>
                      </div>
                      <div className="flex justify-between">
                        <dt className="text-slate-500">Header row:</dt>
                        <dd className="font-medium text-slate-900">{meta.header_row + 1}</dd>
                      </div>
                      {meta.sheet_name && (
                        <div className="flex justify-between">
                          <dt className="text-slate-500">Sheet:</dt>
                          <dd className="font-medium text-slate-900">{meta.sheet_name}</dd>
                        </div>
                      )}
                      {meta.rows_by_feature && Object.keys(meta.rows_by_feature).length > 0 && (
                        <div className="pt-2">
                          <dt className="mb-1 text-slate-500">By feature:</dt>
                          <dd className="space-y-1">
                            {Object.entries(meta.rows_by_feature).map(([feature, count]) => (
                              <div key={feature} className="flex justify-between text-xs">
                                <span className="text-slate-600">{feature}</span>
                                <span className="font-medium">{count}</span>
                              </div>
                            ))}
                          </dd>
                        </div>
                      )}
                    </dl>
                  </div>
                )}
              </>
            )}

            {!mdflowOutput && (
              <div className="surface-muted flex h-[300px] items-center justify-center p-6">
                <div className="text-center">
                  <p className="text-lg font-medium text-slate-400">Output will appear here</p>
                  <p className="mt-1 text-sm text-slate-400">
                    Paste data or upload a file, then click Convert.
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </section>
    </div>
  );
}
