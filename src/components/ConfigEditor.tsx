import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { TpDataSourceOptions, TpSecureJsonData } from '../types';

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

  const onPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        port: parseInt(event.target.value, 10),
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
      <InlineField label="Host" labelWidth={14} interactive tooltip={'Hostname and port'}>
        <Input
          id="config-editor-host"
          onChange={onHostChange}
          value={jsonData.host || "localhost"}
          placeholder="Enter the host, e.g. localhost"
          width={40}
        />
      </InlineField>
      <InlineField label="Port" labelWidth={14} interactive tooltip={'Port'}>
        <Input
          id="config-editor-port"
          type='number'
          onChange={onPortChange}
          value={jsonData.port}
          placeholder="8463"
          width={40}
        />
      </InlineField>
      <InlineField label="Username" labelWidth={14} interactive tooltip={'Username'}>
        <Input
          id="config-editor-username"
          onChange={onUsernameChange}
          value={jsonData.username}
          placeholder="Enter the username, e.g. timeplusd"
          width={40}
        />
      </InlineField>
      <InlineField label="Password" labelWidth={14} interactive tooltip={'Password'}>
        <SecretInput
          required
          id="config-editor-password"
          isConfigured={secureJsonFields.password}
          value={secureJsonData?.password}
          placeholder="Enter your Password"
          width={40}
          onReset={onResetPassword}
          onChange={onPasswordChange}
        />
      </InlineField>
    </>
  );
}
