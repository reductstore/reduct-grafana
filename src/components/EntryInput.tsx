import React, { useRef, useState, useEffect, useMemo } from 'react';
import { IconButton, Tooltip } from '@grafana/ui';
import * as UI from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

const MultiSelect = (UI as any).MultiSelect;
const Select = (UI as any).Select;
const Input = (UI as any).Input;

interface EntryInputProps {
  values: string[];
  options: Array<SelectableValue<string>>;
  onChange: (values: string[]) => void;
  testId?: string;
}

export function EntryInput({ values, options, onChange, testId }: EntryInputProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [textInput, setTextInput] = useState(() => values.join(', '));
  const [forceTextMode, setForceTextMode] = useState(false);

  const optionValues = useMemo(() => new Set(options.map((o) => o.value)), [options]);

  const isTextMode = useMemo(() => {
    if (forceTextMode) {
      return true;
    }
    return values.some((v) => v.includes('*') || !optionValues.has(v));
  }, [values, optionValues, forceTextMode]);

  useEffect(() => {
    if (!isTextMode) {
      setTextInput(values.join(', '));
    }
  }, [values, isTextMode]);

  const handleMultiSelectChange = (selected: Array<SelectableValue<string>>) => {
    onChange(selected.map((s) => s.value).filter((v): v is string => v !== undefined));
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTextInput(e.target.value);
  };

  const handleInputBlur = () => {
    const newEntries = textInput.trim()
      ? textInput
          .split(',')
          .map((s) => s.trim())
          .filter((s) => s.length > 0)
      : [];

    const isSame = newEntries.length === values.length && newEntries.every((v, i) => v === values[i]);

    if (!isSame) {
      onChange(newEntries);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleInputBlur();
    }
  };

  const handleModeToggle = () => {
    if (isTextMode) {
      setForceTextMode(false);
      const validEntries = values.filter((v) => optionValues.has(v) && !v.includes('*'));
      if (validEntries.length !== values.length) {
        onChange(validEntries);
      }
    } else {
      setForceTextMode(true);
      setTextInput(values.join(', '));
    }
  };

  const selectedValues = values
    .map((v) => options.find((o) => o.value === v) || { label: v, value: v })
    .filter((v): v is SelectableValue<string> => v !== undefined);

  return (
    <div ref={containerRef} data-testid={testId} style={{ display: 'flex', gap: '8px', width: '100%' }}>
      <div style={{ flex: 1, minWidth: 0 }}>
        {isTextMode ? (
          <Input
            data-testid={`${testId}-wildcard`}
            value={textInput}
            onChange={handleInputChange}
            onBlur={handleInputBlur}
            onKeyDown={handleKeyDown}
            placeholder="sensor-*, device-1"
            width="100%"
          />
        ) : MultiSelect ? (
          <MultiSelect
            data-testid={`${testId}-multiselect`}
            value={selectedValues}
            options={options}
            onChange={handleMultiSelectChange}
            menuPortalTarget={containerRef.current}
            menuPosition="fixed"
            placeholder="Select entries..."
          />
        ) : (
          <Select
            isMulti
            data-testid={`${testId}-select`}
            value={selectedValues}
            options={options}
            onChange={handleMultiSelectChange}
          />
        )}
      </div>
      <Tooltip content={isTextMode ? 'Switch to entry selection' : 'Switch to text input'}>
        <IconButton
          name={isTextMode ? 'list-ul' : 'pen'}
          onClick={handleModeToggle}
          aria-label={isTextMode ? 'Switch to dropdown' : 'Switch to text input'}
          data-testid={`${testId}-toggle`}
        />
      </Tooltip>
    </div>
  );
}
