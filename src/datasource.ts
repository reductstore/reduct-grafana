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
      entry: query.entry ? templateSrv.replace(query.entry, scopedVars) : undefined,
      entries: query.entries?.map((e) => templateSrv.replace(e, scopedVars)),
      options: {
        ...(query.options ?? {}),
        start: Number(templateSrv.replace(query.options?.start?.toString(), scopedVars)) || undefined,
        stop: Number(templateSrv.replace(query.options?.stop?.toString(), scopedVars)) || undefined,
        when: this.applyTemplateVariablesToWhen(query.options?.when, scopedVars),
      },
    };
  }

  filterQuery(query: ReductQuery): boolean {
    const hasEntries = (query.entries && query.entries.length > 0) || !!query.entry;
    return !!query.bucket && hasEntries;
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
      const sanitized = templateSrv.replace(when, scopedVars);
      try {
        const parsed = JSON.parse(sanitized);
        return applyTemplateToValue(parsed);
      } catch {
        return sanitized;
      }
    }

    return applyTemplateToValue(when);
  }
}
