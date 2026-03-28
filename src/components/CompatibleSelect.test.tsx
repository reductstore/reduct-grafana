import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { CompatibleSelect } from './CompatibleSelect';

const staticOptions = [
  { label: 'First', value: 1 },
  { label: 'Second', value: 2 },
];

const implementations: { combobox: any; select: any } = {
  combobox: ({ onChange, options, 'data-testid': testId }: any) => {
    const handleClick = async () => {
      const opts = typeof options === 'function' ? await options('') : options;
      onChange(opts[1]);
    };
    return (
      <button data-testid={testId} onClick={handleClick}>
        Combobox
      </button>
    );
  },
  select: ({ onChange, options, 'data-testid': testId }: any) => (
    <select data-testid={testId} onChange={(evt) => onChange(options[Number(evt.target.value)])}>
      {options.map((opt: any, idx: number) => (
        <option key={opt.label} value={idx}>
          {opt.label}
        </option>
      ))}
    </select>
  ),
};

let mockGrafanaMajorVersion = 12;

jest.mock('./grafanaVersion', () => ({
  getGrafanaMajorVersion: () => mockGrafanaMajorVersion,
}));

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

describe('CompatibleSelect', () => {
  afterEach(() => {
    mockGrafanaMajorVersion = 12;
    implementations.combobox = ({ onChange, options, 'data-testid': testId }: any) => {
      const handleClick = async () => {
        const opts = typeof options === 'function' ? await options('') : options;
        onChange(opts[1]);
      };
      return (
        <button data-testid={testId} onClick={handleClick}>
          Combobox
        </button>
      );
    };
    implementations.select = ({ onChange, options, 'data-testid': testId }: any) => (
      <select data-testid={testId} onChange={(evt) => onChange(options[Number(evt.target.value)])}>
        {options.map((opt: any, idx: number) => (
          <option key={opt.label} value={idx}>
            {opt.label}
          </option>
        ))}
      </select>
    );
  });

  it('prefers Combobox when available', async () => {
    const onChange = jest.fn();
    implementations.select = () => null;

    render(
      <CompatibleSelect value={staticOptions[0]} options={staticOptions} onChange={onChange} testId="test-select" />
    );

    const wrapper = screen.getByTestId('test-select');
    const button = wrapper.querySelector('button')!;
    expect(button).toBeTruthy();
    fireEvent.click(button);
    await screen.findByTestId('test-select');
    expect(onChange).toHaveBeenCalledWith(staticOptions[1]);
  });

  it('falls back to Select when Combobox unavailable', () => {
    const onChange = jest.fn();
    implementations.combobox = undefined;

    render(
      <CompatibleSelect value={staticOptions[0]} options={staticOptions} onChange={onChange} testId="test-select" />
    );

    const wrapper = screen.getByTestId('test-select');
    const select = wrapper.querySelector('select')!;
    expect(select).toBeTruthy();
    fireEvent.change(select, { target: { value: '1' } });

    expect(onChange).toHaveBeenCalledWith(staticOptions[1]);
  });

  it('falls back to Select when Grafana version is below 12', () => {
    const onChange = jest.fn();
    mockGrafanaMajorVersion = 11;

    render(
      <CompatibleSelect value={staticOptions[0]} options={staticOptions} onChange={onChange} testId="test-select" />
    );

    const wrapper = screen.getByTestId('test-select');
    const select = wrapper.querySelector('select')!;
    expect(select).toBeTruthy();
    fireEvent.change(select, { target: { value: '1' } });

    expect(onChange).toHaveBeenCalledWith(staticOptions[1]);
  });
});
