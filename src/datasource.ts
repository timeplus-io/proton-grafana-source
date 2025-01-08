import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { MyVariableQuery, TpDataSourceOptions, TpQuery } from './types';

export class DataSource extends DataSourceWithBackend<TpQuery, TpDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<TpDataSourceOptions>) {
    super(instanceSettings);
  }


  metricFindQuery(query: MyVariableQuery, options?: any) {
    if (!query || !options.variable.datasource) {
      return Promise.resolve([]);
    }

    const prom = new Promise((resolve) => {
      const req = {
        targets: [{ datasource: options.variable.datasource,
          sql: query.query,
          refId: String(Math.random()) }],
        range: options ? options.range : (getTemplateSrv() as any).timeRange,
      };

      this.query(req).subscribe((res) => {
        const result = res.data[0] || { fields: [] }

        if (result.fields.length > 0)  {
          const labels = result.fields[0].values.map((v) => {
            return {
            text: v,
            value: v
          }});

          resolve(labels);
          return
        }
        resolve([]);
      });
    })

    return prom
  } 

  applyTemplateVariables(query: TpQuery, scopedVars: ScopedVars) {
    const srv = getTemplateSrv()
    const sql = srv.replace(query.sql, scopedVars)
    return {
      ...query,
      sql: sql,
    };
  }

  filterQuery(query: TpQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.sql;
  }

}

