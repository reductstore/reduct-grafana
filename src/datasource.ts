import { DataSourceInstanceSettings, ScopedVars, DataQueryRequest } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { ReductQuery, ReductSourceOptions } from './types';

export class DataSource extends DataSourceWithBackend<ReductQuery, ReductSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<ReductSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: ReductQuery, scopedVars: ScopedVars): ReductQuery {
    const templateSrv = getTemplateSrv();

    return {
      ...query,
      bucket: templateSrv.replace(query.bucket, scopedVars),
      entry: templateSrv.replace(query.entry, scopedVars),
      options: {
        ...(query.options ?? {}),
        start: Number(templateSrv.replace(query.options?.start?.toString(), scopedVars)) || undefined,
        stop: Number(templateSrv.replace(query.options?.stop?.toString(), scopedVars)) || undefined,
        when: this.applyTemplateVariablesToWhen(query.options?.when, scopedVars),
      },
    };
  }

  filterQuery(query: ReductQuery): boolean {
    return !!query.bucket && !!query.entry;
  }

  /**
   * This runs before the query is sent to the backend.
   * We inject from/to here based on current time picker range.
   */
  prepareQuery(query: ReductQuery, options: DataQueryRequest<ReductQuery>): ReductQuery {
    return {
      ...query,
      options: {
        ...(query.options ?? {}),
        start: options.range.from.valueOf(),
        stop: options.range.to.valueOf(),
      },
    };
  }

  private applyTemplateVariablesToWhen(when: any, scopedVars: ScopedVars): any {
    if (when === undefined || when === null) {
      return when;
    }

    const templateSrv = getTemplateSrv();

    const quoteBareInterval = (input: string): string => {
      const macro = '$__interval';
      let inString = false;
      let escaped = false;
      let result = '';

      for (let i = 0; i < input.length; i++) {
        const ch = input[i];

        if (ch === '\\' && !escaped) {
          escaped = true;
          result += ch;
          continue;
        }

        if (ch === '"' && !escaped) {
          inString = !inString;
          result += ch;
          continue;
        }

        escaped = false;

        if (!inString && input.startsWith(macro, i)) {
          result += `"${macro}"`;
          i += macro.length - 1;
          continue;
        }

        result += ch;
      }

      return result;
    };

    const applyTemplateToValue = (value: any): any => {
      if (value === null || value === undefined) {
        return value;
      }

      if (typeof value === 'string') {
        return templateSrv.replace(value, scopedVars);
      }

      if (Array.isArray(value)) {
        return value.map((item) => applyTemplateToValue(item));
      }

      if (typeof value === 'object') {
        return Object.entries(value).reduce((acc, [key, val]) => {
          acc[key] = applyTemplateToValue(val);
          return acc;
        }, {} as Record<string, any>);
      }

      return value;
    };

    if (typeof when === 'string') {
      const sanitized = quoteBareInterval(when);
      try {
        const parsed = JSON.parse(sanitized);
        return applyTemplateToValue(parsed);
      } catch {
        return templateSrv.replace(when, scopedVars);
      }
    }

    return applyTemplateToValue(when);
  }
}
