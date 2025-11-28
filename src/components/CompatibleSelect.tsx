import React, { useRef } from 'react';
import * as UI from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

interface CompatibleSelectProps<T = any> {
  value?: SelectableValue<T>;
  options?: Array<SelectableValue<T>>;
  onChange: (value: SelectableValue<T>) => void;
  testId?: string;
}

export function CompatibleSelect<T>({ value, options = [], onChange, testId }: CompatibleSelectProps<T>) {
  const Combobox = (UI as any).Combobox;
  const Select = (UI as any).Select;
  const hasCombobox = !!Combobox;

  const containerRef = useRef<HTMLDivElement | null>(null);

  // test IDs are only for Select, Combobox uses different structure
  const makeId = (label: any) => `${testId}-option-${String(label).replace(/\s+/g, '-').toLowerCase()}`;

  return (
    <div ref={containerRef} data-testid={testId}>
      {hasCombobox ? (
        <Combobox
          data-testid={testId}
          value={value}
          options={options}
          onChange={onChange}
          menuPortalTarget={containerRef.current}
          menuPosition="fixed"
        />
      ) : (
        <Select
          data-testid={testId}
          value={value}
          options={options}
          onChange={onChange}
          getOptionLabel={(opt: any) => ({
            ...opt,
            'data-testid': makeId(opt.label ?? opt.value),
          })}
        />
      )}
    </div>
  );
}
