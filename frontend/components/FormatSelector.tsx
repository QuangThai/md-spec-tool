'use client';

import { OutputFormat } from '@/lib/types';
import { Select } from '@/components/ui/Select';

interface FormatSelectorProps {
  value: OutputFormat;
  onChange: (format: OutputFormat) => void;
}

export function FormatSelector({ value, onChange }: FormatSelectorProps) {
  return (
    <Select
      value={value}
      onValueChange={(v) => onChange(v as OutputFormat)}
      options={[
        { label: 'Spec Document', value: 'spec' },
        { label: 'Simple Table', value: 'table' },
      ]}
      placeholder="Select format"
      className="w-full"
    />
  );
}
