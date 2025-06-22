import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface MyQuery extends DataQuery {
  queryType: 'listBuckets' | 'getBucketEntries' | 'getBucketSetting' | 'getInfo' | 'listTokens' | 'getReplicationTasks' | 'queryRecords';
  bucket?: string;
  entry?: string;
  options?: QueryOptions;
}

export interface QueryOptions {
  start?: number;
  stop?: number;
  when?: any;
  ext?: any;
  strict?: boolean;
  continuous?: boolean;
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  queryType: 'listBuckets',
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
  serverURL?: string;
  verifySSL?: boolean;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  serverToken?: string;
}
