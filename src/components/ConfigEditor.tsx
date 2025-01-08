import { InlineField, Input, SecretInput } from '@grafana/ui';
import React, { ChangeEvent } from 'react';
import { TpDataSourceOptions, TpSecureJsonData } from '../types';

import { DataSourcePluginOptionsEditorProps } from '@grafana/data';

interface Props extends DataSourcePluginOptionsEditorProps<TpDataSourceOptions, TpSecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        host: event.target.value,
      },
    });
  };

  const onUsernameChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        username: event.target.value,
      },
    });
  };

  const onTCPPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        tcpPort: parseInt(event.target.value, 10),
      },
    });
  };

  const onHTTPPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        httpPort: parseInt(event.target.value, 10),
      },
    });
  };

  // Secure field (only sent to the backend)
  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
      },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  return (
    <>
      <InlineField required={true} label="Host" labelWidth={14} interactive tooltip={'Hostname'}>
        <Input
          required={true}
          id="config-editor-host"
          onChange={onHostChange}
          value={jsonData.host}
          placeholder="Enter the host, e.g. localhost"
          width={40}
        />
      </InlineField>
      <InlineField label="TCP Port" labelWidth={14} interactive tooltip={'TCP Port'}>
        <Input
          id="config-editor-tcp-port"
          type="number"
          onChange={onTCPPortChange}
          value={jsonData.tcpPort}
          placeholder="8463"
          width={40}
        />
      </InlineField>
      <InlineField label="HTTP Port" labelWidth={14} interactive tooltip={'HTTP Port'}>
        <Input
          id="config-editor-http-port"
          type="number"
          onChange={onHTTPPortChange}
          value={jsonData.httpPort}
          placeholder="3218"
          width={40}
        />
      </InlineField>
      <InlineField required={true} label="Username" labelWidth={14} interactive tooltip={'Username'}>
        <Input
          required={true}
          id="config-editor-username"
          onChange={onUsernameChange}
          value={jsonData.username}
          placeholder="Enter the username, e.g. admin"
          width={40}
        />
      </InlineField>
      <InlineField label="Password" labelWidth={14} interactive tooltip={'Password'}>
        <SecretInput
          required
          id="config-editor-password"
          isConfigured={secureJsonFields.password}
          value={secureJsonData?.password}
          placeholder="Enter your password"
          width={40}
          onReset={onResetPassword}
          onChange={onPasswordChange}
        />
      </InlineField>
    </>
  );
}
