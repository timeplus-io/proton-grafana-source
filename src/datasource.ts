import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { TpQuery, TpDataSourceOptions } from './types';

export class DataSource extends DataSourceWithBackend<TpQuery, TpDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<TpDataSourceOptions>) {
    super(instanceSettings);
  }


  applyTemplateVariables(query: TpQuery, scopedVars: ScopedVars) {
    return {
      ...query,
      sql: getTemplateSrv().replace(query.sql, scopedVars),
    };
  }

  filterQuery(query: TpQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.sql;
  }

}

