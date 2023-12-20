import defaults from 'lodash/defaults';

import React, { SyntheticEvent, PureComponent} from 'react';
import { LegacyForms, CodeEditor} from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { TpDataSourceOptions, TpQuery , defaultQuery} from '../types';

const { Switch } = LegacyForms;

type Props = QueryEditorProps<DataSource, TpQuery, TpDataSourceOptions>;

interface State {}

export class QueryEditor extends PureComponent<Props, State> {

  onAddNowChange = (event: SyntheticEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, addNow: event.currentTarget.checked });
    //onRunQuery();
  };

  onQueryTextChange = (value: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, queryText: value });
    //onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText, addNow } = query;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <CodeEditor
            value={queryText || 'select * from <table>'}
            width={600}
            height={200}
            language="sql"
            showLineNumbers={true}
            showMiniMap={false}
            onBlur={this.onQueryTextChange}
          />
        </div>
        <div className="gf-form">
          <Switch
            checked={addNow || false}
            label="Add current time as the first column"
            onChange={this.onAddNowChange}
          />
        </div>
      </div>
    );
  }
}
