import { replaceWhenTemplateVariables } from './templateUtils';

jest.mock('@grafana/runtime', () => {
  const replace = jest.fn((value: string) => value);

  return {
    getTemplateSrv: () => ({ replace }),
  };
});

const templateSrv = (require('@grafana/runtime') as any).getTemplateSrv();
const replace = templateSrv.replace as jest.Mock;

beforeEach(() => {
  replace.mockClear();
  replace.mockImplementation((value: string) => {
    if (value === '${key}') {
      return 'sensor-1';
    }
    if (value === '$var') {
      return 'resolved';
    }
    return value;
  });
});

describe('replaceWhenTemplateVariables', () => {
  it('returns undefined for undefined input', () => {
    expect(replaceWhenTemplateVariables(undefined)).toBeUndefined();
  });

  it('returns undefined for null input', () => {
    expect(replaceWhenTemplateVariables(null)).toBeUndefined();
  });

  it('replaces template variables in nested objects', () => {
    const when = { '&label': { $contains: '${key}' } };
    const result = replaceWhenTemplateVariables(when);
    expect(result).toEqual({ '&label': { $contains: 'sensor-1' } });
  });

  it('replaces template variables in arrays', () => {
    const when = { $and: [{ '&a': '$var' }, { '&b': '${key}' }] };
    const result = replaceWhenTemplateVariables(when);
    expect(result).toEqual({ $and: [{ '&a': 'resolved' }, { '&b': 'sensor-1' }] });
  });

  it('parses JSON string condition after substitution', () => {
    const when = '{ "&label": { "$eq": "${key}" }' + '}';
    const result = replaceWhenTemplateVariables(when as any);
    expect(result).toEqual({ '&label': { $eq: 'sensor-1' } });
  });

  it('returns raw string when JSON parse fails', () => {
    replace.mockImplementation((value: string) => value.replace(/\$\{key\}/g, 'sensor-1'));
    const when = 'not json ${key}';
    const result = replaceWhenTemplateVariables(when as any);
    expect(result).toBe('not json sensor-1');
  });

  it('preserves non-string values (numbers, booleans, null)', () => {
    const when = { '&sensor': { $gt: 42 }, '&active': true, '&removed': null };
    const result = replaceWhenTemplateVariables(when);
    expect(result).toEqual({ '&sensor': { $gt: 42 }, '&active': true, '&removed': null });
  });

  it('passes scopedVars to templateSrv.replace', () => {
    const scopedVars = { key: { text: 'sensor-1', value: 'sensor-1' } };
    replaceWhenTemplateVariables({ '&a': '$var' }, scopedVars);
    expect(replace).toHaveBeenCalledWith('$var', scopedVars);
  });
});
