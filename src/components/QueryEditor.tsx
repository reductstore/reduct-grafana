import React, { useEffect, useState } from 'react';
import { Alert, InlineField, InlineFieldRow } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataMode, ReductQuery, ReductSourceOptions } from '../types';
import { DataSource } from '../datasource';
import { CompatiblePicker } from './CompatiblePicker';
import { CodeEditor } from './CodeEditor';
import { parseJson, stringifyJson } from '../utils/json';

type Props = QueryEditorProps<DataSource, ReductQuery, ReductSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [buckets, setBuckets] = useState<Array<SelectableValue<string>>>([]);
  const [bucket, setBucket] = useState<string | undefined>(query.bucket);
  const [entries, setEntries] = useState<Array<SelectableValue<string>>>([]);
  const [entry, setEntry] = useState<string | undefined>(query.entry);
  const [mode, setMode] = useState<DataMode>(query.options?.mode ?? DataMode.Labels);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const initialWhen = query.options?.when ? stringifyJson(query.options.when) : '{}';
  const [when, setWhen] = useState<string>(initialWhen);
  const [editorWhen, setEditorWhen] = useState<string>(initialWhen);

  const modeOptions: Array<SelectableValue<DataMode>> = [
    { label: 'Labels only', value: DataMode.Labels },
    { label: 'Content only', value: DataMode.Content },
    { label: 'Labels + Content', value: DataMode.Both },
  ];

  // Fetch bucket list on component mounts
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

  // Fetch entry list when a bucket is selected
  useEffect(() => {
    if (!query.bucket) {
      return;
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

  // Control onChange and onRunQuery calls
  useEffect(() => {
    try {
      const whenObj = when.trim() === '' ? {} : parseJson(when);
      onChange({
        ...query,
        bucket: bucket,
        entry: entry,
        options: { ...(query.options ?? {}), mode: mode, when: whenObj },
      });
      if (bucket && entry && !errorMessage) {
        onRunQuery();
      }
      setErrorMessage(null);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bucket, entry, mode, when]);

  const onBucketChange = (v?: SelectableValue<string>) => {
    const newBucket = v?.value;
    setBucket(newBucket);
    setEntry(undefined);
  };

  const onEntryChange = (v?: SelectableValue<string>) => {
    const newEntry = v?.value;
    setEntry(newEntry);
  };

  const onModeChange = (opt: SelectableValue<DataMode>) => {
    const newMode = opt?.value ?? DataMode.Labels;
    setMode(newMode);
  };

  const handleWhenChange = (value: string) => {
    setEditorWhen(value);
  };

  const handleWhenBlur = (value: string) => {
    setEditorWhen(value);
    if (when !== value) {
      setWhen(value);
    }
  };

  const handleWhenError = (error: string) => {
    setErrorMessage(error || null);
  };

  return (
    <>
      {errorMessage && <Alert title="Error: ">{errorMessage}</Alert>}

      <InlineFieldRow>
        <InlineField label="Bucket" grow>
          <CompatiblePicker
            options={buckets}
            value={buckets.find((b) => b.value === bucket)}
            onChange={onBucketChange}
          />
        </InlineField>

        <InlineField label="Entry" grow>
          <CompatiblePicker options={entries} value={entries.find((e) => e.value === entry)} onChange={onEntryChange} />
        </InlineField>

        <InlineField
          label="Scope"
          tooltip="Controls what the query returns: labels (metadata), content (payload), or both."
          grow
        >
          <CompatiblePicker
            options={modeOptions}
            value={modeOptions.find((m) => m.value === mode)}
            onChange={onModeChange}
          />
        </InlineField>
      </InlineFieldRow>

      <InlineField label="When" grow>
        <CodeEditor
          value={editorWhen}
          onChange={handleWhenChange}
          onBlur={handleWhenBlur}
          onError={handleWhenError}
          placeholder="{}"
        />
      </InlineField>
    </>
  );
}
