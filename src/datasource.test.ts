import { ScopedVars } from '@grafana/data';
import { DataSource } from './datasource';

jest.mock('@grafana/runtime', () => {
  const replace = jest.fn((value: any) => {
    if (typeof value === 'string') {
      return value.replace(/\$__interval/g, '5s');
    }
    return value;
  });

  return {
    DataSourceWithBackend: class {},
    getTemplateSrv: () => ({ replace }),
  };
});

describe('DataSource.applyTemplateVariables', () => {
  const scopedVars = {} as ScopedVars;
  const ds = new DataSource({} as any);
  const templateSrv = (require('@grafana/runtime') as any).getTemplateSrv();
  const replace = templateSrv.replace as jest.Mock;

  beforeEach(() => {
    replace.mockClear();
    replace.mockImplementation((value: any) => {
      if (typeof value === 'string') {
        return value.replace(/\$__interval/g, '5s');
      }
      return value;
    });
  });

  it('replaces bucket, entry, and numeric options with templated values', () => {
    replace.mockImplementation((value: any) => {
      if (value === '$bucket') {
        return 'templated-bucket';
      }
      if (value === '$entry') {
        return 'templated-entry';
      }
      if (value === '$start') {
        return '42';
      }
      if (value === '$stop') {
        return '99';
      }
      return value;
    });

    const result = ds.applyTemplateVariables(
      {
        bucket: '$bucket',
        entry: '$entry',
        options: { start: '$start', stop: '$stop', when: { raw: 'keep' } },
      } as any,
      scopedVars
    );

    expect(result.bucket).toBe('templated-bucket');
    expect(result.entry).toBe('templated-entry');
    expect(result.options?.start).toBe(42);
    expect(result.options?.stop).toBe(99);
    expect(result.options?.when).toEqual({ raw: 'keep' });
  });

  it('applies templating to nested when objects and arrays', () => {
    replace.mockImplementation((value: any) => {
      if (value === '$__interval') {
        return '10s';
      }
      if (value === '$label') {
        return 'sensor-1';
      }
      return value;
    });

    const result = ds.applyTemplateVariables(
      {
        options: {
          when: {
            $and: [{ '&sensor': { $eq: '$label' } }, { values: ['$__interval', null, { nested: '$__interval' }] }],
          },
        },
      } as any,
      scopedVars
    );

    expect(result.options?.when).toEqual({
      $and: [{ '&sensor': { $eq: 'sensor-1' } }, { values: ['10s', null, { nested: '10s' }] }],
    });
  });

  it('replaces $__interval when when-condition is JSON string', () => {
    const result = ds.applyTemplateVariables(
      {
        options: { when: '{ "$each_t": "$__interval" }' },
      } as any,
      scopedVars
    );

    expect(result.options?.when).toEqual({ $each_t: '5s' });
  });

  it('applies templating but returns raw string when JSON parse fails', () => {
    const result = ds.applyTemplateVariables(
      {
        options: { when: 'not json $__interval' },
      } as any,
      scopedVars
    );

    expect(result.options?.when).toBe('not json 5s');
  });
});

describe('DataSource helpers', () => {
  const ds = new DataSource({} as any);

  it('filters out queries missing bucket or entry', () => {
    expect(ds.filterQuery({ bucket: 'b', entry: 'e' } as any)).toBe(true);
    expect(ds.filterQuery({ bucket: '', entry: 'e' } as any)).toBe(false);
    expect(ds.filterQuery({ bucket: 'b', entry: '' } as any)).toBe(false);
  });

  it('injects range into query options while preserving other fields', () => {
    const result = ds.prepareQuery(
      {
        bucket: 'b',
        entry: 'e',
        options: { when: { $eq: 1 }, start: 1, stop: 2 },
      } as any,
      {
        range: {
          from: { valueOf: () => 10 } as any,
          to: { valueOf: () => 20 } as any,
        },
      } as any
    );

    expect(result.options?.start).toBe(10);
    expect(result.options?.stop).toBe(20);
    expect(result.options?.when).toEqual({ $eq: 1 });
  });
});
