import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export enum DataMode {
  Labels = 'labels',
  Content = 'content',
  Both = 'both',
}

export interface ReductQuery extends DataQuery {
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
  mode?: DataMode;
}

/**
 * These are options configured for each DataSource instance
 */
export interface ReductSourceOptions extends DataSourceJsonData {
  path?: string;
  serverURL?: string;
  verifySSL?: boolean;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureJsonData {
  serverToken?: string;
}
