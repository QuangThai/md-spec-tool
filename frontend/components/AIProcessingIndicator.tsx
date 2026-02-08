'use client';

import { Brain, AlertCircle } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';

interface AIProcessingIndicatorProps {
  mode: 'off' | 'shadow' | 'on';
  degraded?: boolean;
}

export function AIProcessingIndicator({
  mode,
  degraded,
}: AIProcessingIndicatorProps) {
  const getModeLabel = () => {
    switch (mode) {
      case 'off':
        return 'AI Off';
      case 'shadow':
        return 'Shadow Mode';
      case 'on':
        return 'AI Enabled';
      default:
        return 'Unknown';
    }
  };

  const getModeColor = () => {
    if (degraded) return 'destructive';
    switch (mode) {
      case 'off':
        return 'secondary';
      case 'shadow':
        return 'outline';
      case 'on':
        return 'default';
      default:
        return 'secondary';
    }
  };

  return (
    <div className="flex items-center gap-2">
      {mode === 'on' ? (
        <Brain className="h-4 w-4 text-blue-600 animate-pulse" />
      ) : (
        <Brain className="h-4 w-4 text-muted-foreground" />
      )}
      <Badge className={getModeColor()}>
        {getModeLabel()}
      </Badge>
      {degraded && (
        <div className="flex items-center gap-1 text-xs text-yellow-600">
          <AlertCircle className="h-3 w-3" />
          <span>Fallback active</span>
        </div>
      )}
    </div>
  );
}
