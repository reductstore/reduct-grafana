import {
  DataSourceInstanceSettings,
  ScopedVars,
  DataQueryRequest,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { ReductQuery, ReductSourceOptions } from './types';

export class DataSource extends DataSourceWithBackend<ReductQuery, ReductSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<ReductSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: ReductQuery, scopedVars: ScopedVars): ReductQuery {
    return {
      ...query,
      bucket: getTemplateSrv().replace(query.bucket, scopedVars),
      entry: getTemplateSrv().replace(query.entry, scopedVars),
      options: {
        start: Number(getTemplateSrv().replace(query.options?.start?.toString(), scopedVars)) || undefined,
        stop: Number(getTemplateSrv().replace(query.options?.stop?.toString(), scopedVars)) || undefined,
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
      options:{
        start:options.range.from.valueOf(),
        stop: options.range.to.valueOf(),
        when: query.options?.when || {},
      }
    };
  }
}
