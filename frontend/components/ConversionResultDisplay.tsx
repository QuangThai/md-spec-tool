'use client';

import { Copy, Download } from 'lucide-react';
import { ConversionResponse } from '@/lib/types';
import { Button } from '@/components/ui/Button';
import { AIProcessingIndicator } from './AIProcessingIndicator';
import { ColumnMappingDisplay } from './ColumnMappingDisplay';
import { DegradedModeWarning } from './DegradedModeWarning';
import { LowConfidenceWarning } from './LowConfidenceWarning';

interface ConversionResultDisplayProps {
  result: ConversionResponse;
  onCopy?: () => void;
  onDownload?: () => void;
}

export function ConversionResultDisplay({
  result,
  onCopy,
  onDownload,
}: ConversionResultDisplayProps) {
  const lowConfidenceMappings = result.column_mappings.canonical_fields.filter(
    (m) => m.confidence < 0.7
  );

  return (
    <div className="space-y-4">
      {/* Header with AI indicator */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{result.meta.title}</h2>
          <p className="text-sm text-muted-foreground">
            {result.meta.total_items} items • Schema v{result.meta.schema_version}
          </p>
        </div>
        <AIProcessingIndicator
          mode={result.meta.ai_mode}
          degraded={result.meta.degraded}
        />
      </div>

      {/* Degraded mode warning */}
      {result.meta.degraded && <DegradedModeWarning />}

      {/* Low confidence warnings */}
      {lowConfidenceMappings.length > 0 && (
        <LowConfidenceWarning mappings={lowConfidenceMappings} />
      )}

      {/* Column mapping display */}
      <ColumnMappingDisplay mappings={result.column_mappings} />

      {/* Output preview */}
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium">Output ({result.format})</h3>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="secondary"
              onClick={onCopy}
              className="gap-2"
            >
              <Copy className="h-4 w-4" />
              Copy
            </Button>
            <Button
              size="sm"
              variant="secondary"
              onClick={onDownload}
              className="gap-2"
            >
              <Download className="h-4 w-4" />
              Download
            </Button>
          </div>
        </div>
        <div className="bg-muted p-4 rounded-md max-h-96 overflow-y-auto border border-muted-foreground/20">
          <pre className="text-xs font-mono whitespace-pre-wrap break-words">
            {result.markdown}
          </pre>
        </div>
      </div>

      {/* Warnings */}
      {result.warnings && result.warnings.length > 0 && (
        <div className="space-y-1">
          <h3 className="text-xs font-medium text-muted-foreground">
            Warnings
          </h3>
          {result.warnings.map((warning, i) => (
            <div key={i} className="text-xs text-yellow-700 bg-yellow-50 p-2 rounded">
              {warning}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
