import React, {useEffect, useState} from 'react';
import {Combobox, ComboboxOption, InlineField, InlineFieldRow} from '@grafana/ui';
import {getBackendSrv} from '@grafana/runtime';
import {QueryEditorProps, SelectableValue} from '@grafana/data';
import {ReductSourceOptions, ReductQuery} from '../types';
import {DataSource} from '../datasource';
import { Controlled as CodeMirror } from "react-codemirror2";
import "codemirror/lib/codemirror.css";
import "codemirror/mode/javascript/javascript";

type Props = QueryEditorProps<DataSource, ReductQuery, ReductSourceOptions>;

export function QueryEditor({query, onChange, onRunQuery, datasource}: Props) {
    const [buckets, setBuckets] = useState<Array<ComboboxOption<string>>>([]);
    const [entries, setEntries] = useState<Array<ComboboxOption<string>>>([]);
    const [when, setWhen] = useState<string>("{}");

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
            .post(`/api/datasources/${datasource.id}/resources/listEntries`, {bucket: query.bucket})
            .then((res) => {
                const entryOptions = res.map((e: any) => ({
                    label: e.name,
                    value: e.name,
                }));
                setEntries(entryOptions);
            });
    }, [query.bucket, datasource.id]);

    const onBucketChange = (v?: SelectableValue<string>) => {
        onChange({...query, bucket: v?.value, entry: undefined}); // reset entry on bucket change
    };

    const onEntryChange = (v?: SelectableValue<string>) => {
        onChange({...query, entry: v?.value});
        onRunQuery(); // run immediately when entry changes
    };

    return (
        <>
            <InlineFieldRow>
                <InlineField label="Bucket" grow>
                    <Combobox placeholder="Select bucket" options={buckets} value={query.bucket}
                              onChange={onBucketChange}/>
                </InlineField>

                {query.bucket && (
                    <InlineField label="Entry" grow>
                        <Combobox placeholder="Select entry" options={entries} value={query.entry}
                                  onChange={onEntryChange}/>
                    </InlineField>
                )}
            </InlineFieldRow>

            { query.entry && (
                <InlineField label="When" grow>
                    <CodeMirror
                        className="jsonEditor"
                        value={when}
                        options={{
                            mode: { name: "javascript", json: true },
                            theme: "default",
                            lineNumbers: true,
                            lineWrapping: true,
                            viewportMargin: Infinity,
                            matchBrackets: true,
                            autoCloseBrackets: true,
                            readOnly: false
                        }}
                        onBeforeChange={(editor: any, data: any, value: string) => {
                            setWhen(value);
                        }}
                        onBlur={(editor: any) => {
                            const newState = {...query, when: JSON.parse(editor.getValue())};
                            console.log(newState);
                            onChange(newState); // Update query with new 'when' value
                        }}
                    />
                </InlineField>
            )}

        </>
    );
}
