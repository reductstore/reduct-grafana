import React, { useEffect, useState } from 'react';
import { Alert, Combobox, ComboboxOption, InlineField, InlineFieldRow, useTheme2 } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataMode, ReductQuery, ReductSourceOptions } from '../types';
import { DataSource } from '../datasource';
import { parseJson, stringifyJson } from '../utils/json';
import { Controlled as CodeMirror } from 'react-codemirror2';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/dracula.css';
import 'codemirror/mode/javascript/javascript';
import 'codemirror/addon/edit/matchbrackets';

type Props = QueryEditorProps<DataSource, ReductQuery, ReductSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [buckets, setBuckets] = useState<Array<ComboboxOption<string>>>([]);
  const [bucket, setBucket] = useState<string | undefined>(query.bucket);
  const [entries, setEntries] = useState<Array<ComboboxOption<string>>>([]);
  const [entry, setEntry] = useState<string | undefined>(query.entry);
  const [mode, setMode] = useState<DataMode>(query.options?.mode ?? DataMode.Labels);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const initialWhen = query.options?.when ? stringifyJson(query.options.when) : '{}';
  const [when, setWhen] = useState<string>(initialWhen);
  const [editorWhen, setEditorWhen] = useState<string>(initialWhen);

  const theme = useTheme2();
  const modeOptions: Array<ComboboxOption<DataMode>> = [
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

  const onModeChange = (opt: ComboboxOption<DataMode> | null) => {
    const newMode = opt?.value ?? DataMode.Labels;
    setMode(newMode);
  };

  const onWhenBlur = (editor: any) => {
    const value = editor.getValue().trim();
    if (value === '') {
      setEditorWhen('{}');
      if (when !== '{}') {
        setWhen('{}');
      }
      setErrorMessage(null);
      return;
    }
    try {
      const parsed = parseJson(value);
      const pretty = stringifyJson(parsed);
      setEditorWhen(pretty);
      if (when !== pretty) {
        setWhen(pretty);
      }
      setErrorMessage(null);
    } catch (err: any) {
      setErrorMessage(err.message);
    }
  };

  return (
    <>
      {errorMessage && <Alert title="Error: ">{errorMessage}</Alert>}

      <InlineFieldRow>
        <InlineField label="Bucket" grow>
          <Combobox placeholder="Select bucket" options={buckets} value={bucket} onChange={onBucketChange} />
        </InlineField>

        <InlineField label="Entry" grow>
          <Combobox placeholder="Select entry" options={entries} value={entry} onChange={onEntryChange} />
        </InlineField>

        <InlineField
          label="Scope"
          tooltip="Controls what the query returns: labels (metadata), content (payload), or both."
          grow
        >
          <Combobox placeholder="Select scope" options={modeOptions} value={mode} onChange={onModeChange} />
        </InlineField>
      </InlineFieldRow>

      <InlineField label="When" grow>
        <CodeMirror
          className="jsonEditor"
          value={editorWhen}
          options={{
            mode: { name: 'javascript', json: true },
            theme: theme.isDark ? 'dracula' : 'default',
            lineNumbers: true,
            lineWrapping: true,
            viewportMargin: Infinity,
            matchBrackets: true,
            readOnly: false,
            indentUnit: 2,
            tabSize: 2,
          }}
          onBeforeChange={(_, __, value: string) => setEditorWhen(value)}
          onBlur={(editor: any) => onWhenBlur(editor)}
        />
      </InlineField>
    </>
  );
}
