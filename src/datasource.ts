import { DataQueryRequest, DataSourceInstanceSettings, MetricFindValue, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { TpDataSourceOptions, TpQuery, TpVariableQuery } from './types';

export class DataSource extends DataSourceWithBackend<TpQuery, TpDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<TpDataSourceOptions>) {
    super(instanceSettings);
  }


  metricFindQuery(query: TpVariableQuery, options?: any) {
    let metrics: MetricFindValue[] = []
    if (!query) {
      return Promise.resolve(metrics);
    }

    const prom = new Promise<MetricFindValue[]>((resolve) => {
      const req = {
        targets: [
        {
          sql: query.query,
          refId: String(Math.random())
        }],
        range: options ? options.range : (getTemplateSrv() as any).timeRange,
      } as  DataQueryRequest<TpQuery> ;

      this.query(req).subscribe((res) => {
        const result = res.data[0] || { fields: [] }

        if (result.fields.length === 2) {
          for (let i = 0; i < result.fields[0].values.length; i++) {
            metrics.push({
              text: result.fields[1].values[i],
              value: result.fields[0].values[i]
            })
          }
        } else if (result.fields.length === 1) {
          metrics = result.fields[0].values.map((v: string) => {
            return {
            text: v,
            value: v
          }});
        }
        
        resolve(metrics);
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

