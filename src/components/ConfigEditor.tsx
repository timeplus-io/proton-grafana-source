import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { TpDataSourceOptions } from '../types';

const { FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<TpDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      host: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      port: parseInt(event.target.value, 10),
    };
    onOptionsChange({ ...options, jsonData });
  };

  onPortRESTChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      port2: parseInt(event.target.value, 10),
    };
    onOptionsChange({ ...options, jsonData });
  };

  render() {
    const { options } = this.props;
    const { jsonData } = options;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Host"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onHostChange}
            value={jsonData.host || 'localhost'}
            placeholder="Proton host"
          />
        </div>
        <div className="gf-form">
          <FormField
            label="TCP Port"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPortChange}
            value={jsonData.port || 8463}
            placeholder="8463"
          />
        </div>
        <div className="gf-form">
          <FormField
            label="REST Port"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPortRESTChange}
            value={jsonData.port2 || 3218}
            placeholder="3218"
          />
        </div>        
      </div>
    );
  }
}
