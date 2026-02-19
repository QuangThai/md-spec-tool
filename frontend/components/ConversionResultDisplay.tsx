'use client';

import { useState, useCallback } from 'react';
import { Copy, Download, AlertTriangle } from 'lucide-react';
import { ConversionResponse, ColumnMapping } from '@/lib/types';
import { Button } from '@/components/ui/Button';
import { AIProcessingIndicator } from './AIProcessingIndicator';
import { ColumnMappingDisplay } from './ColumnMappingDisplay';
import { ColumnMappingReview } from './ColumnMappingReview';
import { DegradedModeWarning } from './DegradedModeWarning';
import { LowConfidenceWarning } from './LowConfidenceWarning';

interface ConversionResultDisplayProps {
  result: ConversionResponse;
  onCopy?: () => void;
  onDownload?: () => void;
  onReviewMappings?: (
    columnOverrides?: Record<string, string>
  ) => Promise<ConversionResponse>;
}

export function ConversionResultDisplay({
  result,
  onCopy,
  onDownload,
  onReviewMappings,
}: ConversionResultDisplayProps) {
  const [isReviewOpen, setIsReviewOpen] = useState(false);
  const [isRetrying, setIsRetrying] = useState(false);
  const [currentResult, setCurrentResult] = useState(result);

  const lowConfidenceMappings = currentResult.column_mappings.canonical_fields.filter(
    (m) => m.confidence < 0.7
  );

  const needsReview = (result as any).needs_review === true;
  const avgConfidence = currentResult.column_mappings.meta.avg_confidence;

  const handleRetryMapping = useCallback(
    async (columnOverrides: Record<string, string>) => {
      if (!onReviewMappings) return;
      setIsRetrying(true);
      try {
        const updatedResult = await onReviewMappings(columnOverrides);
        setCurrentResult(updatedResult);
      } finally {
        setIsRetrying(false);
      }
    },
    [onReviewMappings]
  );

  const handleConfirmMappings = useCallback(
    (updatedMappings: ColumnMapping[]) => {
      // Optionally update result with confirmed mappings
      // This could trigger a re-run with confirmed mappings
      setCurrentResult({
        ...currentResult,
        column_mappings: {
          ...currentResult.column_mappings,
          canonical_fields: updatedMappings,
        },
      });
    },
    [currentResult]
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

      {/* Needs Review Alert */}
      {needsReview && (
        <div className="flex items-start gap-3 p-4 bg-amber-50 border border-amber-200 rounded-md">
          <AlertTriangle className="h-5 w-5 text-amber-600 flex-shrink-0 mt-0.5" />
          <div className="flex-1">
            <h3 className="font-medium text-amber-900">
              Quality Gate: Review Recommended
            </h3>
            <p className="text-sm text-amber-800 mt-1">
              Mapping confidence is {Math.round(avgConfidence * 100)}% (below 65% threshold).
              Please review the column mappings to ensure they are correct.
            </p>
            <Button
              variant="secondary"
              size="sm"
              className="mt-3 bg-white hover:bg-amber-50"
              onClick={() => setIsReviewOpen(true)}
              disabled={isRetrying}
            >
              Review Mappings
            </Button>
          </div>
        </div>
      )}

      {/* Low confidence warnings */}
      {!needsReview && lowConfidenceMappings.length > 0 && (
        <LowConfidenceWarning mappings={lowConfidenceMappings} />
      )}

      {/* Mapping Review Modal */}
      {onReviewMappings && (
        <ColumnMappingReview
          mappings={currentResult.column_mappings.canonical_fields}
          sourceHeaders={currentResult.column_mappings.canonical_fields.map(
            (m) => m.source_header
          )}
          warnings={[]}
          avgConfidence={avgConfidence}
          onConfirm={handleConfirmMappings}
          onRetry={handleRetryMapping}
          isOpen={isReviewOpen}
          onClose={() => setIsReviewOpen(false)}
        />
      )}

      {/* Column mapping display */}
      <ColumnMappingDisplay mappings={currentResult.column_mappings} />

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
            {currentResult.markdown}
          </pre>
        </div>
      </div>

      {/* Warnings */}
      {currentResult.warnings && currentResult.warnings.length > 0 && (
         <div className="space-y-1">
           <h3 className="text-xs font-medium text-muted-foreground">
             Warnings
           </h3>
           {currentResult.warnings.map((warning, i) => (
             <div key={i} className="text-xs text-yellow-700 bg-yellow-50 p-2 rounded">
               {warning}
             </div>
           ))}
         </div>
       )}
    </div>
  );
}
