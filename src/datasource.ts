import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { TpQuery, TpDataSourceOptions, defaultQuery } from './types';

export class DataSource extends DataSourceWithBackend<TpQuery, TpDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<TpDataSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: TpQuery, scopedVars: ScopedVars): Record<string, any> {
    return {
      ...query,
      queryText: getTemplateSrv().replace(query.queryText, scopedVars),
    };
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
