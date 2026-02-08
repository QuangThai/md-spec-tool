'use client';

import { useState } from 'react';
import { ConversionResponse } from '@/lib/types';
import { FormatSelector } from './FormatSelector';
import { ConversionResultDisplay } from './ConversionResultDisplay';
import { Button } from '@/components/ui/Button';
import { Loader } from 'lucide-react';

interface ConversionWorkflowProps {
  onConvert?: (format: string) => Promise<ConversionResponse>;
}

export function ConversionWorkflow({ onConvert }: ConversionWorkflowProps) {
  const [format, setFormat] = useState<'spec' | 'table'>('spec');
  const [result, setResult] = useState<ConversionResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleConvert = async () => {
    if (!onConvert) return;

    setLoading(true);
    setError(null);

    try {
      const response = await onConvert(format);
      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Conversion failed');
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = () => {
    if (!result) return;
    navigator.clipboard.writeText(result.markdown);
  };

  const handleDownload = () => {
    if (!result) return;
    const element = document.createElement('a');
    const file = new Blob([result.markdown], { type: 'text/markdown' });
    element.href = URL.createObjectURL(file);
    element.download = `output.md`;
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
  };

  return (
    <div className="space-y-4">
      {/* Format Selection */}
      <div className="space-y-2">
        <label className="text-sm font-medium">Output Format</label>
        <FormatSelector value={format} onChange={setFormat} />
      </div>

      {/* Convert Button */}
      <Button onClick={handleConvert} disabled={loading} className="w-full">
        {loading && <Loader className="mr-2 h-4 w-4 animate-spin" />}
        {loading ? 'Converting...' : 'Convert'}
      </Button>

      {/* Error Display */}
      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Result Display */}
      {result && (
        <ConversionResultDisplay
          result={result}
          onCopy={handleCopy}
          onDownload={handleDownload}
        />
      )}
    </div>
  );
}
