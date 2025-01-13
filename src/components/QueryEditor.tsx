import { CodeEditor, Field } from '@grafana/ui';
import { TpDataSourceOptions, TpQuery } from '../types';

import { DataSource } from '../datasource';
import { QueryEditorProps } from '@grafana/data';
import React from 'react';

type Props = QueryEditorProps<DataSource, TpQuery, TpDataSourceOptions>;

export function QueryEditor({ query, onChange }: Props) {
  const onSQLChange = (sql: string) => {
    onChange({ ...query, sql: sql });
  };

  const { sql } = query;

  return (
    <Field>
      <CodeEditor
        onChange={onSQLChange}
        width="100%"
        height={200}
        language="sql"
        showLineNumbers={true}
        showMiniMap={false}
        value={sql || ''}
      />
    </Field>
  );
}
