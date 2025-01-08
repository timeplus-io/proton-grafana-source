import { TpDataSourceOptions, TpQuery } from './types';

import { ConfigEditor } from './components/ConfigEditor';
import { DataSource } from './datasource';
import { DataSourcePlugin } from '@grafana/data';
import { QueryEditor } from './components/QueryEditor';
import { VariableQueryEditor } from './components/VariableQueryEditor';

export const plugin = new DataSourcePlugin<DataSource, TpQuery, TpDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor)
  .setVariableQueryEditor(VariableQueryEditor)
