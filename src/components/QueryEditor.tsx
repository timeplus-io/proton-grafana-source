import { CodeEditor, InlineField, Stack } from '@grafana/ui';
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
    <Stack gap={0}>
      <InlineField label="Query" labelWidth={16} tooltip="Not used yet">
        <CodeEditor
          onChange={onSQLChange}
          width={600}
          height={200}
          language="sql"
          showLineNumbers={true}
          showMiniMap={false}
          value={sql || ''}
        />
      </InlineField>
    </Stack>
  );
}
