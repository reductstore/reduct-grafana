import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { MyQuery, MyDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<MyQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    return {
      ...query,
      queryType: query.queryType,
      bucket: getTemplateSrv().replace(query.bucket, scopedVars),
      entry: getTemplateSrv().replace(query.entry, scopedVars),
      options: {
        queryType: 'QUERY',
        start: Number(getTemplateSrv().replace(query.options?.start?.toString(), scopedVars)),
        stop: Number(getTemplateSrv().replace(query.options?.stop?.toString(), scopedVars)),
        when: getTemplateSrv().replace(query.options?.when, scopedVars),
        ext: getTemplateSrv().replace(query.options?.ext, scopedVars),
        strict: getTemplateSrv().replace(query.options?.strict?.toString(), scopedVars) === 'true',
        continuous: getTemplateSrv().replace(query.options?.continuous?.toString(), scopedVars) === 'true',
      },
    };
  }

  filterQuery(query: MyQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.queryType;
  }
}
