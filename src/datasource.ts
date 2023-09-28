import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { TpQuery, TpDataSourceOptions, defaultQuery } from './types';

export class DataSource extends DataSourceWithBackend<TpQuery, TpDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<TpDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<TpQuery> {
    return defaultQuery
  }
}
