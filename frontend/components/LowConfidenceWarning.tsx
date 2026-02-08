'use client';

import { AlertTriangle } from 'lucide-react';
import { ColumnMapping } from '@/lib/types';

interface LowConfidenceWarningProps {
  mappings: ColumnMapping[];
}

export function LowConfidenceWarning({ mappings }: LowConfidenceWarningProps) {
  const lowConfidenceMappings = mappings.filter((m) => m.confidence < 0.7);

  if (lowConfidenceMappings.length === 0) return null;

  return (
    <div className="flex items-start gap-3 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
      <AlertTriangle className="h-5 w-5 text-yellow-600 flex-shrink-0 mt-0.5" />
      <div className="flex-1">
        <h3 className="font-medium text-yellow-900">
          Low Confidence Mappings
        </h3>
        <p className="text-sm text-yellow-800 mt-1">
          {lowConfidenceMappings.length} column{' '}
          {lowConfidenceMappings.length === 1 ? 'mapping' : 'mappings'} have low confidence. Please
          review and verify the mappings are correct:
        </p>
        <ul className="mt-2 space-y-1">
          {lowConfidenceMappings.map((m) => (
            <li key={m.column_index} className="text-sm text-yellow-800">
              <span className="font-mono bg-yellow-100 px-1 rounded">
                {m.source_header}
              </span>{' '}
              →{' '}
              <span className="font-mono bg-yellow-100 px-1 rounded">
                {m.canonical_name}
              </span>{' '}
              ({Math.round(m.confidence * 100)}%)
              {m.reasoning && (
                <span className="block text-xs text-yellow-700 mt-0.5">
                  {m.reasoning}
                </span>
              )}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
