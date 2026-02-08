'use client';

import { ArrowRight, AlertTriangle, ChevronDown } from 'lucide-react';
import { ConversionResponse } from '@/lib/types';
import { Badge } from '@/components/ui/Badge';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/Collapsible';

interface ColumnMappingDisplayProps {
  mappings: ConversionResponse['column_mappings'];
}

export function ColumnMappingDisplay({ mappings }: ColumnMappingDisplayProps) {
  return (
    <Collapsible defaultOpen={false}>
      <CollapsibleTrigger className="flex items-center gap-2 text-sm font-medium hover:underline cursor-pointer w-full">
        <ChevronDown className="h-4 w-4" />
        <span>Column Mappings ({mappings.meta.mapped_columns} mapped)</span>
        <Badge
          className={`ml-auto ${mappings.meta.avg_confidence > 0.8 ? 'bg-green-600' : 'bg-orange-600'}`}
        >
          {Math.round(mappings.meta.avg_confidence * 100)}% confidence
        </Badge>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="mt-3 space-y-2 border-l-2 border-muted-foreground pl-4">
          {/* Canonical Field Mappings */}
          {mappings.canonical_fields.map((m) => (
            <div key={m.column_index} className="flex items-center gap-2 text-sm">
              <span className="text-muted-foreground font-medium max-w-[150px] truncate" title={m.source_header}>
                {m.source_header}
              </span>
              <ArrowRight className="h-3 w-3 text-muted-foreground flex-shrink-0" />
              <code className="bg-muted px-2 py-0.5 rounded text-xs font-mono flex-shrink-0">
                {m.canonical_name}
              </code>
              {m.confidence < 0.7 && (
                <AlertTriangle className="h-3 w-3 text-yellow-600 flex-shrink-0" />
              )}
              {m.confidence > 0 && m.confidence <= 1 && (
                <span className="text-xs text-muted-foreground">
                  {Math.round(m.confidence * 100)}%
                </span>
              )}
            </div>
          ))}

          {/* Extra/Unmapped Columns */}
          {mappings.extra_columns.length > 0 && (
            <div className="mt-2 pt-2 border-t border-muted-foreground/30">
              <div className="text-xs font-medium text-muted-foreground mb-1">
                Extra Columns
              </div>
              {mappings.extra_columns.map((e) => (
                <div
                  key={e.column_index}
                  className="flex items-center gap-2 text-xs text-muted-foreground"
                >
                  <span className="font-medium">{e.name}</span>
                  <span className="italic">({e.semantic_role})</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
