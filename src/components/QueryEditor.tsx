import React, { useEffect, useState } from 'react';
import { Combobox, ComboboxOption, InlineField, InlineFieldRow } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { MyQuery, MyDataSourceOptions } from '../types';
import { DataSource } from '../datasource';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [buckets, setBuckets] = useState<Array<ComboboxOption<string>>>([]);
  const [entries, setEntries] = useState<Array<ComboboxOption<string>>>([]);

  // 1. Fetch bucket list when component mounts
  useEffect(() => {
    getBackendSrv()
      .get(`/api/datasources/${datasource.id}/resources/listBuckets`)
      .then((res) => {
        const options = res.map((b: any) => ({
          label: b.name,
          value: b.name,
        }));
        setBuckets(options);
      });
  }, [datasource.id]);

  // 2. Fetch entry list when a bucket is selected
  useEffect(() => {
    if (!query.bucket) {
        return
    }

    getBackendSrv()
      .post(`/api/datasources/${datasource.id}/resources/listEntries`, { bucket: query.bucket })
      .then((res) => {
        const entryOptions = res.map((e: any) => ({
          label: e.name,
          value: e.name,
        }));
        setEntries(entryOptions);
      });
  }, [query.bucket, datasource.id]);

  const onBucketChange = (v?: SelectableValue<string>) => {
    onChange({ ...query, bucket: v?.value, entry: undefined }); // reset entry on bucket change
  };

  const onEntryChange = (v?: SelectableValue<string>) => {
    onChange({ ...query, entry: v?.value });
    onRunQuery(); // run immediately when entry changes
  };

  return (
    <InlineFieldRow>
      <InlineField label="Bucket" grow>
        <Combobox placeholder="Select bucket" options={buckets} value={query.bucket} onChange={onBucketChange} />
      </InlineField>

      {query.bucket && (
        <InlineField label="Entry" grow>
          <Combobox placeholder="Select entry" options={entries} value={query.entry} onChange={onEntryChange} />
        </InlineField>
      )}
    </InlineFieldRow>
  );
}
