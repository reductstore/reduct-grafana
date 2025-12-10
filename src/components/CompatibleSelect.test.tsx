import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { CompatibleSelect } from './CompatibleSelect';

const implementations: { combobox: any; select: any } = {
  combobox: ({ onChange, options, 'data-testid': testId }: any) => (
    <button data-testid={testId} onClick={() => onChange(options[1])}>
      Combobox
    </button>
  ),
  select: ({ onChange, options, 'data-testid': testId, getOptionLabel }: any) => (
    <select data-testid={testId} onChange={(evt) => onChange(options[Number(evt.target.value)])}>
      {options.map((opt: any, idx: number) => {
        const labelProps = getOptionLabel ? getOptionLabel(opt) : opt;
        return (
          <option key={opt.label} value={idx} data-testid={labelProps['data-testid']}>
            {opt.label}
          </option>
        );
      })}
    </select>
  ),
};

jest.mock('@grafana/ui', () => ({
  get Combobox() {
    return implementations.combobox;
  },
  set Combobox(value) {
    implementations.combobox = value;
  },
  get Select() {
    return implementations.select;
  },
  set Select(value) {
    implementations.select = value;
  },
}));

const options = [
  { label: 'First', value: 1 },
  { label: 'Second', value: 2 },
];

describe('CompatibleSelect', () => {
  afterEach(() => {
    implementations.combobox = ({ onChange, options, 'data-testid': testId }: any) => (
      <button data-testid={testId} onClick={() => onChange(options[1])}>
        Combobox
      </button>
    );
    implementations.select = ({ onChange, options, 'data-testid': testId, getOptionLabel }: any) => (
      <select data-testid={testId} onChange={(evt) => onChange(options[Number(evt.target.value)])}>
        {options.map((opt: any, idx: number) => {
          const labelProps = getOptionLabel ? getOptionLabel(opt) : opt;
          return (
            <option key={opt.label} value={idx} data-testid={labelProps['data-testid']}>
              {opt.label}
            </option>
          );
        })}
      </select>
    );
  });

  it('prefers Combobox when available', () => {
    const onChange = jest.fn();
    implementations.select = () => null;

    render(<CompatibleSelect value={options[0]} options={options} onChange={onChange} testId="test-select" />);

    const button = screen.getAllByTestId('test-select').find((el: HTMLElement) => el.tagName === 'BUTTON');
    expect(button).toBeDefined();
    fireEvent.click(button!);
    expect(onChange).toHaveBeenCalledWith(options[1]);
  });

  it('falls back to Select and decorates option test IDs', () => {
    const onChange = jest.fn();
    implementations.combobox = undefined;

    render(<CompatibleSelect value={options[0]} options={options} onChange={onChange} testId="test-select" />);

    const select = screen.getAllByTestId('test-select').find((el: HTMLElement) => el.tagName === 'SELECT');
    expect(select).toBeDefined();
    fireEvent.change(select!, { target: { value: '1' } });

    expect(screen.getByTestId('test-select-option-first')).toBeInTheDocument();
    expect(onChange).toHaveBeenCalledWith(options[1]);
  });
});
