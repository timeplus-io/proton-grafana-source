import { CodeEditor, Field } from '@grafana/ui';
import React, { useState } from 'react';

import { VariableQueryProps } from '../types';

export const VariableQueryEditor = ({ onChange, query }: VariableQueryProps) => {
  const [state, setState] = useState(query);

  const saveQuery = () => {
    onChange(state, `${state.query}`);
  };

  const onSQLChange = (sql: string) => {
    setState({
      query: sql,
    });
  };

  return (
    <Field
      label=""
      description="Make sure your query returns either 1 or 2 columns. If your query returns 1 column only, it will be used as the value and label. If it returns 2 columns, the first column will be used as the value while the second column will be used as the label."
    >
      <CodeEditor
        onChange={onSQLChange}
        onBlur={saveQuery}
        width="100%"
        height={200}
        language="sql"
        showLineNumbers={true}
        showMiniMap={false}
        value={state.query}
      />
    </Field>
  );
};
