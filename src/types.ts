import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface TpQuery extends DataQuery {
  sql: string;
}

/**
 * These are options configured for each DataSource instance
 */
export interface TpDataSourceOptions extends DataSourceJsonData {
  host?: string;
  port?: number;
  username?: string
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface TpSecureJsonData {
  password?: string;
}
