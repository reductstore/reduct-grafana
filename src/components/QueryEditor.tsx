import React, { ChangeEvent } from 'react';
import { CodeEditor, InlineField, Input, Select, Stack, Switch } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery, QueryOptions } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const queryTypes: Array<SelectableValue<string>> = [
  { label: 'List Buckets', value: 'listBuckets' },
  { label: 'Get Bucket Entries', value: 'getBucketEntries' },
  { label: 'Get Bucket Setting', value: 'getBucketSetting' },
  { label: 'Get Info', value: 'getInfo' },
  { label: 'List Tokens', value: 'listTokens' },
  { label: 'Get Replication Tasks', value: 'getReplicationTasks' },
  { label: 'Query Records', value: 'queryRecords' },
];

const DEFAULT_WHEN_QUERY = `{
  "&label_name": { "$gt": 10 }
}`;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryTypeChange = (value: SelectableValue<string>) => {
    onChange({ ...query, queryType: value.value as 'listBuckets' | 'getBucketEntries' | 'getBucketSetting' | 'getInfo' | 'listTokens' | 'getReplicationTasks' | 'queryRecords' });
    onRunQuery();
  };

  const onBucketChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, bucket: event.target.value });
    onRunQuery();
  };

  const onEntryChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, entry: event.target.value });
    onRunQuery();
  };

  const onStartChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, options: { ...query.options, start: Number(event.target.value) } as QueryOptions });
    onRunQuery();
  };

  const onStopChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, options: { ...query.options, stop: Number(event.target.value) } as QueryOptions });
    onRunQuery();
  };

  const onWhenChange = (newValue: string) => {
    try {
      // Validate JSON
      JSON.parse(newValue);
      onChange({ ...query, options: { ...query.options, when: newValue } as QueryOptions });
      onRunQuery();
    } catch (e) {
      // Don't update if JSON is invalid
      console.log('Invalid JSON in When condition');
    }
  };

  const onExtChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, options: { ...query.options, ext: event.target.value } as QueryOptions });
    onRunQuery();
  };

  const onStrictChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, options: { ...query.options, strict: event.target.checked } as QueryOptions });
    onRunQuery();
  };

  const onContinuousChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, options: { ...query.options, continuous: event.target.checked } as QueryOptions });
    onRunQuery();
  };

  return (
    <Stack direction="column" gap={2}>
      <Stack direction="row" gap={2}>
        <InlineField label="Query Type" labelWidth={16}>
          <Select
            options={queryTypes}
            value={query.queryType}
            onChange={onQueryTypeChange}
            width={20}
          />
        </InlineField>
        {query.queryType === 'getBucketEntries' && (
          <InlineField label="Bucket" labelWidth={16}>
            <Input
              value={query.bucket || ''}
              onChange={onBucketChange}
              placeholder="Enter bucket name"
              width={20}
            />
          </InlineField>
        )}
      </Stack>
      {query.queryType === 'queryRecords' && (
        <>
         <InlineField label="Bucket" labelWidth={16}>
              <Input
                value={query.bucket || ''}
                onChange={onBucketChange}
                placeholder="Bucket name"
                width={50}
              />
          </InlineField>
          <InlineField label="Entry" labelWidth={16}>
            <Input
              value={query.entry || ''}
              onChange={onEntryChange}
              placeholder="Entry name"
              width={50}
            />
          </InlineField>
          <InlineField label="Query Options" labelWidth={16} grow>
            <Stack direction="row" gap={2}>
             
              <Input
                value={query.options?.start || ''}
                type="number"
                onChange={onStartChange}
                placeholder="Start time"
                width={20}
              />
              <Input
                value={query.options?.stop || ''}
                type="number"
                onChange={onStopChange}
                placeholder="Stop time"
                width={20}
              />
              <Input
                value={query.options?.ext || ''}
                onChange={onExtChange}
                placeholder="Extension"
                width={20}
              />
              <Switch
                value={query.options?.strict || false}
                onChange={onStrictChange}
                label="Strict"
              />
              <Switch
                value={query.options?.continuous || false}
                onChange={onContinuousChange}
                label="Continuous"
              />
            </Stack>
          </InlineField>
          <InlineField label="When Condition" labelWidth={16} grow>
            <CodeEditor
              language="json"
              showLineNumbers={true}
              height="200px"
              value={query.options?.when || DEFAULT_WHEN_QUERY}
              onBlur={onWhenChange}
              showMiniMap={false}
            />
          </InlineField>
        </>
      )}
    </Stack>
  );
}
