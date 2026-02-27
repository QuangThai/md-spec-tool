'use client';

import { OutputFormat } from '@/lib/types';
import { Select } from '@/components/ui/Select';

interface FormatSelectorProps {
  value: OutputFormat;
  onChange: (format: OutputFormat) => void;
  id?: string;
}

export function FormatSelector({ value, onChange, id }: FormatSelectorProps) {
  return (
    <Select
      id={id}
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
