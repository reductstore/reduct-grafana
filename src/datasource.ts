import {
  DataSourceInstanceSettings,
  ScopedVars,
  DataQueryRequest,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { MyQuery, MyDataSourceOptions } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
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
  filterQuery(query: MyQuery): boolean {
    return !!query.bucket && !!query.entry;
  }
  

  /**
   * This runs before the query is sent to the backend.
   * We inject from/to here based on current time picker range.
   */
  prepareQuery(query: MyQuery, options: DataQueryRequest<MyQuery>): MyQuery {
    return {
      ...query,
      options:{
        start:options.range.from.valueOf(),
        stop: options.range.to.valueOf(),
      }
    };
  }
}
