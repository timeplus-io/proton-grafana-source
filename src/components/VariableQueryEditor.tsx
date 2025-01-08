import React, { useState } from 'react';

import { VariableQueryProps } from '../types';

export const VariableQueryEditor = ({ onChange, query }: VariableQueryProps) => {
  const [state, setState] = useState(query);

  const saveQuery = () => {
    onChange(state, `${state.query}`);
  };

  const handleChange = (event: React.FormEvent<HTMLInputElement>) =>
    setState({
      ...state,
      [event.currentTarget.name]: event.currentTarget.value,
    });

  return (
    <>
      <div className="gf-form">
        <span className="gf-form-label width-10">Query</span>
        <input name="query" className="gf-form-input" onBlur={saveQuery} onChange={handleChange} value={state.query} />
      </div>
    </>
  );
};
