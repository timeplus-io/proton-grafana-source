import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { TpQuery, TpDataSourceOptions, defaultQuery } from './types';

export class DataSource extends DataSourceWithBackend<TpQuery, TpDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<TpDataSourceOptions>) {
    super(instanceSettings);
  }

  filterQuery(query: TpQuery): boolean {
    if (query.hide || query.queryText === '') {
      return false;
    }
    return true;
  }

  getDefaultQuery(_: CoreApp): Partial<TpQuery> {
    return defaultQuery
  }
}
