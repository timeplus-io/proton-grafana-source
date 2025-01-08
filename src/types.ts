import { DataQuery } from '@grafana/schema';
import { DataSourceJsonData } from '@grafana/data';

export interface TpQuery extends DataQuery {
  sql: string;
}

/**
 * These are options configured for each DataSource instance
 */
export interface TpDataSourceOptions extends DataSourceJsonData {
  host?: string;
  tcpPort?: number;
  httpPort?: number;
  username?: string
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface TpSecureJsonData {
  password?: string;
}

export interface MyVariableQuery {
  query: string;
}

export interface VariableQueryProps {
  query: MyVariableQuery;
  onChange: (query: MyVariableQuery, definition: string) => void;
}
