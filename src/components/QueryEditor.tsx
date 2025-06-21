import React, { ChangeEvent } from 'react';
import { InlineField, Input, Select, Stack } from '@grafana/ui';
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

    return (
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
      {query.queryType === 'queryRecords' && (
        // add bucket name and options
        <>  
        <InlineField label="Bucket" labelWidth={16}>
          <Input
            value={query.bucket || ''}
            onChange={onBucketChange}
            placeholder="Enter bucket name"
            width={20}
          />    
        </InlineField>
        <InlineField label="Entry" labelWidth={16}>
          <Input
            value={query.entry || ''}
            onChange={onEntryChange}
            placeholder="Enter entry name"
            width={20}
          />
        </InlineField>
        <InlineField label="Start" labelWidth={16}>
          <Input
            value={query.options?.start || ''}
            onChange={onStartChange}
            placeholder="Enter start"
            width={20}
          />
        </InlineField>
        <InlineField label="Stop" labelWidth={16}>
          <Input
            value={query.options?.stop || ''}
            onChange={onStopChange}
            placeholder="Enter stop"
            width={20}
          />
        </InlineField>
        
        </>
      )}
    </Stack>
  );
}
