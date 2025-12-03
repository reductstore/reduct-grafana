import React, { useEffect, useState, useCallback } from 'react';
import { InlineField, InlineFieldRow } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataMode, ReductQuery, ReductSourceOptions } from '../types';
import { DataSource } from '../datasource';
import { CompatibleSelect } from './CompatibleSelect';
import { JsonEditor } from './json-editor/JsonEditor';
import { QueryHeader } from './QueryHeader';

type Props = QueryEditorProps<DataSource, ReductQuery, ReductSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [buckets, setBuckets] = useState<Array<SelectableValue<string>>>([]);
  const [entries, setEntries] = useState<Array<SelectableValue<string>>>([]);

  const bucket = query.bucket;
  const entry = query.entry;
  const mode = query.options?.mode ?? DataMode.LabelOnly;

  const modeOptions: Array<SelectableValue<DataMode>> = [
    { label: 'Label Only', value: DataMode.LabelOnly },
    { label: 'Content Only', value: DataMode.ContentOnly },
    { label: 'Label & Content', value: DataMode.LabelAndContent },
  ];

  // Load list of buckets
  useEffect(() => {
    getBackendSrv()
      .get(`/api/datasources/${datasource.id}/resources/listBuckets`, undefined, undefined, {
        showErrorAlert: false,
      })
      .then((res) => {
        setBuckets(res.map((b: any) => ({ label: b.name, value: b.name })));
      })
      .catch((error) => {
        console.warn('Failed to load buckets:', error);
        setBuckets([]);
      });
  }, [datasource.id]);

  // Load entries when bucket changes
  useEffect(() => {
    if (!bucket) {
      setEntries([]);
      return;
    }

    getBackendSrv()
      .post(
        `/api/datasources/${datasource.id}/resources/listEntries`,
        { bucket },
        {
          showErrorAlert: false,
        }
      )
      .then((res) => {
        setEntries(res.map((e: any) => ({ label: e.name, value: e.name })));
      })
      .catch((error) => {
        console.warn('Failed to load entries:', error);
        setEntries([]);
      });
  }, [bucket, datasource.id]);

  // Helper to trigger Grafana update + run query when valid
  const updateQuery = useCallback(
    (newBucket: string | undefined, newEntry: string | undefined, newMode: DataMode) => {
      onChange({
        ...query,
        bucket: newBucket,
        entry: newEntry,
        options: { ...(query.options ?? {}), mode: newMode },
      });

      if (newBucket && newEntry) {
        onRunQuery();
      }
    },
    [query, onChange, onRunQuery]
  );

  const onBucketChange = useCallback(
    (v?: SelectableValue<string>) => {
      updateQuery(v?.value, undefined, mode);
    },
    [mode, updateQuery]
  );

  const onEntryChange = useCallback(
    (v?: SelectableValue<string>) => {
      updateQuery(bucket, v?.value, mode);
    },
    [bucket, mode, updateQuery]
  );

  const onModeChange = useCallback(
    (opt: SelectableValue<DataMode>) => {
      updateQuery(bucket, entry, opt?.value ?? DataMode.LabelOnly);
    },
    [bucket, entry, updateQuery]
  );

  // Handle changes from JSON editor
  const handleEditorChange = useCallback(
    (newQuery: ReductQuery, process: boolean) => {
      onChange(newQuery);
      if (process && newQuery.bucket && newQuery.entry) {
        onRunQuery();
      }
    },
    [onChange, onRunQuery]
  );

  return (
    <>
      <QueryHeader query={query} datasource={datasource} onRunQuery={onRunQuery} />
      <InlineFieldRow>
        <InlineField label="Bucket" tooltip="The bucket to query from" grow>
          <CompatibleSelect
            testId="bucket-picker"
            options={buckets}
            value={buckets.find((b) => b.value === bucket)}
            onChange={onBucketChange}
          />
        </InlineField>
        <InlineField label="Entry" tooltip="The entry within the bucket to query" grow>
          <CompatibleSelect
            testId="entry-picker"
            options={entries}
            value={entries.find((e) => e.value === entry)}
            onChange={onEntryChange}
          />
        </InlineField>
        <InlineField label="Scope" tooltip="Controls what the query returns: labels only, content only, or both" grow>
          <CompatibleSelect
            testId="scope-picker"
            options={modeOptions}
            value={modeOptions.find((m) => m.value === mode)}
            onChange={onModeChange}
          />
        </InlineField>
      </InlineFieldRow>
      <InlineFieldRow>
        <InlineField grow>
          <JsonEditor query={query} onChange={handleEditorChange} datasource={datasource} />
        </InlineField>
      </InlineFieldRow>
    </>
  );
}
