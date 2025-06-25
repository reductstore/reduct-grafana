import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput, Switch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;


  const onServerTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        serverToken: event.target.value,
      },
    });
  };

  const onServerURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        serverURL: event.target.value,
      },
    });
  };

  const onVerifySSLChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        verifySSL: event.target.checked,
      },
    });
  };

  const onResetServerToken = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        serverToken: false,
      },
      secureJsonData: {
        serverToken: '',
      },
    });
  };

  return (
    <>
      <InlineField label="URL" labelWidth={20} required tooltip="The URL of your ReductStore server">
        <Input
          id="config-editor-server-url"
          value={jsonData.serverURL || ''}
          placeholder="http://localhost:8383"
          onChange={onServerURLChange}
          width={40}
        />
      </InlineField>
      <InlineField label="Token" labelWidth={20} tooltip="Your ReductStore API token">
        <SecretInput
          id="config-editor-server-token"
          isConfigured={secureJsonFields.serverToken}
          value={secureJsonData?.serverToken || ''}
          placeholder="Enter your ReductStore server token"
          width={40}
          onReset={onResetServerToken}
          onChange={onServerTokenChange}
        />
      </InlineField>
      <InlineField label="Verify SSL" labelWidth={20} tooltip="Enable SSL certificate verification">
        <div style={{ display: 'flex', alignItems: 'center', height: '32px' }}>
          <Switch
            id="config-editor-verify-ssl"
            defaultChecked={true}
            value={jsonData.verifySSL}
            width={40}
            onChange={onVerifySSLChange}
          />
        </div>
      </InlineField>
    </>
  );
}
