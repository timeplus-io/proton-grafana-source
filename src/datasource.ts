import { DataSourceInstanceSettings, ScopedVars, DataQueryRequest, DataQueryResponse, LiveChannelScope } from '@grafana/data';
import { DataSourceWithBackend, getGrafanaLiveSrv, getTemplateSrv } from '@grafana/runtime';

import { TpQuery, TpDataSourceOptions } from './types';
import { merge, Observable } from 'rxjs';

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

  query(request: DataQueryRequest<TpQuery>): Observable<DataQueryResponse> {
    const observables = request.targets.map((query, index) => {

      return getGrafanaLiveSrv().getDataStream({
        addr: {
          scope: LiveChannelScope.DataSource,
          namespace: this.uid,
          path: `timeplus/${this.uid}/${uuidv4()}`, // this will allow each new query to create a new connection
          data: {
            ...query,
          },
        },
      });
    });

    return merge(...observables);
  }
}

function uuidv4() {
  return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, c =>
    (+c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> +c / 4).toString(16)
  );
}
