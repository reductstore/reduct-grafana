import React, { useRef, useCallback } from 'react';
import * as UI from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

const MultiCombobox = (UI as any).MultiCombobox;
const MultiSelect = (UI as any).MultiSelect;

interface EntryInputProps {
  values: string[];
  options: Array<SelectableValue<string>>;
  onChange: (values: string[]) => void;
  testId?: string;
}

export function EntryInput({ values, options, onChange, testId }: EntryInputProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const hasMultiCombobox = !!MultiCombobox;

  const handleMultiComboboxChange = useCallback(
    (selected: Array<{ value: string; label?: string }>) => {
      onChange(selected.map((s) => s.value));
    },
    [onChange]
  );

  const handleMultiSelectChange = useCallback(
    (selected: Array<SelectableValue<string>>) => {
      onChange(selected.map((s) => s.value).filter((v): v is string => v !== undefined));
    },
    [onChange]
  );

  const selectedValues = values.map((v) => options.find((o) => o.value === v) || { label: v, value: v });

  return (
    <div ref={containerRef} data-testid={testId} style={{ width: '100%', minWidth: 200 }}>
      {hasMultiCombobox ? (
        <MultiCombobox
          data-testid={testId}
          value={values}
          options={options}
          onChange={handleMultiComboboxChange}
          createCustomValue
          placeholder="Select or type entries..."
        />
      ) : (
        <MultiSelect
          data-testid={testId}
          value={selectedValues}
          options={options}
          onChange={handleMultiSelectChange}
          allowCustomValue
          placeholder="Select or type entries..."
        />
      )}
    </div>
  );
}
