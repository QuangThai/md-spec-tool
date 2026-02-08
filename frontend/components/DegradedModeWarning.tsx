'use client';

import { AlertTriangle, X } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/components/ui/Button';

interface DegradedModeWarningProps {
  onDismiss?: () => void;
}

export function DegradedModeWarning({ onDismiss }: DegradedModeWarningProps) {
  const [visible, setVisible] = useState(true);

  if (!visible) return null;

  const handleDismiss = () => {
    setVisible(false);
    onDismiss?.();
  };

  return (
    <div className="flex items-start gap-3 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
      <AlertTriangle className="h-5 w-5 text-yellow-600 flex-shrink-0 mt-0.5" />
      <div className="flex-1">
        <h3 className="font-medium text-yellow-900">
          Running in Degraded Mode
        </h3>
        <p className="text-sm text-yellow-800 mt-1">
          AI processing failed. The system is using fallback heuristics for column mapping.
          Results may be less accurate.
        </p>
      </div>
      <Button
        variant="ghost"
        size="sm"
        onClick={handleDismiss}
        className="flex-shrink-0"
      >
        <X className="h-4 w-4" />
      </Button>
    </div>
  );
}
