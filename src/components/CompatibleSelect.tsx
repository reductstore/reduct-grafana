import React, { useRef } from 'react';
import * as UI from '@grafana/ui';
import { SelectableValue } from '@grafana/data';
import { getGrafanaMajorVersion } from './grafanaVersion';

interface CompatibleSelectProps<T = any> {
  value?: SelectableValue<T>;
  options?: Array<SelectableValue<T>>;
  onChange: (value: SelectableValue<T>) => void;
  testId?: string;
  loading?: boolean;
}

export function CompatibleSelect<T>({ value, options = [], onChange, testId, loading }: CompatibleSelectProps<T>) {
  const Combobox = (UI as any).Combobox;
  const Select = (UI as any).Select;
  const hasCombobox = !!Combobox && getGrafanaMajorVersion() >= 12;

  const containerRef = useRef<HTMLDivElement | null>(null);

  const handleComboboxChange = (option: { value: T; label?: string }) => {
    const selected = options.find((o) => o.value === option.value) || { value: option.value, label: option.label };
    onChange(selected);
  };

  return (
    <div ref={containerRef} data-testid={testId}>
      {hasCombobox ? (
        <Combobox
          value={value?.value ?? null}
          options={options}
          onChange={handleComboboxChange}
          loading={loading}
        />
      ) : (
        <Select value={value} options={options} onChange={onChange} isLoading={loading} />
      )}
    </div>
  );
}
