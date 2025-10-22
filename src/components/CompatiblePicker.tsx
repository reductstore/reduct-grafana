import React, { useMemo } from 'react';
import * as UI from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

type Opt<T> = SelectableValue<T>;

let hasCombobox: boolean | null = null;
let hasWarnedDeprecation = false;

export function CompatiblePicker<T>(props: {
  value?: Opt<T>;
  options?: Array<Opt<T>>;
  onChange: (v: Opt<T>) => void;
  loadOptions?: (q?: string) => Promise<Array<Opt<T>>>;
}) {
  const { value, options, onChange, loadOptions } = props;
  const components = useMemo(() => {
    if (hasCombobox === null) {
      hasCombobox = !!(UI as any).Combobox;

      if (!hasCombobox && !hasWarnedDeprecation) {
        console.warn(
          'Using deprecated Select/AsyncSelect \
          component as fallback for older Grafana versions.'
        );
        hasWarnedDeprecation = true;
      }
    }

    return {
      Combobox: (UI as any).Combobox,
      Select: (UI as any).Select,
      AsyncSelect: (UI as any).AsyncSelect,
      hasCombobox,
    };
  }, []);

  if (components.hasCombobox) {
    return <components.Combobox value={value} options={options} onChange={onChange} loadOptions={loadOptions} />;
  }

  // Fallback for Grafana 9.5: use (Async)Select
  if (loadOptions) {
    // eslint-disable-next-line deprecation/deprecation
    return <components.AsyncSelect value={value} defaultOptions loadOptions={loadOptions} onChange={onChange} />;
  }
  // eslint-disable-next-line deprecation/deprecation
  return <components.Select value={value} options={options} onChange={onChange} />;
}
