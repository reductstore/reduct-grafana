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
    DataSourceWithBackend: class { },
    getTemplateSrv: () => ({ replace }),
  };
});

describe('DataSource.applyTemplateVariables', () => {
  const scopedVars = {} as ScopedVars;
  const ds = new DataSource({} as any);

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
