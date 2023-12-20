import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface TpQuery extends DataQuery {
  queryText?: string;
  addNow: boolean;
}

export const defaultQuery: Partial<TpQuery> = {
  //queryText: 'select min(number) as min, max(number) as max, count(*) as count \nfrom raw_iot_data \nemit periodic 1s',
  queryText: 'select now()',
  addNow: false,
};

/**
 * These are options configured for each DataSource instance
 */
export interface TpDataSourceOptions extends DataSourceJsonData {
  host?: string;
  port?: number;
  port2?: number;
}
