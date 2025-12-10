import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export enum DataMode {
  LabelOnly = 'LabelOnly',
  ContentOnly = 'ContentOnly',
  LabelAndContent = 'LabelAndContent',
}

export interface ReductQuery extends DataQuery {
  bucket?: string;
  entry?: string;
  options?: QueryOptions;
}

export type ReductWhenCondition = Record<string, any>;

export interface QueryOptions {
  start?: number;
  stop?: number;
  when?: ReductWhenCondition;
  ext?: any;
  strict?: boolean;
  continuous?: boolean;
  mode?: DataMode;
}

/**
 * These are options configured for each DataSource instance
 */
export interface ReductSourceOptions extends DataSourceJsonData {
  serverURL?: string;
  verifySSL?: boolean;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureJsonData {
  serverToken?: string;
}

export type QuotaType = 'NONE' | 'FIFO' | 'HARD';

export interface BucketSetting {
  max_block_size?: number;
  max_block_records?: number;
  quota_type?: QuotaType;
  quota_size?: number;
}

export interface ServerDefaults {
  bucket: BucketSetting;
}

export interface LicenseInfo {
  licensee: string;
  invoice: string;
  expiry_date: string;
  plan: string;
  device_number: number;
}

export interface ServerInfo {
  version: string;
  bucket_count: number;
  usage: number;
  uptime: number;
  oldest_record: number;
  latest_record: number;
  license?: LicenseInfo;
  defaults: ServerDefaults;
}
