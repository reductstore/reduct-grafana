import React, {useEffect, useState} from 'react';
import {Alert, Combobox, ComboboxOption, InlineField, InlineFieldRow, InlineSwitch, useTheme2} from '@grafana/ui';
import {getBackendSrv} from '@grafana/runtime';
import {QueryEditorProps, SelectableValue} from '@grafana/data';
import {ReductQuery, ReductSourceOptions} from '../types';
import {DataSource} from '../datasource';
import {Controlled as CodeMirror} from "react-codemirror2"
import "codemirror/lib/codemirror.css";
import "codemirror/theme/dracula.css";
import "codemirror/mode/javascript/javascript";

type Props = QueryEditorProps<DataSource, ReductQuery, ReductSourceOptions>;

export function QueryEditor({query, onChange, onRunQuery, datasource}: Props) {
    const [buckets, setBuckets] = useState<Array<ComboboxOption<string>>>([]);
    const [entries, setEntries] = useState<Array<ComboboxOption<string>>>([]);
    const [when, setWhen] = useState<string>("{}");
    const [parseContent, setParseContent] = useState<boolean>(false);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);

    const theme = useTheme2()
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
        onRunQuery();
    };

    return (
        <>
            {errorMessage &&
                <Alert title="Error: ">{errorMessage}</Alert>}

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

            {query.entry && (
                <>
                    <InlineField
                        label="Parse content"
                        tooltip="If enabled, the backend parses JSON bodies and exposes fields as series"
                    >
                        <InlineSwitch
                            value={parseContent}
                            onChange={(e) => {
                                const newState = {...query, parseContent: e.currentTarget.checked};
                                onChange(newState);
                                onRunQuery();
                                setParseContent(e.currentTarget.checked);
                            }}
                        />
                    </InlineField>

                    <InlineField label="When" grow>
                        <CodeMirror
                            className="jsonEditor"
                            value={when}
                            options={{
                                mode: {name: "javascript", json: true},
                                theme: theme.isDark ? "dracula" : "default",
                                lineNumbers: true,
                                lineWrapping: true,
                                viewportMargin: Infinity,
                                matchBrackets: true,
                                autoCloseBrackets: true,
                                readOnly: false,
                            }}
                            onBeforeChange={(editor: any, data: any, value: string) => {
                                setWhen(value);
                            }}
                            onBlur={(editor: any) => {
                                try {
                                    const newState = {...query, when: JSON.parse(editor.getValue())};
                                    onChange(newState); // Update query with new 'when' value
                                    setErrorMessage(null);
                                    onRunQuery();
                                } catch (e) {
                                    setErrorMessage(e instanceof Error ? e.message : String(e));
                                }
                            }}/>
                    </InlineField>
                </>
            )}

        </>
    );
}
